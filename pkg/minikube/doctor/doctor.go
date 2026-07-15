package doctor

import (
	"k8s.io/minikube/pkg/libmachine"
	"k8s.io/minikube/pkg/minikube/config"
)

func Run(api libmachine.API, cc *config.ClusterConfig) []Result {
	var results []Result

	// Configuration Checks
	results = append(results, ProfileCheck())
	results = append(results, DriverCheck())
	results = append(results, KubernetesVersionCheck())
	results = append(results, ContainerRuntimeCheck())

	// Environment Checks
	results = append(results, DriverBinaryCheck(cc.Driver))
	results = append(results, KubectlCheck())
	results = append(results, DriverDaemonCheck(cc.Driver))

	// Cluster Health Checks (Phase 2)
	results = append(results, ClusterCheck(api, cc))
	results = append(results, APIServerCheck(api, cc))
	results = append(results, NodesCheck(api, cc))
	results = append(results, KubeconfigCheck(api, cc))

	// Configuration Validation Checks (Phase 5)
	results = append(results, ProfileValidationCheck())

	// Resource Validation Checks (Phase 6)
	results = append(results, ResourceValidationCheck(cc)...)

	return results
}
