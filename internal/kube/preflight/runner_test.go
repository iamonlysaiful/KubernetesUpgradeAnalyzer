package preflight

import (
	"errors"
	"strings"
	"testing"
)

type fakeResolver struct {
	selection ContextSelection
	err       error
}

func (f fakeResolver) Resolve(KubeconfigOptions) (ContextSelection, error) {
	if f.err != nil {
		return ContextSelection{}, f.err
	}
	return f.selection, nil
}

type fakeChecker struct {
	serverVersion string
	serverErr     error
	discoveryErr  error
	permissions   []PermissionCheck
	permissionErr error
}

func (f fakeChecker) ServerVersion(ContextSelection) (string, error) {
	if f.serverErr != nil {
		return "", f.serverErr
	}
	return f.serverVersion, nil
}

func (f fakeChecker) Discovery(ContextSelection) error {
	return f.discoveryErr
}

func (f fakeChecker) Permissions(ContextSelection) ([]PermissionCheck, error) {
	if f.permissionErr != nil {
		return nil, f.permissionErr
	}
	return f.permissions, nil
}

func TestRunnerRunReturnsPassingResult(t *testing.T) {
	runner := Runner{
		Resolver: fakeResolver{selection: ContextSelection{Name: "ctx-001"}},
		Checker: fakeChecker{
			serverVersion: "1.30.0",
			permissions: []PermissionCheck{
				{Resource: "pods", Verb: "list", EvidenceClass: EvidenceRequired, Status: StatusPass},
				{Resource: "ingresses", Verb: "list", EvidenceClass: EvidenceOptional, Status: StatusPass},
			},
		},
	}

	result, err := runner.Run(KubeconfigOptions{})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.Context.Name != "ctx-001" {
		t.Fatalf("Context.Name = %q, want ctx-001", result.Context.Name)
	}
	if result.ServerVersion != "1.30.0" {
		t.Fatalf("ServerVersion = %q, want 1.30.0", result.ServerVersion)
	}
	if result.DiscoveryStatus != StatusPass {
		t.Fatalf("DiscoveryStatus = %q, want %q", result.DiscoveryStatus, StatusPass)
	}
	if result.HasRequiredFailure() {
		t.Fatal("HasRequiredFailure() = true, want false")
	}
	if len(result.Limitations) != 0 {
		t.Fatalf("Limitations = %#v, want none", result.Limitations)
	}
}

func TestRunnerRunRecordsOptionalPermissionLimitation(t *testing.T) {
	runner := Runner{
		Resolver: fakeResolver{selection: ContextSelection{Name: "ctx-001"}},
		Checker: fakeChecker{
			serverVersion: "1.30.0",
			permissions: []PermissionCheck{
				{Resource: "pods", Verb: "list", EvidenceClass: EvidenceRequired, Status: StatusPass},
				{Resource: "events", Verb: "list", EvidenceClass: EvidenceOptional, Status: StatusFail},
			},
		},
	}

	result, err := runner.Run(KubeconfigOptions{})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.HasRequiredFailure() {
		t.Fatal("HasRequiredFailure() = true, want false for optional denial")
	}
	if len(result.Limitations) != 1 {
		t.Fatalf("len(Limitations) = %d, want 1", len(result.Limitations))
	}
	if result.Limitations[0].Severity != "WARN" {
		t.Fatalf("Limitation severity = %q, want WARN", result.Limitations[0].Severity)
	}
}

func TestRunnerRunRecordsRequiredPermissionFailure(t *testing.T) {
	runner := Runner{
		Resolver: fakeResolver{selection: ContextSelection{Name: "ctx-001"}},
		Checker: fakeChecker{
			serverVersion: "1.30.0",
			permissions: []PermissionCheck{
				{Resource: "pods", Verb: "list", EvidenceClass: EvidenceRequired, Status: StatusFail},
			},
		},
	}

	result, err := runner.Run(KubeconfigOptions{})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if !result.HasRequiredFailure() {
		t.Fatal("HasRequiredFailure() = false, want true for required denial")
	}
	if len(result.Limitations) != 1 {
		t.Fatalf("len(Limitations) = %d, want 1", len(result.Limitations))
	}
	if result.Limitations[0].Severity != "ERROR" {
		t.Fatalf("Limitation severity = %q, want ERROR", result.Limitations[0].Severity)
	}
}

func TestRunnerRunReturnsContextResolutionError(t *testing.T) {
	runner := Runner{
		Resolver: fakeResolver{err: errors.New("missing context")},
		Checker:  fakeChecker{},
	}

	_, err := runner.Run(KubeconfigOptions{})
	if err == nil {
		t.Fatal("Run returned nil error, want context resolution error")
	}
	if !strings.Contains(err.Error(), "resolve kubeconfig context") {
		t.Fatalf("error = %q, want context resolution context", err.Error())
	}
}

func TestRunnerRunRecordsDiscoveryFailure(t *testing.T) {
	runner := Runner{
		Resolver: fakeResolver{selection: ContextSelection{Name: "ctx-001"}},
		Checker: fakeChecker{
			serverVersion: "1.30.0",
			discoveryErr:  errors.New("discovery failed"),
		},
	}

	result, err := runner.Run(KubeconfigOptions{})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.DiscoveryStatus != StatusFail {
		t.Fatalf("DiscoveryStatus = %q, want %q", result.DiscoveryStatus, StatusFail)
	}
	if !result.HasRequiredFailure() {
		t.Fatal("HasRequiredFailure() = false, want true for discovery failure")
	}
}
