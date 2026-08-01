// Package gotcha provides a catalog of version-specific Kubernetes breaking
// changes and AKS operational gotchas that are not captured by API-removal
// scanning (kubent), but which operators should be aware of before upgrading.
package gotcha

// Gotcha describes a behavioral or operational breaking change introduced at
// a specific Kubernetes minor version boundary.
type Gotcha struct {
	// Code is a stable unique identifier for this gotcha.
	Code string
	// CrossesMinor is the minor version where this change lands (e.g. 32 for 1.32).
	// A gotcha is returned when fromMinor < CrossesMinor <= toMinor.
	// Use 0 with AlwaysApply:true for gotchas that apply to every upgrade.
	CrossesMinor int
	// AlwaysApply includes this gotcha regardless of version range when true.
	AlwaysApply bool
	// Title is a short description shown in the report.
	Title string
	// Summary explains the change and its potential impact.
	Summary string
	// Action is the recommended operator action.
	Action string
	// Reference is the authoritative upstream documentation URL.
	Reference string
	// AKSOnly restricts the gotcha to AKS clusters only.
	AKSOnly bool
}

// catalog is the built-in set of version-specific gotchas.
// Entries cover behavioral changes NOT already caught by API-removal rules.
// Each entry must cite an authoritative source in Reference.
var catalog = []Gotcha{
	{
		Code:         "SIDECAR_CONTAINERS_GA_1_29",
		CrossesMinor: 29,
		Title:        "Sidecar containers (KEP-753) graduated to GA in 1.29",
		Summary:      "Init containers with `restartPolicy: Always` are now treated as sidecar containers and restart independently of the main containers. Pods that previously relied on init-container sequencing assumptions may behave differently if any init container declares this field.",
		Action:       "Audit init container definitions for `restartPolicy: Always`; verify startup ordering for service-mesh injected or logging sidecars.",
		Reference:    "https://kubernetes.io/docs/concepts/workloads/pods/sidecar-containers/",
	},
	{
		Code:         "SA_TOKEN_LEGACY_CLEANUP_1_31",
		CrossesMinor: 31,
		Title:        "Legacy ServiceAccount token secrets cleaned up from 1.31",
		Summary:      "The LegacyServiceAccountTokenCleanUp feature gate is GA and enabled by default from 1.31. Auto-created `kubernetes.io/service-account-token` Secrets that have not been used for more than one year are deleted automatically. Applications holding a reference to a long-lived static token Secret may lose access after upgrade.",
		Action:       "Identify `type: kubernetes.io/service-account-token` Secrets; migrate to projected bound service account tokens (`automountServiceAccountToken: true` at the Pod level or via projected volumes).",
		Reference:    "https://kubernetes.io/docs/reference/access-authn-authz/service-accounts-admin/#token-controller",
	},
	{
		Code:         "FLOWCONTROL_V1BETA3_REMOVED_1_32",
		CrossesMinor: 32,
		Title:        "flowcontrol.apiserver.k8s.io/v1beta3 removed in 1.32",
		Summary:      "FlowSchema and PriorityLevelConfiguration v1beta3 were deprecated in 1.30 and are removed in 1.32. Any client, operator, or monitoring tool that calls the v1beta3 endpoints directly will receive 404 errors after upgrade.",
		Action:       "Use `flowcontrol.apiserver.k8s.io/v1` (GA since 1.29). Run `kubectl api-resources | grep flowcontrol` to confirm v1 availability before upgrading.",
		Reference:    "https://kubernetes.io/docs/reference/using-api/deprecation-guide/#v1-32",
	},
	{
		Code:         "AKS_LTS_1_32",
		CrossesMinor: 32,
		AKSOnly:      true,
		Title:        "AKS Long Term Support (LTS) is available for 1.32",
		Summary:      "Kubernetes 1.32 is an AKS Long Term Support release, backed for 2 years instead of the standard ~1 year. If your organisation requires extended support or a slower upgrade cadence, 1.32 LTS may be a preferable destination compared to later short-term versions.",
		Action:       "Evaluate whether 1.32 LTS better matches your support requirements before targeting a later version.",
		Reference:    "https://learn.microsoft.com/azure/aks/supported-kubernetes-versions#long-term-support-lts",
	},
	{
		// AKS always separates control-plane and node-pool upgrades; surface this
		// as an informational note for any AKS upgrade.
		Code:        "AKS_NODE_POOL_UPGRADE_SEPARATE",
		AlwaysApply: true,
		AKSOnly:     true,
		Title:       "AKS: control plane and node pool upgrades are separate operations",
		Summary:     "Upgrading the AKS control plane does NOT automatically upgrade node pools. After the control plane upgrade completes, each node pool must be upgraded separately. Running a mixed-version cluster (control plane N+1, nodes at N) is temporarily supported within API skew limits but is not a stable long-term state.",
		Action:      "After control-plane upgrade, run `az aks nodepool list -g <rg> --cluster-name <cluster>` to check node pool versions and upgrade each pool explicitly.",
		Reference:   "https://learn.microsoft.com/azure/aks/upgrade-aks-control-plane",
	},
}

// ScanPath returns gotchas relevant for an upgrade from fromMinor to toMinor
// (both inclusive of the 1.X minor version number, e.g. 34 for 1.34.x).
// AKS-only gotchas are excluded when isAKS is false.
func ScanPath(fromMinor, toMinor int, isAKS bool) []Gotcha {
	var out []Gotcha
	for _, g := range catalog {
		if g.AKSOnly && !isAKS {
			continue
		}
		if g.AlwaysApply || (g.CrossesMinor > fromMinor && g.CrossesMinor <= toMinor) {
			out = append(out, g)
		}
	}
	return out
}
