package tui

import (
	"context"

	"github.com/rancher/axe/throwing"
	"github.com/rancher/rio/cli/pkg/clicontext"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/leader"
	"k8s.io/apimachinery/pkg/runtime"
	runtime2 "k8s.io/apimachinery/pkg/util/runtime"
)

type Tui struct {
}

func (t *Tui) Run(ctx *clicontext.CLIContext) error {
	ss := map[string]chan struct{}{
		appKind:             make(chan struct{}, 0),
		serviceKind:         make(chan struct{}, 0),
		routeKind:           make(chan struct{}, 0),
		externalServiceKind: make(chan struct{}, 0),
		configKind:          make(chan struct{}, 0),
		podKind:             make(chan struct{}, 0),
		publicdomainKind:    make(chan struct{}, 0),
	}

	h := &handler{
		signals: ss,
	}

	rioContext := types.NewContext(ctx.SystemNamespace, ctx.RestConfig)
	go func() {
		leader.RunOrDie(ctx.Ctx, ctx.SystemNamespace, "rio-cli", rioContext.K8s, func(context context.Context) {
			register(context, rioContext, h)
			runtime2.Must(rioContext.Start(context))
			<-context.Done()
		})
	}()

	tui := throwing.NewAppView(ctx.K8s, drawer, tableEventHandler, ss)
	if err := tui.Init(); err != nil {
		return err
	}
	return tui.Run()
}

func register(ctx context.Context, rioContext *types.Context, h *handler) {
	rioContext.Rio.Rio().V1().App().AddGenericHandler(ctx, "rio-app-tui", h.syncObject)
	rioContext.Rio.Rio().V1().Service().AddGenericHandler(ctx, "rio-service-tui", h.syncObject)
	rioContext.Rio.Rio().V1().Router().AddGenericHandler(ctx, "rio-router-tui", h.syncObject)
	rioContext.Rio.Rio().V1().PublicDomain().AddGenericHandler(ctx, "rio-domain-tui", h.syncObject)
	rioContext.Rio.Rio().V1().ExternalService().AddGenericHandler(ctx, "rio-external-tui", h.syncObject)
}

type handler struct {
	signals map[string]chan struct{}
}

func (h handler) syncObject(k string, object runtime.Object) (runtime.Object, error) {
	switch object.(type) {
	case *riov1.Service:
		s := h.signals[serviceKind]
		go func() {
			s <- struct{}{}
		}()
	case *riov1.App:
		s1 := h.signals[appKind]
		s2 := h.signals[serviceKind]
		s3 := h.signals[podKind]

		go func() {
			s1 <- struct{}{}
			s2 <- struct{}{}
			s3 <- struct{}{}
		}()
	case *riov1.Router:
		s := h.signals[routeKind]
		go func() {
			s <- struct{}{}
		}()
	case *riov1.PublicDomain:
		s := h.signals[publicdomainKind]
		go func() {
			s <- struct{}{}
		}()
	case *riov1.ExternalService:
		s := h.signals[externalServiceKind]
		go func() {
			s <- struct{}{}
		}()
	}
	return object, nil
}
