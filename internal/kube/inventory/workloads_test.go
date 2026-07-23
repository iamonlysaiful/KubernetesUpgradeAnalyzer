package inventory

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCollectWorkloadsBuildsDeterministicSummaries(t *testing.T) {
	replicas := int32(3)
	parallelism := int32(2)
	suspended := true

	client := fake.NewSimpleClientset(
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "team-b"},
			Spec: appsv1.DeploymentSpec{
				Replicas: &replicas,
				Template: podTemplate(
					container("worker", "registry-001/app-worker:2.1.0"),
					container("api", "registry-001/app-api:1.2.3"),
				),
			},
			Status: appsv1.DeploymentStatus{ReadyReplicas: 2},
		},
		&appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{Name: "agent", Namespace: "team-a"},
			Spec: appsv1.DaemonSetSpec{
				Template: podTemplate(container("agent", "registry-001/agent@sha256:abc123")),
			},
			Status: appsv1.DaemonSetStatus{DesiredNumberScheduled: 4, NumberReady: 3},
		},
		&appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{Name: "db", Namespace: "team-a"},
			Spec: appsv1.StatefulSetSpec{
				Template: podTemplate(container("db", "registry-001/db:12")),
			},
			Status: appsv1.StatefulSetStatus{ReadyReplicas: 1},
		},
		&appsv1.ReplicaSet{
			ObjectMeta: metav1.ObjectMeta{Name: "legacy", Namespace: "team-a"},
			Spec: appsv1.ReplicaSetSpec{
				Replicas: &replicas,
				Template: podTemplate(container("legacy", "registry-001/legacy")),
			},
			Status: appsv1.ReplicaSetStatus{ReadyReplicas: 3},
		},
		&batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{Name: "migrate", Namespace: "team-b"},
			Spec: batchv1.JobSpec{
				Parallelism: &parallelism,
				Template:    podTemplate(container("migrate", "registry-001/migrate:2026.07")),
			},
			Status: batchv1.JobStatus{Succeeded: 1},
		},
		&batchv1.CronJob{
			ObjectMeta: metav1.ObjectMeta{Name: "nightly", Namespace: "team-c"},
			Spec: batchv1.CronJobSpec{
				Suspend:     &suspended,
				Schedule:    "0 0 * * *",
				JobTemplate: batchv1.JobTemplateSpec{Spec: batchv1.JobSpec{Template: podTemplate(container("nightly", "registry-001/nightly:1.0.0"))}},
			},
		},
	)

	workloads, err := NewCollector(client).collectWorkloads(context.Background())
	if err != nil {
		t.Fatalf("collectWorkloads returned error: %v", err)
	}

	if got := workloadKeys(workloads); got != "team-a/DaemonSet/agent,team-a/ReplicaSet/legacy,team-a/StatefulSet/db,team-b/Deployment/api,team-b/Job/migrate,team-c/CronJob/nightly" {
		t.Fatalf("workload sort/order = %q", got)
	}

	deployment := workloads[3]
	if deployment.DesiredReplicas != 3 || deployment.ReadyReplicas != 2 || deployment.Critical != "UNKNOWN" {
		t.Fatalf("deployment summary = %#v", deployment)
	}
	if got := containerKeys(deployment.Containers); got != "api:1.2.3,worker:2.1.0" {
		t.Fatalf("deployment containers = %q", got)
	}

	daemonSet := workloads[0]
	if daemonSet.DesiredReplicas != 4 || daemonSet.ReadyReplicas != 3 {
		t.Fatalf("daemonSet summary = %#v", daemonSet)
	}
	if daemonSet.Containers[0].ImageTag != "" {
		t.Fatalf("digest image tag = %q, want empty", daemonSet.Containers[0].ImageTag)
	}

	statefulSet := workloads[2]
	if statefulSet.DesiredReplicas != 1 {
		t.Fatalf("nil statefulSet replicas = %d, want default 1", statefulSet.DesiredReplicas)
	}

	cronJob := workloads[5]
	if cronJob.DesiredReplicas != 0 || cronJob.ReadyReplicas != 0 {
		t.Fatalf("suspended cronJob summary = %#v", cronJob)
	}
}

func podTemplate(containers ...corev1.Container) corev1.PodTemplateSpec {
	return corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{Containers: containers},
	}
}

func container(name string, image string) corev1.Container {
	return corev1.Container{Name: name, Image: image}
}

func workloadKeys(workloads []Workload) string {
	var result string
	for i, workload := range workloads {
		if i > 0 {
			result += ","
		}
		result += workload.Ref.Namespace + "/" + workload.Ref.Kind + "/" + workload.Ref.Name
	}
	return result
}

func containerKeys(containers []Container) string {
	var result string
	for i, container := range containers {
		if i > 0 {
			result += ","
		}
		result += container.Name + ":" + container.ImageTag
	}
	return result
}
