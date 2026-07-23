package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/kube/preflight"
)

func TestRunVersion(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"version"}, &stdout, &stderr, BuildInfo{
		Version:        "test-version",
		Commit:         "test-commit",
		BuildDate:      "2026-07-22T00:00:00Z",
		CatalogVersion: "0.1.0",
	})

	if code != ExitReady {
		t.Fatalf("Run(version) exit code = %d, want %d", code, ExitReady)
	}
	if stderr.Len() != 0 {
		t.Fatalf("Run(version) stderr = %q, want empty", stderr.String())
	}

	output := stdout.String()
	for _, want := range []string{
		"kua version: test-version",
		"commit: test-commit",
		"assessmentSchema: kua.assessment.v1",
		"catalogVersion: 0.1.0",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("Run(version) output missing %q in:\n%s", want, output)
		}
	}
}

func TestRunVersionWithGlobalFlags(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{
		"--log-level=debug",
		"--format", "json",
		"--provider-source", "offline",
		"--context", "ctx-001",
		"version",
	}, &stdout, &stderr, BuildInfo{})

	if code != ExitReady {
		t.Fatalf("Run(version with flags) exit code = %d, want %d", code, ExitReady)
	}
	if stderr.Len() != 0 {
		t.Fatalf("Run(version with flags) stderr = %q, want empty", stderr.String())
	}
	if !strings.Contains(stdout.String(), "kua version: 0.0.0-dev") {
		t.Fatalf("Run(version with flags) output = %q, want version text", stdout.String())
	}
}

func TestRunUnimplementedCommands(t *testing.T) {
	for _, command := range []string{"analyze", "health", "compatibility", "report"} {
		t.Run(command, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			code := Run([]string{command}, &stdout, &stderr, BuildInfo{})

			if code != ExitExecution {
				t.Fatalf("Run(%s) exit code = %d, want %d", command, code, ExitExecution)
			}
			if stdout.Len() != 0 {
				t.Fatalf("Run(%s) stdout = %q, want empty", command, stdout.String())
			}
			if !strings.Contains(stderr.String(), "not implemented yet") {
				t.Fatalf("Run(%s) stderr = %q, want unimplemented message", command, stderr.String())
			}
		})
	}
}

func TestRunInventoryPreflight(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := RunWithDependencies([]string{
		"--context", "ctx-001",
		"--kubeconfig", "/tmp/synthetic-kubeconfig",
		"inventory",
	}, &stdout, &stderr, BuildInfo{}, Dependencies{
		PreflightRunner: fakePreflightRunner{
			result: preflight.Result{
				Context: preflight.ContextSelection{
					Name:             "ctx-001",
					KubeconfigSource: preflight.KubeconfigSourceExplicit,
				},
				ServerVersion:   "v1.30.0",
				DiscoveryStatus: preflight.StatusPass,
				PermissionChecks: []preflight.PermissionCheck{
					{Resource: "pods", Verb: "list", EvidenceClass: preflight.EvidenceRequired, Status: preflight.StatusPass},
				},
			},
		},
	})

	if code != ExitReady {
		t.Fatalf("Run(inventory) exit code = %d, want %d", code, ExitReady)
	}
	if stderr.Len() != 0 {
		t.Fatalf("Run(inventory) stderr = %q, want empty", stderr.String())
	}
	for _, want := range []string{
		"inventory preflight only",
		"context: ctx-001",
		"kubeconfigSource: EXPLICIT",
		"serverVersion: v1.30.0",
		"discovery: PASS",
		"requiredFailure: false",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("Run(inventory) output missing %q in:\n%s", want, stdout.String())
		}
	}
}

func TestRunInventoryPreflightJSON(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := RunWithDependencies([]string{
		"--format=json",
		"inventory",
	}, &stdout, &stderr, BuildInfo{}, Dependencies{
		PreflightRunner: fakePreflightRunner{
			result: preflight.Result{
				Context: preflight.ContextSelection{
					Name:             "ctx-json",
					KubeconfigSource: preflight.KubeconfigSourceDefault,
				},
				ServerVersion:   "v1.31.4",
				DiscoveryStatus: preflight.StatusPass,
				PermissionChecks: []preflight.PermissionCheck{
					{Resource: "pods", Verb: "list", EvidenceClass: preflight.EvidenceRequired, Status: preflight.StatusPass},
					{Resource: "events", Verb: "list", EvidenceClass: preflight.EvidenceOptional, Status: preflight.StatusUnknown, Reason: "not checked"},
				},
				Limitations: []preflight.Limitation{
					{Code: "OPTIONAL_UNKNOWN", Severity: "warning", Summary: "events permission was not checked"},
				},
			},
		},
	})

	if code != ExitReady {
		t.Fatalf("Run(inventory json) exit code = %d, want %d", code, ExitReady)
	}
	if stderr.Len() != 0 {
		t.Fatalf("Run(inventory json) stderr = %q, want empty", stderr.String())
	}
	if strings.Contains(stdout.String(), "inventory preflight only") {
		t.Fatalf("Run(inventory json) emitted console text:\n%s", stdout.String())
	}

	var got inventoryPreflightDocument
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("Run(inventory json) output is not JSON: %v\n%s", err, stdout.String())
	}
	if !got.PreflightOnly {
		t.Fatalf("Run(inventory json) preflightOnly = false, want true")
	}
	if got.Kind != "InventoryPreflight" {
		t.Fatalf("Run(inventory json) kind = %q, want InventoryPreflight", got.Kind)
	}
	if got.Context != "ctx-json" || got.KubeconfigSource != "DEFAULT" || got.ServerVersion != "v1.31.4" {
		t.Fatalf("Run(inventory json) context/source/version = %#v", got)
	}
	if len(got.PermissionChecks) != 2 {
		t.Fatalf("Run(inventory json) permissionChecks = %d, want 2", len(got.PermissionChecks))
	}
	if got.PermissionChecks[1].Status != preflight.StatusUnknown {
		t.Fatalf("Run(inventory json) optional unknown status = %s, want UNKNOWN", got.PermissionChecks[1].Status)
	}
	if len(got.Limitations) != 1 {
		t.Fatalf("Run(inventory json) limitations = %d, want 1", len(got.Limitations))
	}
}

func TestRunInventoryPreflightFailure(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := RunWithDependencies([]string{"inventory"}, &stdout, &stderr, BuildInfo{}, Dependencies{
		PreflightRunner: fakePreflightRunner{err: errors.New("missing context")},
	})

	if code != ExitExecution {
		t.Fatalf("Run(inventory failure) exit code = %d, want %d", code, ExitExecution)
	}
	if stdout.Len() != 0 {
		t.Fatalf("Run(inventory failure) stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), "inventory preflight failed") {
		t.Fatalf("Run(inventory failure) stderr = %q, want preflight failure", stderr.String())
	}
}

func TestRunUsageErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "no args", args: nil},
		{name: "unknown command", args: []string{"unknown"}},
		{name: "unknown flag", args: []string{"--unknown", "version"}},
		{name: "missing flag value", args: []string{"--log-level"}},
		{name: "invalid log level", args: []string{"--log-level", "trace", "version"}},
		{name: "invalid format", args: []string{"--format", "yaml", "version"}},
		{name: "invalid provider source", args: []string{"--provider-source", "internet", "version"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			code := Run(test.args, &stdout, &stderr, BuildInfo{})

			if code != ExitUsage {
				t.Fatalf("Run(%s) exit code = %d, want %d", test.name, code, ExitUsage)
			}
			if !strings.Contains(stderr.String(), "Usage:") {
				t.Fatalf("Run(%s) stderr = %q, want usage text", test.name, stderr.String())
			}
		})
	}
}

type fakePreflightRunner struct {
	result preflight.Result
	err    error
}

func (f fakePreflightRunner) Run(preflight.KubeconfigOptions) (preflight.Result, error) {
	if f.err != nil {
		return preflight.Result{}, f.err
	}
	return f.result, nil
}

func TestParseArgsStoresConfig(t *testing.T) {
	cfg, positional, err := parseArgs([]string{
		"--log-level", "warn",
		"--format=markdown",
		"--provider-source=file",
		"--context", "ctx-001",
		"--kubeconfig", "/tmp/kubeconfig",
		"--config", "/tmp/kua.yaml",
		"--output", "/tmp/report.md",
		"analyze",
	})

	if err != nil {
		t.Fatalf("parseArgs returned error: %v", err)
	}
	if len(positional) != 1 || positional[0] != "analyze" {
		t.Fatalf("parseArgs positional = %#v, want analyze", positional)
	}

	want := Config{
		LogLevel:       "warn",
		Format:         "markdown",
		ProviderSource: "file",
		Context:        "ctx-001",
		Kubeconfig:     "/tmp/kubeconfig",
		ConfigPath:     "/tmp/kua.yaml",
		OutputPath:     "/tmp/report.md",
	}
	if cfg != want {
		t.Fatalf("parseArgs config = %#v, want %#v", cfg, want)
	}
}
