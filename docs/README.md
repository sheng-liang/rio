# Documentation

## Usage

- [Install Options](#install-options)
- [Concept](#concept)
- [Running workloads](#running-workload)
  - [Canary Deployment](#canary-deployment)
  - [Automatic DNS and HTTPS](#automatic-dns-and-https)
  - [Adding external services](#adding-external-services)
  - [Adding Router](#adding-router)
  - [Adding Public domain](#adding-public-domain)
  - [Using Riofile](#using-riofile)
- [Monitoring](#Monitoring)
- [AutoScaling based on QPS](#autoscaling)
- [Continuous Delivery](#continuous-deliverysource-code-to-deployment)
  - [Example](#example)
  - [Setting Private repository](#setup-credential-for-private-repository)
  - [Setting github webhook](#setup-github-webhook-experimental)
  - [Setting private registry](#set-custom-build-arguments-and-docker-registry)
  - [Setting Pull Request Feature](#enable-pull-request-experimental)
  - [View Build logs](#view-build-logs)
- [Using Riofile to build and develop](#using-riofile-to-build-and-develop-application)
- [Advanced options](#advanced-options)
- [FAQ](#faq)

### Install Options
Rio provides three install options for users. 

* `ingress`: Rio will use existing ingress controller and ingress resource to expose gateway services. All the traffic will go through ingress then inside cluster. Starting v0.4.0 this is the default mode.
* `svclb`: Rio will use service loadbalancer to expose gateway services. 
* `hostport`: Rio will expose hostport for gateway services.

## Concept

### Service

The main unit that is being dealt with in Rio are services. Services are just a scalable set of containers that provide a
similar function. When you run containers in Rio you are really creating a Service. `rio run` and `rio create` will
create a service. You can later scale that service with `rio scale`. Services are assigned a DNS name so that group
of containers can be accessed from other services.

### Apps

An App contains multiple service revisions. Each service in Rio is assigned an app and version. Services that have the same app but
different versions are reference to as revisions. The group of all revisions for an app is what is called an App or application in Rio.
An application named `foo` will be given a DNS name like `foo.clusterdomain.on-rio.io` and each version is assigned it's own DNS name. If the app was
`foo` and the version is `v2` the assigned DNS name for that revision would be similar to `foo-v2.clusterdomain.on-rio.io`. `rio ps` and `rio revision` will
list the assigned DNS names.

### Router

Router is a virtual service that load balances and routes traffic to other services. Routing rules can route based
on hostname, path, HTTP headers, protocol, and source.

### External Service

External Service provides a way to register external IPs or hostnames in the service mesh so they can be accessed by Rio services.

### Public Domain

Public Domain can be configured to assign a service or router a vanity domain like www.myproductionsite.com.

### Configs

ConfigMaps(Kubernetes resource) can be referenced by Rio services. It is a piece of configuration which can be mounted into pods so that it can be separated from image artifacts.
It can be created separated in the existing cluster and referenced by Rio service.

### Secrets

Secrets(Kubernetes resource) can be referenced by rio services. It contains sensitive data which can be mounted into pods to consume. Secrets can also be created separated in the existing
cluster and referenced by rio services. 

### Running workload

To deploy workload to rio:
```bash
# ibuildthecloud/demo:v1 is a docker image that listens on 80 and print "hello world"
$ rio run -p 80/http --name svc ibuildthecloud/demo:v1
default/svc:v0

# See the endpoint of your workload
$ rio ps
Name          CREATED          ENDPOINT                                    REVISIONS   SCALE     WEIGHT    DETAIL
default/svc   53 seconds ago   https://svc-default.5yt5mw.on-rio.io:9443   v0          1         100%      

### Access your workload
$ curl https://svc-default.5yt5mw.on-rio.io:9443
Hello World
```

Rio provides a similar experience as Docker CLI when running a container. Run `rio run --help` to see more options.

##### Canary Deployment
Rio allows you to easily configure canary deployment by staging services and shifting traffic between revisions.

```bash
# Create a new service
$ rio run -p 80/http --name demo1 ibuildthecloud/demo:v1

# Stage a new version, updating just the docker image and assigning it to "v3" version. If you want to change options other than just image, run with --edit.
$ rio stage --image=ibuildthecloud/demo:v3 default/demo1:v3
$ rio stage --edit default/svc:v3

# Notice a new URL was created for your staged service. For each revision you will get a unique URL.
$ rio revision default/demo1
Name               IMAGE                    CREATED          SCALE     ENDPOINT                                         WEIGHT    DETAIL
default/demo1:v3   ibuildthecloud/demo:v3   19 seconds ago   1         https://demo1-v3-default.5yt5mw.on-rio.io:9443   0         
default/demo1:v0   ibuildthecloud/demo:v1   2 minutes ago    1         https://demo1-v0-default.5yt5mw.on-rio.io:9443   100   

# Access the current revision
$ curl -s https://demo1-v0-default.5yt5mw.on-rio.io:9443
Hello World

# Access the staged service under the new URL
$ curl -s https://demo1-v3-default.5yt5mw.on-rio.io:9443
Hello World v3

# Promote v3 service. The traffic will be shifted to v3 gradually. By default we apply a 5% shift every 5 seconds, but it can be configured
# using the flags `--rollout-increment` and `--rollout-interval`. To turn off rollout(the traffic percentage will be changed to
# the desired value immediately), run `--no-rollout`.
$ rio promote default/demo1:v3

Name               IMAGE                    CREATED              SCALE     ENDPOINT                                         WEIGHT    DETAIL
default/demo1:v3   ibuildthecloud/demo:v3   About a minute ago   1         https://demo1-v3-default.5yt5mw.on-rio.io:9443   5         
default/demo1:v0   ibuildthecloud/demo:v1   3 minutes ago        1         https://demo1-v0-default.5yt5mw.on-rio.io:9443   95   

# Access the app. You should be able to see traffic routing to the new revision
$ curl https://demo1-default.5yt5mw.on-rio.io:9443
Hello World

$ curl https://demo1-default.5yt5mw.on-rio.io:9443
Hello World v3

# Wait for v3 to be 100% weight. Access the app, all traffic should be routed to new revision right now.
$ rio revision default/svc
Name               IMAGE                    CREATED         SCALE     ENDPOINT                                         WEIGHT    DETAIL
default/demo1:v3   ibuildthecloud/demo:v3   4 minutes ago   1         https://demo1-v3-default.5yt5mw.on-rio.io:9443   100       
default/demo1:v0   ibuildthecloud/demo:v1   6 minutes ago   1         https://demo1-v0-default.5yt5mw.on-rio.io:9443   0         

$ curl https://demo1-default.5yt5mw.on-rio.io:9443
Hello World v3

# Manually adjusting weight between revisions
$ rio weight default/demo1:v0=5% default/demo1:v3=95%

$ rio ps
Name            CREATED             ENDPOINT                                      REVISIONS   SCALE     WEIGHT    DETAIL
default/demo1   7 minutes ago       https://demo1-default.5yt5mw.on-rio.io:9443   v0,v3       3,3       5%,95%    
```

##### Automatic DNS and HTTPS
By default Rio will create a DNS record pointing to your cluster. Rio also uses Let's Encrypt to create
a certificate for the cluster domain so that all services support HTTPS by default.
For example, when you deploy your workload, you can access your workload in HTTPS. The domain always follows the format
of ${app}-${namespace}.\${cluster-domain}. You can see your cluster domain by running `rio info`.

Some name servers provide protection against DNS rebinding attacks. Dnsmasq is a popular example running on many
 routers. This may break endpoint name resolution (`on-rio.io`) for you. Luckily dnsmasq also provides whitelisting.

##### Adding external services
ExternalService is a service(databases, legacy apps) that is outside of your cluster, and can be added into service discovery.
It can be IPs, FQDN or service in another namespace. Once added, external service can be discovered by short name within the same namespace.

```bash
$ rio external create ${namespace/name} mydb.com

$ rio external create ${namespace/name} 8.8.8.8

$ rio external create ${namespace/name} ${another_svc/another_namespace}

```

##### Adding Router
Router is a set of L7 load-balancing rules that can route between your services. It can add Header-based, path-based routing, cookies
and other rules.

To create router in a different namespace:
```bash
$ rio route add $namespace.$name to $target_namespace/target_service 
```

To insert a router(rule)
```bash
$ rio route insert $namespace.$name to $target_namespace/target_service  
```

To create route based path match
```bash
$ rio route add $namespace.$name/path to $target_namespace/target_service 
```

To create router to a different port:
```bash
$ rio route add $namespace.$name to $target_namespace/target_service ,port=8080
```

To create router based on header(supports exact match: `foo`, prefix match: `foo*`, regular expression match: `regexp(foo.*)`)
```bash
$ rio route add --header USER=$format $namespace.$name to $target_namespace/target_service 
```

To create router based on cookies(supports exact match: `foo`, prefix match: `foo*`, regular expression match: `regexp(foo.*)`)
```bash
$ rio route add --cookie USER=$format $namespace.$name to $target_namespace/target_service 
```

To create route based on HTTP method(supports exact match: `foo`, prefix match: `foo*`, regular expression match: `regexp(foo.*)`)
```bash
$ rio route add --method GET $namespace.$name to $target_namespace/target_service
```

To add, set or remove headers:
```bash
$ rio route add --add-header FOO=BAR $namespace.$name to $target_namespace/target_service   
$ rio route add --set-header FOO=BAR $namespace.$name to $target_namespace/target_service  
$ rio route add --remove-header FOO=BAR $namespace.$name to $target_namespace/target_service  
```

To mirror traffic:
```bash
$ rio route add $namespace.$name mirror $target_namespace/target_service 
```

To rewrite host header and path
```bash
$ rio route add $namespace.$name rewrite $target_namespace/target_service 
```

To redirect to another service
```bash
$ rio route add $namespace.$name redirect $target_namespace/target_service/path  
```

To add timeout
```bash
$ rio route add --timeout $namespace.$name to $target_namespace/target_service  
```

To add fault injection
```bash
$ rio route add --fault-httpcode 502 --fault-delay 1s --fault-percentage 80 $namespace.$name to $target_namespace/target_service 
```

To add retry logic
```bash
$ rio route add --retry-attempts 5 --retry-timeout 1s $namespace.$name to $target_namespace/target_service 
```

To create router to different revision and different weight
```bash
$ rio route add $namespace.$name to $service:v0,weight=50 $service:v1,weight=50 
```

##### Adding Public domain
Rio allows you to add a vanity domain to your workloads. For example, to add a domain `www.myproductionsite.com` to your workload,
run
```bash
# Create a domain that points to route1. You have to setup a cname record from your domain to cluster domain.
# For example, foo.bar -> CNAME -> iazlia.on-rio.io
$ rio domain register www.myproductionsite.com default/route1
default/foo-bar

# Use your own certs by providing a secret that contain tls cert and key instead of provisioning by letsencrypts. The secret has to be created first in system namespace.
$ rio domain register --secret $name www.myproductionsite.com default/route1

# Access your domain 
```

Note: By default Rio will automatically configure Letsencrypt HTTP-01 challenge to provision certs for your publicdomain. This needs you to install rio on standard ports.
If you are install rio with svclb or hostport mode, try `rio install --http-port 80 --https-port 443`.

##### Using Riofile

###### Riofile example

Rio allows you to define a file called `Riofile`. `Riofile` allows you define rio services, configmap is a friendly way with `docker-compose` syntax.
For example, to define a nginx application with conf

```yaml
configs:
  conf:
    index.html: |-
      <!DOCTYPE html>
      <html>
      <body>
      
      <h1>Hello World</h1>
      
      </body>
      </html>
services:
  nginx:
    image: nginx
    ports:
    - 80/http
    configs:
    - conf/index.html:/usr/share/nginx/html/index.html
```

Once you have defined `Riofile`, simply run `rio up`. Any change you made for `Riofile`, re-run `rio up` to pick the change.

###### Riofile reference
```yaml
# Configmap
configs:          
  config-foo:     # specify name in the section 
    key1: |-      # specify key and data in the section 
      {{ config1 }}
    key2: |-
      {{ config2 }}
      
# Service
services:
  service-foo:
    disableServiceMesh: true # disable service mesh side injection for service
    
    # scale setting
    scale: 2 # specify scale of the service. If you pass range `1-10`, it will enable autoscaling which can be scale from 1 to 10.
    updateBatchSize: 1 # specify the update batch size.
    
    # revision setting
    app: my-app # specify app name. Defaults to service name. This is used to aggregate services that belongs to the same app.
    version: v0 # specify revision name
    weight: 80 # weight assigned to this revision. Value: 0-100
    
    # autoscaling setting
    concurrency: 10 # specify concurrent request each pod can handle(soft limit, used to scale service)
    
    # rollout config
    rollout: true# whether rollout traffic gradually
    rolloutIncrement: 5 # traffic percentage increment(%) for each interval. Will not work if rollout is false
    rolloutInterval: 2 # traffic increment interval(seconds). Will not work if rollout is false
    
    # Permission for service
    # 
    #   global_permissions:
    #   - 'create,get,list certmanager.k8s.io/*'
    #  
    #   this will give workload abilities to **create, get, list** **all** resources in api group **certmanager.k8s.io**.
    #
    #   If you want to hook up with an existing role:
    #
    #   
    #   global_permissions:
    #   - 'role=cluster-admin'
    #   
    #
    #   - `permisions`: Specify current namespace permission of workload
    #
    #   Example: 
    #   
    #   permissions:
    #   - 'create,get,list certmanager.k8s.io/*'
    #  
    #
    #   This will give workload abilities to **create, get, list** **all** resources in api group **certmanager.k8s.io** in **current** namespace. 
    #   
    #   Example: 
    #   
    #   permissions:
    #   - 'create,get,list /node/proxy'
    #   
    #    This will give subresource for node/proxy 
    global_permissions:
    - 'create,get,list certmanager.k8s.io/*'
    permissions:
    - 'create,get,list certmanager.k8s.io/*'
    
    # container configuration
    image: # container image
    imagePullPolicy: # image pull policy. Options: (always/never/ifNotProsent)
    build:
      repo: https://github.com/rancher/rio # git repository to build
      branch: master # git repository branch
      revision: v0.1.0 # revision digest to build. If set, image will be built based on this revision. Otherwise it will take head revision in repo. Also if revision is not set, it will be served as the base revision to watch any change in repo and create new revision based changes from repo.
      buildArgs: # build arguments to pass to buildkit https://docs.docker.com/engine/reference/builder/#understand-how-arg-and-from-interact
      - foo=bar
      dockerFile: Dockerfile # the name of Dockerfile to look for
      dockerFilePath: ./ # the path of Dockerfile to look for
      buildContext: ./  # docker build context
      noCache: true # build without cache
      buildImageName: myname/image:tag # specify custom image name(excluding registry name). Default name: $namespace/name:$revision_digest
      pushRegistry: docker.io # specify push registry. Example: docker.io, gcr.io
      stageOnly: true # if set, newly created revision will get any traffic
      githubSecretName: secretGithub # specify github webhook secretName to setup github webhook
      gitSecretName: secretGit # specify git secret name for private git repository
      pushRegistrySecretName: secretDocker # specify secret name for pushing to docker registry
      enablePr: true # enable pull request feature
    command: # container entrypoint, not executed within a shell. The docker image's ENTRYPOINT is used if this is not provided.
    - echo
    args: # arguments to the entrypoint. The docker image's CMD is used if this is not provided.
    - "hello world"
    workingDir: /home # container working directory
    ports: # container ports, format: `$(servicePort:)containerPort/protocol`
    - 8080:80/http,web # service port 8080 will be mapped to container port 80 with protocol http, named `web`
    - 8080/http,admin,internal=true # service port 8080 will be mapped to container port 8080 with protocol http, named `admin`, internal port(will not be exposed through gateway) 
    env: # specify environment variable
    - POD_NAME=$(self/name) # mapped to "metadata.name" 
    # 
    # "self/name":           "metadata.name",
    # "self/namespace":      "metadata.namespace",
    # "self/labels":         "metadata.labels",
    # "self/annotations":    "metadata.annotations",
    # "self/node":           "spec.nodeName",
    # "self/serviceAccount": "spec.serviceAccountName",
    # "self/hostIp":         "status.hostIP",
    # "self/nodeIp":         "status.hostIP",
    # "self/ip":             "status.podIP",
    # 
    cpus: 100m # cpu request, format 0.5 or 500m. 500m = 0.5 core. https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/
    memory: 100 mi # memory request. 100mi, available options https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/
    secrets: # specify secret to mount. Format: `$name/$key:/path/to/file`. Secret has to be pre-created in the same namespace
    - foo/bar:/my/password
    configs: # specify configmap to mount. Format: `$name/$key:/path/to/file`. 
    - foo/bar:/my/config
    livenessProbe: # livenessProbe setting. https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-probes/
      httpGet:
        path: /ping
        port: 9997
      initialDelaySeconds: 10
    readinessProbe: # readinessProbe https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-probes/
      failureThreshold: 7
      httpGet:
        path: /ready
        port: 9997
    stdin: true # whether this container should allocate a buffer for stdin in the container runtime
    stdinOnce: true # whether the container runtime should close the stdin channel after it has been opened by a single attach. When stdin is true the stdin stream will remain open across multiple attach sessions.
    tty: true # whether this container should allocate a TTY for itself
    user: 1000 # the UID to run the entrypoint of the container process.
    group: 1000 # the GID to run the entrypoint of the container process
    readOnly: true # whether this container has a read-only root filesystem
    
    nodeAffinity: # Describes node affinity scheduling rules for the pod.
    podAffinity:  # Describes pod affinity scheduling rules (e.g. co-locate this pod in the same node, zone, etc. as some other pod(s)).
    podAntiAffinity: # Describes pod anti-affinity scheduling rules (e.g. avoid putting this pod in the same node, zone, etc. as some other pod(s)).
    
    addHost: # hostname alias
    net: host # host networking
    imagePullSecrets: # image pull secret https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod
    - secret1
    - secret2 
    
    containers: # specify sidecars
    - init: true # init container
      image: ubuntu
      args:
      - "echo"
      - "hello world"
      # other options are available in container section above
 
# Router
routers:
  foo:
    routes:
    - matches: # match rules, the first rule matching an incoming request is used
      - path: # match path, can specify regxp, prefix or exact match 
          regxp: /bar.*
          # prefix: /bar
          # exact: /bar
        scheme:
          regxp: /bar.*
          # prefix: /bar
          # exact: /bar
        method:
          regxp: /bar.*
          #prefix: /bar
          #exact: /bar
        headers:
          FOO:
            regxp: /bar.*
            #prefix: /bar
            #exact: /bar
        cookie:
          USER:
            regxp: /bar.*
            #prefix: /bar
            #exact: /bar
      to:  # specify destination
      - service: service-foo
        revision: v0
        namespace: default
        weight: 50
      - service: service-foo
        revision: v1
        namespace: default
        weight: 50
      redirect: # specify redirect rule
        host: www.foo.bar
        path: /redirect
      rewrite:
        host: www.foo.bar
        path: /rewrite
      headers: # header operations
        add:
          foo: bar
        set:
          foo: bar
        remove:
        - foo
      fault:
        percentage: 80 # inject fault percentage(%)
        delayMillis: 100 # adding delay before injecting fault (millseconds)
        abort:
          httpStatus: 502 # injecting http code
      mirror:   # sending mirror traffic
        service: mirror-foo
        revision: v0
        namespace: default
      timeoutMillis: 100 # setting request timeout (milli-seconds)
      retry:
        attempts: 10 # retry attempts
        timeoutMillis: 100 # retry timeout (milli-seconds)
        
# externalservices
externalservices:
  foo:
    ipAddresses: # pointing to external IP addresses
    - 1.1.1.1
    - 2.2.2.2
    fqdn: www.foo.bar # pointing to fqdn
    service: $namespace/$name # pointing to services in another namespace
``` 

###### Watching Riofile
You can setup github repository to watch Riofile changes and re-apply Riofile changes. Here is the example:
```bash
$ rio up https://github.com/username/repo
```
If you want to setup webhook to watch, go to [here](#setup-github-webhook-experimental)


### Monitoring
By default, Rio will deploy [Grafana](https://grafana.com/) and [Kiali](https://www.kiali.io/) to give users the ability to watch all metrics of the service mesh.
You can find endpoints of both services by running `rio -s ps`.

```bash
Name                          CREATED       ENDPOINT                                           REVISIONS   SCALE     WEIGHT    DETAIL
rio-system/controller         7 hours ago                                                      v0          1         100%      
rio-system/activator          7 hours ago                                                      v0          1         100%      
rio-system/kiali              9 hours ago   https://kiali-rio-system.5yt5mw.on-rio.io:9443     v0          1         100%      
rio-system/cert-manager       9 hours ago                                                      v0          1         100%      
rio-system/istio-pilot        9 hours ago                                                      v0          1         100%      
rio-system/istio-gateway      9 hours ago                                                      v0          1         100%      
rio-system/istio-citadel      9 hours ago                                                      v0          1         100%      
rio-system/istio-telemetry    9 hours ago                                                      v0          1         100%      
rio-system/grafana            9 hours ago   https://grafana-rio-system.5yt5mw.on-rio.io:9443   v0          1         100%      
rio-system/registry           9 hours ago                                                      v0          1         100%      
rio-system/webhook            9 hours ago   https://webhook-rio-system.5yt5mw.on-rio.io:9443   v0          1         100%      
rio-system/autoscaler         9 hours ago                                                      v0          1         100%      
rio-system/build-controller   9 hours ago                                                      v0          1         100%      
rio-system/prometheus         9 hours ago                                                      v0          1         100%  
```

### Autoscaling
By default each workload is enabled with autoscaling(min scale 1, max scale 10), which means the workload can be scaled from 1 instance to 10 instances
depending on how much traffic it receives. To change the scale range, run `rio run --scale=$min-$max ${args}`. To disable autoscaling,
 run `rio run --scale=${num} ${args}`
 
```bash
# Run a workload, set the minimal and maximum scale
$ rio run -p 8080/http --name autoscale --scale=1-20 strongmonkey1992/autoscale:v0
default/autoscale:v0

# Put some load to the workload. We use [hey](https://github.com/rakyll/hey) to create traffic
$ hey -z 600s -c 60 http://autoscale-v0-default.5yt5mw.on-rio.io:9080

# Note that the service has been scaled to 6 instances
$ rio revision default/autoscale
Name                   IMAGE                           CREATED          SCALE     ENDPOINT                                             WEIGHT    DETAIL
default/autoscale:v0   strongmonkey1992/autoscale:v0   49 seconds ago   1         https://autoscale-v0-default.5yt5mw.on-rio.io:9443   100       

# Run a workload that can be scaled to zero
$ rio run -p 8080/http --name autoscale-zero --scale=0-20 strongmonkey1992/autoscale:v0
default/autoscale-zero:v0

# Wait a couple of minutes for the workload to scale to zero
$ rio revision default/autoscale-zero
Name                        IMAGE                           CREATED         SCALE     ENDPOINT                                                  WEIGHT    DETAIL
default/autoscale-zero:v0   strongmonkey1992/autoscale:v0   9 seconds ago   1         https://autoscale-zero-v0-default.5yt5mw.on-rio.io:9443   100       

# Access the workload. Once there is an active request, the workload will be re-scaled to active.
$ rio ps
Name                     CREATED          ENDPOINT                                               REVISIONS   SCALE     WEIGHT    DETAIL
default/autoscale-zero   13 minutes ago   https://autoscale-zero-default.5yt5mw.on-rio.io:9443   v0          0/1       100%     

$ curl -s https://autoscale-zero-v0-default.5yt5mw.on-rio.io:9443
Hi there, I am StrongMonkey:v13

# Verify that the workload has been re-scaled to 1
$ rio revision default/autoscale-zero
Name                        IMAGE                           CREATED         SCALE     ENDPOINT                                                  WEIGHT    DETAIL
default/autoscale-zero:v0   strongmonkey1992/autoscale:v0   9 seconds ago   1         https://autoscale-zero-v0-default.5yt5mw.on-rio.io:9443   100       
```

### Continuous Delivery(Source code to Deployment)

##### Example 
Rio supports configuration of a Git-based source code repository to deploy the actual workload. It can be as easy
as giving Rio a valid Git repository repo URL.

```bash
# Run a workload from a git repo. We assume the repo has a Dockerfile at root directory to build the image
$ rio run -n build https://github.com/StrongMonkey/demo.git
default/build:v0

# Waiting for the image to be built. Note that the image column is empty. Once the image is ready service will be active
$ rio revision
Name               IMAGE     CREATED         SCALE     ENDPOINT                                         WEIGHT    DETAIL
default/build:v0             6 seconds ago   0/1       https://build-v0-default.5yt5mw.on-rio.io:9443   100     

# The image is ready. Note that we deploy from the default docker registry into the cluster.
# The image name has the format of ${registry-domain}/${namespace}/${name}:${commit}
$ rio revision
Name               IMAGE                                                    CREATED              SCALE     ENDPOINT                                         WEIGHT    DETAIL
default/build:v0   default-build:ff564b7058e15c3e6813f06feb965af7787f0b28   About a minute ago   1         https://build-v0-default.5yt5mw.on-rio.io:9443   100  


# Show the endpoint of your workload
$ rio ps
Name            CREATED              ENDPOINT                                      REVISIONS   SCALE     WEIGHT    DETAIL
default/build   About a minute ago   https://build-default.5yt5mw.on-rio.io:9443   v0          1         100%      

# Access the endpoint
$ curl -s https://build-default.5yt5mw.on-rio.io:9443
Hi there, I am StrongMonkey:v1
```

When you point your workload to a git repo, Rio will automatically watch any commit or tag pushed to
a specific branch (default is master). By default, Rio will pull and check the branch at a certain interval, but
can be configured to use a webhook instead.

```bash
# edit the code, change v1 to v3, push the code
$ vim main.go | git add -u | git commit -m "change to v3" | git push $remote

# A new revision has been automatically created. Noticed that once the new revision is created, the traffic will
# be automatically shifted from the old revision to the new revision.
$ rio revision default/build
NAME                   IMAGE                                                                                                       CREATED          STATE     SCALE     ENDPOINT                                             WEIGHT                               DETAIL
default/build:v0       registry-rio-system.iazlia.on-rio.io/default/build:32a4e453ca3bf0672ece9abf6901fa307d951add                 11 minutes ago   active    1         https://build-v0-default.iazlia.on-rio.io:9443
default/build:vc6d4c   registry-rio-system.iazlia.on-rio.io/default/build-e46cfb4-1d207:c6d4c4452b064e476940de7b33c7a70ac0d9e153   22 seconds ago   active    1         https://build-vc6d4c-default.iazlia.on-rio.io:9443   =============================> 100

# Access the endpoint
$ curl https://build-default.8axlxl.on-rio.io
Hi there, I am StrongMonkey:v1
$ curl https://build-default.8axlxl.on-rio.io
Hi there, I am StrongMonkey:v3

# Wait until all traffic has been shifted to the new revision
$ rio revision default/build
NAME                  IMAGE                                                                                                       CREATED          STATE     SCALE     ENDPOINT                                       WEIGHT                               DETAIL
default/build:v0      registry-rio-system.8axlxl.on-rio.io/default/build:34512dddba18781fb6909c303eb206a73d41d9ba                 24 minutes ago   active    1         https://build-v0-default.8axlxl.on-rio.io
default/build:25a0a   registry-rio-system.8axlxl.on-rio.io/default/build-e46cfb4-08a3b:25a0acda54812619f8063c121f6ed5ed2bfb968f   4 minutes ago    active    1         https://build-25a0a-default.8axlxl.on-rio.io   =============================> 100

# Access the workload. Note that all the traffic is routed to the new revision
$ curl https://build-default.8axlxl.on-rio.io
Hi there, I am StrongMonkey:v3
```

#### Setup credential for private repository
1. Set up git basic auth.(Currently ssh key is not supported and will be added soon). Here is an exmaple of adding a github repo.
```bash
$ rio secret add --git-basic-auth
Select namespace[default]: $(put the same namespace with your workload)
git url: https://github.com/username
username: $username
password: $password
```
2. Run your workload and point it to your private git repo. It will automatically use the secret you just configured.

#### Setup Github webhook (experimental)
By default, rio will automatically pull git repo and check if repo code has changed. You can also configure a webhook to automatically push any events to Rio to trigger the build.

1. Set up Github webhook token.
```bash
$ rio secret add --github-webhook
Select namespace[default]: $(put the same namespace with your workload)
accessToken: $(github_accesstoken) # the token has to be able create webhook in your github repo.
```

2. Create workload and point to your repo.

3. Go to your Github repo, it should have webhook configured to point to one of our webhook service.

#### Set Custom build arguments and docker registry
You can also push to your own registry for images that rio has built.

1. Setup docker registry auth. Here is an example of how to setup docker registry.
```bash
$ rio secret add --docker
Select namespace[default]: $(put the same namespace with your workload)
Registry url[]: https://index.docker.io/v1/
username[]: $(your_docker_hub_username)
password[]: $(password)
```

2. Create your workload. Set the correct push registry.

```bash
$ rio run --build-registry docker.io --build-image-name $(username)/yourimagename $(repo)
```
`docker.io/$(username)/yourimagename` will be pushed into dockerhub registry. 

#### Enable Pull request (experimental)
Rio also allows you to configure pull request builds. This needs you to configure github webhook token correctly.

1. Set up github webhook token in the previous session

2. Run workload with pull-request enabled.

```bash
$ rio run --build-enable-pr $(repo)
```

After this, if there is any pull request, Rio will create a deployment based on this pull request, and you will get a unique link
to see the change this pull request introduced in the actual deployment.

#### View build logs
To view logs from your builds
```bash
$ rio builds
NAME                                                                     SERVICE                   REVISION                                   CREATED        SUCCEED   REASON
default/fervent-swartz6-ee709-786b366d5d44de6b547939f51d467437e45c5ee1   default/fervent-swartz6   786b366d5d44de6b547939f51d467437e45c5ee1   23 hours ago   True    

$ rio logs -f default/fervent-swartz6-ee709-786b366d5d44de6b547939f51d467437e45c5ee1

# restart any builds that failed
$ rio build restart default/fervent-swartz6-ee709-786b366d5d44de6b547939f51d467437e45c5ee1
```

### Using Riofile to build and develop application

Rio allows developer to build and develop applications from local source code. Rio will by default use buildkit to build application.

Requirements:
1. Local repo must have `Dockerfile` and `Riofile`.
2. Developer have rio installed in a **single-node k3s** cluster. We will support minikube later, but as today buildkit is not supported in minikube.(https://github.com/kubernetes/minikube/issues/4143 )

Use cases:
1. `git clone https://github.com/StrongMonkey/riofile-demo.git`
2. `cd riofile-demo`
3. `rio up`. It will build the project and bring up services.
4. `rio ps`. 
5. `vim main.go && change "Hi there, I am demoing Riofile" to "Hi there, I am demoing something"`
6. Re-run `rio up`. It will rebuild. After it is done, revisit service endpoint to see if content is changed.

If you want more complicated build arguments, rio supports the following format
```yaml
services:
  demo:
   ports:
   - 8080/http
   build:
    buildArgs:
    - foo=bar
    dockerFile: Dockerfile
    dockerFilePath: ./
    buildContext: ./
    noCache: true
    push: true
    buildImageName: docker.io/foo/bar
```

## Advanced Options

There are other install options:

* `http-port`: HTTP port gateway service will listen. If install mode is svclb or hostport, defaults to 9080. If install mode is ingress, it will 80.
* `https-port`: HTTPS port gateway service will listen. If install mode is svclb or hostport, defaults to 9443. If install mode is ingress, it will 443.
* `ip-address`: Manually specify worker IP addresses to generate DNS domain. By default Rio will detect based on install mode.
* `service-cidr`: Manually specify service-cidr for service mesh to intercept traffic. By default Rio will try to detect.
* `disable-features`: Specify feature to disable during install. Here are the available feature list.

| Feature | Description |
|----------|----------------|
| autoscaling | Auto-scaling services based on QPS and requests load
| build | Rio Build, from source code to deployment
| grafana | Grafana Dashboard
| istio | Service routing using Istio
| kiali | Kiali Dashboard
| letsencrypt | Let's Encrypt
| mixer | Istio Mixer telemetry
| prometheus | Enable prometheus
| rdns | Assign cluster a hostname from public Rancher DNS service

* `httpproxy`: Specify HTTP_PROXY environment variable for control plane.
* `lite`: install with lite mode.


## FAQ

* How can I upgrade rio?
```
Upgrading rio just needs the latest release of rio binary. Re-run `rio install` with your install options.
```

* How can I swap out letsencrypt certificate with my own certs?
```
Create a TLS secret in `rio-system` namespace that contains your tls cert and key. Edit cluster domain by running `k edit clusterdomain cluster-domain -n rio-system`.
Change spec.secretRef.name to the name of your TLS secret.
```

* How can I use my own DNS domain?
```
Disable rdns and letsencrypt features by running `rio install --disable-features rdns,letsencrypt`. Edit cluster domain by running `k edit clusterdomain cluster-domain -n rio-system`.
Change status.domain to your own wildcard doamin. You are responsible to manage your dns record to gateway IP or worker nodes.
```

* How can I reference persist volume?
```
Rio only supports stateless workloads at this point.
```

* How to manually specify IP addresses?
```
Rio will automatically detect work node ip addresses based on install mode. If your host has multiple IP addresses, you can manually specify which IP address Rio should use for creating external DNS records with the `--ip-address` flag. 
For instance to advertise the external IP of an AWS instance: `rio install --ip-address $(curl -s http://169.254.169.254/latest/meta-data/public-ipv4)`
By doing this, you lose the ability to dynamic updating IP addresses to DNS.
```

