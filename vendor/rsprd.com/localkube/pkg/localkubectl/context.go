package localkubectl

import (
	"fmt"

	kubecfg "k8s.io/kubernetes/pkg/client/unversioned/clientcmd/api"
	kubectlcfg "k8s.io/kubernetes/pkg/kubectl/cmd/config"
)

// getConfig gets the kubectl client configuration using defaults
func getConfig(pathOpts *kubectlcfg.PathOptions) (*kubecfg.Config, error) {
	config, err := pathOpts.GetStartingConfig()
	if err != nil {
		return nil, fmt.Errorf("could not get config: %v", err)
	}

	return config, nil
}

// SetupContext creates a new cluster and context in ~/.kube using the provided API Server. If setCurrent is true, it is made the current context.
func SetupContext(clusterName, contextName, kubeAPIServer string, setCurrent bool) error {
	pathOpts := kubectlcfg.NewDefaultPathOptions()
	config, err := getConfig(pathOpts)
	if err != nil {
		return err
	}

	cluster, exists := config.Clusters[clusterName]
	if !exists {
		cluster = kubecfg.NewCluster()
	}

	// configure cluster
	cluster.Server = kubeAPIServer
	cluster.InsecureSkipTLSVerify = true
	config.Clusters[clusterName] = cluster

	context, exists := config.Contexts[contextName]
	if !exists {
		context = kubecfg.NewContext()
	}

	// configure context
	context.Cluster = clusterName
	config.Contexts[contextName] = context

	// set as current if requested
	if setCurrent {
		config.CurrentContext = contextName
	}

	return kubectlcfg.ModifyConfig(pathOpts, *config, true)
}

// GetCurrentContext returns the context currently being used by kubectl
func GetCurrentContext() (string, error) {
	pathOpts := kubectlcfg.NewDefaultPathOptions()
	config, err := getConfig(pathOpts)
	if err != nil {
		return "", err
	}

	return config.CurrentContext, nil
}

// SetCurrentContext changes the kubectl context to the given value
func SetCurrentContext(context string) error {
	pathOpts := kubectlcfg.NewDefaultPathOptions()
	config, err := getConfig(pathOpts)
	if err != nil {
		return err
	}

	config.CurrentContext = context
	return kubectlcfg.ModifyConfig(pathOpts, *config, true)
}

// SwitchContextInstructions returns instructions about how to switch kubectl to use the given context
func SwitchContextInstructions(contextName string) string {
	l1 := "To setup kubectl to use localkube run:\n"
	l2 := fmt.Sprintf("kubectl config use-context %s\n", contextName)
	return l1 + l2
}
