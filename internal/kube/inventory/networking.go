package inventory

import (
	"context"
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c Collector) collectNetworking(ctx context.Context) ([]ResourceRef, error) {
	services, err := c.Client.CoreV1().Services("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	ingresses, err := c.Client.NetworkingV1().Ingresses("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	networking := make([]ResourceRef, 0, len(services.Items)+len(ingresses.Items))
	for _, service := range services.Items {
		networking = append(networking, ResourceRef{
			APIVersion: "v1",
			Kind:       "Service",
			Namespace:  service.Namespace,
			Name:       service.Name,
		})
	}
	for _, ingress := range ingresses.Items {
		networking = append(networking, ResourceRef{
			APIVersion: "networking.k8s.io/v1",
			Kind:       "Ingress",
			Namespace:  ingress.Namespace,
			Name:       ingress.Name,
		})
	}

	sort.Slice(networking, func(i, j int) bool {
		if networking[i].Namespace != networking[j].Namespace {
			return networking[i].Namespace < networking[j].Namespace
		}
		if networking[i].Kind != networking[j].Kind {
			return networking[i].Kind < networking[j].Kind
		}
		return networking[i].Name < networking[j].Name
	})
	return networking, nil
}
