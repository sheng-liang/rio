package create

import (
	"fmt"
	"strings"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/kvfile"
	"github.com/rancher/rio/cli/pkg/stack"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Create struct {
	AddHost            []string          `desc:"Add a custom host-to-IP mapping (host:ip)"`
	BuildBranch        string            `desc:"Build repository branch" default:"master"`
	BuildTag           string            `desc:"Build repository tag"`
	BuildCommit        string            `desc:"Build repository commit"`
	BuildSecret        string            `desc:"Set webhook secret"`
	BuildHook          bool              `desc:"Enable webhook"`
	CapAdd             []string          `desc:"Add Linux capabilities"`
	CapDrop            []string          `desc:"Drop Linux capabilities"`
	Config             []string          `desc:"Configs to expose to the service (format: name:target)"`
	Cpus               string            `desc:"Number of CPUs"`
	DeploymentStrategy string            `desc:"Approach to creating containers (parallel|ordered)" default:"parallel"`
	Detach             bool              `desc:"Do not attach after when -it is specified"`
	Device             []string          `desc:"Add a host device to the container"`
	DnsOption          []string          `desc:"Set DNS options"`
	DnsSearch          []string          `desc:"Set custom DNS search domains"`
	Dns                []string          `desc:"Set custom DNS servers"`
	Entrypoint         []string          `desc:"Overwrite the default ENTRYPOINT of the image"`
	E_Env              []string          `desc:"Set environment variables"`
	EnvFile            []string          `desc:"Read in a file of environment variables"`
	Expose             []string          `desc:"Expose a container's port(s) internally"`
	Concurrency        int               `desc:"The maximum concurrent request a container can handle(autoscaling)" default:"10"`
	GlobalPermission   []string          `desc:"Permissions to grant to container's service account for all stacks"`
	Group              string            `desc:"The GID to run the entrypoint of the container process"`
	HealthCmd          string            `desc:"Command to run to check health"`
	HealthInterval     string            `desc:"Time between running the check (ms|s|m|h)" default:"0s"`
	HealthRetries      int               `desc:"Consecutive successes needed to report healthy"`
	HealthStartPeriod  string            `desc:"Start period for the container to initialize before starting healthchecks (ms|s|m|h)" default:"0s"`
	HealthTimeout      string            `desc:"Maximum time to allow one check to run (ms|s|m|h)" default:"0s"`
	HealthURL          string            `desc:"URL to hit to check health (example: http://localhost:8080/ping)"`
	Hostname           string            `desc:"Container host name"`
	ImagePullPolicy    string            `desc:"Behavior determining when to pull the image (never|always|not-present)" default:"not-present"`
	Init               bool              `desc:"Run an init inside the container that forwards signals and reaps processes"`
	I_Interactive      bool              `desc:"Keep STDIN open even if not attached"`
	Ipc                string            `desc:"IPC mode to use"`
	L_Label            map[string]string `desc:"Set meta data on a container"`
	LabelFile          []string          `desc:"Read in a line delimited file of labels"`
	M_Memory           string            `desc:"Memory reservation (format: <number>[<unit>], where unit = b, k, m or g)"`
	MemoryLimit        string            `desc:"Memory hard limit (format: <number>[<unit>], where unit = b, k, m or g)"`
	Metadata           map[string]string `desc:"Metadata to attach to this service"`
	N_Name             string            `desc:"Assign a name to the container"`
	Net_Network        string            `desc:"Connect a container to a network (default|host)" default:"default"`
	Permission         []string          `desc:"Permissions to grant to container's service account in current stack"`
	Pid                string            `desc:"PID namespace to use"`
	Privileged         bool              `desc:"Give extended privileges to this container"`
	P_Publish          []string          `desc:"Publish a container's port(s) externally"`
	ReadOnly           bool              `desc:"Mount the container's root filesystem as read only"`
	ReadyCmd           string            `desc:"Command to run to check readiness"`
	ReadyInterval      string            `desc:"Time between running the check (ms|s|m|h)" default:"0s"`
	ReadyRetries       int               `desc:"Consecutive successes needed to report ready"`
	ReadyStartPeriod   string            `desc:"Start period for the container to initialize before starting readychecks (ms|s|m|h)" default:"0s"`
	ReadyTimeout       string            `desc:"Maximum time to allow one check to run (ms|s|m|h)" default:"0s"`
	ReadyURL           string            `desc:"URL to hit to check readiness (example: http://localhost:8080/ping)"`
	Restart            string            `desc:"Restart policy to apply when a container exits" default:"always"`
	Secret             []string          `desc:"Secrets to inject to the service (format: name:target)"`
	SecurityOpt        []string          `desc:"Security Options"`
	StopTimeout        string            `desc:"Timeout (in seconds) to stop a container"`
	Tmpfs              []string          `desc:"Mount a tmpfs directory"`
	T_Tty              bool              `desc:"Allocate a pseudo-TTY"`
	UnhealthyRetries   int               `desc:"Consecutive failures needed to report unhealthy"`
	UnreadyRetries     int               `desc:"Consecutive failures needed to report unready"`
	UpdateOrder        string            `desc:"Update order when doing batched rolling container updates (start-first|stop-first)"`
	UpdateStrategy     string            `desc:"Approach to updating containers (rolling|on-delete)" default:"rolling"`
	U_User             string            `desc:"UID[:GID] Sets the UID used and optionally GID for entrypoint process (format: <uid>[:<gid>])"`
	VolumeDriver       string            `desc:"Optional volume driver for the container"`
	VolumesFrom        []string          `desc:"Mount volumes from the specified container(s)"`
	V_Volume           []string          `desc:"Bind mount a volume"`
	W_Workdir          string            `desc:"Working directory inside the container"`

	Scheduling
}

type Scheduling struct {
	Global         bool     `desc:"Run one container per node (or some nodes depending on scheduling)"`
	Node           string   `desc:"Skip scheduling and run service on specified node"`
	NodePreferred  []string `desc:"Node running containers if possible should match expression"`
	NodeRequireAny []string `desc:"Node running containers must match one expression"`
	NodeRequire    []string `desc:"Node running containers must match all expressions"`
	Scheduler      string   `desc:"Use a custom scheduler of the given name"`
}

func (c *Create) Run(ctx *clicontext.CLIContext) error {
	_, err := c.RunCallback(ctx, func(s *riov1.Service) *riov1.Service {
		return s
	})
	return err
}

func (c *Create) RunCallback(ctx *clicontext.CLIContext, cb func(service *riov1.Service) *riov1.Service) (*riov1.Service, error) {
	var err error

	service, err := c.ToService(ctx.CLI.Args())
	if err != nil {
		return nil, err
	}

	service.Spec.ProjectName, service.Spec.StackName, service.Name, err = stack.ResolveSpaceStackForName(ctx, service.Name)
	if err != nil {
		return nil, err
	}

	client, err := ctx.KubeClient()
	if err != nil {
		return nil, err
	}

	service = cb(service)

	s, err := client.Rio.Services(service.Spec.StackName).Create(service)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (c *Create) ToService(args []string) (*riov1.Service, error) {
	var (
		err error
	)

	if len(args) == 0 {
		return nil, fmt.Errorf("at least one (1) argument is required")
	}

	service := &riov1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   c.N_Name,
			Labels: c.L_Label,
		},
		Spec: riov1.ServiceSpec{
			ServiceUnversionedSpec: riov1.ServiceUnversionedSpec{
				ContainerConfig: riov1.ContainerConfig{
					CPUs:                c.Cpus,
					CapAdd:              c.CapAdd,
					CapDrop:             c.CapDrop,
					Command:             args[1:],
					DefaultVolumeDriver: c.VolumeDriver,
					Entrypoint:          c.Entrypoint,
					ImagePullPolicy:     c.ImagePullPolicy,
					Init:                c.Init,
					OpenStdin:           c.I_Interactive,
					ContainerPrivilegedConfig: riov1.ContainerPrivilegedConfig{
						Privileged: c.Privileged,
					},
					ReadonlyRootfs: c.ReadOnly,
					VolumesFrom:    c.VolumesFrom,
					WorkingDir:     c.W_Workdir,
					Tty:            c.T_Tty,
				},
				PodConfig: riov1.PodConfig{
					ExtraHosts:    c.AddHost,
					Global:        c.Global,
					Hostname:      c.Hostname,
					DNS:           c.Dns,
					DNSOptions:    c.DnsOption,
					DNSSearch:     c.DnsSearch,
					RestartPolicy: c.Restart,
					Scheduling: riov1.Scheduling{
						Scheduler: c.Scheduler,
						Node: riov1.NodeScheduling{
							NodeName:   c.Node,
							RequireAll: c.NodeRequire,
							RequireAny: c.NodeRequireAny,
							Preferred:  c.NodePreferred,
						},
					},
				},
				PrivilegedConfig: riov1.PrivilegedConfig{
					IpcMode:     c.Ipc,
					PidMode:     c.Pid,
					NetworkMode: c.Net_Network,
				},
				DeploymentStrategy: c.DeploymentStrategy,
				Labels:             c.L_Label,
				UpdateOrder:        c.UpdateOrder,
				UpdateStrategy:     c.UpdateStrategy,
			},
		},
	}

	if strings.HasSuffix(args[0], ".git") {
		service.Spec.ImageBuild = &riov1.ImageBuild{
			Branch: c.BuildBranch,
			Url:    args[0],
			Tag:    c.BuildTag,
			Commit: c.BuildCommit,
			Secret: c.BuildSecret,
			Hook:   c.BuildHook,
		}
	} else {
		service.Spec.Image = args[0]
	}

	if c.U_User != "" {
		uidAndGid := strings.Split(c.U_User, ":")
		service.Spec.User = uidAndGid[0]
		if len(uidAndGid) == 2 {
			service.Spec.Group = uidAndGid[1]
		}
	}

	if c.Group != "" {
		service.Spec.Group = c.Group
	}

	service.Spec.Volumes, err = ParseMounts(c.V_Volume)
	if err != nil {
		return nil, err
	}

	service.Spec.Devices, err = ParseDevices(c.Device)
	if err != nil {
		return nil, err
	}

	service.Spec.Configs, err = ParseConfigs(c.Config)
	if err != nil {
		return nil, err
	}

	service.Spec.Secrets, err = ParseSecrets(c.Secret)
	if err != nil {
		return nil, err
	}

	service.Spec.Metadata = map[string]string{}
	for k, v := range c.Metadata {
		service.Spec.Metadata[k] = v
	}

	service.Spec.GlobalPermissions, err = ParsePermissions(c.GlobalPermission)
	if err != nil {
		return nil, err
	}

	service.Spec.Permissions, err = ParsePermissions(c.Permission)
	if err != nil {
		return nil, err
	}

	service.Spec.Environment, err = kvfile.ReadKVEnvStrings(c.EnvFile, c.E_Env)
	if err != nil {
		return nil, err
	}

	service.Labels, err = parseLabels(c.LabelFile, service.Labels)
	if err != nil {
		return nil, err
	}

	if err := populateHealthCheck(c, service); err != nil {
		return nil, err
	}

	if err := populateMemory(c, service); err != nil {
		return nil, err
	}

	service.Spec.Tmpfs, err = ParseTmpfs(c.Tmpfs)
	if err != nil {
		return nil, err
	}

	service.Spec.PortBindings, err = ParsePorts(c.P_Publish)
	if err != nil {
		return nil, err
	}

	service.Spec.ExposedPorts, err = ParseExposedPorts(c.Expose)
	if err != nil {
		return nil, err
	}

	return service, nil
}
