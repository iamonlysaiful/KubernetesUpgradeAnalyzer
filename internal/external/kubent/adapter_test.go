package kubent

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestValidateVersionAcceptsSupportedVersion(t *testing.T) {
	runner := &fakeRunner{result: Result{Stdout: []byte("kubent version 0.7.3\n")}}
	adapter := Adapter{Runner: runner}

	version, err := adapter.ValidateVersion(context.Background())
	if err != nil {
		t.Fatalf("ValidateVersion returned error: %v", err)
	}
	if version != SupportedVersion {
		t.Fatalf("version = %q, want %q", version, SupportedVersion)
	}
	if got := runner.commands[0].Args; !reflect.DeepEqual(got, []string{"--version"}) {
		t.Fatalf("args = %#v, want --version", got)
	}
}

func TestValidateVersionRejectsUnsupportedVersion(t *testing.T) {
	adapter := Adapter{Runner: &fakeRunner{result: Result{Stdout: []byte("kubent version 0.7.2\n")}}}

	version, err := adapter.ValidateVersion(context.Background())
	if !errors.Is(err, ErrUnsupportedVersion) {
		t.Fatalf("error = %v, want ErrUnsupportedVersion", err)
	}
	if version != "0.7.2" {
		t.Fatalf("version = %q, want 0.7.2", version)
	}
}

func TestRunJSONBuildsControlledArguments(t *testing.T) {
	runner := &fakeRunner{result: Result{Stdout: []byte(`{"DeprecatedAPIs":[]}`)}}
	adapter := Adapter{Runner: runner, Path: "/usr/local/bin/kubent", Timeout: 5 * time.Second, MaxStdout: 1024, MaxStderr: 256}

	_, err := adapter.RunJSON(context.Background(), "1.33.0", "/tmp/kubeconfig", "ctx-a")
	if err != nil {
		t.Fatalf("RunJSON returned error: %v", err)
	}

	command := runner.commands[0]
	if command.Path != "/usr/local/bin/kubent" {
		t.Fatalf("Path = %q, want explicit path", command.Path)
	}
	wantArgs := []string{"--output", "json", "--helm3=false", "--target-version", "1.33.0", "--kubeconfig", "/tmp/kubeconfig", "--context", "ctx-a"}
	if !reflect.DeepEqual(command.Args, wantArgs) {
		t.Fatalf("Args = %#v, want %#v", command.Args, wantArgs)
	}
	if command.Timeout != 5*time.Second || command.MaxStdout != 1024 || command.MaxStderr != 256 {
		t.Fatalf("bounds = %#v", command)
	}
}

func TestRunJSONParsesDeprecatedAPIs(t *testing.T) {
	runner := &fakeRunner{result: Result{Stdout: []byte(`{"DeprecatedAPIs":[{"Name":"api","Namespace":"default","Kind":"Ingress","APIVersion":"extensions/v1beta1","ReplaceWith":"networking.k8s.io/v1","Since":"1.22","Deleted":true}]}`)}}
	adapter := Adapter{Runner: runner}

	report, err := adapter.RunJSON(context.Background(), "1.33.0", "", "")
	if err != nil {
		t.Fatalf("RunJSON returned error: %v", err)
	}
	if len(report.DeprecatedAPIs) != 1 {
		t.Fatalf("DeprecatedAPIs = %d, want 1", len(report.DeprecatedAPIs))
	}
	api := report.DeprecatedAPIs[0]
	if api.Kind != "Ingress" || api.APIVersion != "extensions/v1beta1" || !api.Deleted {
		t.Fatalf("api = %#v, want parsed ingress removal", api)
	}
}

func TestRunJSONRejectsMalformedOutput(t *testing.T) {
	adapter := Adapter{Runner: &fakeRunner{result: Result{Stdout: []byte(`not json`)}}}

	_, err := adapter.RunJSON(context.Background(), "1.33.0", "", "")
	if !errors.Is(err, ErrMalformedOutput) {
		t.Fatalf("error = %v, want ErrMalformedOutput", err)
	}
}

func TestRunJSONRejectsExecutionFailure(t *testing.T) {
	adapter := Adapter{Runner: &fakeRunner{result: Result{ExitCode: 2, Stderr: []byte("boom")}}}

	_, err := adapter.RunJSON(context.Background(), "1.33.0", "", "")
	if !errors.Is(err, ErrExecutionFailed) {
		t.Fatalf("error = %v, want ErrExecutionFailed", err)
	}
}

func TestRunJSONRejectsOversizedOutput(t *testing.T) {
	adapter := Adapter{Runner: &fakeRunner{result: Result{Stdout: []byte(`{"DeprecatedAPIs":[]}`)}}, MaxStdout: 5}

	_, err := adapter.RunJSON(context.Background(), "1.33.0", "", "")
	if !errors.Is(err, ErrOutputLimitExceeded) {
		t.Fatalf("error = %v, want ErrOutputLimitExceeded", err)
	}
}

func TestAdapterWithoutRunnerIsMissingBinary(t *testing.T) {
	adapter := Adapter{}

	_, err := adapter.ValidateVersion(context.Background())
	if !errors.Is(err, ErrMissingBinary) {
		t.Fatalf("error = %v, want ErrMissingBinary", err)
	}
}

type fakeRunner struct {
	result   Result
	err      error
	commands []Command
}

func (runner *fakeRunner) Run(ctx context.Context, command Command) (Result, error) {
	runner.commands = append(runner.commands, command)
	return runner.result, runner.err
}
