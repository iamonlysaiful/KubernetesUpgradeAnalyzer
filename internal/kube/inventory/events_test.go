package inventory

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCollectEventsBuildsSanitizedDeterministicSummaries(t *testing.T) {
	client := fake.NewSimpleClientset(
		eventObject("event-b", "team-b", "v1", "Pod", "api-pod", corev1.EventTypeWarning, "BackOff", time.Date(2026, 7, 23, 9, 0, 0, 0, time.UTC)),
		eventObject("event-a", "team-a", "apps/v1", "Deployment", "api", corev1.EventTypeNormal, "ScalingReplicaSet", time.Date(2026, 7, 23, 8, 0, 0, 0, time.UTC)),
		eventObject("event-c", "team-a", "v1", "Service", "edge", "Custom", "UnknownType", time.Date(2026, 7, 23, 10, 0, 0, 0, time.UTC)),
	)

	events, err := NewCollector(client).collectEvents(context.Background())
	if err != nil {
		t.Fatalf("collectEvents returned error: %v", err)
	}

	if got := eventKeys(events); got != "2026-07-23T08:00:00Z/team-a/Deployment/api/ScalingReplicaSet/NORMAL,2026-07-23T09:00:00Z/team-b/Pod/api-pod/BackOff/WARNING,2026-07-23T10:00:00Z/team-a/Service/edge/UnknownType/UNKNOWN" {
		t.Fatalf("events = %q", got)
	}
}

func eventObject(name string, namespace string, apiVersion string, kind string, involvedName string, eventType string, reason string, lastSeenAt time.Time) *corev1.Event {
	return &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		InvolvedObject: corev1.ObjectReference{
			APIVersion: apiVersion,
			Kind:       kind,
			Namespace:  namespace,
			Name:       involvedName,
		},
		Type:          eventType,
		Reason:        reason,
		Message:       "raw message must not be copied",
		LastTimestamp: metav1.NewTime(lastSeenAt),
	}
}

func eventKeys(events []Event) string {
	var result string
	for i, event := range events {
		if i > 0 {
			result += ","
		}
		result += event.LastSeenAt + "/" + event.Ref.Namespace + "/" + event.Ref.Kind + "/" + event.Ref.Name + "/" + event.Reason + "/" + event.Type
	}
	return result
}
