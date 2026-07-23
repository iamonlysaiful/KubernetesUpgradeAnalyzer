package inventory

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCollectStorageBuildsDeterministicRefs(t *testing.T) {
	client := fake.NewSimpleClientset(
		&corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: "data", Namespace: "team-b"}},
		&corev1.PersistentVolume{ObjectMeta: metav1.ObjectMeta{Name: "pv-001"}},
		&storagev1.StorageClass{ObjectMeta: metav1.ObjectMeta{Name: "managed-001"}},
		&corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: "cache", Namespace: "team-a"}},
	)

	storage, err := NewCollector(client).collectStorage(context.Background())
	if err != nil {
		t.Fatalf("collectStorage returned error: %v", err)
	}

	if got := storageKeys(storage); got != "/PersistentVolume/pv-001,/StorageClass/managed-001,team-a/PersistentVolumeClaim/cache,team-b/PersistentVolumeClaim/data" {
		t.Fatalf("storage refs = %q", got)
	}
	if storage[0].APIVersion != "v1" || storage[1].APIVersion != "storage.k8s.io/v1" || storage[2].APIVersion != "v1" {
		t.Fatalf("storage apiVersions = %#v", storage)
	}
}

func storageKeys(refs []ResourceRef) string {
	var result string
	for i, ref := range refs {
		if i > 0 {
			result += ","
		}
		result += ref.Namespace + "/" + ref.Kind + "/" + ref.Name
	}
	return result
}
