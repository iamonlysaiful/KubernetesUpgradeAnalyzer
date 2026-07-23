package inventory

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/kube/preflight"
)

const SchemaVersion = "kua.cluster-snapshot.v1"

type Collector struct {
	Client kubernetes.Interface
	Clock  func() time.Time
}

func NewCollector(client kubernetes.Interface) Collector {
	return Collector{
		Client: client,
		Clock:  time.Now,
	}
}

func (c Collector) CollectCore(ctx context.Context, preflightResult preflight.Result) (Snapshot, error) {
	if c.Client == nil {
		return Snapshot{}, fmt.Errorf("kubernetes client is required")
	}

	namespaces, err := c.collectNamespaces(ctx)
	if err != nil {
		return Snapshot{}, fmt.Errorf("collect namespaces: %w", err)
	}

	nodes, err := c.collectNodes(ctx)
	if err != nil {
		return Snapshot{}, fmt.Errorf("collect nodes: %w", err)
	}

	capturedAt := c.now().UTC()
	return Snapshot{
		SchemaVersion: SchemaVersion,
		SnapshotID:    snapshotID(preflightResult.Context.Name, capturedAt),
		CapturedAt:    capturedAt.Format(time.RFC3339),
		Cluster: Cluster{
			Identity: ResourceRef{Kind: "Cluster", Name: preflightResult.Context.Name},
			Provider: Provider{
				Type:       "UNKNOWN",
				Confidence: "UNKNOWN",
			},
			Context: Context{
				Name:             preflightResult.Context.Name,
				KubeconfigSource: string(preflightResult.Context.KubeconfigSource),
			},
		},
		Kubernetes: Kubernetes{
			ServerVersion: normalizeServerVersion(preflightResult.ServerVersion),
		},
		Inventory: Inventory{
			Namespaces: namespaces,
			Nodes:      nodes,
			Workloads:  []Workload{},
			Storage:    []ResourceRef{},
			Networking: []ResourceRef{},
			CRDs:       []ResourceRef{},
			Events:     []Event{},
		},
		Limitations: []Limitation{
			{
				Code:     "PARTIAL_INVENTORY_P2_02",
				Severity: "WARN",
				Summary:  "P2-02 collects namespaces and nodes only; workloads, storage, networking, CRDs, and events are intentionally not collected yet.",
			},
		},
	}, nil
}

func (c Collector) now() time.Time {
	if c.Clock == nil {
		return time.Now()
	}
	return c.Clock()
}

func (c Collector) collectNamespaces(ctx context.Context) ([]ResourceRef, error) {
	list, err := c.Client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	namespaces := make([]ResourceRef, 0, len(list.Items))
	for _, namespace := range list.Items {
		namespaces = append(namespaces, ResourceRef{
			APIVersion: "v1",
			Kind:       "Namespace",
			Name:       namespace.Name,
		})
	}
	sort.Slice(namespaces, func(i, j int) bool {
		return namespaces[i].Name < namespaces[j].Name
	})
	return namespaces, nil
}

func (c Collector) collectNodes(ctx context.Context) ([]Node, error) {
	list, err := c.Client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	nodes := make([]Node, 0, len(list.Items))
	for _, node := range list.Items {
		nodes = append(nodes, Node{
			Ref: ResourceRef{
				APIVersion: "v1",
				Kind:       "Node",
				Name:       node.Name,
			},
			KubeletVersion:    normalizeServerVersion(node.Status.NodeInfo.KubeletVersion),
			ProviderIDPresent: node.Spec.ProviderID != "",
			NodePool:          nodePool(node),
			Conditions:        nodeConditions(node.Status.Conditions),
		})
	}
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Ref.Name < nodes[j].Ref.Name
	})
	return nodes, nil
}

func nodePool(node corev1.Node) string {
	for _, key := range []string{
		"kubernetes.azure.com/agentpool",
		"agentpool",
	} {
		if value := node.Labels[key]; value != "" {
			return value
		}
	}
	return ""
}

func nodeConditions(conditions []corev1.NodeCondition) []Condition {
	result := make([]Condition, 0, len(conditions))
	for _, condition := range conditions {
		result = append(result, Condition{
			Type:   string(condition.Type),
			Status: conditionStatus(condition.Status),
			Reason: condition.Reason,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Type < result[j].Type
	})
	return result
}

func conditionStatus(status corev1.ConditionStatus) string {
	switch status {
	case corev1.ConditionTrue:
		return "TRUE"
	case corev1.ConditionFalse:
		return "FALSE"
	default:
		return "UNKNOWN"
	}
}

func normalizeServerVersion(version string) string {
	return strings.TrimPrefix(version, "v")
}

func snapshotID(contextName string, capturedAt time.Time) string {
	cleaned := strings.Builder{}
	for _, r := range contextName {
		switch {
		case r >= 'a' && r <= 'z':
			cleaned.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			cleaned.WriteRune(r)
		case r >= '0' && r <= '9':
			cleaned.WriteRune(r)
		case r == '.' || r == '_' || r == ':' || r == '-':
			cleaned.WriteRune(r)
		default:
			cleaned.WriteRune('-')
		}
	}
	prefix := cleaned.String()
	if prefix == "" {
		prefix = "cluster"
	}
	id := fmt.Sprintf("%s-%s", prefix, capturedAt.UTC().Format("20060102T150405Z"))
	if len(id) > 128 {
		return id[:128]
	}
	return id
}
