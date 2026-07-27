package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/external/kubent"
	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/kube/inventory"
	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/kube/preflight"
	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/provider"
	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/recommendation"
	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/report"
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

func TestRunAnalyzeJSONProducesInconclusiveAssessment(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	now := time.Date(2026, 7, 27, 4, 0, 0, 0, time.UTC)

	code := RunWithDependencies([]string{
		"--format=json",
		"--provider-source=none",
		"analyze",
	}, &stdout, &stderr, BuildInfo{}, Dependencies{
		PreflightRunner: fakePreflightRunner{result: preflight.Result{
			Context:         preflight.ContextSelection{Name: "ctx-analyze", KubeconfigSource: preflight.KubeconfigSourceDefault},
			ServerVersion:   "v1.30.0",
			DiscoveryStatus: preflight.StatusPass,
		}},
		InventoryCollector: fakeInventoryCollector{snapshot: validCoreSnapshot("ctx-analyze", "1.30.0")},
		ProviderFactory:    fakeProviderFactory{provider: fakeProvider{}},
		APIAnalyzer:        fakeAPIAnalyzer{limitation: recommendation.Limitation{Code: "API_TARGET_UNAVAILABLE", Summary: "target missing"}},
		Clock:              func() time.Time { return now },
	})

	if code != ExitInconclusive {
		t.Fatalf("Run(analyze json) exit code = %d, want %d", code, ExitInconclusive)
	}
	if stderr.Len() != 0 {
		t.Fatalf("Run(analyze json) stderr = %q, want empty", stderr.String())
	}

	var got report.Document
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("Run(analyze json) output is not report JSON: %v\n%s", err, stdout.String())
	}
	if got.SchemaVersion != "kua.assessment.v1" {
		t.Fatalf("schemaVersion = %q, want kua.assessment.v1", got.SchemaVersion)
	}
	if got.Current != "1.30.0" {
		t.Fatalf("current = %q, want 1.30.0", got.Current)
	}
	if got.Readiness != "INCONCLUSIVE" || got.Risk != "UNKNOWN" {
		t.Fatalf("readiness/risk = %s/%s, want INCONCLUSIVE/UNKNOWN", got.Readiness, got.Risk)
	}
	if !hasLimitation(got.Limitations, "API_TARGET_UNAVAILABLE") {
		t.Fatalf("limitations = %#v, want API_TARGET_UNAVAILABLE", got.Limitations)
	}
	if strings.Contains(stdout.String(), "ctx-analyze") {
		t.Fatalf("analyze JSON leaked context name:\n%s", stdout.String())
	}
}

func TestRunAnalyzeRedactsResourceNames(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	snapshot := validCoreSnapshot("ctx-redact", "1.30.0")
	snapshot.Inventory.Nodes[0].Ref.Name = "node-private"
	snapshot.Inventory.Nodes[0].Conditions = []inventory.Condition{}

	code := RunWithDependencies([]string{
		"--format=json",
		"--redacted",
		"--provider-source=none",
		"analyze",
	}, &stdout, &stderr, BuildInfo{}, Dependencies{
		PreflightRunner:    fakePreflightRunner{result: validPreflight("ctx-redact", "v1.30.0")},
		InventoryCollector: fakeInventoryCollector{snapshot: snapshot},
		ProviderFactory:    fakeProviderFactory{provider: fakeProvider{}},
		APIAnalyzer:        fakeAPIAnalyzer{limitation: recommendation.Limitation{Code: "API_TARGET_UNAVAILABLE", Summary: "target missing"}},
		Clock:              func() time.Time { return time.Date(2026, 7, 27, 4, 1, 0, 0, time.UTC) },
	})

	if code != ExitInconclusive {
		t.Fatalf("Run(analyze redacted) exit code = %d, want %d", code, ExitInconclusive)
	}
	if strings.Contains(stdout.String(), "node-private") {
		t.Fatalf("redacted output leaked node name:\n%s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "\"redacted\": true") {
		t.Fatalf("redacted output missing redacted marker:\n%s", stdout.String())
	}
}

func TestRunReportRendersInputDocument(t *testing.T) {
	tmp := t.TempDir() + "/assessment.json"
	input := `{"schemaVersion":"kua.assessment.v1","assessmentId":"assessment-test","generatedAt":"2026-07-27T04:02:00Z","redacted":false,"currentVersion":"1.30.0","readiness":"INCONCLUSIVE","risk":"UNKNOWN","findings":[],"limitations":[]}`
	if err := os.WriteFile(tmp, []byte(input), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"--format=console", "--input", tmp, "report"}, &stdout, &stderr, BuildInfo{})

	if code != ExitInconclusive {
		t.Fatalf("Run(report) exit code = %d, want %d", code, ExitInconclusive)
	}
	if stderr.Len() != 0 {
		t.Fatalf("Run(report) stderr = %q, want empty", stderr.String())
	}
	if !strings.Contains(stdout.String(), "Assessment: assessment-test") {
		t.Fatalf("Run(report) output = %q, want rendered assessment", stdout.String())
	}
}

func TestProcessRunnerCapturesSuccessfulStderr(t *testing.T) {
	result, err := processRunner{}.Run(context.Background(), kubent.Command{
		Path:      "sh",
		Args:      []string{"-c", "printf 'version 0.7.3' >&2"},
		Timeout:   time.Second,
		MaxStdout: 1024,
		MaxStderr: 1024,
	})
	if err != nil {
		t.Fatalf("processRunner.Run returned error: %v", err)
	}
	if string(result.Stderr) != "version 0.7.3" {
		t.Fatalf("stderr = %q, want version text", string(result.Stderr))
	}
}

func TestProcessRunnerCapturesFailedStderr(t *testing.T) {
	result, err := processRunner{}.Run(context.Background(), kubent.Command{
		Path:      "sh",
		Args:      []string{"-c", "printf boom >&2; exit 2"},
		Timeout:   time.Second,
		MaxStdout: 1024,
		MaxStderr: 1024,
	})
	if err != nil {
		t.Fatalf("processRunner.Run returned error: %v", err)
	}
	if result.ExitCode != 2 || string(result.Stderr) != "boom" {
		t.Fatalf("result = %#v, want exit 2 and stderr boom", result)
	}
}

func TestRunAnalyzeAggregatesAPIFindings(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := RunWithDependencies([]string{
		"--format=json",
		"--provider-source=none",
		"--target-version", "1.33.0",
		"analyze",
	}, &stdout, &stderr, BuildInfo{}, Dependencies{
		PreflightRunner:    fakePreflightRunner{result: validPreflight("ctx-api", "v1.30.0")},
		InventoryCollector: fakeInventoryCollector{snapshot: validCoreSnapshot("ctx-api", "1.30.0")},
		ProviderFactory:    fakeProviderFactory{provider: fakeProvider{}},
		APIAnalyzer: fakeAPIAnalyzer{findings: []kubent.Finding{{
			AnalyzerVersion: "0.7.3",
			TargetVersion:   "1.33.0",
			Status:          kubent.FindingFail,
			Resource:        kubent.ResourceRef{Kind: "Ingress", Namespace: "default", Name: "legacy"},
			APIVersion:      "extensions/v1beta1",
			Kind:            "Ingress",
			RemovedIn:       "1.22",
			Replacement:     "networking.k8s.io/v1",
		}}},
		Clock: func() time.Time { return time.Date(2026, 7, 27, 4, 3, 0, 0, time.UTC) },
	})

	if code != ExitNotReady {
		t.Fatalf("Run(analyze API) exit code = %d, want %d", code, ExitNotReady)
	}
	var got report.Document
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("Run(analyze API) output is not JSON: %v\n%s", err, stdout.String())
	}
	if len(got.Findings) == 0 || got.Findings[0].Category != recommendation.CategoryAPI {
		t.Fatalf("findings = %#v, want API finding", got.Findings)
	}
}

func TestRunAnalyzeCanReturnReadyWhenEvidencePasses(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := RunWithDependencies([]string{
		"--format=json",
		"--provider-source=file",
		"--target-version", "1.33.12",
		"analyze",
	}, &stdout, &stderr, BuildInfo{}, Dependencies{
		PreflightRunner:    fakePreflightRunner{result: validPreflight("ctx-ready", "v1.30.0")},
		InventoryCollector: fakeInventoryCollector{snapshot: validCoreSnapshot("ctx-ready", "1.30.0")},
		ProviderFactory: fakeProviderFactory{provider: fakeProvider{evidence: &provider.ProviderEvidence{
			SchemaVersion:  "kua.provider-evidence.aks.v1",
			EvidenceID:     "synthetic",
			CurrentVersion: "1.30.0",
			Cluster: provider.ClusterIdentity{
				Provider:           provider.ProviderAKS,
				IdentityConfidence: provider.ConfidenceHigh,
			},
			AvailableUpgrades: []provider.UpgradeOption{
				{Version: "1.31.9"},
				{Version: "1.32.7"},
				{Version: "1.33.12"},
			},
			Limitations: []provider.Limitation{},
		}}},
		APIAnalyzer: fakeAPIAnalyzer{findings: []kubent.Finding{{
			AnalyzerVersion: "0.7.3",
			TargetVersion:   "1.33.12",
			Status:          kubent.FindingPass,
		}}},
		Clock: func() time.Time { return time.Date(2026, 7, 27, 4, 4, 0, 0, time.UTC) },
	})

	if code != ExitReady {
		t.Fatalf("Run(analyze ready) exit code = %d, want %d; stderr=%q stdout=%s", code, ExitReady, stderr.String(), stdout.String())
	}
	var got report.Document
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("Run(analyze ready) output is not JSON: %v\n%s", err, stdout.String())
	}
	if got.Readiness != recommendation.ReadinessReady || got.Risk != recommendation.RiskLow {
		t.Fatalf("readiness/risk = %s/%s, want READY/LOW", got.Readiness, got.Risk)
	}
	if got.Destination != "1.33.12" || len(got.Path) != 3 {
		t.Fatalf("destination/path = %q/%#v, want 1.33.12 with 3 stages", got.Destination, got.Path)
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
		InventoryCollector: fakeInventoryCollector{
			snapshot: inventory.Snapshot{
				SchemaVersion: inventory.SchemaVersion,
				SnapshotID:    "ctx-golden-20260723T010203Z",
				CapturedAt:    "2026-07-23T01:02:03Z",
				Cluster: inventory.Cluster{
					Identity: inventory.ResourceRef{Kind: "Cluster", Name: "ctx-golden"},
					Provider: inventory.Provider{Type: "UNKNOWN", Confidence: "UNKNOWN"},
					Context:  inventory.Context{Name: "ctx-golden", KubeconfigSource: "DEFAULT"},
				},
				Kubernetes: inventory.Kubernetes{ServerVersion: "1.30.7"},
				Inventory: inventory.Inventory{
					Namespaces: []inventory.ResourceRef{
						{APIVersion: "v1", Kind: "Namespace", Name: "alpha"},
						{APIVersion: "v1", Kind: "Namespace", Name: "zeta"},
					},
					Nodes: []inventory.Node{
						{
							Ref:               inventory.ResourceRef{APIVersion: "v1", Kind: "Node", Name: "node-a"},
							KubeletVersion:    "1.30.6",
							ProviderIDPresent: false,
							Conditions:        []inventory.Condition{},
						},
						{
							Ref:               inventory.ResourceRef{APIVersion: "v1", Kind: "Node", Name: "node-b"},
							KubeletVersion:    "1.30.7",
							ProviderIDPresent: true,
							NodePool:          "pool-b",
							Conditions: []inventory.Condition{
								{Type: "MemoryPressure", Status: "FALSE", Reason: "SufficientMemory"},
								{Type: "Ready", Status: "TRUE", Reason: "KubeletReady"},
							},
						},
					},
					Workloads:  []inventory.Workload{},
					Storage:    []inventory.ResourceRef{},
					Networking: []inventory.ResourceRef{},
					CRDs:       []inventory.ResourceRef{},
					Events:     []inventory.Event{},
				},
				Limitations: []inventory.Limitation{{
					Code:     "PARTIAL_INVENTORY_P2_02",
					Severity: "WARN",
					Summary:  "P2-02 collects namespaces and nodes only; workloads, storage, networking, CRDs, and events are intentionally not collected yet.",
				}},
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
	want, err := os.ReadFile("../../schemas/fixtures/cluster-snapshot/valid/p2-02-core-inventory.json")
	if err != nil {
		t.Fatalf("ReadFile golden fixture returned error: %v", err)
	}
	if stdout.String() != string(want) {
		t.Fatalf("Run(inventory json) output does not match golden fixture.\nGot:\n%s\nWant:\n%s", stdout.String(), string(want))
	}

	var got inventory.Snapshot
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("Run(inventory json) output is not JSON: %v\n%s", err, stdout.String())
	}
	if got.SchemaVersion != inventory.SchemaVersion {
		t.Fatalf("Run(inventory json) schemaVersion = %q, want %q", got.SchemaVersion, inventory.SchemaVersion)
	}
	if got.Cluster.Context.Name != "ctx-golden" {
		t.Fatalf("Run(inventory json) context = %q, want ctx-golden", got.Cluster.Context.Name)
	}
	if len(got.Inventory.Namespaces) != 2 || got.Inventory.Namespaces[0].Name != "alpha" {
		t.Fatalf("Run(inventory json) namespaces = %#v, want alpha,zeta", got.Inventory.Namespaces)
	}
	if len(got.Inventory.Nodes) != 2 || got.Inventory.Nodes[1].ProviderIDPresent != true {
		t.Fatalf("Run(inventory json) nodes = %#v, want two nodes with node-b provider ID present", got.Inventory.Nodes)
	}
	if len(got.Limitations) != 1 || got.Limitations[0].Code != "PARTIAL_INVENTORY_P2_02" {
		t.Fatalf("Run(inventory json) limitations = %d, want 1", len(got.Limitations))
	}
}

func TestRunInventoryJSONPreflightRequiredFailureDoesNotCollect(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := RunWithDependencies([]string{"--format=json", "inventory"}, &stdout, &stderr, BuildInfo{}, Dependencies{
		PreflightRunner: fakePreflightRunner{
			result: preflight.Result{
				Context:         preflight.ContextSelection{Name: "ctx-fail", KubeconfigSource: preflight.KubeconfigSourceDefault},
				ServerVersion:   "v1.31.4",
				DiscoveryStatus: preflight.StatusFail,
			},
		},
		InventoryCollector: fakeInventoryCollector{
			err: errors.New("collector should not run"),
		},
	})

	if code != ExitInconclusive {
		t.Fatalf("Run(inventory json preflight failure) exit code = %d, want %d", code, ExitInconclusive)
	}
	if stderr.Len() != 0 {
		t.Fatalf("Run(inventory json preflight failure) stderr = %q, want empty", stderr.String())
	}
	var got inventoryPreflightDocument
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("Run(inventory json preflight failure) output is not preflight JSON: %v\n%s", err, stdout.String())
	}
	if got.Kind != "InventoryPreflight" || !got.RequiredFailure {
		t.Fatalf("Run(inventory json preflight failure) document = %#v, want failed preflight document", got)
	}
}

func TestRunInventoryJSONCollectionFailure(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := RunWithDependencies([]string{"--format=json", "inventory"}, &stdout, &stderr, BuildInfo{}, Dependencies{
		PreflightRunner: fakePreflightRunner{
			result: preflight.Result{
				Context:         preflight.ContextSelection{Name: "ctx-collect", KubeconfigSource: preflight.KubeconfigSourceDefault},
				ServerVersion:   "v1.31.4",
				DiscoveryStatus: preflight.StatusPass,
			},
		},
		InventoryCollector: fakeInventoryCollector{
			err: errors.New("list nodes failed"),
		},
	})

	if code != ExitExecution {
		t.Fatalf("Run(inventory json collection failure) exit code = %d, want %d", code, ExitExecution)
	}
	if stdout.Len() != 0 {
		t.Fatalf("Run(inventory json collection failure) stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), "inventory collection failed") {
		t.Fatalf("Run(inventory json collection failure) stderr = %q, want collection failure", stderr.String())
	}
}

func TestRunInventoryJSONValidationFailure(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := RunWithDependencies([]string{"--format=json", "inventory"}, &stdout, &stderr, BuildInfo{}, Dependencies{
		PreflightRunner: fakePreflightRunner{
			result: preflight.Result{
				Context:         preflight.ContextSelection{Name: "ctx-invalid", KubeconfigSource: preflight.KubeconfigSourceDefault},
				ServerVersion:   "v1.31.4",
				DiscoveryStatus: preflight.StatusPass,
			},
		},
		InventoryCollector: fakeInventoryCollector{
			snapshot: inventory.Snapshot{
				SchemaVersion: "kua.cluster-snapshot.v2",
				SnapshotID:    "bad",
				CapturedAt:    "not-a-time",
			},
		},
	})

	if code != ExitExecution {
		t.Fatalf("Run(inventory json validation failure) exit code = %d, want %d", code, ExitExecution)
	}
	if stdout.Len() != 0 {
		t.Fatalf("Run(inventory json validation failure) stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), "inventory snapshot validation failed") {
		t.Fatalf("Run(inventory json validation failure) stderr = %q, want validation failure", stderr.String())
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

type fakeInventoryCollector struct {
	snapshot inventory.Snapshot
	err      error
}

type fakeProviderFactory struct {
	provider provider.Provider
}

func (f fakeProviderFactory) NewProvider(inventory.Snapshot, Config) provider.Provider {
	return f.provider
}

type fakeProvider struct {
	evidence *provider.ProviderEvidence
	err      error
}

type fakeAPIAnalyzer struct {
	findings   []kubent.Finding
	limitation recommendation.Limitation
}

func (f fakeAPIAnalyzer) Analyze(context.Context, Config, string) ([]kubent.Finding, recommendation.Limitation) {
	return f.findings, f.limitation
}

func (f fakeProvider) Identity() (provider.ProviderType, provider.Confidence) {
	return provider.ProviderAKS, provider.ConfidenceHigh
}

func (f fakeProvider) Evidence(context.Context, provider.EvidenceOptions) (*provider.ProviderEvidence, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.evidence, nil
}

func validPreflight(contextName string, serverVersion string) preflight.Result {
	return preflight.Result{
		Context:         preflight.ContextSelection{Name: contextName, KubeconfigSource: preflight.KubeconfigSourceDefault},
		ServerVersion:   serverVersion,
		DiscoveryStatus: preflight.StatusPass,
	}
}

func validCoreSnapshot(contextName string, serverVersion string) inventory.Snapshot {
	return inventory.Snapshot{
		SchemaVersion: inventory.SchemaVersion,
		SnapshotID:    "synthetic-20260727T040000Z",
		CapturedAt:    "2026-07-27T04:00:00Z",
		Cluster: inventory.Cluster{
			Identity: inventory.ResourceRef{Kind: "Cluster", Name: contextName},
			Provider: inventory.Provider{Type: "UNKNOWN", Confidence: "UNKNOWN"},
			Context:  inventory.Context{Name: contextName, KubeconfigSource: "DEFAULT"},
		},
		Kubernetes: inventory.Kubernetes{ServerVersion: serverVersion},
		Inventory: inventory.Inventory{
			Namespaces: []inventory.ResourceRef{{APIVersion: "v1", Kind: "Namespace", Name: "default"}},
			Nodes: []inventory.Node{{
				Ref:               inventory.ResourceRef{APIVersion: "v1", Kind: "Node", Name: "node-a"},
				KubeletVersion:    serverVersion,
				ProviderIDPresent: false,
				Conditions:        []inventory.Condition{{Type: "Ready", Status: "TRUE", Reason: "KubeletReady"}},
			}},
			Workloads:  []inventory.Workload{},
			Storage:    []inventory.ResourceRef{},
			Networking: []inventory.ResourceRef{},
			CRDs:       []inventory.ResourceRef{},
			Events:     []inventory.Event{},
		},
		Limitations: []inventory.Limitation{{Code: "PARTIAL_INVENTORY_P2_02", Severity: "WARN", Summary: "partial inventory"}},
	}
}

func hasLimitation(limitations []recommendation.Limitation, code string) bool {
	for _, limitation := range limitations {
		if limitation.Code == code {
			return true
		}
	}
	return false
}

func (f fakeInventoryCollector) CollectCore(preflight.KubeconfigOptions, preflight.Result) (inventory.Snapshot, error) {
	if f.err != nil {
		return inventory.Snapshot{}, f.err
	}
	return f.snapshot, nil
}

func (f fakeInventoryCollector) CollectAssessment(preflight.KubeconfigOptions, preflight.Result) (inventory.Snapshot, error) {
	if f.err != nil {
		return inventory.Snapshot{}, f.err
	}
	return f.snapshot, nil
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
