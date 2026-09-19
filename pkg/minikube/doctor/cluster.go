package doctor

import (
	"fmt"

	"k8s.io/minikube/pkg/libmachine"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/config"
)

func ClusterCheck(api libmachine.API, cc *config.ClusterConfig) Result {
	statuses, err := cluster.GetStatus(api, cc)
	if err != nil {
		return Result{
			Name:           "Cluster Status",
			Status:         "FAIL",
			Message:        "Cluster is not running",
			Details:        err.Error(),
			Recommendation: "Start the cluster by running: minikube start",
		}
	}
	if len(statuses) == 0 {
		return Result{
			Name:           "Cluster Status",
			Status:         "FAIL",
			Message:        "No active nodes found in profile configuration",
			Recommendation: "Start the cluster by running: minikube start",
		}
	}
	hostStatus := statuses[0].Host
	if hostStatus == "Running" {
		return Result{
			Name:    "Cluster Status",
			Status:  "PASS",
			Message: "Running",
		}
	}
	return Result{
		Name:           "Cluster Status",
		Status:         "FAIL",
		Message:        "Cluster is in " + hostStatus + " state",
		Recommendation: "Start the cluster by running: minikube start",
	}
}

func APIServerCheck(api libmachine.API, cc *config.ClusterConfig) Result {
	statuses, err := cluster.GetStatus(api, cc)
	if err != nil {
		return Result{
			Name:           "API Server",
			Status:         "FAIL",
			Message:        "API Server is unreachable",
			Details:        err.Error(),
			Recommendation: "Ensure the cluster is running: minikube status",
		}
	}
	if len(statuses) == 0 {
		return Result{
			Name:           "API Server",
			Status:         "FAIL",
			Message:        "No active control plane node found",
			Recommendation: "Start the cluster by running: minikube start",
		}
	}
	apiServerStatus := statuses[0].APIServer
	if apiServerStatus == "Running" {
		return Result{
			Name:    "API Server",
			Status:  "PASS",
			Message: "Running",
		}
	}
	return Result{
		Name:           "API Server",
		Status:         "FAIL",
		Message:        "API Server is in " + apiServerStatus + " state",
		Recommendation: "Restart minikube by running: minikube start",
	}
}

func NodesCheck(api libmachine.API, cc *config.ClusterConfig) Result {
	statuses, err := cluster.GetStatus(api, cc)
	if err != nil {
		return Result{
			Name:           "Nodes Ready",
			Status:         "FAIL",
			Message:        "Unable to retrieve node status",
			Details:        err.Error(),
			Recommendation: "Check node status by running: kubectl get nodes",
		}
	}
	allReady := true
	var notReadyNodes []string
	for _, st := range statuses {
		if st.Kubelet != "Running" {
			allReady = false
			notReadyNodes = append(notReadyNodes, st.Name)
		}
	}
	if allReady {
		return Result{
			Name:    "Nodes Ready",
			Status:  "PASS",
			Message: "All nodes ready",
		}
	}
	return Result{
		Name:           "Nodes Ready",
		Status:         "FAIL",
		Message:        fmt.Sprintf("Nodes not ready: %v", notReadyNodes),
		Recommendation: "Ensure kubelet service is active on the node(s)",
	}
}

func KubeconfigCheck(api libmachine.API, cc *config.ClusterConfig) Result {
	statuses, err := cluster.GetStatus(api, cc)
	if err != nil {
		return Result{
			Name:           "Kubeconfig",
			Status:         "FAIL",
			Message:        "Unable to verify kubeconfig status",
			Details:        err.Error(),
			Recommendation: "Recreate the config by running: minikube update-context",
		}
	}
	if len(statuses) == 0 {
		return Result{
			Name:           "Kubeconfig",
			Status:         "FAIL",
			Message:        "No cluster nodes configured",
			Recommendation: "Recreate the config by running: minikube update-context",
		}
	}
	kubeconfigStatus := statuses[0].Kubeconfig
	if kubeconfigStatus == "Configured" {
		return Result{
			Name:    "Kubeconfig",
			Status:  "PASS",
			Message: "Connected to cluster",
		}
	}
	return Result{
		Name:           "Kubeconfig",
		Status:         "FAIL",
		Message:        "Kubeconfig is in " + kubeconfigStatus + " state",
		Recommendation: "Configure kubectl context by running: minikube update-context",
	}
}
