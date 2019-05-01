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
	v1 "github.com/rancher/rio/pkg/apis/autoscale.rio.cattle.io/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// ServiceScaleRecommendationLister helps list ServiceScaleRecommendations.
type ServiceScaleRecommendationLister interface {
	// List lists all ServiceScaleRecommendations in the indexer.
	List(selector labels.Selector) (ret []*v1.ServiceScaleRecommendation, err error)
	// ServiceScaleRecommendations returns an object that can list and get ServiceScaleRecommendations.
	ServiceScaleRecommendations(namespace string) ServiceScaleRecommendationNamespaceLister
	ServiceScaleRecommendationListerExpansion
}

// serviceScaleRecommendationLister implements the ServiceScaleRecommendationLister interface.
type serviceScaleRecommendationLister struct {
	indexer cache.Indexer
}

// NewServiceScaleRecommendationLister returns a new ServiceScaleRecommendationLister.
func NewServiceScaleRecommendationLister(indexer cache.Indexer) ServiceScaleRecommendationLister {
	return &serviceScaleRecommendationLister{indexer: indexer}
}

// List lists all ServiceScaleRecommendations in the indexer.
func (s *serviceScaleRecommendationLister) List(selector labels.Selector) (ret []*v1.ServiceScaleRecommendation, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.ServiceScaleRecommendation))
	})
	return ret, err
}

// ServiceScaleRecommendations returns an object that can list and get ServiceScaleRecommendations.
func (s *serviceScaleRecommendationLister) ServiceScaleRecommendations(namespace string) ServiceScaleRecommendationNamespaceLister {
	return serviceScaleRecommendationNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// ServiceScaleRecommendationNamespaceLister helps list and get ServiceScaleRecommendations.
type ServiceScaleRecommendationNamespaceLister interface {
	// List lists all ServiceScaleRecommendations in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1.ServiceScaleRecommendation, err error)
	// Get retrieves the ServiceScaleRecommendation from the indexer for a given namespace and name.
	Get(name string) (*v1.ServiceScaleRecommendation, error)
	ServiceScaleRecommendationNamespaceListerExpansion
}

// serviceScaleRecommendationNamespaceLister implements the ServiceScaleRecommendationNamespaceLister
// interface.
type serviceScaleRecommendationNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all ServiceScaleRecommendations in the indexer for a given namespace.
func (s serviceScaleRecommendationNamespaceLister) List(selector labels.Selector) (ret []*v1.ServiceScaleRecommendation, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.ServiceScaleRecommendation))
	})
	return ret, err
}

// Get retrieves the ServiceScaleRecommendation from the indexer for a given namespace and name.
func (s serviceScaleRecommendationNamespaceLister) Get(name string) (*v1.ServiceScaleRecommendation, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1.Resource("servicescalerecommendation"), name)
	}
	return obj.(*v1.ServiceScaleRecommendation), nil
}
