package services

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/rancher/rio/modules/linkerd/controllers/services/populate"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/constructors"
	adminv1controller "github.com/rancher/rio/pkg/generated/controllers/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	corev1controller "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	"github.com/rancher/wrangler/pkg/objectset"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	c := stackobject.NewGeneratingController(ctx, rContext, "routing-service", rContext.Rio.Rio().V1().Service())
	c.Apply = c.Apply.WithStrictCaching().
		WithCacheTypes(rContext.K8sNetworking.Networking().V1beta1().Ingress())

	sh := &serviceHandler{
		systemNamespace:    rContext.Namespace,
		clusterDomainCache: rContext.Global.Admin().V1().ClusterDomain().Cache(),
		secrets:            rContext.Core.Core().V1().Secret(),
	}

	c.Populator = sh.populate
	return nil
}

type serviceHandler struct {
	systemNamespace    string
	clusterDomainCache adminv1controller.ClusterDomainCache
	secrets            corev1controller.SecretController
}

func (s *serviceHandler) populate(obj runtime.Object, namespace *corev1.Namespace, os *objectset.ObjectSet) error {
	service := obj.(*riov1.Service)

	clusterDomain, err := s.clusterDomainCache.Get(s.systemNamespace, constants.ClusterDomainName)
	if err != nil {
		return err
	}

	if clusterDomain.Status.ClusterDomain == "" {
		return nil
	}

	certName := ""
	if constants.InstallMode != constants.InstallModeIngress {
		certName = clusterDomain.Spec.SecretRef.Name
	}

	if _, err := s.secrets.Cache().Get(service.Namespace, certName); err != nil {
		if errors.IsNotFound(err) {
			if existing, err := s.secrets.Cache().Get(s.systemNamespace, certName); err == nil {
				secret := constructors.NewSecret(service.Namespace, certName, corev1.Secret{
					Data: existing.Data,
				})
				if _, err := s.secrets.Create(secret); err != nil && !errors.IsAlreadyExists(err) {
					return err
				}
			}
		}
	}

	populate.IngressForService(clusterDomain.Status.ClusterDomain, certName, service, os)

	return err
}
