package ingress

import (
	"context"

	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Searcher struct {
	client kubernetes.Interface
}

type apiVersion string

var (
	networkingV1      apiVersion = "networking.k8s.io/v1"
	networkingBetaV1  apiVersion = "networking.k8s.io/v1beta1"
	extensionsV1beta1 apiVersion = "extensions/v1beta1"
)

func NewSearcher(client kubernetes.Interface) *Searcher {
	return &Searcher{
		client: client,
	}
}

func (s *Searcher) ListIngresses(ctx context.Context, namespace string, opts metav1.ListOptions) (*v1.IngressList, error) {
	// TOOD: add cache
	vers, err := s.getSupportedIngressVersion()
	if err != nil {
		return nil, err
	}
	if versionsContain(vers, networkingV1) {
		return listIngress(ctx, s.client, namespace, opts, fetchIngressV1)
	} else {
		return listIngress(ctx, s.client, namespace, opts, fetchIngressBetaV1)
	}
}

func (s *Searcher) getSupportedIngressVersion() ([]apiVersion, error) {
	groups, err := s.client.Discovery().ServerGroups()
	if err != nil {
		return nil, err
	}
	allVersions := getApiVersions(groups.Groups)
	var versions []apiVersion
	for _, ver := range allVersions {
		switch ver {
		case networkingV1:
			versions = append(versions, networkingV1)
			break
		case networkingBetaV1:
			versions = append(versions, networkingBetaV1)
			break
		case extensionsV1beta1:
			versions = append(versions, extensionsV1beta1)
			break
		}
	}
	return versions, nil
}

func getApiVersions(groups []metav1.APIGroup) []apiVersion {
	var versions []apiVersion
	for _, g := range groups {
		for _, v := range g.Versions {
			ver := v.GroupVersion
			versions = append(versions, apiVersion(ver))
		}
	}
	return versions
}

func versionsContain(versions []apiVersion, ver apiVersion) bool {
	for _, v := range versions {
		if ver == v {
			return true
		}
	}
	return false
}
