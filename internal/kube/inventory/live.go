package inventory

import (
	"context"
	"fmt"

	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/kube/preflight"
)

type LiveCollector struct{}

func (LiveCollector) CollectCore(options preflight.KubeconfigOptions, result preflight.Result) (Snapshot, error) {
	collector, err := liveCollector(options)
	if err != nil {
		return Snapshot{}, err
	}
	return collector.CollectCore(context.Background(), result)
}

func (LiveCollector) CollectAssessment(options preflight.KubeconfigOptions, result preflight.Result) (Snapshot, error) {
	collector, err := liveCollector(options)
	if err != nil {
		return Snapshot{}, err
	}
	return collector.CollectSnapshot(context.Background(), result, CollectionOptions{
		Workloads:  true,
		Storage:    true,
		Networking: true,
		CRDs:       true,
		Events:     true,
		Limitation: Limitation{
			Code:     "LIVE_INVENTORY_PHASE_8_5",
			Severity: "INFO",
			Summary:  "Phase 8.5 analyze collection includes live read-only workloads, storage, networking, CRDs, and events.",
		},
	})
}

func liveCollector(options preflight.KubeconfigOptions) (Collector, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if options.Path != "" {
		loadingRules.ExplicitPath = options.Path
	}

	overrides := &clientcmd.ConfigOverrides{}
	if options.Context != "" {
		overrides.CurrentContext = options.Context
	}

	restConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		overrides,
	).ClientConfig()
	if err != nil {
		return Collector{}, fmt.Errorf("build kubernetes rest config: %w", err)
	}

	client, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return Collector{}, fmt.Errorf("create kubernetes client: %w", err)
	}

	apiExtensionsClient, err := apiextensionsclient.NewForConfig(restConfig)
	if err != nil {
		return Collector{}, fmt.Errorf("create kubernetes apiextensions client: %w", err)
	}

	return NewCollectorWithAPIExtensions(client, apiExtensionsClient), nil
}
