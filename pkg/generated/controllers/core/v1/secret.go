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
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/watch"
	informers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes/typed/core/v1"
	listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

type SecretHandler func(string, *v1.Secret) (*v1.Secret, error)

type SecretController interface {
	SecretClient

	OnChange(ctx context.Context, name string, sync SecretHandler)
	OnRemove(ctx context.Context, name string, sync SecretHandler)
	Enqueue(namespace, name string)

	Cache() SecretCache

	Informer() cache.SharedIndexInformer
	GroupVersionKind() schema.GroupVersionKind

	AddGenericHandler(ctx context.Context, name string, handler generic.Handler)
	AddGenericRemoveHandler(ctx context.Context, name string, handler generic.Handler)
	Updater() generic.Updater
}

type SecretClient interface {
	Create(*v1.Secret) (*v1.Secret, error)
	Update(*v1.Secret) (*v1.Secret, error)

	Delete(namespace, name string, options *metav1.DeleteOptions) error
	Get(namespace, name string, options metav1.GetOptions) (*v1.Secret, error)
	List(namespace string, opts metav1.ListOptions) (*v1.SecretList, error)
	Watch(namespace string, opts metav1.ListOptions) (watch.Interface, error)
	Patch(namespace, name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Secret, err error)
}

type SecretCache interface {
	Get(namespace, name string) (*v1.Secret, error)
	List(namespace string, selector labels.Selector) ([]*v1.Secret, error)

	AddIndexer(indexName string, indexer SecretIndexer)
	GetByIndex(indexName, key string) ([]*v1.Secret, error)
}

type SecretIndexer func(obj *v1.Secret) ([]string, error)

type secretController struct {
	controllerManager *generic.ControllerManager
	clientGetter      clientset.SecretsGetter
	informer          informers.SecretInformer
	gvk               schema.GroupVersionKind
}

func NewSecretController(gvk schema.GroupVersionKind, controllerManager *generic.ControllerManager, clientGetter clientset.SecretsGetter, informer informers.SecretInformer) SecretController {
	return &secretController{
		controllerManager: controllerManager,
		clientGetter:      clientGetter,
		informer:          informer,
		gvk:               gvk,
	}
}

func FromSecretHandlerToHandler(sync SecretHandler) generic.Handler {
	return func(key string, obj runtime.Object) (ret runtime.Object, err error) {
		var v *v1.Secret
		if obj == nil {
			v, err = sync(key, nil)
		} else {
			v, err = sync(key, obj.(*v1.Secret))
		}
		if v == nil {
			return nil, err
		}
		return v, err
	}
}

func (c *secretController) Updater() generic.Updater {
	return func(obj runtime.Object) (runtime.Object, error) {
		newObj, err := c.Update(obj.(*v1.Secret))
		if newObj == nil {
			return nil, err
		}
		return newObj, err
	}
}

func UpdateSecretOnChange(updater generic.Updater, handler SecretHandler) SecretHandler {
	return func(key string, obj *v1.Secret) (*v1.Secret, error) {
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
				copyObj = newObj.(*v1.Secret)
			}
		}

		return copyObj, err
	}
}

func (c *secretController) AddGenericHandler(ctx context.Context, name string, handler generic.Handler) {
	c.controllerManager.AddHandler(ctx, c.gvk, c.informer.Informer(), name, handler)
}

func (c *secretController) AddGenericRemoveHandler(ctx context.Context, name string, handler generic.Handler) {
	removeHandler := generic.NewRemoveHandler(name, c.Updater(), handler)
	c.controllerManager.AddHandler(ctx, c.gvk, c.informer.Informer(), name, removeHandler)
}

func (c *secretController) OnChange(ctx context.Context, name string, sync SecretHandler) {
	c.AddGenericHandler(ctx, name, FromSecretHandlerToHandler(sync))
}

func (c *secretController) OnRemove(ctx context.Context, name string, sync SecretHandler) {
	removeHandler := generic.NewRemoveHandler(name, c.Updater(), FromSecretHandlerToHandler(sync))
	c.AddGenericHandler(ctx, name, removeHandler)
}

func (c *secretController) Enqueue(namespace, name string) {
	c.controllerManager.Enqueue(c.gvk, namespace, name)
}

func (c *secretController) Informer() cache.SharedIndexInformer {
	return c.informer.Informer()
}

func (c *secretController) GroupVersionKind() schema.GroupVersionKind {
	return c.gvk
}

func (c *secretController) Cache() SecretCache {
	return &secretCache{
		lister:  c.informer.Lister(),
		indexer: c.informer.Informer().GetIndexer(),
	}
}

func (c *secretController) Create(obj *v1.Secret) (*v1.Secret, error) {
	return c.clientGetter.Secrets(obj.Namespace).Create(obj)
}

func (c *secretController) Update(obj *v1.Secret) (*v1.Secret, error) {
	return c.clientGetter.Secrets(obj.Namespace).Update(obj)
}

func (c *secretController) Delete(namespace, name string, options *metav1.DeleteOptions) error {
	return c.clientGetter.Secrets(namespace).Delete(name, options)
}

func (c *secretController) Get(namespace, name string, options metav1.GetOptions) (*v1.Secret, error) {
	return c.clientGetter.Secrets(namespace).Get(name, options)
}

func (c *secretController) List(namespace string, opts metav1.ListOptions) (*v1.SecretList, error) {
	return c.clientGetter.Secrets(namespace).List(opts)
}

func (c *secretController) Watch(namespace string, opts metav1.ListOptions) (watch.Interface, error) {
	return c.clientGetter.Secrets(namespace).Watch(opts)
}

func (c *secretController) Patch(namespace, name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Secret, err error) {
	return c.clientGetter.Secrets(namespace).Patch(name, pt, data, subresources...)
}

type secretCache struct {
	lister  listers.SecretLister
	indexer cache.Indexer
}

func (c *secretCache) Get(namespace, name string) (*v1.Secret, error) {
	return c.lister.Secrets(namespace).Get(name)
}

func (c *secretCache) List(namespace string, selector labels.Selector) ([]*v1.Secret, error) {
	return c.lister.Secrets(namespace).List(selector)
}

func (c *secretCache) AddIndexer(indexName string, indexer SecretIndexer) {
	utilruntime.Must(c.indexer.AddIndexers(map[string]cache.IndexFunc{
		indexName: func(obj interface{}) (strings []string, e error) {
			return indexer(obj.(*v1.Secret))
		},
	}))
}

func (c *secretCache) GetByIndex(indexName, key string) (result []*v1.Secret, err error) {
	objs, err := c.indexer.ByIndex(indexName, key)
	if err != nil {
		return nil, err
	}
	for _, obj := range objs {
		result = append(result, obj.(*v1.Secret))
	}
	return result, nil
}
