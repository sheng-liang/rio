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

	v1 "github.com/rancher/rio/pkg/apis/webhookinator.rio.cattle.io/v1"
	clientset "github.com/rancher/rio/pkg/generated/clientset/versioned/typed/webhookinator.rio.cattle.io/v1"
	informers "github.com/rancher/rio/pkg/generated/informers/externalversions/webhookinator.rio.cattle.io/v1"
	listers "github.com/rancher/rio/pkg/generated/listers/webhookinator.rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/generic"
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

type GitWebHookExecutionHandler func(string, *v1.GitWebHookExecution) (*v1.GitWebHookExecution, error)

type GitWebHookExecutionController interface {
	GitWebHookExecutionClient

	OnChange(ctx context.Context, name string, sync GitWebHookExecutionHandler)
	OnRemove(ctx context.Context, name string, sync GitWebHookExecutionHandler)
	Enqueue(namespace, name string)

	Cache() GitWebHookExecutionCache

	Informer() cache.SharedIndexInformer
	GroupVersionKind() schema.GroupVersionKind

	AddGenericHandler(ctx context.Context, name string, handler generic.Handler)
	AddGenericRemoveHandler(ctx context.Context, name string, handler generic.Handler)
	Updater() generic.Updater
}

type GitWebHookExecutionClient interface {
	Create(*v1.GitWebHookExecution) (*v1.GitWebHookExecution, error)
	Update(*v1.GitWebHookExecution) (*v1.GitWebHookExecution, error)
	UpdateStatus(*v1.GitWebHookExecution) (*v1.GitWebHookExecution, error)
	Delete(namespace, name string, options *metav1.DeleteOptions) error
	Get(namespace, name string, options metav1.GetOptions) (*v1.GitWebHookExecution, error)
	List(namespace string, opts metav1.ListOptions) (*v1.GitWebHookExecutionList, error)
	Watch(namespace string, opts metav1.ListOptions) (watch.Interface, error)
	Patch(namespace, name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.GitWebHookExecution, err error)
}

type GitWebHookExecutionCache interface {
	Get(namespace, name string) (*v1.GitWebHookExecution, error)
	List(namespace string, selector labels.Selector) ([]*v1.GitWebHookExecution, error)

	AddIndexer(indexName string, indexer GitWebHookExecutionIndexer)
	GetByIndex(indexName, key string) ([]*v1.GitWebHookExecution, error)
}

type GitWebHookExecutionIndexer func(obj *v1.GitWebHookExecution) ([]string, error)

type gitWebHookExecutionController struct {
	controllerManager *generic.ControllerManager
	clientGetter      clientset.GitWebHookExecutionsGetter
	informer          informers.GitWebHookExecutionInformer
	gvk               schema.GroupVersionKind
}

func NewGitWebHookExecutionController(gvk schema.GroupVersionKind, controllerManager *generic.ControllerManager, clientGetter clientset.GitWebHookExecutionsGetter, informer informers.GitWebHookExecutionInformer) GitWebHookExecutionController {
	return &gitWebHookExecutionController{
		controllerManager: controllerManager,
		clientGetter:      clientGetter,
		informer:          informer,
		gvk:               gvk,
	}
}

func FromGitWebHookExecutionHandlerToHandler(sync GitWebHookExecutionHandler) generic.Handler {
	return func(key string, obj runtime.Object) (ret runtime.Object, err error) {
		var v *v1.GitWebHookExecution
		if obj == nil {
			v, err = sync(key, nil)
		} else {
			v, err = sync(key, obj.(*v1.GitWebHookExecution))
		}
		if v == nil {
			return nil, err
		}
		return v, err
	}
}

func (c *gitWebHookExecutionController) Updater() generic.Updater {
	return func(obj runtime.Object) (runtime.Object, error) {
		newObj, err := c.Update(obj.(*v1.GitWebHookExecution))
		if newObj == nil {
			return nil, err
		}
		return newObj, err
	}
}

func UpdateGitWebHookExecutionOnChange(updater generic.Updater, handler GitWebHookExecutionHandler) GitWebHookExecutionHandler {
	return func(key string, obj *v1.GitWebHookExecution) (*v1.GitWebHookExecution, error) {
		if obj == nil {
			return handler(key, nil)
		}

		copyObj := obj.DeepCopy()
		newObj, err := handler(key, copyObj)
		if newObj != nil {
			copyObj = newObj
		}
		if obj.ResourceVersion == copyObj.ResourceVersion && !equality.Semantic.DeepEqual(obj, copyObj) {
			newObj, err := updater(copyObj)
			if newObj != nil && err == nil {
				copyObj = newObj.(*v1.GitWebHookExecution)
			}
		}

		return copyObj, err
	}
}

func (c *gitWebHookExecutionController) AddGenericHandler(ctx context.Context, name string, handler generic.Handler) {
	c.controllerManager.AddHandler(ctx, c.gvk, c.informer.Informer(), name, handler)
}

func (c *gitWebHookExecutionController) AddGenericRemoveHandler(ctx context.Context, name string, handler generic.Handler) {
	removeHandler := generic.NewRemoveHandler(name, c.Updater(), handler)
	c.controllerManager.AddHandler(ctx, c.gvk, c.informer.Informer(), name, removeHandler)
}

func (c *gitWebHookExecutionController) OnChange(ctx context.Context, name string, sync GitWebHookExecutionHandler) {
	c.AddGenericHandler(ctx, name, FromGitWebHookExecutionHandlerToHandler(sync))
}

func (c *gitWebHookExecutionController) OnRemove(ctx context.Context, name string, sync GitWebHookExecutionHandler) {
	removeHandler := generic.NewRemoveHandler(name, c.Updater(), FromGitWebHookExecutionHandlerToHandler(sync))
	c.AddGenericHandler(ctx, name, removeHandler)
}

func (c *gitWebHookExecutionController) Enqueue(namespace, name string) {
	c.controllerManager.Enqueue(c.gvk, namespace, name)
}

func (c *gitWebHookExecutionController) Informer() cache.SharedIndexInformer {
	return c.informer.Informer()
}

func (c *gitWebHookExecutionController) GroupVersionKind() schema.GroupVersionKind {
	return c.gvk
}

func (c *gitWebHookExecutionController) Cache() GitWebHookExecutionCache {
	return &gitWebHookExecutionCache{
		lister:  c.informer.Lister(),
		indexer: c.informer.Informer().GetIndexer(),
	}
}

func (c *gitWebHookExecutionController) Create(obj *v1.GitWebHookExecution) (*v1.GitWebHookExecution, error) {
	return c.clientGetter.GitWebHookExecutions(obj.Namespace).Create(obj)
}

func (c *gitWebHookExecutionController) Update(obj *v1.GitWebHookExecution) (*v1.GitWebHookExecution, error) {
	return c.clientGetter.GitWebHookExecutions(obj.Namespace).Update(obj)
}

func (c *gitWebHookExecutionController) UpdateStatus(obj *v1.GitWebHookExecution) (*v1.GitWebHookExecution, error) {
	return c.clientGetter.GitWebHookExecutions(obj.Namespace).UpdateStatus(obj)
}

func (c *gitWebHookExecutionController) Delete(namespace, name string, options *metav1.DeleteOptions) error {
	return c.clientGetter.GitWebHookExecutions(namespace).Delete(name, options)
}

func (c *gitWebHookExecutionController) Get(namespace, name string, options metav1.GetOptions) (*v1.GitWebHookExecution, error) {
	return c.clientGetter.GitWebHookExecutions(namespace).Get(name, options)
}

func (c *gitWebHookExecutionController) List(namespace string, opts metav1.ListOptions) (*v1.GitWebHookExecutionList, error) {
	return c.clientGetter.GitWebHookExecutions(namespace).List(opts)
}

func (c *gitWebHookExecutionController) Watch(namespace string, opts metav1.ListOptions) (watch.Interface, error) {
	return c.clientGetter.GitWebHookExecutions(namespace).Watch(opts)
}

func (c *gitWebHookExecutionController) Patch(namespace, name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.GitWebHookExecution, err error) {
	return c.clientGetter.GitWebHookExecutions(namespace).Patch(name, pt, data, subresources...)
}

type gitWebHookExecutionCache struct {
	lister  listers.GitWebHookExecutionLister
	indexer cache.Indexer
}

func (c *gitWebHookExecutionCache) Get(namespace, name string) (*v1.GitWebHookExecution, error) {
	return c.lister.GitWebHookExecutions(namespace).Get(name)
}

func (c *gitWebHookExecutionCache) List(namespace string, selector labels.Selector) ([]*v1.GitWebHookExecution, error) {
	return c.lister.GitWebHookExecutions(namespace).List(selector)
}

func (c *gitWebHookExecutionCache) AddIndexer(indexName string, indexer GitWebHookExecutionIndexer) {
	utilruntime.Must(c.indexer.AddIndexers(map[string]cache.IndexFunc{
		indexName: func(obj interface{}) (strings []string, e error) {
			return indexer(obj.(*v1.GitWebHookExecution))
		},
	}))
}

func (c *gitWebHookExecutionCache) GetByIndex(indexName, key string) (result []*v1.GitWebHookExecution, err error) {
	objs, err := c.indexer.ByIndex(indexName, key)
	if err != nil {
		return nil, err
	}
	for _, obj := range objs {
		result = append(result, obj.(*v1.GitWebHookExecution))
	}
	return result, nil
}
