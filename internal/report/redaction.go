package report

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/plan"
)

// Redactor generates deterministic assessment-local aliases.
type Redactor struct {
	assessmentID string
}

// NewRedactor creates a deterministic redactor.
func NewRedactor(assessmentID string) *Redactor {
	id := assessmentID
	if id == "" {
		id = "assessment-unknown"
	}
	return &Redactor{assessmentID: id}
}

// RedactDocument redacts sensitive identifiers while preserving decision logic.
func (r *Redactor) RedactDocument(doc Document) Document {
	redacted := doc
	redacted.Redacted = true

	for i := range redacted.Findings {
		f := &redacted.Findings[i]
		if f.Resource != nil {
			originalNamespace := f.Resource.Namespace
			originalName := f.Resource.Name
			if f.Resource.Namespace != "" {
				f.Resource.Namespace = r.alias("ns", f.Resource.Namespace)
			}
			if f.Resource.Name != "" {
				f.Resource.Name = r.alias("res", f.Resource.Name)
			}
			f.Summary = r.redactResourceText(f.Summary, originalNamespace, originalName, f.Resource.Namespace, f.Resource.Name)
			f.Detail = r.redactResourceText(f.Detail, originalNamespace, originalName, f.Resource.Namespace, f.Resource.Name)
			f.Remediation = r.redactResourceText(f.Remediation, originalNamespace, originalName, f.Resource.Namespace, f.Resource.Name)
		}

		// Redact common provider identity tokens that can appear in free text.
		f.Summary = r.redactFreeText(f.Summary)
		f.Detail = r.redactFreeText(f.Detail)
		f.Remediation = r.redactFreeText(f.Remediation)
	}

	for i := range redacted.Limitations {
		redacted.Limitations[i].Summary = r.redactFreeText(redacted.Limitations[i].Summary)
		redacted.Limitations[i].Impact = r.redactFreeText(redacted.Limitations[i].Impact)
	}

	if redacted.UpgradePlan != nil {
		p := *redacted.UpgradePlan
		p.Steps = r.redactPlanSteps(redacted.UpgradePlan.Steps)
		p.ValidationSteps = r.redactPlanSteps(redacted.UpgradePlan.ValidationSteps)
		redacted.UpgradePlan = &p
	}

	return redacted
}

func (r *Redactor) redactPlanSteps(steps []plan.PlanStep) []plan.PlanStep {
	out := make([]plan.PlanStep, len(steps))
	for i, s := range steps {
		s.Command = r.redactCommand(s.Command)
		out[i] = s
	}
	return out
}

// redactCommand redacts resource group and cluster/resource name values
// following common Azure CLI and kubectl flags (e.g. "-g <rg> -n <name>").
func (r *Redactor) redactCommand(in string) string {
	if in == "" {
		return in
	}
	out := in
	for _, flag := range []string{"-g", "--resource-group"} {
		out = redactFlagValue(out, flag, r, "rg")
	}
	for _, flag := range []string{"-n", "--name", "--cluster-name"} {
		out = redactFlagValue(out, flag, r, "res")
	}
	return r.redactFreeText(out)
}

func redactFlagValue(in, flag string, r *Redactor, prefix string) string {
	fields := strings.Fields(in)
	changed := false
	for i := 0; i < len(fields)-1; i++ {
		if fields[i] == flag {
			fields[i+1] = r.alias(prefix, fields[i+1])
			changed = true
		}
	}
	if !changed {
		return in
	}
	return strings.Join(fields, " ")
}

func (r *Redactor) redactResourceText(in string, originalNamespace string, originalName string, namespaceAlias string, nameAlias string) string {
	out := in
	if originalNamespace != "" {
		out = strings.ReplaceAll(out, originalNamespace, namespaceAlias)
	}
	if originalName != "" {
		out = strings.ReplaceAll(out, originalName, nameAlias)
	}
	return out
}

func (r *Redactor) alias(prefix, value string) string {
	sum := sha1.Sum([]byte(r.assessmentID + "|" + prefix + "|" + value))
	return fmt.Sprintf("%s-%s", prefix, hex.EncodeToString(sum[:])[:8])
}

func (r *Redactor) redactFreeText(in string) string {
	if in == "" {
		return in
	}
	out := in
	out = redactToken(out, "subscriptions/", r, "sub")
	out = redactToken(out, "resourceGroups/", r, "rg")
	out = redactToken(out, "namespaces/", r, "ns")
	return out
}

func redactToken(in, marker string, r *Redactor, prefix string) string {
	if !strings.Contains(in, marker) {
		return in
	}
	parts := strings.Split(in, marker)
	if len(parts) < 2 {
		return in
	}
	b := strings.Builder{}
	b.WriteString(parts[0])
	for i := 1; i < len(parts); i++ {
		tail := parts[i]
		end := strings.IndexAny(tail, " /,;:\n\t")
		if end < 0 {
			end = len(tail)
		}
		token := tail[:end]
		if token != "" {
			b.WriteString(marker)
			b.WriteString(r.alias(prefix, token))
			b.WriteString(tail[end:])
		} else {
			b.WriteString(marker)
			b.WriteString(tail)
		}
	}
	return b.String()
}
