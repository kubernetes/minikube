package doctor

import (
	"fmt"

	"k8s.io/minikube/pkg/minikube/config"
)

func ResourceValidationCheck(cc *config.ClusterConfig) []Result {
	var results []Result

	// CPU Check
	cpu := cc.CPUs
	if cpu < 2 {
		results = append(results, Result{
			Name:           "CPU Allocation",
			Status:         "FAIL",
			Message:        fmt.Sprintf("%d CPUs allocated (min 2 required)", cpu),
			Recommendation: "Increase allocated CPUs by running: minikube config set cpus 2",
		})
	} else if cpu < 4 {
		results = append(results, Result{
			Name:           "CPU Allocation",
			Status:         "WARNING",
			Message:        fmt.Sprintf("%d CPUs allocated (4 recommended)", cpu),
			Recommendation: "Increase allocated CPUs by running: minikube config set cpus 4",
		})
	} else {
		results = append(results, Result{
			Name:    "CPU Allocation",
			Status:  "PASS",
			Message: fmt.Sprintf("%d CPUs", cpu),
		})
	}

	// Memory Check
	memory := cc.Memory
	if memory < 2048 {
		results = append(results, Result{
			Name:           "Memory Allocation",
			Status:         "FAIL",
			Message:        fmt.Sprintf("%d MB allocated (min 2048 MB required)", memory),
			Recommendation: "Increase allocated memory by running: minikube config set memory 2048",
		})
	} else if memory < 4096 {
		results = append(results, Result{
			Name:           "Memory Allocation",
			Status:         "WARNING",
			Message:        fmt.Sprintf("%d MB allocated (4096 MB recommended)", memory),
			Recommendation: "Increase allocated memory by running: minikube config set memory 4096",
		})
	} else {
		results = append(results, Result{
			Name:    "Memory Allocation",
			Status:  "PASS",
			Message: fmt.Sprintf("%d MB", memory),
		})
	}

	// Disk Size Check
	disk := cc.DiskSize
	if disk < 20000 {
		results = append(results, Result{
			Name:           "Disk Allocation",
			Status:         "FAIL",
			Message:        fmt.Sprintf("%d MB allocated (min 20000 MB recommended)", disk),
			Recommendation: "Increase allocated disk space by running: minikube config set disk-size 20000",
		})
	} else {
		results = append(results, Result{
			Name:    "Disk Allocation",
			Status:  "PASS",
			Message: fmt.Sprintf("%d MB", disk),
		})
	}

	return results
}
