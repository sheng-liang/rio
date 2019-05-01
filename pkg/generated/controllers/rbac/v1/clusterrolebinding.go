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
	"context"

	"github.com/rancher/wrangler/pkg/generic"
	v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/watch"
	informers "k8s.io/client-go/informers/rbac/v1"
	clientset "k8s.io/client-go/kubernetes/typed/rbac/v1"
	listers "k8s.io/client-go/listers/rbac/v1"
	"k8s.io/client-go/tools/cache"
)

type ClusterRoleBindingHandler func(string, *v1.ClusterRoleBinding) (*v1.ClusterRoleBinding, error)

type ClusterRoleBindingController interface {
	ClusterRoleBindingClient

	OnChange(ctx context.Context, name string, sync ClusterRoleBindingHandler)
	OnRemove(ctx context.Context, name string, sync ClusterRoleBindingHandler)
	Enqueue(name string)

	Cache() ClusterRoleBindingCache

	Informer() cache.SharedIndexInformer
	GroupVersionKind() schema.GroupVersionKind

	AddGenericHandler(ctx context.Context, name string, handler generic.Handler)
	AddGenericRemoveHandler(ctx context.Context, name string, handler generic.Handler)
	Updater() generic.Updater
}

type ClusterRoleBindingClient interface {
	Create(*v1.ClusterRoleBinding) (*v1.ClusterRoleBinding, error)
	Update(*v1.ClusterRoleBinding) (*v1.ClusterRoleBinding, error)

	Delete(name string, options *metav1.DeleteOptions) error
	Get(name string, options metav1.GetOptions) (*v1.ClusterRoleBinding, error)
	List(opts metav1.ListOptions) (*v1.ClusterRoleBindingList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.ClusterRoleBinding, err error)
}

type ClusterRoleBindingCache interface {
	Get(name string) (*v1.ClusterRoleBinding, error)
	List(selector labels.Selector) ([]*v1.ClusterRoleBinding, error)

	AddIndexer(indexName string, indexer ClusterRoleBindingIndexer)
	GetByIndex(indexName, key string) ([]*v1.ClusterRoleBinding, error)
}

type ClusterRoleBindingIndexer func(obj *v1.ClusterRoleBinding) ([]string, error)

type clusterRoleBindingController struct {
	controllerManager *generic.ControllerManager
	clientGetter      clientset.ClusterRoleBindingsGetter
	informer          informers.ClusterRoleBindingInformer
	gvk               schema.GroupVersionKind
}

func NewClusterRoleBindingController(gvk schema.GroupVersionKind, controllerManager *generic.ControllerManager, clientGetter clientset.ClusterRoleBindingsGetter, informer informers.ClusterRoleBindingInformer) ClusterRoleBindingController {
	return &clusterRoleBindingController{
		controllerManager: controllerManager,
		clientGetter:      clientGetter,
		informer:          informer,
		gvk:               gvk,
	}
}

func FromClusterRoleBindingHandlerToHandler(sync ClusterRoleBindingHandler) generic.Handler {
	return func(key string, obj runtime.Object) (ret runtime.Object, err error) {
		var v *v1.ClusterRoleBinding
		if obj == nil {
			v, err = sync(key, nil)
		} else {
			v, err = sync(key, obj.(*v1.ClusterRoleBinding))
		}
		if v == nil {
			return nil, err
		}
		return v, err
	}
}

func (c *clusterRoleBindingController) Updater() generic.Updater {
	return func(obj runtime.Object) (runtime.Object, error) {
		newObj, err := c.Update(obj.(*v1.ClusterRoleBinding))
		if newObj == nil {
			return nil, err
		}
		return newObj, err
	}
}

func UpdateClusterRoleBindingOnChange(updater generic.Updater, handler ClusterRoleBindingHandler) ClusterRoleBindingHandler {
	return func(key string, obj *v1.ClusterRoleBinding) (*v1.ClusterRoleBinding, error) {
		if obj == nil {
			return handler(key, nil)
		}

		copyObj := obj.DeepCopy()
		newObj, err := handler(key, copyObj)
		if newObj != nil {
			copyObj = newObj
		}
		if obj.ResourceVersion == copyObj.ResourceVersion && !equality.Semantic.DeepEqual(obj, copyObj) {
			newObj, _ := updater(copyObj)
			if newObj != nil {
				copyObj = newObj.(*v1.ClusterRoleBinding)
			}
		}

		return copyObj, err
	}
}

func (c *clusterRoleBindingController) AddGenericHandler(ctx context.Context, name string, handler generic.Handler) {
	c.controllerManager.AddHandler(ctx, c.gvk, c.informer.Informer(), name, handler)
}

func (c *clusterRoleBindingController) AddGenericRemoveHandler(ctx context.Context, name string, handler generic.Handler) {
	removeHandler := generic.NewRemoveHandler(name, c.Updater(), handler)
	c.controllerManager.AddHandler(ctx, c.gvk, c.informer.Informer(), name, removeHandler)
}

func (c *clusterRoleBindingController) OnChange(ctx context.Context, name string, sync ClusterRoleBindingHandler) {
	c.AddGenericHandler(ctx, name, FromClusterRoleBindingHandlerToHandler(sync))
}

func (c *clusterRoleBindingController) OnRemove(ctx context.Context, name string, sync ClusterRoleBindingHandler) {
	removeHandler := generic.NewRemoveHandler(name, c.Updater(), FromClusterRoleBindingHandlerToHandler(sync))
	c.AddGenericHandler(ctx, name, removeHandler)
}

func (c *clusterRoleBindingController) Enqueue(name string) {
	c.controllerManager.Enqueue(c.gvk, "", name)
}

func (c *clusterRoleBindingController) Informer() cache.SharedIndexInformer {
	return c.informer.Informer()
}

func (c *clusterRoleBindingController) GroupVersionKind() schema.GroupVersionKind {
	return c.gvk
}

func (c *clusterRoleBindingController) Cache() ClusterRoleBindingCache {
	return &clusterRoleBindingCache{
		lister:  c.informer.Lister(),
		indexer: c.informer.Informer().GetIndexer(),
	}
}

func (c *clusterRoleBindingController) Create(obj *v1.ClusterRoleBinding) (*v1.ClusterRoleBinding, error) {
	return c.clientGetter.ClusterRoleBindings().Create(obj)
}

func (c *clusterRoleBindingController) Update(obj *v1.ClusterRoleBinding) (*v1.ClusterRoleBinding, error) {
	return c.clientGetter.ClusterRoleBindings().Update(obj)
}

func (c *clusterRoleBindingController) Delete(name string, options *metav1.DeleteOptions) error {
	return c.clientGetter.ClusterRoleBindings().Delete(name, options)
}

func (c *clusterRoleBindingController) Get(name string, options metav1.GetOptions) (*v1.ClusterRoleBinding, error) {
	return c.clientGetter.ClusterRoleBindings().Get(name, options)
}

func (c *clusterRoleBindingController) List(opts metav1.ListOptions) (*v1.ClusterRoleBindingList, error) {
	return c.clientGetter.ClusterRoleBindings().List(opts)
}

func (c *clusterRoleBindingController) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return c.clientGetter.ClusterRoleBindings().Watch(opts)
}

func (c *clusterRoleBindingController) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.ClusterRoleBinding, err error) {
	return c.clientGetter.ClusterRoleBindings().Patch(name, pt, data, subresources...)
}

type clusterRoleBindingCache struct {
	lister  listers.ClusterRoleBindingLister
	indexer cache.Indexer
}

func (c *clusterRoleBindingCache) Get(name string) (*v1.ClusterRoleBinding, error) {
	return c.lister.Get(name)
}

func (c *clusterRoleBindingCache) List(selector labels.Selector) ([]*v1.ClusterRoleBinding, error) {
	return c.lister.List(selector)
}

func (c *clusterRoleBindingCache) AddIndexer(indexName string, indexer ClusterRoleBindingIndexer) {
	utilruntime.Must(c.indexer.AddIndexers(map[string]cache.IndexFunc{
		indexName: func(obj interface{}) (strings []string, e error) {
			return indexer(obj.(*v1.ClusterRoleBinding))
		},
	}))
}

func (c *clusterRoleBindingCache) GetByIndex(indexName, key string) (result []*v1.ClusterRoleBinding, err error) {
	objs, err := c.indexer.ByIndex(indexName, key)
	if err != nil {
		return nil, err
	}
	for _, obj := range objs {
		result = append(result, obj.(*v1.ClusterRoleBinding))
	}
	return result, nil
}
