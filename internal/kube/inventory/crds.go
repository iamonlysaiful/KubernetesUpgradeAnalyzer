package inventory

import (
	"context"
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c Collector) collectCRDs(ctx context.Context) ([]ResourceRef, error) {
	list, err := c.APIExtensionsClient.ApiextensionsV1().CustomResourceDefinitions().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	crds := make([]ResourceRef, 0, len(list.Items))
	for _, crd := range list.Items {
		crds = append(crds, ResourceRef{
			APIVersion: "apiextensions.k8s.io/v1",
			Kind:       "CustomResourceDefinition",
			Name:       crd.Name,
		})
	}
	sort.Slice(crds, func(i, j int) bool {
		return crds[i].Name < crds[j].Name
	})
	return crds, nil
}
