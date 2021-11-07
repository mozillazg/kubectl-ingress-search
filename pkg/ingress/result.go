package ingress

import (
	"fmt"

	"github.com/mozillazg/kubectl-ingress-search/pkg/types"
	v1 "k8s.io/api/networking/v1"
)

type Backend struct {
	v1.IngressBackend
}

type Rule struct {
	Namespace types.Value
	Name      types.Value
	Host      types.Value
	Path      types.Value
	Backend   types.Value
}

func ParseRules(ingress v1.Ingress) []Rule {
	var rs []Rule
	for _, rule := range ingress.Spec.Rules {
		if rule.IngressRuleValue.HTTP == nil {
			continue
		}
		for _, p := range rule.IngressRuleValue.HTTP.Paths {
			rs = append(rs, Rule{
				Namespace: types.String(ingress.Namespace).ToValue(),
				Name:      types.String(ingress.Name).ToValue(),
				Host:      types.String(rule.Host).ToValue(),
				Path:      types.String(p.Path).ToValue(),
				Backend: (Backend{
					IngressBackend: p.Backend,
				}).ToValue(),
			})
		}
	}
	return rs
}

func (b Backend) String() string {
	if b.Service != nil {
		service := b.Service
		port := service.Port.Name
		if service.Port.Number > 0 {
			port = fmt.Sprintf("%d", service.Port.Number)
		}
		return fmt.Sprintf("Service/%s:%s", service.Name, port)
	}

	if b.Resource != nil {
		resource := b.Resource
		if resource.APIGroup != nil {
			return fmt.Sprintf("%s/%s/%s", *resource.APIGroup, resource.Kind, resource.Name)
		} else {
			return fmt.Sprintf("%s/%s", resource.Kind, resource.Name)
		}
	}
	return ""
}

func (b Backend) ToValue() types.Value {
	return types.Value{V: b}
}
