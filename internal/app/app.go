package app

import (
	"fmt"
	"io"
	"runtime"
)

const (
	ExitReady        = 0
	ExitNotReady     = 2
	ExitInconclusive = 3
	ExitUsage        = 4
	ExitExecution    = 5
)

const (
	schemaVersion = "kua.assessment.v1"
	binaryName    = "kua"
)

type BuildInfo struct {
	Version        string
	Commit         string
	BuildDate      string
	CatalogVersion string
}

func Run(args []string, stdout io.Writer, stderr io.Writer, build BuildInfo) int {
	if len(args) == 0 {
		printUsage(stderr)
		return ExitUsage
	}

	switch args[0] {
	case "version":
		printVersion(stdout, build)
		return ExitReady
	case "analyze", "inventory", "health", "compatibility", "report":
		fmt.Fprintf(stderr, "%s %s is not implemented yet\n", binaryName, args[0])
		return ExitExecution
	case "help", "-h", "--help":
		printUsage(stdout)
		return ExitReady
	default:
		fmt.Fprintf(stderr, "unknown command %q\n", args[0])
		printUsage(stderr)
		return ExitUsage
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintf(w, `%s analyzes Kubernetes upgrade readiness.

Usage:
  %s <command>

Commands:
  analyze        Run full upgrade readiness assessment
  inventory      Collect and summarize inventory
  health         Run health checks
  compatibility  Run API and component compatibility checks
  report         Render a saved assessment
  version        Print build and contract versions
`, binaryName, binaryName)
}

func printVersion(w io.Writer, build BuildInfo) {
	fmt.Fprintf(w, "%s version: %s\n", binaryName, defaultString(build.Version, "0.0.0-dev"))
	fmt.Fprintf(w, "commit: %s\n", defaultString(build.Commit, "unknown"))
	fmt.Fprintf(w, "buildDate: %s\n", defaultString(build.BuildDate, "unknown"))
	fmt.Fprintf(w, "go: %s\n", runtime.Version())
	fmt.Fprintf(w, "assessmentSchema: %s\n", schemaVersion)
	fmt.Fprintf(w, "catalogVersion: %s\n", defaultString(build.CatalogVersion, "unavailable"))
}

func defaultString(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
