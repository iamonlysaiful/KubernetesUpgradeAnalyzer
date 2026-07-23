package preflight

import (
	"strings"
	"testing"
)

const syntheticKubeconfig = `
apiVersion: v1
kind: Config
clusters:
- name: cluster-a
  cluster:
    server: https://cluster-a.example.invalid
- name: cluster-b
  cluster:
    server: https://cluster-b.example.invalid
users:
- name: user-a
  user: {}
- name: user-b
  user: {}
contexts:
- name: context-a
  context:
    cluster: cluster-a
    user: user-a
    namespace: ns-a
- name: context-b
  context:
    cluster: cluster-b
    user: user-b
current-context: context-a
`

func TestResolveContextFromBytesUsesCurrentContext(t *testing.T) {
	selected, err := ResolveContextFromBytes([]byte(syntheticKubeconfig), KubeconfigSourceDefault, "")
	if err != nil {
		t.Fatalf("ResolveContextFromBytes returned error: %v", err)
	}

	if selected.Name != "context-a" {
		t.Fatalf("selected.Name = %q, want context-a", selected.Name)
	}
	if selected.ClusterName != "cluster-a" {
		t.Fatalf("selected.ClusterName = %q, want cluster-a", selected.ClusterName)
	}
	if selected.UserName != "user-a" {
		t.Fatalf("selected.UserName = %q, want user-a", selected.UserName)
	}
	if selected.Namespace != "ns-a" {
		t.Fatalf("selected.Namespace = %q, want ns-a", selected.Namespace)
	}
	if selected.KubeconfigSource != KubeconfigSourceDefault {
		t.Fatalf("selected.KubeconfigSource = %q, want %q", selected.KubeconfigSource, KubeconfigSourceDefault)
	}
}

func TestResolveContextFromBytesUsesExplicitContext(t *testing.T) {
	selected, err := ResolveContextFromBytes([]byte(syntheticKubeconfig), KubeconfigSourceExplicit, "context-b")
	if err != nil {
		t.Fatalf("ResolveContextFromBytes returned error: %v", err)
	}

	if selected.Name != "context-b" {
		t.Fatalf("selected.Name = %q, want context-b", selected.Name)
	}
	if selected.ClusterName != "cluster-b" {
		t.Fatalf("selected.ClusterName = %q, want cluster-b", selected.ClusterName)
	}
	if selected.KubeconfigSource != KubeconfigSourceExplicit {
		t.Fatalf("selected.KubeconfigSource = %q, want %q", selected.KubeconfigSource, KubeconfigSourceExplicit)
	}
}

func TestResolveContextFromBytesRejectsMissingContext(t *testing.T) {
	_, err := ResolveContextFromBytes([]byte(syntheticKubeconfig), KubeconfigSourceDefault, "missing")
	if err == nil {
		t.Fatal("ResolveContextFromBytes returned nil error, want missing context error")
	}
	if !strings.Contains(err.Error(), `context "missing" not found`) {
		t.Fatalf("error = %q, want missing context", err.Error())
	}
}

func TestResolveContextFromBytesRejectsMissingCurrentContext(t *testing.T) {
	kubeconfig := strings.Replace(syntheticKubeconfig, "current-context: context-a", "current-context: \"\"", 1)

	_, err := ResolveContextFromBytes([]byte(kubeconfig), KubeconfigSourceDefault, "")
	if err == nil {
		t.Fatal("ResolveContextFromBytes returned nil error, want missing current context error")
	}
	if !strings.Contains(err.Error(), "no current context") {
		t.Fatalf("error = %q, want missing current context", err.Error())
	}
}
