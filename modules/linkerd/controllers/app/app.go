package app

import (
	"context"
	"fmt"
	"strconv"

	splitv1alpha1 "github.com/deislabs/smi-sdk-go/pkg/apis/split/v1alpha1"
	"github.com/rancher/rio/modules/istio/pkg/domains"
	"github.com/rancher/rio/modules/linkerd/controllers/services/populate"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/constructors"
	projectv1controller "github.com/rancher/rio/pkg/generated/controllers/admin.rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	corev1controller "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	"github.com/rancher/wrangler/pkg/objectset"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	c := stackobject.NewGeneratingController(ctx, rContext, "routing-serviceset", rContext.Rio.Rio().V1().App())
	c.Apply = c.Apply.WithStrictCaching().
		WithCacheTypes(rContext.K8sNetworking.Networking().V1beta1().Ingress(),
			rContext.SMI.Split().V1alpha1().TrafficSplit()).WithRateLimiting(10)

	sh := &appHandler{
		systemNamespace:    rContext.Namespace,
		clusterDomainCache: rContext.Global.Admin().V1().ClusterDomain().Cache(),
		serviceCache:       rContext.Rio.Rio().V1().Service().Cache(),
		secrets:            rContext.Core.Core().V1().Secret(),
	}

	c.Populator = sh.populate
	return nil
}

type appHandler struct {
	systemNamespace    string
	clusterDomainCache projectv1controller.ClusterDomainCache
	serviceCache       v1.ServiceCache
	secrets            corev1controller.SecretController
}

func (a appHandler) populate(obj runtime.Object, namespace *corev1.Namespace, os *objectset.ObjectSet) error {
	app := obj.(*riov1.App)

	clusterDomain, err := a.clusterDomainCache.Get(a.systemNamespace, constants.ClusterDomainName)
	if err != nil {
		return err
	}

	if clusterDomain.Status.ClusterDomain == "" {
		return nil
	}

	if app.Namespace != a.systemNamespace {
		split := constructors.NewTrafficSplit(app.Namespace, app.Name, splitv1alpha1.TrafficSplit{
			Spec: splitv1alpha1.TrafficSplitSpec{
				Service: app.Name,
			},
		})
		for ver, rev := range app.Status.RevisionWeight {
			split.Spec.Backends = append(split.Spec.Backends, splitv1alpha1.TrafficSplitBackend{
				Service: fmt.Sprintf("%s-%s", app.Name, ver),
				Weight:  resource.MustParse(strconv.Itoa(rev.Weight)),
			})
		}
		os.Add(split)
	}

	var revisions []*riov1.Service
	for i := len(app.Spec.Revisions) - 1; i >= 0; i-- {
		revision, err := a.serviceCache.Get(app.Namespace, app.Spec.Revisions[i].ServiceName)
		if err != nil && !errors.IsNotFound(err) {
			return err
		} else if errors.IsNotFound(err) {
			continue
		}
		revisions = append(revisions, revision)
	}
	if len(revisions) == 0 || !domains.IsPublic(revisions[0]) {
		return nil
	}

	certName := ""
	if constants.InstallMode != constants.InstallModeIngress {
		certName = clusterDomain.Spec.SecretRef.Name
	}

	if _, err := a.secrets.Cache().Get(revisions[0].Namespace, certName); err != nil {
		if errors.IsNotFound(err) {
			if existing, err := a.secrets.Cache().Get(a.systemNamespace, certName); err == nil {
				secret := constructors.NewSecret(revisions[0].Namespace, certName, corev1.Secret{
					Data: existing.Data,
				})
				if _, err := a.secrets.Create(secret); err != nil && !errors.IsAlreadyExists(err) {
					return err
				}
			}
		}
	}

	populate.IngressForApp(clusterDomain.Status.ClusterDomain, certName, app, revisions, os)

	return err
}
