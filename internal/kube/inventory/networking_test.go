package inventory

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCollectNetworkingBuildsDeterministicRefs(t *testing.T) {
	client := fake.NewSimpleClientset(
		&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "team-b"}},
		&networkingv1.Ingress{ObjectMeta: metav1.ObjectMeta{Name: "edge", Namespace: "team-a"}},
		&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "db", Namespace: "team-a"}},
	)

	networking, err := NewCollector(client).collectNetworking(context.Background())
	if err != nil {
		t.Fatalf("collectNetworking returned error: %v", err)
	}

	if got := resourceKeys(networking); got != "team-a/Ingress/edge,team-a/Service/db,team-b/Service/api" {
		t.Fatalf("networking refs = %q", got)
	}
	if networking[0].APIVersion != "networking.k8s.io/v1" || networking[1].APIVersion != "v1" {
		t.Fatalf("networking apiVersions = %#v", networking)
	}
}

func resourceKeys(refs []ResourceRef) string {
	var result string
	for i, ref := range refs {
		if i > 0 {
			result += ","
		}
		result += ref.Namespace + "/" + ref.Kind + "/" + ref.Name
	}
	return result
}
