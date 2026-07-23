package inventory

import (
	"context"
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/kube/preflight"
)

type LiveCollector struct{}

func (LiveCollector) CollectCore(options preflight.KubeconfigOptions, result preflight.Result) (Snapshot, error) {
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
		return Snapshot{}, fmt.Errorf("build kubernetes rest config: %w", err)
	}

	client, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return Snapshot{}, fmt.Errorf("create kubernetes client: %w", err)
	}

	return NewCollector(client).CollectCore(context.Background(), result)
}
