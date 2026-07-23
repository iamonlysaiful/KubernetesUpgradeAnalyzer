package inventory

import (
	"context"
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c Collector) collectStorage(ctx context.Context) ([]ResourceRef, error) {
	pvcs, err := c.Client.CoreV1().PersistentVolumeClaims("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	pvs, err := c.Client.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	storageClasses, err := c.Client.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	storage := make([]ResourceRef, 0, len(pvcs.Items)+len(pvs.Items)+len(storageClasses.Items))
	for _, pvc := range pvcs.Items {
		storage = append(storage, ResourceRef{
			APIVersion: "v1",
			Kind:       "PersistentVolumeClaim",
			Namespace:  pvc.Namespace,
			Name:       pvc.Name,
		})
	}
	for _, pv := range pvs.Items {
		storage = append(storage, ResourceRef{
			APIVersion: "v1",
			Kind:       "PersistentVolume",
			Name:       pv.Name,
		})
	}
	for _, storageClass := range storageClasses.Items {
		storage = append(storage, ResourceRef{
			APIVersion: "storage.k8s.io/v1",
			Kind:       "StorageClass",
			Name:       storageClass.Name,
		})
	}

	sort.Slice(storage, func(i, j int) bool {
		if storage[i].Namespace != storage[j].Namespace {
			return storage[i].Namespace < storage[j].Namespace
		}
		if storage[i].Kind != storage[j].Kind {
			return storage[i].Kind < storage[j].Kind
		}
		return storage[i].Name < storage[j].Name
	})
	return storage, nil
}
