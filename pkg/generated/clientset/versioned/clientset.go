/*
Copyright 2019 Rancher Labs.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by main. DO NOT EDIT.

package versioned

import (
	autoscalev1 "github.com/rancher/rio/pkg/generated/clientset/versioned/typed/autoscale.rio.cattle.io/v1"
	gitv1 "github.com/rancher/rio/pkg/generated/clientset/versioned/typed/git.rio.cattle.io/v1"
	projectv1 "github.com/rancher/rio/pkg/generated/clientset/versioned/typed/project.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/generated/clientset/versioned/typed/rio.cattle.io/v1"
	webhookinatorv1 "github.com/rancher/rio/pkg/generated/clientset/versioned/typed/webhookinator.rio.cattle.io/v1"
	discovery "k8s.io/client-go/discovery"
	rest "k8s.io/client-go/rest"
	flowcontrol "k8s.io/client-go/util/flowcontrol"
)

type Interface interface {
	Discovery() discovery.DiscoveryInterface
	AutoscaleV1() autoscalev1.AutoscaleV1Interface
	ProjectV1() projectv1.ProjectV1Interface
	RioV1() riov1.RioV1Interface
	WebhookinatorV1() webhookinatorv1.WebhookinatorV1Interface
	GitV1() gitv1.GitV1Interface
}

// Clientset contains the clients for groups. Each group has exactly one
// version included in a Clientset.
type Clientset struct {
	*discovery.DiscoveryClient
	autoscaleV1     *autoscalev1.AutoscaleV1Client
	projectV1       *projectv1.ProjectV1Client
	rioV1           *riov1.RioV1Client
	webhookinatorV1 *webhookinatorv1.WebhookinatorV1Client
	gitV1           *gitv1.GitV1Client
}

// AutoscaleV1 retrieves the AutoscaleV1Client
func (c *Clientset) AutoscaleV1() autoscalev1.AutoscaleV1Interface {
	return c.autoscaleV1
}

// ProjectV1 retrieves the ProjectV1Client
func (c *Clientset) ProjectV1() projectv1.ProjectV1Interface {
	return c.projectV1
}

// RioV1 retrieves the RioV1Client
func (c *Clientset) RioV1() riov1.RioV1Interface {
	return c.rioV1
}

// WebhookinatorV1 retrieves the WebhookinatorV1Client
func (c *Clientset) WebhookinatorV1() webhookinatorv1.WebhookinatorV1Interface {
	return c.webhookinatorV1
}

// GitV1 retrieves the GitV1Client
func (c *Clientset) GitV1() gitv1.GitV1Interface {
	return c.gitV1
}

// Discovery retrieves the DiscoveryClient
func (c *Clientset) Discovery() discovery.DiscoveryInterface {
	if c == nil {
		return nil
	}
	return c.DiscoveryClient
}

// NewForConfig creates a new Clientset for the given config.
func NewForConfig(c *rest.Config) (*Clientset, error) {
	configShallowCopy := *c
	if configShallowCopy.RateLimiter == nil && configShallowCopy.QPS > 0 {
		configShallowCopy.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(configShallowCopy.QPS, configShallowCopy.Burst)
	}
	var cs Clientset
	var err error
	cs.autoscaleV1, err = autoscalev1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.projectV1, err = projectv1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.rioV1, err = riov1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.webhookinatorV1, err = webhookinatorv1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.gitV1, err = gitv1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}

	cs.DiscoveryClient, err = discovery.NewDiscoveryClientForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	return &cs, nil
}

// NewForConfigOrDie creates a new Clientset for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *Clientset {
	var cs Clientset
	cs.autoscaleV1 = autoscalev1.NewForConfigOrDie(c)
	cs.projectV1 = projectv1.NewForConfigOrDie(c)
	cs.rioV1 = riov1.NewForConfigOrDie(c)
	cs.webhookinatorV1 = webhookinatorv1.NewForConfigOrDie(c)
	cs.gitV1 = gitv1.NewForConfigOrDie(c)

	cs.DiscoveryClient = discovery.NewDiscoveryClientForConfigOrDie(c)
	return &cs
}

// New creates a new Clientset for the given RESTClient.
func New(c rest.Interface) *Clientset {
	var cs Clientset
	cs.autoscaleV1 = autoscalev1.New(c)
	cs.projectV1 = projectv1.New(c)
	cs.rioV1 = riov1.New(c)
	cs.webhookinatorV1 = webhookinatorv1.New(c)
	cs.gitV1 = gitv1.New(c)

	cs.DiscoveryClient = discovery.NewDiscoveryClient(c)
	return &cs
}
