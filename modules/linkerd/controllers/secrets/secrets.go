package secrets

import (
	"context"

	"github.com/rancher/rio/modules/system/features/letsencrypt/pkg/issuers"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/rancher/rio/types"
	corev1controller "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
)

func Register(ctx context.Context, rContext *types.Context) error {
	h := handler{
		systemNamespace: rContext.Namespace,
		secrets:         rContext.Core.Core().V1().Secret(),
		namespaces:      rContext.Core.Core().V1().Namespace().Cache(),
	}

	rContext.Core.Core().V1().Secret().OnChange(ctx, "sync-rio-wildcards-linkerd", h.syncSecrets)
	return nil
}

type handler struct {
	systemNamespace string
	secrets         corev1controller.SecretController
	namespaces      corev1controller.NamespaceCache
}

func (h handler) syncSecrets(key string, secret *v1.Secret) (*v1.Secret, error) {
	if secret == nil || secret.DeletionTimestamp != nil {
		return secret, nil
	}

	if secret.Name == issuers.RioWildcardCerts && secret.Namespace == h.systemNamespace {
		nss, err := h.namespaces.List(labels.NewSelector())
		if err != nil {
			return secret, err
		}
		for _, ns := range nss {
			if existing, err := h.secrets.Cache().Get(ns.Name, secret.Name); err == nil {
				existing.Data = secret.Data
				if _, err := h.secrets.Update(existing); err != nil {
					return secret, err
				}
			}
		}
	}
	return secret, nil
}
