package recommendation

// FindingEnhancer provides Impact, Action, and IfIgnored data for findings.
// This implements Phase 10.5 - mandatory action items for all findings.

// healthFindingEnhancement returns enhanced data for health findings.
func healthFindingEnhancement(ruleID string) (impact *FindingImpact, action *ActionItem, ifIgnored string) {
	switch ruleID {
	case "health.node.notReady":
		return &FindingImpact{
				Level:       ImpactHigh,
				Explanation: "Node is not Ready and cannot schedule workloads",
			},
			&ActionItem{
				Description: "Investigate node status and resolve issues",
				Command:     "kubectl describe node <node-name>",
				Effort:      "15-60 minutes",
			},
			"Upgrade may fail or leave workloads unscheduled"

	case "health.node.pressure":
		return &FindingImpact{
				Level:       ImpactMedium,
				Explanation: "Node has resource pressure which may affect workload stability",
			},
			&ActionItem{
				Description: "Address resource constraints on the node",
				Command:     "kubectl describe node <node-name> | grep -A5 Conditions",
				Effort:      "10-30 minutes",
			},
			"Workloads may be evicted during upgrade causing instability"

	case "health.node.kubeletSkew":
		return &FindingImpact{
				Level:       ImpactLow,
				Explanation: "Kubelet version differs from API server version",
			},
			&ActionItem{
				Description: "Version skew will resolve after upgrade completes",
				Effort:      "0 minutes (self-resolving)",
			},
			"Minor: kubectl may show version warnings"

	case "health.workload.unavailable":
		return &FindingImpact{
				Level:       ImpactMedium,
				Explanation: "Workload has fewer ready replicas than desired",
			},
			&ActionItem{
				Description: "Investigate and restore workload health",
				Command:     "kubectl describe deployment/<name> -n <namespace>",
				Effort:      "10-30 minutes",
			},
			"Degraded application availability during upgrade"

	case "health.pvc.unbound":
		return &FindingImpact{
				Level:       ImpactHigh,
				Explanation: "PVC is not bound to a PersistentVolume",
			},
			&ActionItem{
				Description: "Resolve PVC binding issues before upgrade",
				Command:     "kubectl describe pvc <name> -n <namespace>",
				Effort:      "15-45 minutes",
			},
			"Pods requiring this storage will fail to start after upgrade"

	case "health.event.warning":
		return &FindingImpact{
				Level:       ImpactLow,
				Explanation: "Recent warning events may indicate underlying issues",
			},
			&ActionItem{
				Description: "Review events and address any concerning patterns",
				Command:     "kubectl get events -A --sort-by='.lastTimestamp' | grep Warning",
				Effort:      "10-20 minutes",
			},
			"Underlying issues may be exacerbated by upgrade"

	case "health.event.unknownType":
		return &FindingImpact{
				Level:       ImpactLow,
				Explanation: "Unknown event types may indicate custom controllers",
			},
			&ActionItem{
				Description: "Review unknown events for any issues",
				Command:     "kubectl get events -A --field-selector type!=Normal,type!=Warning",
				Effort:      "5-10 minutes",
			},
			"Minor: custom controller behavior may be unpredictable"

	default:
		return &FindingImpact{
				Level:       ImpactMedium,
				Explanation: "Health issue detected that may affect upgrade",
			},
			&ActionItem{
				Description: "Review and address the health issue",
				Effort:      "15-30 minutes",
			},
			"Upgrade may encounter issues or instability"
	}
}

// apiFindingEnhancement returns enhanced data for API compatibility findings.
func apiFindingEnhancement(isRemoved bool, replacement, removedIn string) (impact *FindingImpact, action *ActionItem, ifIgnored string) {
	if isRemoved {
		return &FindingImpact{
				Level:       ImpactHigh,
				Explanation: "API is removed in target version and will cause failures",
			},
			&ActionItem{
				Description: migrateAPIAction(replacement),
				Command:     "kubectl convert -f <manifest> --output-version <new-api>",
				Effort:      "5-30 minutes per resource",
			},
			"Workloads using this API will fail to deploy after upgrade"
	}

	// Deprecated but not yet removed
	return &FindingImpact{
			Level:       ImpactMedium,
			Explanation: "API is deprecated and will be removed in a future version",
		},
		&ActionItem{
			Description: migrateAPIAction(replacement),
			Command:     "kubectl convert -f <manifest> --output-version <new-api>",
			Effort:      "5-15 minutes per resource",
		},
		"API will stop working in future upgrades; plan migration"
}

func migrateAPIAction(replacement string) string {
	if replacement != "" {
		return "Migrate to " + replacement
	}
	return "Update API version to current stable version"
}

// componentFindingEnhancement returns enhanced data for component findings.
func componentFindingEnhancement(componentName string, isUnknown bool, isPartial bool) (impact *FindingImpact, action *ActionItem, ifIgnored string) {
	if isUnknown {
		return &FindingImpact{
				Level:       ImpactMedium,
				Explanation: "Component version cannot be determined from image tags",
			},
			&ActionItem{
				Description: "Verify " + componentName + " version and compatibility manually",
				Command:     "kubectl get pods -A -o jsonpath='{.items[*].spec.containers[*].image}' | grep -i " + componentName,
				Effort:      "10-15 minutes",
			},
			"Component may have compatibility issues not detected by analysis"
	}

	if isPartial {
		return &FindingImpact{
				Level:       ImpactLow,
				Explanation: "Some instances have clear versions, others are ambiguous",
			},
			&ActionItem{
				Description: "Review workloads with ambiguous image tags for " + componentName,
				Effort:      "5-10 minutes",
			},
			"Minor: some component instances may have undetected issues"
	}

	// Known version (INFO finding)
	return &FindingImpact{
			Level:       ImpactLow,
			Explanation: "Component detected with known version",
		},
		&ActionItem{
			Description: "Verify " + componentName + " compatibility with target Kubernetes version",
			Command:     "Check component release notes and compatibility matrix",
			Effort:      "5 minutes",
		},
		"None: component version is known"
}

// providerFindingEnhancement returns enhanced data for provider findings.
func providerFindingEnhancement(code string) (impact *FindingImpact, action *ActionItem, ifIgnored string) {
	switch code {
	case "PROVIDER_UNAVAILABLE":
		return &FindingImpact{
				Level:       ImpactMedium,
				Explanation: "Cannot verify upgrade availability from cloud provider",
			},
			&ActionItem{
				Description: "Check upgrade availability via Azure Portal or CLI",
				Command:     "az aks get-upgrades -g <rg> -n <cluster> -o table",
				Effort:      "2-5 minutes",
			},
			"Upgrade path not validated; manual verification required"

	case "FILE_EVIDENCE_SOURCE":
		return &FindingImpact{
				Level:       ImpactLow,
				Explanation: "Provider evidence loaded from file may not reflect current state",
			},
			&ActionItem{
				Description: "Re-run with live provider access for current state",
				Command:     "kua analyze --provider-source=auto",
				Effort:      "2 minutes",
			},
			"Minor: evidence may be stale"

	default:
		return &FindingImpact{
				Level:       ImpactMedium,
				Explanation: "Provider evidence issue may affect recommendation accuracy",
			},
			&ActionItem{
				Description: "Address provider evidence limitation",
				Effort:      "5-10 minutes",
			},
			"Recommendation confidence reduced"
	}
}

// EnhanceHealthFinding adds Impact, Action, and IfIgnored to a health finding.
func EnhanceHealthFinding(f *Finding) {
	impact, action, ifIgnored := healthFindingEnhancement(f.ID)
	f.Impact = impact
	f.Action = action
	f.IfIgnored = ifIgnored
}

// EnhanceAPIFinding adds Impact, Action, and IfIgnored to an API finding.
func EnhanceAPIFinding(f *Finding, isRemoved bool, replacement, removedIn string) {
	impact, action, ifIgnored := apiFindingEnhancement(isRemoved, replacement, removedIn)
	f.Impact = impact
	f.Action = action
	f.IfIgnored = ifIgnored
}

// EnhanceComponentFinding adds Impact, Action, and IfIgnored to a component finding.
func EnhanceComponentFinding(f *Finding, componentName string, isUnknown, isPartial bool) {
	impact, action, ifIgnored := componentFindingEnhancement(componentName, isUnknown, isPartial)
	f.Impact = impact
	f.Action = action
	f.IfIgnored = ifIgnored
}

// EnhanceProviderFinding adds Impact, Action, and IfIgnored to a provider finding.
func EnhanceProviderFinding(f *Finding, code string) {
	impact, action, ifIgnored := providerFindingEnhancement(code)
	f.Impact = impact
	f.Action = action
	f.IfIgnored = ifIgnored
}
