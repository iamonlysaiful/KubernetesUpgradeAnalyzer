package preflight

import (
	"testing"

	authorizationv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/kubernetes"
	kubefake "k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"
)

func TestKubernetesCheckerServerVersion(t *testing.T) {
	client := fakeKubernetesClient()
	client.Discovery().(*fake.FakeDiscovery).FakedServerVersion = &version.Info{GitVersion: "v1.30.0"}

	got, err := NewKubernetesChecker(client).ServerVersion(ContextSelection{})
	if err != nil {
		t.Fatalf("ServerVersion returned error: %v", err)
	}
	if got != "v1.30.0" {
		t.Fatalf("ServerVersion = %q, want v1.30.0", got)
	}
}

func TestKubernetesCheckerDiscovery(t *testing.T) {
	client := fakeKubernetesClient()

	if err := NewKubernetesChecker(client).Discovery(ContextSelection{}); err != nil {
		t.Fatalf("Discovery returned error: %v", err)
	}
}

func TestKubernetesCheckerPermissions(t *testing.T) {
	client := fakeKubernetesClient()
	denyResource := "events"
	addAccessReviewReactor(t, client, denyResource)

	checks, err := NewKubernetesChecker(client).Permissions(ContextSelection{})
	if err != nil {
		t.Fatalf("Permissions returned error: %v", err)
	}
	if len(checks) == 0 {
		t.Fatal("Permissions returned no checks")
	}

	var sawDeniedOptional bool
	for _, check := range checks {
		if check.Resource == denyResource {
			sawDeniedOptional = true
			if check.Status != StatusFail {
				t.Fatalf("events Status = %q, want %q", check.Status, StatusFail)
			}
			if check.EvidenceClass != EvidenceOptional {
				t.Fatalf("events EvidenceClass = %q, want %q", check.EvidenceClass, EvidenceOptional)
			}
			continue
		}
		if check.Status != StatusPass {
			t.Fatalf("%s Status = %q, want %q", check.Resource, check.Status, StatusPass)
		}
	}
	if !sawDeniedOptional {
		t.Fatal("Permissions did not include denied events check")
	}
}

func fakeKubernetesClient() kubernetes.Interface {
	client := kubefake.NewSimpleClientset()
	discoveryClient := client.Discovery().(*fake.FakeDiscovery)
	discoveryClient.Resources = []*metav1.APIResourceList{
		{
			GroupVersion: "v1",
			APIResources: []metav1.APIResource{
				{Name: "pods", Verbs: []string{"get", "list"}},
				{Name: "nodes", Verbs: []string{"get", "list"}},
			},
		},
		{
			GroupVersion: "apps/v1",
			APIResources: []metav1.APIResource{
				{Name: "deployments", Verbs: []string{"get", "list"}},
			},
		},
	}
	return client
}

func addAccessReviewReactor(t *testing.T, client kubernetes.Interface, denyResource string) {
	t.Helper()

	client.(*kubefake.Clientset).PrependReactor("create", "selfsubjectaccessreviews", func(action clienttesting.Action) (bool, runtime.Object, error) {
		createAction, ok := action.(clienttesting.CreateAction)
		if !ok {
			t.Fatalf("action is %T, want CreateAction", action)
		}
		review, ok := createAction.GetObject().(*authorizationv1.SelfSubjectAccessReview)
		if !ok {
			t.Fatalf("object is %T, want SelfSubjectAccessReview", createAction.GetObject())
		}
		resource := review.Spec.ResourceAttributes.Resource
		allowed := resource != denyResource

		return true, &authorizationv1.SelfSubjectAccessReview{
			ObjectMeta: metav1.ObjectMeta{
				Name: "synthetic-review",
			},
			Status: authorizationv1.SubjectAccessReviewStatus{
				Allowed: allowed,
				Reason:  "synthetic fake-client response",
			},
		}, nil
	})
}

func TestRequiredPermissionChecksDoNotUseSecretOrWatch(t *testing.T) {
	for _, check := range requiredPermissionChecks() {
		if check.Resource == "secrets" {
			t.Fatal("requiredPermissionChecks includes secrets")
		}
		if check.Verb == "watch" {
			t.Fatalf("requiredPermissionChecks includes watch for %s", check.Resource)
		}
		if check.Resource == "" {
			t.Fatal("requiredPermissionChecks includes empty resource")
		}
	}
}
