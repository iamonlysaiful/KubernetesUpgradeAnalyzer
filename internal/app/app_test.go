package app

import (
	"bytes"
	"strings"
	"testing"
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

func TestRunUnimplementedCommands(t *testing.T) {
	for _, command := range []string{"analyze", "inventory", "health", "compatibility", "report"} {
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

func TestRunUsageErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "no args", args: nil},
		{name: "unknown command", args: []string{"unknown"}},
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
