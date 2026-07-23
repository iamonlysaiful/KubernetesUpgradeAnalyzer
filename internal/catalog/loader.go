package catalog

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

var semverPattern = regexp.MustCompile(`^(0|[1-9][0-9]*)\.[0-9]+\.[0-9]+(-[A-Za-z0-9.-]+)?$`)

type ValidationError struct {
	Problems []string
}

func (err ValidationError) Error() string {
	return "catalog validation failed: " + strings.Join(err.Problems, "; ")
}

func LoadFile(path string) (Bundle, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Bundle{}, err
	}
	return LoadBytes(data, SourceFile)
}

func LoadBytes(data []byte, source SourceKind) (Bundle, error) {
	var bundle Bundle
	if err := json.Unmarshal(data, &bundle); err != nil {
		return Bundle{}, fmt.Errorf("decode catalog: %w", err)
	}
	bundle.Source = source
	bundle.ChecksumSHA256 = checksum(data)
	if err := Validate(bundle); err != nil {
		return Bundle{}, err
	}
	return bundle, nil
}

func Validate(bundle Bundle) error {
	var problems []string
	if bundle.SchemaVersion != SchemaVersion {
		problems = append(problems, "schemaVersion must be "+SchemaVersion)
	}
	if !semverPattern.MatchString(bundle.CatalogVersion) {
		problems = append(problems, "catalogVersion must be semantic version")
	}
	if _, err := time.Parse(time.RFC3339, bundle.CreatedAt); err != nil {
		problems = append(problems, "createdAt must be RFC3339")
	}
	if bundle.ReviewAfter != "" {
		if _, err := time.Parse(time.RFC3339, bundle.ReviewAfter); err != nil {
			problems = append(problems, "reviewAfter must be RFC3339")
		}
	}
	if bundle.Kubernetes.ValidatedRange.MinMinor == "" || bundle.Kubernetes.ValidatedRange.MaxMinor == "" {
		problems = append(problems, "kubernetes.validatedRange minMinor and maxMinor are required")
	}

	sourceIDs := map[string]bool{}
	for i, source := range bundle.Sources {
		prefix := fmt.Sprintf("sources[%d]", i)
		if source.SourceID == "" {
			problems = append(problems, prefix+".sourceId is required")
		}
		if sourceIDs[source.SourceID] {
			problems = append(problems, prefix+".sourceId is duplicate")
		}
		sourceIDs[source.SourceID] = true
		if source.Title == "" {
			problems = append(problems, prefix+".title is required")
		}
		if source.URL == "" {
			problems = append(problems, prefix+".url is required")
		}
		if _, err := time.Parse(time.RFC3339, source.RetrievedAt); err != nil {
			problems = append(problems, prefix+".retrievedAt must be RFC3339")
		}
		if source.Claim == "" {
			problems = append(problems, prefix+".claim is required")
		}
	}

	problems = append(problems, validateAPIRules(bundle.Kubernetes.APIRules, sourceIDs)...)
	problems = append(problems, validateProviders(bundle.Providers, sourceIDs)...)
	problems = append(problems, validateComponents(bundle.Components, sourceIDs)...)

	if len(problems) > 0 {
		sort.Strings(problems)
		return ValidationError{Problems: problems}
	}
	return nil
}

func validateAPIRules(rules []APIRule, sourceIDs map[string]bool) []string {
	var problems []string
	seen := map[string]bool{}
	for i, rule := range rules {
		prefix := fmt.Sprintf("kubernetes.apiRules[%d]", i)
		if rule.RuleID == "" {
			problems = append(problems, prefix+".ruleId is required")
		}
		if seen[rule.RuleID] {
			problems = append(problems, prefix+".ruleId is duplicate")
		}
		seen[rule.RuleID] = true
		if rule.APIVersion == "" {
			problems = append(problems, prefix+".apiVersion is required")
		}
		if rule.Kind == "" {
			problems = append(problems, prefix+".kind is required")
		}
		if rule.Status != "removed" && rule.Status != "deprecated" {
			problems = append(problems, prefix+".status is invalid")
		}
		if rule.RemovedIn == "" {
			problems = append(problems, prefix+".removedIn is required")
		}
		if !sourceIDs[rule.SourceID] {
			problems = append(problems, prefix+".sourceId must reference a source")
		}
	}
	return problems
}

func validateProviders(providers []Provider, sourceIDs map[string]bool) []string {
	var problems []string
	for i, provider := range providers {
		prefix := fmt.Sprintf("providers[%d]", i)
		if provider.Provider != "AKS" {
			problems = append(problems, prefix+".provider is invalid")
		}
		if provider.CandidateSource != "LIVE_PROVIDER_EVIDENCE" && provider.CandidateSource != "CATALOG_POLICY" {
			problems = append(problems, prefix+".candidateSource is invalid")
		}
		if !sourceIDs[provider.SourceID] {
			problems = append(problems, prefix+".sourceId must reference a source")
		}
	}
	return problems
}

func validateComponents(components []Component, sourceIDs map[string]bool) []string {
	var problems []string
	idsAndAliases := map[string]string{}
	for i, component := range components {
		prefix := fmt.Sprintf("components[%d]", i)
		if component.ProductID == "" {
			problems = append(problems, prefix+".productId is required")
		}
		if previous, ok := idsAndAliases[component.ProductID]; ok {
			problems = append(problems, fmt.Sprintf("%s.productId duplicates %s", prefix, previous))
		}
		idsAndAliases[component.ProductID] = prefix + ".productId"
		for aliasIndex, alias := range component.Aliases {
			if alias == "" {
				problems = append(problems, fmt.Sprintf("%s.aliases[%d] is required", prefix, aliasIndex))
				continue
			}
			if previous, ok := idsAndAliases[alias]; ok {
				problems = append(problems, fmt.Sprintf("%s.aliases[%d] duplicates %s", prefix, aliasIndex, previous))
			}
			idsAndAliases[alias] = fmt.Sprintf("%s.aliases[%d]", prefix, aliasIndex)
		}
		if component.ProductVersionRange == "" {
			problems = append(problems, prefix+".productVersionRange is required")
		}
		if component.KubernetesRange == "" {
			problems = append(problems, prefix+".kubernetesRange is required")
		}
		if !oneOf(component.Status, "supported", "unsupported", "conditional", "unknown") {
			problems = append(problems, prefix+".status is invalid")
		}
		if !oneOf(component.Confidence, "authoritative", "inferred", "community", "unknown") {
			problems = append(problems, prefix+".confidence is invalid")
		}
		if component.ExpiresAt != "" {
			if _, err := time.Parse(time.RFC3339, component.ExpiresAt); err != nil {
				problems = append(problems, prefix+".expiresAt must be RFC3339")
			}
		}
		if !sourceIDs[component.SourceID] {
			problems = append(problems, prefix+".sourceId must reference a source")
		}
	}
	return problems
}

func checksum(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func oneOf(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}
