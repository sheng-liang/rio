package populate

import (
	"fmt"
	"strings"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/rancher/rio/modules/istio/pkg/domains"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/rio/pkg/services"
	"github.com/rancher/wrangler/pkg/objectset"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func IngressForService(domain, certName string, svc *riov1.Service, os *objectset.ObjectSet) {
	if !domains.IsPublic(svc) {
		return
	}
	app, version := services.AppAndVersion(svc)
	serviceName := app + "-" + version
	host := domains.GetExternalDomain(serviceName, svc.Namespace, domain)

	var port int32
	for _, p := range svc.Spec.Ports {
		if !p.InternalOnly {
			port = p.TargetPort
			break
		}
	}

	ingress := constructors.NewIngress(svc.Namespace, serviceName, networkingv1beta1.Ingress{
		ObjectMeta: v1.ObjectMeta{
			Annotations: map[string]string{
				"kubernetes.io/ingress.class": "linkerd-gateway",
			},
		},
		Spec: networkingv1beta1.IngressSpec{
			Rules: []networkingv1beta1.IngressRule{
				{
					Host: host,
					IngressRuleValue: networkingv1beta1.IngressRuleValue{
						HTTP: &networkingv1beta1.HTTPIngressRuleValue{
							Paths: []networkingv1beta1.HTTPIngressPath{
								{
									Backend: networkingv1beta1.IngressBackend{
										ServiceName: serviceName,
										ServicePort: intstr.FromInt(int(port)),
									},
								},
							},
						},
					},
				},
			},
		},
	})

	if certName != "" {
		ingress.Spec.TLS = []networkingv1beta1.IngressTLS{
			{
				Hosts:      []string{fmt.Sprintf("*.%s", domain)},
				SecretName: certName,
			},
		}
	}

	os.Add(ingress)
	return
}

func IngressForApp(domain, certName string, app *riov1.App, revisions []*riov1.Service, os *objectset.ObjectSet) {
	host := domains.GetExternalDomain(app.Name, app.Namespace, domain)
	ingress := constructors.NewIngress(app.Namespace, app.Name, networkingv1beta1.Ingress{
		ObjectMeta: v1.ObjectMeta{
			Annotations: map[string]string{
				"kubernetes.io/ingress.class": "linkerd-gateway",
			},
		},
		Spec: networkingv1beta1.IngressSpec{
			Rules: []networkingv1beta1.IngressRule{
				{
					Host: host,
					IngressRuleValue: networkingv1beta1.IngressRuleValue{
						HTTP: &networkingv1beta1.HTTPIngressRuleValue{
							Paths: []networkingv1beta1.HTTPIngressPath{},
						},
					},
				},
			},
		},
	})

	weightBuffer := strings.Builder{}
	for _, rev := range revisions {
		_, ver := services.AppAndVersion(rev)
		if w, ok := app.Status.RevisionWeight[ver]; ok {
			var port int32
			for _, p := range rev.Spec.Ports {
				if !p.InternalOnly {
					port = p.TargetPort
					break
				}
			}

			ingress.Spec.Rules[0].HTTP.Paths = append(ingress.Spec.Rules[0].HTTP.Paths, networkingv1beta1.HTTPIngressPath{
				Backend: networkingv1beta1.IngressBackend{
					ServiceName: app.Name + "-" + ver,
					ServicePort: intstr.FromInt(int(port)),
				},
			})
			weightBuffer.WriteString(fmt.Sprintf("%s-%s: %v", app.Name, ver, w.Weight))
			weightBuffer.WriteString("\n")
		}
	}

	ingress.Annotations["traefik.ingress.kubernetes.io/service-weights"] = weightBuffer.String()
	if certName != "" {
		ingress.Spec.TLS = []networkingv1beta1.IngressTLS{
			{
				Hosts:      []string{fmt.Sprintf("*.%s", domain)},
				SecretName: certName,
			},
		}
	}

	os.Add(ingress)
	return
}
