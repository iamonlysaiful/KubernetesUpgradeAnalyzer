package inventory

import (
	"context"
	"sort"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c Collector) collectWorkloads(ctx context.Context) ([]Workload, error) {
	var workloads []Workload

	deployments, err := c.Client.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, deployment := range deployments.Items {
		workloads = append(workloads, workloadFromDeployment(deployment))
	}

	daemonSets, err := c.Client.AppsV1().DaemonSets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, daemonSet := range daemonSets.Items {
		workloads = append(workloads, workloadFromDaemonSet(daemonSet))
	}

	statefulSets, err := c.Client.AppsV1().StatefulSets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, statefulSet := range statefulSets.Items {
		workloads = append(workloads, workloadFromStatefulSet(statefulSet))
	}

	replicaSets, err := c.Client.AppsV1().ReplicaSets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, replicaSet := range replicaSets.Items {
		// Skip scaled-to-zero ReplicaSets owned by a Deployment: Kubernetes
		// retains these as rollout history (old revisions) so rollbacks are
		// possible, but they are not running and would otherwise be double
		// counted (and misdetected as distinct component versions) alongside
		// the active Deployment they belong to.
		if isStaleDeploymentRevision(replicaSet) {
			continue
		}
		workloads = append(workloads, workloadFromReplicaSet(replicaSet))
	}

	jobs, err := c.Client.BatchV1().Jobs("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, job := range jobs.Items {
		workloads = append(workloads, workloadFromJob(job))
	}

	cronJobs, err := c.Client.BatchV1().CronJobs("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, cronJob := range cronJobs.Items {
		workloads = append(workloads, workloadFromCronJob(cronJob))
	}

	sortWorkloads(workloads)
	return workloads, nil
}

func workloadFromDeployment(deployment appsv1.Deployment) Workload {
	return Workload{
		Ref:             workloadRef("apps/v1", "Deployment", deployment.Namespace, deployment.Name),
		DesiredReplicas: int32Value(deployment.Spec.Replicas, 1),
		ReadyReplicas:   int(deployment.Status.ReadyReplicas),
		Critical:        "UNKNOWN",
		VersionLabel:    podVersionLabel(deployment.Spec.Template.Labels),
		Containers:      containers(deployment.Spec.Template.Spec.Containers),
	}
}

func workloadFromDaemonSet(daemonSet appsv1.DaemonSet) Workload {
	return Workload{
		Ref:             workloadRef("apps/v1", "DaemonSet", daemonSet.Namespace, daemonSet.Name),
		DesiredReplicas: int(daemonSet.Status.DesiredNumberScheduled),
		ReadyReplicas:   int(daemonSet.Status.NumberReady),
		Critical:        "UNKNOWN",
		VersionLabel:    podVersionLabel(daemonSet.Spec.Template.Labels),
		Containers:      containers(daemonSet.Spec.Template.Spec.Containers),
	}
}

func workloadFromStatefulSet(statefulSet appsv1.StatefulSet) Workload {
	return Workload{
		Ref:             workloadRef("apps/v1", "StatefulSet", statefulSet.Namespace, statefulSet.Name),
		DesiredReplicas: int32Value(statefulSet.Spec.Replicas, 1),
		ReadyReplicas:   int(statefulSet.Status.ReadyReplicas),
		Critical:        "UNKNOWN",
		VersionLabel:    podVersionLabel(statefulSet.Spec.Template.Labels),
		Containers:      containers(statefulSet.Spec.Template.Spec.Containers),
	}
}

func workloadFromReplicaSet(replicaSet appsv1.ReplicaSet) Workload {
	return Workload{
		Ref:             workloadRef("apps/v1", "ReplicaSet", replicaSet.Namespace, replicaSet.Name),
		DesiredReplicas: int32Value(replicaSet.Spec.Replicas, 1),
		ReadyReplicas:   int(replicaSet.Status.ReadyReplicas),
		Critical:        "UNKNOWN",
		VersionLabel:    podVersionLabel(replicaSet.Spec.Template.Labels),
		Containers:      containers(replicaSet.Spec.Template.Spec.Containers),
	}
}

// isStaleDeploymentRevision reports whether a ReplicaSet is a scaled-to-zero
// revision kept by a Deployment for rollback history rather than an active
// or standalone workload.
func isStaleDeploymentRevision(replicaSet appsv1.ReplicaSet) bool {
	if int32Value(replicaSet.Spec.Replicas, 1) != 0 {
		return false
	}
	for _, owner := range replicaSet.OwnerReferences {
		if owner.Kind == "Deployment" {
			return true
		}
	}
	return false
}

func workloadFromJob(job batchv1.Job) Workload {
	return Workload{
		Ref:             workloadRef("batch/v1", "Job", job.Namespace, job.Name),
		DesiredReplicas: int32Value(job.Spec.Parallelism, 1),
		ReadyReplicas:   int(job.Status.Succeeded),
		Critical:        "UNKNOWN",
		VersionLabel:    podVersionLabel(job.Spec.Template.Labels),
		Containers:      containers(job.Spec.Template.Spec.Containers),
	}
}

func workloadFromCronJob(cronJob batchv1.CronJob) Workload {
	desiredReplicas := 1
	if cronJob.Spec.Suspend != nil && *cronJob.Spec.Suspend {
		desiredReplicas = 0
	}
	return Workload{
		Ref:             workloadRef("batch/v1", "CronJob", cronJob.Namespace, cronJob.Name),
		DesiredReplicas: desiredReplicas,
		ReadyReplicas:   0,
		Critical:        "UNKNOWN",
		VersionLabel:    podVersionLabel(cronJob.Spec.JobTemplate.Spec.Template.Labels),
		Containers:      containers(cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers),
	}
}

// podVersionLabel returns the value of app.kubernetes.io/version from
// pod-template labels, which is set by Helm charts and operators to the actual
// application version (e.g. "5.8.8" for EMQX). Returns empty string if absent.
func podVersionLabel(labels map[string]string) string {
	return labels["app.kubernetes.io/version"]
}

func workloadRef(apiVersion string, kind string, namespace string, name string) ResourceRef {
	return ResourceRef{
		APIVersion: apiVersion,
		Kind:       kind,
		Namespace:  namespace,
		Name:       name,
	}
}

func containers(kubernetesContainers []corev1.Container) []Container {
	result := make([]Container, 0, len(kubernetesContainers))
	for _, container := range kubernetesContainers {
		result = append(result, Container{
			Name:     container.Name,
			Image:    container.Image,
			ImageTag: imageTag(container.Image),
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

func int32Value(value *int32, defaultValue int) int {
	if value == nil {
		return defaultValue
	}
	return int(*value)
}

func imageTag(image string) string {
	if strings.Contains(image, "@") {
		return ""
	}
	lastSlash := strings.LastIndex(image, "/")
	lastColon := strings.LastIndex(image, ":")
	if lastColon <= lastSlash {
		return ""
	}
	tag := image[lastColon+1:]
	if tag == "" {
		return ""
	}
	return tag
}

func sortWorkloads(workloads []Workload) {
	sort.Slice(workloads, func(i, j int) bool {
		if workloads[i].Ref.Namespace != workloads[j].Ref.Namespace {
			return workloads[i].Ref.Namespace < workloads[j].Ref.Namespace
		}
		if workloads[i].Ref.Kind != workloads[j].Ref.Kind {
			return workloads[i].Ref.Kind < workloads[j].Ref.Kind
		}
		return workloads[i].Ref.Name < workloads[j].Ref.Name
	})
}
