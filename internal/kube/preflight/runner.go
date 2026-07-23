package preflight

import "fmt"

type ContextResolver interface {
	Resolve(KubeconfigOptions) (ContextSelection, error)
}

type ClusterChecker interface {
	ServerVersion(ContextSelection) (string, error)
	Discovery(ContextSelection) error
	Permissions(ContextSelection) ([]PermissionCheck, error)
}

type Runner struct {
	Resolver ContextResolver
	Checker  ClusterChecker
}

type defaultResolver struct{}

func (defaultResolver) Resolve(options KubeconfigOptions) (ContextSelection, error) {
	return ResolveContext(options)
}

func NewRunner(checker ClusterChecker) Runner {
	return Runner{
		Resolver: defaultResolver{},
		Checker:  checker,
	}
}

func (r Runner) Run(options KubeconfigOptions) (Result, error) {
	if r.Resolver == nil {
		return Result{}, fmt.Errorf("preflight context resolver is required")
	}
	if r.Checker == nil {
		return Result{}, fmt.Errorf("preflight cluster checker is required")
	}

	context, err := r.Resolver.Resolve(options)
	if err != nil {
		return Result{}, fmt.Errorf("resolve kubeconfig context: %w", err)
	}

	result := Result{
		Context:         context,
		DiscoveryStatus: StatusPass,
	}

	serverVersion, err := r.Checker.ServerVersion(context)
	if err != nil {
		result.Limitations = append(result.Limitations, Limitation{
			Code:     "kube.server_version.unavailable",
			Severity: "ERROR",
			Summary:  "Kubernetes server version could not be read.",
		})
		return result, fmt.Errorf("read server version: %w", err)
	}
	result.ServerVersion = serverVersion

	if err := r.Checker.Discovery(context); err != nil {
		result.DiscoveryStatus = StatusFail
		result.Limitations = append(result.Limitations, Limitation{
			Code:     "kube.discovery.unavailable",
			Severity: "ERROR",
			Summary:  "Kubernetes API discovery failed.",
		})
		return result, nil
	}

	checks, err := r.Checker.Permissions(context)
	if err != nil {
		result.Limitations = append(result.Limitations, Limitation{
			Code:     "kube.rbac.unavailable",
			Severity: "ERROR",
			Summary:  "Kubernetes permission checks failed.",
		})
		return result, fmt.Errorf("check permissions: %w", err)
	}
	result.PermissionChecks = checks
	result.Limitations = append(result.Limitations, limitationsForPermissionChecks(checks)...)

	return result, nil
}

func limitationsForPermissionChecks(checks []PermissionCheck) []Limitation {
	var limitations []Limitation
	for _, check := range checks {
		if check.Status != StatusFail {
			continue
		}

		severity := "WARN"
		if check.EvidenceClass == EvidenceRequired {
			severity = "ERROR"
		}
		limitations = append(limitations, Limitation{
			Code:     "kube.rbac.denied",
			Severity: severity,
			Summary:  fmt.Sprintf("Kubernetes %s access denied for %s.", check.Verb, check.Resource),
		})
	}
	return limitations
}
