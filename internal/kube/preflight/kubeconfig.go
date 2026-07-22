package preflight

import (
	"fmt"
	"os"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

type KubeconfigSource string

const (
	KubeconfigSourceDefault  KubeconfigSource = "DEFAULT"
	KubeconfigSourceExplicit KubeconfigSource = "EXPLICIT"
)

type ContextSelection struct {
	Name             string
	ClusterName      string
	UserName         string
	Namespace        string
	KubeconfigSource KubeconfigSource
}

type KubeconfigOptions struct {
	Path    string
	Context string
}

func ResolveContext(options KubeconfigOptions) (ContextSelection, error) {
	config, source, err := loadConfig(options.Path)
	if err != nil {
		return ContextSelection{}, err
	}
	return selectContext(config, source, options.Context)
}

func ResolveContextFromBytes(data []byte, source KubeconfigSource, contextName string) (ContextSelection, error) {
	config, err := clientcmd.Load(data)
	if err != nil {
		return ContextSelection{}, fmt.Errorf("load kubeconfig: %w", err)
	}
	return selectContext(config, source, contextName)
}

func loadConfig(path string) (*api.Config, KubeconfigSource, error) {
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, KubeconfigSourceExplicit, fmt.Errorf("read kubeconfig: %w", err)
		}
		config, err := clientcmd.Load(data)
		if err != nil {
			return nil, KubeconfigSourceExplicit, fmt.Errorf("load kubeconfig: %w", err)
		}
		return config, KubeconfigSourceExplicit, nil
	}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	config, err := loadingRules.Load()
	if err != nil {
		return nil, KubeconfigSourceDefault, fmt.Errorf("load default kubeconfig: %w", err)
	}
	return config, KubeconfigSourceDefault, nil
}

func selectContext(config *api.Config, source KubeconfigSource, contextName string) (ContextSelection, error) {
	if config == nil {
		return ContextSelection{}, fmt.Errorf("kubeconfig is empty")
	}

	selected := contextName
	if selected == "" {
		selected = config.CurrentContext
	}
	if selected == "" {
		return ContextSelection{}, fmt.Errorf("kubeconfig has no current context")
	}

	context, ok := config.Contexts[selected]
	if !ok || context == nil {
		return ContextSelection{}, fmt.Errorf("kubeconfig context %q not found", selected)
	}
	if context.Cluster == "" {
		return ContextSelection{}, fmt.Errorf("kubeconfig context %q has no cluster", selected)
	}
	if _, ok := config.Clusters[context.Cluster]; !ok {
		return ContextSelection{}, fmt.Errorf("kubeconfig cluster %q not found", context.Cluster)
	}

	return ContextSelection{
		Name:             selected,
		ClusterName:      context.Cluster,
		UserName:         context.AuthInfo,
		Namespace:        context.Namespace,
		KubeconfigSource: source,
	}, nil
}
