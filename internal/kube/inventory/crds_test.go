package inventory

import (
	"context"
	"testing"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsfake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCollectCRDsBuildsDeterministicRefs(t *testing.T) {
	extensionsClient := apiextensionsfake.NewSimpleClientset(
		&apiextensionsv1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{Name: "widgets.example.test"},
		},
		&apiextensionsv1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{Name: "alerts.example.test"},
		},
	)

	crds, err := NewCollectorWithAPIExtensions(fake.NewSimpleClientset(), extensionsClient).collectCRDs(context.Background())
	if err != nil {
		t.Fatalf("collectCRDs returned error: %v", err)
	}

	if got := names(crds); got != "alerts.example.test,widgets.example.test" {
		t.Fatalf("CRDs = %q, want alerts.example.test,widgets.example.test", got)
	}
	if crds[0].APIVersion != "apiextensions.k8s.io/v1" ||
		crds[0].Kind != "CustomResourceDefinition" ||
		crds[0].Namespace != "" {
		t.Fatalf("CRD ref = %#v", crds[0])
	}
}
