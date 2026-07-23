package kubent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const SupportedVersion = "0.7.3"

var (
	ErrMissingBinary       = errors.New("kubent binary is missing")
	ErrUnsupportedVersion  = errors.New("kubent version is unsupported")
	ErrExecutionFailed     = errors.New("kubent execution failed")
	ErrMalformedOutput     = errors.New("kubent output is malformed")
	ErrOutputLimitExceeded = errors.New("kubent output limit exceeded")
)

type Runner interface {
	Run(ctx context.Context, command Command) (Result, error)
}

type Command struct {
	Path       string
	Args       []string
	Timeout    time.Duration
	MaxStdout  int
	MaxStderr  int
	RedactHint string
}

type Result struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
}

type Adapter struct {
	Runner    Runner
	Path      string
	Timeout   time.Duration
	MaxStdout int
	MaxStderr int
}

func (adapter Adapter) ValidateVersion(ctx context.Context) (string, error) {
	result, err := adapter.run(ctx, []string{"--version"})
	if err != nil {
		return "", err
	}
	version := parseVersion(string(result.Stdout))
	if version == "" {
		version = parseVersion(string(result.Stderr))
	}
	if version == "" {
		return "", fmt.Errorf("%w: version not found", ErrMalformedOutput)
	}
	if version != SupportedVersion {
		return version, fmt.Errorf("%w: got %s want %s", ErrUnsupportedVersion, version, SupportedVersion)
	}
	return version, nil
}

func (adapter Adapter) RunJSON(ctx context.Context, targetVersion string, kubeconfig string, kubeContext string) (Report, error) {
	args := []string{"--output", "json", "--helm3=false", "--target-version", targetVersion}
	if kubeconfig != "" {
		args = append(args, "--kubeconfig", kubeconfig)
	}
	if kubeContext != "" {
		args = append(args, "--context", kubeContext)
	}
	result, err := adapter.run(ctx, args)
	if err != nil {
		return Report{}, err
	}
	var report Report
	if err := json.Unmarshal(result.Stdout, &report); err != nil {
		return Report{}, fmt.Errorf("%w: %v", ErrMalformedOutput, err)
	}
	return report, nil
}

func (adapter Adapter) BuildCommand(args []string) Command {
	timeout := adapter.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	maxStdout := adapter.MaxStdout
	if maxStdout == 0 {
		maxStdout = 4 * 1024 * 1024
	}
	maxStderr := adapter.MaxStderr
	if maxStderr == 0 {
		maxStderr = 256 * 1024
	}
	path := adapter.Path
	if path == "" {
		path = "kubent"
	}
	return Command{
		Path:      path,
		Args:      append([]string(nil), args...),
		Timeout:   timeout,
		MaxStdout: maxStdout,
		MaxStderr: maxStderr,
	}
}

func (adapter Adapter) run(ctx context.Context, args []string) (Result, error) {
	if adapter.Runner == nil {
		return Result{}, ErrMissingBinary
	}
	command := adapter.BuildCommand(args)
	result, err := adapter.Runner.Run(ctx, command)
	if err != nil {
		if errors.Is(err, ErrOutputLimitExceeded) {
			return Result{}, err
		}
		return Result{}, fmt.Errorf("%w: %v", ErrExecutionFailed, err)
	}
	if len(result.Stdout) > command.MaxStdout || len(result.Stderr) > command.MaxStderr {
		return Result{}, ErrOutputLimitExceeded
	}
	if result.ExitCode != 0 {
		return Result{}, fmt.Errorf("%w: exit code %d", ErrExecutionFailed, result.ExitCode)
	}
	return result, nil
}

func parseVersion(value string) string {
	fields := strings.Fields(value)
	for _, field := range fields {
		candidate := strings.TrimPrefix(strings.TrimSpace(field), "v")
		if candidate == SupportedVersion {
			return candidate
		}
		if strings.Count(candidate, ".") == 2 {
			return candidate
		}
	}
	return ""
}
