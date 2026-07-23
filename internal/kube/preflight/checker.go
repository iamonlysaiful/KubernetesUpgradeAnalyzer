package preflight

import (
	"context"
	"fmt"

	authorizationv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
)

type KubernetesChecker struct {
	Client kubernetes.Interface
}

func NewKubernetesChecker(client kubernetes.Interface) KubernetesChecker {
	return KubernetesChecker{Client: client}
}

func (c KubernetesChecker) ServerVersion(ContextSelection) (string, error) {
	if c.Client == nil {
		return "", fmt.Errorf("kubernetes client is required")
	}

	info, err := c.Client.Discovery().ServerVersion()
	if err != nil {
		return "", err
	}
	if info == nil || info.GitVersion == "" {
		return "", fmt.Errorf("kubernetes server version is empty")
	}
	return info.GitVersion, nil
}

func (c KubernetesChecker) Discovery(ContextSelection) error {
	if c.Client == nil {
		return fmt.Errorf("kubernetes client is required")
	}

	_, _, err := c.Client.Discovery().ServerGroupsAndResources()
	return normalizeDiscoveryError(err)
}

func (c KubernetesChecker) Permissions(ContextSelection) ([]PermissionCheck, error) {
	if c.Client == nil {
		return nil, fmt.Errorf("kubernetes client is required")
	}

	checks := requiredPermissionChecks()
	for i := range checks {
		allowed, reason, err := c.selfSubjectAccessReview(checks[i].Resource, checks[i].Verb)
		if err != nil {
			return nil, err
		}
		checks[i].Status = StatusFail
		if allowed {
			checks[i].Status = StatusPass
		}
		checks[i].Reason = reason
	}
	return checks, nil
}

func (c KubernetesChecker) selfSubjectAccessReview(resource string, verb string) (bool, string, error) {
	review := &authorizationv1.SelfSubjectAccessReview{
		Spec: authorizationv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authorizationv1.ResourceAttributes{
				Verb:     verb,
				Resource: resource,
			},
		},
	}

	result, err := c.Client.AuthorizationV1().SelfSubjectAccessReviews().Create(context.Background(), review, metav1.CreateOptions{})
	if err != nil {
		return false, "", err
	}
	return result.Status.Allowed, result.Status.Reason, nil
}

func normalizeDiscoveryError(err error) error {
	if err == nil {
		return nil
	}
	if discovery.IsGroupDiscoveryFailedError(err) {
		return err
	}
	return err
}

func requiredPermissionChecks() []PermissionCheck {
	return []PermissionCheck{
		{Resource: "namespaces", Verb: "list", EvidenceClass: EvidenceRequired, Status: StatusUnknown},
		{Resource: "nodes", Verb: "list", EvidenceClass: EvidenceRequired, Status: StatusUnknown},
		{Resource: "pods", Verb: "list", EvidenceClass: EvidenceRequired, Status: StatusUnknown},
		{Resource: "deployments", Verb: "list", EvidenceClass: EvidenceRequired, Status: StatusUnknown},
		{Resource: "daemonsets", Verb: "list", EvidenceClass: EvidenceRequired, Status: StatusUnknown},
		{Resource: "statefulsets", Verb: "list", EvidenceClass: EvidenceRequired, Status: StatusUnknown},
		{Resource: "jobs", Verb: "list", EvidenceClass: EvidenceRequired, Status: StatusUnknown},
		{Resource: "cronjobs", Verb: "list", EvidenceClass: EvidenceRequired, Status: StatusUnknown},
		{Resource: "persistentvolumeclaims", Verb: "list", EvidenceClass: EvidenceRequired, Status: StatusUnknown},
		{Resource: "customresourcedefinitions", Verb: "list", EvidenceClass: EvidenceRequired, Status: StatusUnknown},
		{Resource: "ingresses", Verb: "list", EvidenceClass: EvidenceOptional, Status: StatusUnknown},
		{Resource: "storageclasses", Verb: "list", EvidenceClass: EvidenceOptional, Status: StatusUnknown},
		{Resource: "csidrivers", Verb: "list", EvidenceClass: EvidenceOptional, Status: StatusUnknown},
		{Resource: "events", Verb: "list", EvidenceClass: EvidenceOptional, Status: StatusUnknown},
	}
}
