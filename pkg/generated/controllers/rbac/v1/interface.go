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

package v1

import (
	"github.com/rancher/wrangler/pkg/generic"
	v1 "k8s.io/api/rbac/v1"
	informers "k8s.io/client-go/informers/rbac/v1"
	clientset "k8s.io/client-go/kubernetes/typed/rbac/v1"
)

type Interface interface {
	ClusterRole() ClusterRoleController
	ClusterRoleBinding() ClusterRoleBindingController
	Role() RoleController
	RoleBinding() RoleBindingController
}

func New(controllerManager *generic.ControllerManager, client clientset.RbacV1Interface,
	informers informers.Interface) Interface {
	return &version{
		controllerManager: controllerManager,
		client:            client,
		informers:         informers,
	}
}

type version struct {
	controllerManager *generic.ControllerManager
	informers         informers.Interface
	client            clientset.RbacV1Interface
}

func (c *version) ClusterRole() ClusterRoleController {
	return NewClusterRoleController(v1.SchemeGroupVersion.WithKind("ClusterRole"), c.controllerManager, c.client, c.informers.ClusterRoles())
}
func (c *version) ClusterRoleBinding() ClusterRoleBindingController {
	return NewClusterRoleBindingController(v1.SchemeGroupVersion.WithKind("ClusterRoleBinding"), c.controllerManager, c.client, c.informers.ClusterRoleBindings())
}
func (c *version) Role() RoleController {
	return NewRoleController(v1.SchemeGroupVersion.WithKind("Role"), c.controllerManager, c.client, c.informers.Roles())
}
func (c *version) RoleBinding() RoleBindingController {
	return NewRoleBindingController(v1.SchemeGroupVersion.WithKind("RoleBinding"), c.controllerManager, c.client, c.informers.RoleBindings())
}
