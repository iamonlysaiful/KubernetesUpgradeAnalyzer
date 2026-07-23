package preflight

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

type LiveRunner struct{}

func (LiveRunner) Run(options KubeconfigOptions) (Result, error) {
	config, source, err := loadConfig(options.Path)
	if err != nil {
		return Result{}, err
	}

	selection, err := selectContext(config, source, options.Context)
	if err != nil {
		return Result{}, err
	}

	restConfig, err := restConfigForContext(config, options.Context)
	if err != nil {
		return Result{}, err
	}

	client, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return Result{}, fmt.Errorf("create kubernetes client: %w", err)
	}

	return Runner{
		Resolver: staticResolver{selection: selection},
		Checker:  NewKubernetesChecker(client),
	}.Run(options)
}

type staticResolver struct {
	selection ContextSelection
}

func (r staticResolver) Resolve(KubeconfigOptions) (ContextSelection, error) {
	return r.selection, nil
}

func restConfigForContext(config *api.Config, contextName string) (*rest.Config, error) {
	overrides := &clientcmd.ConfigOverrides{}
	if contextName != "" {
		overrides.CurrentContext = contextName
	}
	clientConfig := clientcmd.NewDefaultClientConfig(*config, overrides)
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("build kubernetes rest config: %w", err)
	}
	return restConfig, nil
}
