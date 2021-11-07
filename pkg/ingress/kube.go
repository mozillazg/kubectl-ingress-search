package ingress

import (
	"context"

	v1 "k8s.io/api/networking/v1"
	"k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type fetchIngressFunc func(ctx context.Context, client kubernetes.Interface, namespace string, opts metav1.ListOptions) (*v1.IngressList, error)

func listIngress(ctx context.Context, client kubernetes.Interface, namespace string, opts metav1.ListOptions, fetchFunc fetchIngressFunc) (*v1.IngressList, error) {
	result := &v1.IngressList{}
	continueMark := ""
	limit := int64(500)
	for {
		if opts.Continue == "" {
			opts.Continue = continueMark
		}
		if opts.Limit == 0 {
			opts.Limit = limit
		}
		ret, err := fetchFunc(ctx, client, namespace, opts)
		if err != nil {
			return nil, err
		}
		result.Items = append(result.Items, ret.Items...)
		if ret.Continue == "" {
			break
		}
	}
	return result, nil
}

func fetchIngressV1(ctx context.Context, client kubernetes.Interface, namespace string, opts metav1.ListOptions) (*v1.IngressList, error) {
	ret, err := client.NetworkingV1().Ingresses(namespace).List(ctx, opts)
	return ret, err
}

func fetchIngressBetaV1(ctx context.Context, client kubernetes.Interface, namespace string, opts metav1.ListOptions) (*v1.IngressList, error) {
	ret, err := client.NetworkingV1beta1().Ingresses(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}
	newResp := convertIngressV1beta1ToV1(*ret)
	return &newResp, err
}

func convertIngressV1beta1ToV1(old v1beta1.IngressList) v1.IngressList {
	v1List := v1.IngressList{
		TypeMeta: old.TypeMeta,
		ListMeta: old.ListMeta,
		Items:    nil,
	}
	for _, item := range old.Items {
		spec := item.Spec
		var tls []v1.IngressTLS
		var rules []v1.IngressRule
		for _, x := range spec.TLS {
			tls = append(tls, v1.IngressTLS{
				Hosts:      x.Hosts,
				SecretName: x.SecretName,
			})
		}
		for _, x := range spec.Rules {
			r := v1.IngressRule{
				Host: x.Host,
			}
			if x.HTTP != nil {
				var paths []v1.HTTPIngressPath
				for _, p := range x.HTTP.Paths {
					path := v1.HTTPIngressPath{
						Path: p.Path,
					}
					if p.PathType != nil {
						t := v1.PathType(*p.PathType)
						path.PathType = &t
					}
					backend := v1.IngressBackend{
						Service:  nil,
						Resource: nil,
					}
					if p.Backend.ServiceName != "" {
						port := int32(p.Backend.ServicePort.IntValue())
						portName := p.Backend.ServicePort.String()
						if port > 0 {
							portName = ""
						}
						backend.Service = &v1.IngressServiceBackend{
							Name: p.Backend.ServiceName,
							Port: v1.ServiceBackendPort{
								Number: port,
								Name:   portName,
							},
						}
					}
					if p.Backend.Resource != nil {
						backend.Resource = p.Backend.Resource
					}
					path.Backend = backend
					paths = append(paths, path)
				}
				r.IngressRuleValue = v1.IngressRuleValue{
					HTTP: &v1.HTTPIngressRuleValue{
						Paths: paths,
					},
				}
			}
			rules = append(rules, r)
		}
		newItem := v1.Ingress{
			TypeMeta:   item.TypeMeta,
			ObjectMeta: item.ObjectMeta,
			Spec: v1.IngressSpec{
				IngressClassName: spec.IngressClassName,
				DefaultBackend:   nil,
				TLS:              tls,
				Rules:            rules,
			},
			Status: v1.IngressStatus{
				LoadBalancer: item.Status.LoadBalancer,
			},
		}
		v1List.Items = append(v1List.Items, newItem)
	}
	return v1List
}
