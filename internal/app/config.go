package app

import "fmt"

type Config struct {
	LogLevel           string
	Format             string
	ProviderSource     string
	Context            string
	Kubeconfig         string
	ConfigPath         string
	OutputPath         string
	InputPath          string
	ProviderEvidence   string
	Subscription       string
	ResourceGroup      string
	ClusterName        string
	TargetVersion      string
	ComponentOverrides string
	Redacted           bool
	Yes                bool              // skip interactive confirmation
	inlineOverrides    map[string]string // component ID -> version, set by interactive prompts
}

func DefaultConfig() Config {
	return Config{
		LogLevel:       "info",
		Format:         "console",
		ProviderSource: "auto",
	}
}

func parseArgs(args []string) (Config, []string, *AppError) {
	cfg := DefaultConfig()
	var positional []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "" {
			continue
		}
		if arg == "--" {
			positional = append(positional, args[i+1:]...)
			break
		}
		if len(arg) < 2 || arg[:2] != "--" {
			positional = append(positional, arg)
			continue
		}

		name, value, hasInlineValue := splitFlag(arg)
		if !isKnownFlag(name) {
			return cfg, nil, UsageError(fmt.Sprintf("unknown flag %q", name))
		}
		if isBoolFlag(name) && !hasInlineValue {
			value = "true"
		} else if !hasInlineValue {
			if i+1 >= len(args) || args[i+1] == "" || isFlag(args[i+1]) {
				return cfg, nil, UsageError(fmt.Sprintf("missing value for %s", name))
			}
			i++
			value = args[i]
		}
		if err := applyFlag(&cfg, name, value); err != nil {
			return cfg, nil, err
		}
	}

	return cfg, positional, nil
}

func splitFlag(arg string) (name string, value string, hasInlineValue bool) {
	for i := 2; i < len(arg); i++ {
		if arg[i] == '=' {
			return arg[:i], arg[i+1:], true
		}
	}
	return arg, "", false
}

func isFlag(value string) bool {
	return len(value) >= 2 && value[:2] == "--"
}

func isKnownFlag(name string) bool {
	switch name {
	case "--log-level", "--format", "--provider-source", "--context", "--kubeconfig", "--config", "--output", "--input", "--provider-evidence", "--subscription", "--resource-group", "--cluster-name", "--target-version", "--component-overrides", "--redacted", "--yes":
		return true
	default:
		return false
	}
}

func isBoolFlag(name string) bool {
	return name == "--redacted" || name == "--yes"
}

func applyFlag(cfg *Config, name string, value string) *AppError {
	if value == "" {
		return UsageError(fmt.Sprintf("missing value for %s", name))
	}

	switch name {
	case "--log-level":
		if !oneOf(value, "debug", "info", "warn", "error") {
			return UsageError("invalid --log-level; expected debug, info, warn, or error")
		}
		cfg.LogLevel = value
	case "--format":
		if !oneOf(value, "console", "json", "markdown", "html") {
			return UsageError("invalid --format; expected console, json, markdown, or html")
		}
		cfg.Format = value
	case "--provider-source":
		if !oneOf(value, "auto", "azure", "file", "offline", "none") {
			return UsageError("invalid --provider-source; expected auto, azure, file, offline, or none")
		}
		cfg.ProviderSource = value
	case "--context":
		cfg.Context = value
	case "--kubeconfig":
		cfg.Kubeconfig = value
	case "--config":
		cfg.ConfigPath = value
	case "--output":
		cfg.OutputPath = value
	case "--input":
		cfg.InputPath = value
	case "--provider-evidence":
		cfg.ProviderEvidence = value
	case "--subscription":
		cfg.Subscription = value
	case "--resource-group":
		cfg.ResourceGroup = value
	case "--cluster-name":
		cfg.ClusterName = value
	case "--target-version":
		cfg.TargetVersion = value
	case "--component-overrides":
		cfg.ComponentOverrides = value
	case "--redacted":
		if !oneOf(value, "true", "false") {
			return UsageError("invalid --redacted; expected true or false")
		}
		cfg.Redacted = value == "true"
	case "--yes":
		if !oneOf(value, "true", "false") {
			return UsageError("invalid --yes; expected true or false")
		}
		cfg.Yes = value == "true"
	}

	return nil
}

func oneOf(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}
