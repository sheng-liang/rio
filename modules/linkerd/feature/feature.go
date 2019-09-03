package feature

import (
	"context"

	"github.com/rancher/wrangler/pkg/start"

	"github.com/rancher/rio/modules/linkerd/controllers/secrets"

	"github.com/rancher/rio/modules/linkerd/controllers/app"
	"github.com/rancher/rio/modules/linkerd/controllers/services"
	v1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/pkg/stack"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	feature := &features.FeatureController{
		FeatureName: "linkerd",
		FeatureSpec: v1.FeatureSpec{
			Description: "linkerd service mesh",
			Enabled:     constants.ServiceMeshMode == "linkerd",
		},
		SystemStacks: []*stack.SystemStack{
			stack.NewSystemStack(rContext.Apply, rContext.Namespace, "linkerd-crd"),
			stack.NewSystemStack(rContext.Apply, rContext.Namespace, "linkerd"),
		},
		Controllers: []features.ControllerRegister{
			app.Register,
			services.Register,
			secrets.Register,
		},
		FixedAnswers: map[string]string{
			"NAMESPACE":    rContext.Namespace,
			"INSTALL_MODE": constants.InstallMode,
			"HTTP_PORT":    constants.DefaultHTTPOpenPort,
			"HTTPS_PORT":   constants.DefaultHTTPSOpenPort,
		},
		OnStart: func(feature *v1.Feature) error {
			return start.All(ctx, 5,
				rContext.SMI,
				rContext.K8sNetworking)
		},
	}
	return feature.Register()
}
