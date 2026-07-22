package main

import (
	"os"

	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/app"
)

var (
	version        = "0.0.0-dev"
	commit         = "unknown"
	buildDate      = "unknown"
	catalogVersion = "unavailable"
)

func main() {
	os.Exit(app.Run(os.Args[1:], os.Stdout, os.Stderr, app.BuildInfo{
		Version:        version,
		Commit:         commit,
		BuildDate:      buildDate,
		CatalogVersion: catalogVersion,
	}))
}
