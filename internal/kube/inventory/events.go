package inventory

import (
	"context"
	"sort"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c Collector) collectEvents(ctx context.Context) ([]Event, error) {
	list, err := c.Client.CoreV1().Events("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	events := make([]Event, 0, len(list.Items))
	for _, event := range list.Items {
		lastSeenAt := eventLastSeenAt(event)
		if lastSeenAt.IsZero() {
			lastSeenAt = c.now().UTC()
		}
		events = append(events, Event{
			Ref: ResourceRef{
				APIVersion: event.InvolvedObject.APIVersion,
				Kind:       event.InvolvedObject.Kind,
				Namespace:  event.InvolvedObject.Namespace,
				Name:       event.InvolvedObject.Name,
			},
			Type:       eventType(event.Type),
			Reason:     event.Reason,
			LastSeenAt: lastSeenAt.UTC().Format(time.RFC3339),
		})
	}

	sort.Slice(events, func(i, j int) bool {
		if events[i].LastSeenAt != events[j].LastSeenAt {
			return events[i].LastSeenAt < events[j].LastSeenAt
		}
		if events[i].Ref.Namespace != events[j].Ref.Namespace {
			return events[i].Ref.Namespace < events[j].Ref.Namespace
		}
		if events[i].Ref.Kind != events[j].Ref.Kind {
			return events[i].Ref.Kind < events[j].Ref.Kind
		}
		if events[i].Ref.Name != events[j].Ref.Name {
			return events[i].Ref.Name < events[j].Ref.Name
		}
		if events[i].Reason != events[j].Reason {
			return events[i].Reason < events[j].Reason
		}
		return events[i].Type < events[j].Type
	})
	return events, nil
}

func eventType(value string) string {
	switch value {
	case corev1.EventTypeNormal:
		return "NORMAL"
	case corev1.EventTypeWarning:
		return "WARNING"
	default:
		return "UNKNOWN"
	}
}

func eventLastSeenAt(event corev1.Event) time.Time {
	if !event.LastTimestamp.IsZero() {
		return event.LastTimestamp.Time
	}
	if !event.EventTime.IsZero() {
		return event.EventTime.Time
	}
	if !event.FirstTimestamp.IsZero() {
		return event.FirstTimestamp.Time
	}
	return time.Time{}
}
