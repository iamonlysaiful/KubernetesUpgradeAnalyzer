package inventory

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	snapshotIDPattern        = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{2,127}$`)
	kubernetesVersionPattern = regexp.MustCompile(`^1\.(30|31|32|33)(\.([0-9]+))?([+-][A-Za-z0-9._-]+)?$`)
)

func ValidateCoreSnapshot(snapshot Snapshot) error {
	var problems []string

	if snapshot.SchemaVersion != SchemaVersion {
		problems = append(problems, "schemaVersion must be "+SchemaVersion)
	}
	if !snapshotIDPattern.MatchString(snapshot.SnapshotID) {
		problems = append(problems, "snapshotId must match cluster snapshot id format")
	}
	if _, err := time.Parse(time.RFC3339, snapshot.CapturedAt); err != nil {
		problems = append(problems, "capturedAt must be RFC3339")
	}

	problems = append(problems, validateCluster(snapshot.Cluster)...)
	problems = append(problems, validateKubernetes(snapshot.Kubernetes)...)
	problems = append(problems, validateInventory(snapshot.Inventory)...)
	problems = append(problems, validateLimitations(snapshot.Limitations)...)

	if len(problems) > 0 {
		return fmt.Errorf("core snapshot validation failed: %s", strings.Join(problems, "; "))
	}
	return nil
}

func validateCluster(cluster Cluster) []string {
	var problems []string
	problems = append(problems, validateResourceRef("cluster.identity", cluster.Identity, false)...)
	if !oneOf(cluster.Provider.Type, "AKS", "EKS", "GKE", "OPENSHIFT", "GENERIC", "UNKNOWN") {
		problems = append(problems, "cluster.provider.type is invalid")
	}
	if !oneOf(cluster.Provider.Confidence, "HIGH", "MEDIUM", "LOW", "UNKNOWN") {
		problems = append(problems, "cluster.provider.confidence is invalid")
	}
	if cluster.Context.Name == "" {
		problems = append(problems, "cluster.context.name is required")
	}
	if cluster.Context.KubeconfigSource != "" && !oneOf(cluster.Context.KubeconfigSource, "DEFAULT", "EXPLICIT", "UNKNOWN") {
		problems = append(problems, "cluster.context.kubeconfigSource is invalid")
	}
	return problems
}

func validateKubernetes(kubernetes Kubernetes) []string {
	if !kubernetesVersionPattern.MatchString(kubernetes.ServerVersion) {
		return []string{"kubernetes.serverVersion is invalid or outside supported range"}
	}
	return nil
}

func validateInventory(inventory Inventory) []string {
	var problems []string
	if inventory.Namespaces == nil {
		problems = append(problems, "inventory.namespaces must be an array")
	}
	if inventory.Nodes == nil {
		problems = append(problems, "inventory.nodes must be an array")
	}
	if inventory.Workloads == nil {
		problems = append(problems, "inventory.workloads must be an array")
	}
	if inventory.Storage == nil {
		problems = append(problems, "inventory.storage must be an array")
	}
	if inventory.Networking == nil {
		problems = append(problems, "inventory.networking must be an array")
	}
	if inventory.CRDs == nil {
		problems = append(problems, "inventory.crds must be an array")
	}
	if inventory.Events == nil {
		problems = append(problems, "inventory.events must be an array")
	}

	for i, namespace := range inventory.Namespaces {
		problems = append(problems, validateResourceRef(fmt.Sprintf("inventory.namespaces[%d]", i), namespace, false)...)
	}
	for i, node := range inventory.Nodes {
		problems = append(problems, validateNode(i, node)...)
	}
	return problems
}

func validateNode(index int, node Node) []string {
	var problems []string
	prefix := fmt.Sprintf("inventory.nodes[%d]", index)
	problems = append(problems, validateResourceRef(prefix+".ref", node.Ref, false)...)
	if !kubernetesVersionPattern.MatchString(node.KubeletVersion) {
		problems = append(problems, prefix+".kubeletVersion is invalid or outside supported range")
	}
	if node.Conditions == nil {
		problems = append(problems, prefix+".conditions must be an array")
	}
	for conditionIndex, condition := range node.Conditions {
		conditionPrefix := fmt.Sprintf("%s.conditions[%d]", prefix, conditionIndex)
		if condition.Type == "" {
			problems = append(problems, conditionPrefix+".type is required")
		}
		if !oneOf(condition.Status, "TRUE", "FALSE", "UNKNOWN") {
			problems = append(problems, conditionPrefix+".status is invalid")
		}
	}
	return problems
}

func validateLimitations(limitations []Limitation) []string {
	var problems []string
	if limitations == nil {
		return []string{"limitations must be an array"}
	}
	for i, limitation := range limitations {
		prefix := fmt.Sprintf("limitations[%d]", i)
		if !snapshotIDPattern.MatchString(limitation.Code) {
			problems = append(problems, prefix+".code is invalid")
		}
		if !oneOf(limitation.Severity, "INFO", "WARN", "ERROR") {
			problems = append(problems, prefix+".severity is invalid")
		}
		if limitation.Summary == "" {
			problems = append(problems, prefix+".summary is required")
		}
		for resourceIndex, resource := range limitation.Resources {
			problems = append(problems, validateResourceRef(fmt.Sprintf("%s.resources[%d]", prefix, resourceIndex), resource, false)...)
		}
	}
	return problems
}

func validateResourceRef(prefix string, ref ResourceRef, namespaced bool) []string {
	var problems []string
	if ref.Kind == "" {
		problems = append(problems, prefix+".kind is required")
	}
	if ref.Name == "" {
		problems = append(problems, prefix+".name is required")
	}
	if namespaced && ref.Namespace == "" {
		problems = append(problems, prefix+".namespace is required")
	}
	return problems
}

func oneOf(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}
