// Package plan provides upgrade plan generation and validation checklists.
package plan

import (
	"fmt"
	"time"
)

// UpgradePlan contains the complete upgrade execution plan.
type UpgradePlan struct {
	// Steps are the ordered upgrade execution steps.
	Steps []PlanStep `json:"steps"`
	// EstimatedTime is the total estimated time for the upgrade.
	EstimatedTime time.Duration `json:"estimatedTime"`
	// ValidationSteps are post-upgrade validation checks.
	ValidationSteps []PlanStep `json:"validationSteps"`
	// RollbackGuidance provides rollback instructions if issues occur.
	RollbackGuidance string `json:"rollbackGuidance,omitempty"`
}

// PlanStep represents a single step in the upgrade plan.
type PlanStep struct {
	// Order is the step sequence number (1-based).
	Order int `json:"order"`
	// Description explains what this step does.
	Description string `json:"description"`
	// Command is the CLI command to execute (optional).
	Command string `json:"command,omitempty"`
	// Expected describes the expected outcome.
	Expected string `json:"expected,omitempty"`
	// EstimatedTime is the estimated duration for this step.
	EstimatedTime time.Duration `json:"estimatedTime,omitempty"`
}

// PlanInput contains information needed to generate an upgrade plan.
type PlanInput struct {
	// Provider is the cloud provider type (e.g., "AKS").
	Provider string
	// ClusterName is the cluster identifier.
	ClusterName string
	// ResourceGroup is the Azure resource group (AKS-specific).
	ResourceGroup string
	// CurrentVersion is the current Kubernetes version.
	CurrentVersion string
	// TargetVersion is the target Kubernetes version.
	TargetVersion string
	// NodePoolCount is the number of node pools.
	NodePoolCount int
	// TotalNodeCount is the total number of nodes.
	TotalNodeCount int
	// HasStatefulSets indicates if StatefulSets are present.
	HasStatefulSets bool
	// HasIngress indicates if Ingress resources are present.
	HasIngress bool
	// DetectedComponents lists detected third-party components.
	DetectedComponents []string
	// RiskLevel indicates the assessed risk level.
	RiskLevel string
}

// Generator creates upgrade plans based on cluster configuration.
type Generator struct{}

// NewGenerator creates a new plan generator.
func NewGenerator() *Generator {
	return &Generator{}
}

// Generate creates an upgrade plan from the input.
func (g *Generator) Generate(input PlanInput) *UpgradePlan {
	switch input.Provider {
	case "AKS", "aks":
		return g.generateAKSPlan(input)
	default:
		return g.generateGenericPlan(input)
	}
}

// generateAKSPlan creates an AKS-specific upgrade plan.
func (g *Generator) generateAKSPlan(input PlanInput) *UpgradePlan {
	plan := &UpgradePlan{
		Steps:           make([]PlanStep, 0),
		ValidationSteps: make([]PlanStep, 0),
	}

	order := 1

	// Step 1: Backup
	plan.Steps = append(plan.Steps, PlanStep{
		Order:         order,
		Description:   "Take AKS backup or etcd snapshot",
		Command:       fmt.Sprintf("az aks update -g %s -n %s --enable-backup", input.ResourceGroup, input.ClusterName),
		Expected:      "Backup created successfully",
		EstimatedTime: 5 * time.Minute,
	})
	order++

	// Step 2: Upgrade control plane
	plan.Steps = append(plan.Steps, PlanStep{
		Order:         order,
		Description:   fmt.Sprintf("Upgrade control plane to %s", input.TargetVersion),
		Command:       fmt.Sprintf("az aks upgrade -g %s -n %s --kubernetes-version %s --control-plane-only -y", input.ResourceGroup, input.ClusterName, input.TargetVersion),
		Expected:      "Control plane upgraded successfully",
		EstimatedTime: 15 * time.Minute,
	})
	order++

	// Step 3: Upgrade system node pool
	plan.Steps = append(plan.Steps, PlanStep{
		Order:         order,
		Description:   "Upgrade system node pool",
		Command:       fmt.Sprintf("az aks nodepool upgrade -g %s --cluster-name %s -n systempool --kubernetes-version %s", input.ResourceGroup, input.ClusterName, input.TargetVersion),
		Expected:      "System node pool upgraded",
		EstimatedTime: 5 * time.Minute,
	})
	order++

	// Step 4: Upgrade user node pools (one at a time recommended)
	if input.NodePoolCount > 1 {
		plan.Steps = append(plan.Steps, PlanStep{
			Order:         order,
			Description:   fmt.Sprintf("Upgrade user node pools (one at a time, %d pools)", input.NodePoolCount-1),
			Command:       fmt.Sprintf("az aks nodepool upgrade -g %s --cluster-name %s -n <pool-name> --kubernetes-version %s", input.ResourceGroup, input.ClusterName, input.TargetVersion),
			Expected:      "All user node pools upgraded",
			EstimatedTime: time.Duration(input.NodePoolCount-1) * 5 * time.Minute,
		})
		order++
	}

	// Step 5: Wait for all nodes Ready
	plan.Steps = append(plan.Steps, PlanStep{
		Order:         order,
		Description:   "Wait for all nodes to be Ready",
		Command:       "kubectl get nodes -w",
		Expected:      "All nodes showing Ready status",
		EstimatedTime: 2 * time.Minute,
	})
	order++

	// Step 6: Verify StatefulSets (if present)
	if input.HasStatefulSets {
		plan.Steps = append(plan.Steps, PlanStep{
			Order:         order,
			Description:   "Verify StatefulSets are healthy",
			Command:       "kubectl get statefulsets -A -o wide",
			Expected:      "All StatefulSets showing ready replicas",
			EstimatedTime: 2 * time.Minute,
		})
		order++
	}

	// Step 7: Run application smoke tests
	plan.Steps = append(plan.Steps, PlanStep{
		Order:         order,
		Description:   "Run application smoke tests",
		Expected:      "All critical application endpoints responding",
		EstimatedTime: 5 * time.Minute,
	})
	order++

	// Step 8: Monitor for 30 minutes
	plan.Steps = append(plan.Steps, PlanStep{
		Order:         order,
		Description:   "Monitor cluster for 30 minutes",
		Command:       "kubectl get events -A --watch",
		Expected:      "No unusual errors or restarts",
		EstimatedTime: 30 * time.Minute,
	})

	// Calculate estimated time
	plan.EstimatedTime = g.estimateTime(input)

	// Generate validation steps
	plan.ValidationSteps = g.generateValidationSteps(input)

	// Generate rollback guidance
	plan.RollbackGuidance = g.generateRollbackGuidance(input)

	return plan
}

// generateGenericPlan creates a provider-agnostic upgrade plan.
func (g *Generator) generateGenericPlan(input PlanInput) *UpgradePlan {
	plan := &UpgradePlan{
		Steps:           make([]PlanStep, 0),
		ValidationSteps: make([]PlanStep, 0),
	}

	plan.Steps = append(plan.Steps, PlanStep{
		Order:         1,
		Description:   "Take cluster backup",
		Expected:      "Backup created successfully",
		EstimatedTime: 10 * time.Minute,
	})

	plan.Steps = append(plan.Steps, PlanStep{
		Order:         2,
		Description:   fmt.Sprintf("Upgrade cluster to %s", input.TargetVersion),
		Expected:      "Cluster upgraded successfully",
		EstimatedTime: 30 * time.Minute,
	})

	plan.Steps = append(plan.Steps, PlanStep{
		Order:         3,
		Description:   "Verify cluster health",
		Command:       "kubectl get nodes && kubectl get pods -A",
		Expected:      "All nodes Ready, all pods Running",
		EstimatedTime: 5 * time.Minute,
	})

	plan.EstimatedTime = g.estimateTime(input)
	plan.ValidationSteps = g.generateValidationSteps(input)
	plan.RollbackGuidance = g.generateRollbackGuidance(input)

	return plan
}

// estimateTime calculates the estimated upgrade duration.
// Formula: Base (15 min) + 5 min per node pool + 1 min per 10 nodes
func (g *Generator) estimateTime(input PlanInput) time.Duration {
	base := 15 * time.Minute
	nodePoolTime := time.Duration(input.NodePoolCount) * 5 * time.Minute
	nodeTime := time.Duration(input.TotalNodeCount/10) * time.Minute
	monitoring := 30 * time.Minute

	return base + nodePoolTime + nodeTime + monitoring
}

// generateValidationSteps creates post-upgrade validation checks.
func (g *Generator) generateValidationSteps(input PlanInput) []PlanStep {
	steps := []PlanStep{
		{
			Order:       1,
			Description: "All nodes Ready",
			Command:     "kubectl get nodes",
			Expected:    "All nodes showing Ready status",
		},
		{
			Order:       2,
			Description: "All DaemonSets at desired count",
			Command:     "kubectl get daemonsets -A",
			Expected:    "DESIRED equals READY for all DaemonSets",
		},
		{
			Order:       3,
			Description: "All Deployments available",
			Command:     "kubectl get deployments -A",
			Expected:    "All Deployments show available replicas",
		},
		{
			Order:       4,
			Description: "All StatefulSets ready",
			Command:     "kubectl get statefulsets -A",
			Expected:    "All StatefulSets show ready replicas",
		},
		{
			Order:       5,
			Description: "No CrashLoopBackOff pods",
			Command:     "kubectl get pods -A | grep -v Running | grep -v Completed",
			Expected:    "No pods in CrashLoopBackOff or Error state",
		},
	}

	nextOrder := 6

	if input.HasIngress {
		steps = append(steps, PlanStep{
			Order:       nextOrder,
			Description: "Ingress responding",
			Command:     "curl -I https://your-ingress-url",
			Expected:    "HTTP 200 OK response",
		})
		nextOrder++
	}

	// Add one validation step per distinct component type (multiple detections
	// of the same component share one post-upgrade check).
	seen := make(map[string]bool)
	for _, comp := range input.DetectedComponents {
		switch comp {
		case "EMQX", "emqx":
			if !seen["emqx"] {
				seen["emqx"] = true
				steps = append(steps, PlanStep{
					Order:       nextOrder,
					Description: "EMQX cluster healthy",
					Command:     "kubectl get emqxbrokers -A -o wide",
					Expected:    "EMQX cluster showing Running status",
				})
				nextOrder++
			}
		case "CoreDNS", "coredns":
			if !seen["coredns"] {
				seen["coredns"] = true
				steps = append(steps, PlanStep{
					Order:       nextOrder,
					Description: "CoreDNS responding",
					Command:     "kubectl run test-dns --image=busybox --rm -it --restart=Never -- nslookup kubernetes",
					Expected:    "DNS resolution succeeds",
				})
				nextOrder++
			}
		}
	}

	return steps
}

// generateRollbackGuidance creates rollback instructions based on risk level.
func (g *Generator) generateRollbackGuidance(input PlanInput) string {
	switch input.RiskLevel {
	case "HIGH":
		return fmt.Sprintf(`ROLLBACK PROCEDURE (HIGH RISK)
If critical issues occur during or after upgrade:

1. IMMEDIATE: Stop any ongoing node pool upgrades
   az aks nodepool operation-abort -g %s --cluster-name %s -n <pool-name>

2. For control plane issues:
   - AKS does not support control plane rollback
   - Contact Azure Support immediately
   - Have your backup ready for restoration

3. For node pool issues:
   - Consider adding a new node pool with previous version
   - Cordon/drain affected nodes
   - Migrate workloads to healthy nodes

4. For application issues:
   - Roll back application deployments to known-good versions
   - Check pod logs for compatibility issues
   - Review recent events: kubectl get events -A --sort-by='.lastTimestamp'

Keep Azure Support contact ready: https://portal.azure.com/#blade/Microsoft_Azure_Support/HelpAndSupportBlade`, input.ResourceGroup, input.ClusterName)

	case "MEDIUM":
		return fmt.Sprintf(`ROLLBACK GUIDANCE (MEDIUM RISK)
If issues occur during upgrade:

1. Stop ongoing operations if upgrade appears stuck
2. For node issues: drain and replace affected nodes
3. For application issues: roll back deployments to previous versions
4. Monitor events: kubectl get events -A --sort-by='.lastTimestamp'

Note: AKS control plane cannot be rolled back once upgraded.
Node pools can be replaced with previous-version pools if needed.

Resource Group: %s | Cluster: %s`, input.ResourceGroup, input.ClusterName)

	default:
		return `ROLLBACK GUIDANCE
Standard upgrade with low risk. In case of unexpected issues:

1. Monitor pod health: kubectl get pods -A --watch
2. Check events for errors: kubectl get events -A --sort-by='.lastTimestamp'
3. For application issues: roll back specific deployments as needed

Note: AKS control plane upgrades cannot be rolled back.`
	}
}

// FormatDuration formats a duration for display.
func FormatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}
