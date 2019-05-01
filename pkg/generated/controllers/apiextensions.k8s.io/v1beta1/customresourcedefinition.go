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

package v1beta1

import (
	"context"

	"github.com/rancher/wrangler/pkg/generic"
	v1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	clientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	informers "k8s.io/apiextensions-apiserver/pkg/client/informers/externalversions/apiextensions/v1beta1"
	listers "k8s.io/apiextensions-apiserver/pkg/client/listers/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

type CustomResourceDefinitionHandler func(string, *v1beta1.CustomResourceDefinition) (*v1beta1.CustomResourceDefinition, error)

type CustomResourceDefinitionController interface {
	CustomResourceDefinitionClient

	OnChange(ctx context.Context, name string, sync CustomResourceDefinitionHandler)
	OnRemove(ctx context.Context, name string, sync CustomResourceDefinitionHandler)
	Enqueue(name string)

	Cache() CustomResourceDefinitionCache

	Informer() cache.SharedIndexInformer
	GroupVersionKind() schema.GroupVersionKind

	AddGenericHandler(ctx context.Context, name string, handler generic.Handler)
	AddGenericRemoveHandler(ctx context.Context, name string, handler generic.Handler)
	Updater() generic.Updater
}

type CustomResourceDefinitionClient interface {
	Create(*v1beta1.CustomResourceDefinition) (*v1beta1.CustomResourceDefinition, error)
	Update(*v1beta1.CustomResourceDefinition) (*v1beta1.CustomResourceDefinition, error)
	UpdateStatus(*v1beta1.CustomResourceDefinition) (*v1beta1.CustomResourceDefinition, error)
	Delete(name string, options *metav1.DeleteOptions) error
	Get(name string, options metav1.GetOptions) (*v1beta1.CustomResourceDefinition, error)
	List(opts metav1.ListOptions) (*v1beta1.CustomResourceDefinitionList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.CustomResourceDefinition, err error)
}

type CustomResourceDefinitionCache interface {
	Get(name string) (*v1beta1.CustomResourceDefinition, error)
	List(selector labels.Selector) ([]*v1beta1.CustomResourceDefinition, error)

	AddIndexer(indexName string, indexer CustomResourceDefinitionIndexer)
	GetByIndex(indexName, key string) ([]*v1beta1.CustomResourceDefinition, error)
}

type CustomResourceDefinitionIndexer func(obj *v1beta1.CustomResourceDefinition) ([]string, error)

type customResourceDefinitionController struct {
	controllerManager *generic.ControllerManager
	clientGetter      clientset.CustomResourceDefinitionsGetter
	informer          informers.CustomResourceDefinitionInformer
	gvk               schema.GroupVersionKind
}

func NewCustomResourceDefinitionController(gvk schema.GroupVersionKind, controllerManager *generic.ControllerManager, clientGetter clientset.CustomResourceDefinitionsGetter, informer informers.CustomResourceDefinitionInformer) CustomResourceDefinitionController {
	return &customResourceDefinitionController{
		controllerManager: controllerManager,
		clientGetter:      clientGetter,
		informer:          informer,
		gvk:               gvk,
	}
}

func FromCustomResourceDefinitionHandlerToHandler(sync CustomResourceDefinitionHandler) generic.Handler {
	return func(key string, obj runtime.Object) (ret runtime.Object, err error) {
		var v *v1beta1.CustomResourceDefinition
		if obj == nil {
			v, err = sync(key, nil)
		} else {
			v, err = sync(key, obj.(*v1beta1.CustomResourceDefinition))
		}
		if v == nil {
			return nil, err
		}
		return v, err
	}
}

func (c *customResourceDefinitionController) Updater() generic.Updater {
	return func(obj runtime.Object) (runtime.Object, error) {
		newObj, err := c.Update(obj.(*v1beta1.CustomResourceDefinition))
		if newObj == nil {
			return nil, err
		}
		return newObj, err
	}
}

func UpdateCustomResourceDefinitionOnChange(updater generic.Updater, handler CustomResourceDefinitionHandler) CustomResourceDefinitionHandler {
	return func(key string, obj *v1beta1.CustomResourceDefinition) (*v1beta1.CustomResourceDefinition, error) {
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
				copyObj = newObj.(*v1beta1.CustomResourceDefinition)
			}
		}

		return copyObj, err
	}
}

func (c *customResourceDefinitionController) AddGenericHandler(ctx context.Context, name string, handler generic.Handler) {
	c.controllerManager.AddHandler(ctx, c.gvk, c.informer.Informer(), name, handler)
}

func (c *customResourceDefinitionController) AddGenericRemoveHandler(ctx context.Context, name string, handler generic.Handler) {
	removeHandler := generic.NewRemoveHandler(name, c.Updater(), handler)
	c.controllerManager.AddHandler(ctx, c.gvk, c.informer.Informer(), name, removeHandler)
}

func (c *customResourceDefinitionController) OnChange(ctx context.Context, name string, sync CustomResourceDefinitionHandler) {
	c.AddGenericHandler(ctx, name, FromCustomResourceDefinitionHandlerToHandler(sync))
}

func (c *customResourceDefinitionController) OnRemove(ctx context.Context, name string, sync CustomResourceDefinitionHandler) {
	removeHandler := generic.NewRemoveHandler(name, c.Updater(), FromCustomResourceDefinitionHandlerToHandler(sync))
	c.AddGenericHandler(ctx, name, removeHandler)
}

func (c *customResourceDefinitionController) Enqueue(name string) {
	c.controllerManager.Enqueue(c.gvk, "", name)
}

func (c *customResourceDefinitionController) Informer() cache.SharedIndexInformer {
	return c.informer.Informer()
}

func (c *customResourceDefinitionController) GroupVersionKind() schema.GroupVersionKind {
	return c.gvk
}

func (c *customResourceDefinitionController) Cache() CustomResourceDefinitionCache {
	return &customResourceDefinitionCache{
		lister:  c.informer.Lister(),
		indexer: c.informer.Informer().GetIndexer(),
	}
}

func (c *customResourceDefinitionController) Create(obj *v1beta1.CustomResourceDefinition) (*v1beta1.CustomResourceDefinition, error) {
	return c.clientGetter.CustomResourceDefinitions().Create(obj)
}

func (c *customResourceDefinitionController) Update(obj *v1beta1.CustomResourceDefinition) (*v1beta1.CustomResourceDefinition, error) {
	return c.clientGetter.CustomResourceDefinitions().Update(obj)
}

func (c *customResourceDefinitionController) UpdateStatus(obj *v1beta1.CustomResourceDefinition) (*v1beta1.CustomResourceDefinition, error) {
	return c.clientGetter.CustomResourceDefinitions().UpdateStatus(obj)
}

func (c *customResourceDefinitionController) Delete(name string, options *metav1.DeleteOptions) error {
	return c.clientGetter.CustomResourceDefinitions().Delete(name, options)
}

func (c *customResourceDefinitionController) Get(name string, options metav1.GetOptions) (*v1beta1.CustomResourceDefinition, error) {
	return c.clientGetter.CustomResourceDefinitions().Get(name, options)
}

func (c *customResourceDefinitionController) List(opts metav1.ListOptions) (*v1beta1.CustomResourceDefinitionList, error) {
	return c.clientGetter.CustomResourceDefinitions().List(opts)
}

func (c *customResourceDefinitionController) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return c.clientGetter.CustomResourceDefinitions().Watch(opts)
}

func (c *customResourceDefinitionController) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.CustomResourceDefinition, err error) {
	return c.clientGetter.CustomResourceDefinitions().Patch(name, pt, data, subresources...)
}

type customResourceDefinitionCache struct {
	lister  listers.CustomResourceDefinitionLister
	indexer cache.Indexer
}

func (c *customResourceDefinitionCache) Get(name string) (*v1beta1.CustomResourceDefinition, error) {
	return c.lister.Get(name)
}

func (c *customResourceDefinitionCache) List(selector labels.Selector) ([]*v1beta1.CustomResourceDefinition, error) {
	return c.lister.List(selector)
}

func (c *customResourceDefinitionCache) AddIndexer(indexName string, indexer CustomResourceDefinitionIndexer) {
	utilruntime.Must(c.indexer.AddIndexers(map[string]cache.IndexFunc{
		indexName: func(obj interface{}) (strings []string, e error) {
			return indexer(obj.(*v1beta1.CustomResourceDefinition))
		},
	}))
}

func (c *customResourceDefinitionCache) GetByIndex(indexName, key string) (result []*v1beta1.CustomResourceDefinition, err error) {
	objs, err := c.indexer.ByIndex(indexName, key)
	if err != nil {
		return nil, err
	}
	for _, obj := range objs {
		result = append(result, obj.(*v1beta1.CustomResourceDefinition))
	}
	return result, nil
}
