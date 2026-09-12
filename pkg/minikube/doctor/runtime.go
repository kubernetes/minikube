package doctor

import (
	"k8s.io/minikube/pkg/minikube/config"
)

func ContainerRuntimeCheck() Result {
	profiles, err := config.ListValidProfiles()
	if err != nil {
		return Result{
			Name:    "Runtime",
			Status:  "FAIL",
			Message: err.Error(),
		}
	}

	if len(profiles) == 0 {
		return Result{
			Name:    "Runtime",
			Status:  "FAIL",
			Message: "No profiles found",
		}
	}

	// Use active or first profile
	activeProfile := profiles[0]
	for _, p := range profiles {
		if p.Active {
			activeProfile = p
			break
		}
	}

	if activeProfile.Config == nil {
		return Result{
			Name:    "Runtime",
			Status:  "FAIL",
			Message: "Profile configuration is nil",
		}
	}

	runtime := activeProfile.Config.KubernetesConfig.ContainerRuntime
	if runtime == "" {
		runtime = "docker" // Default container runtime in older versions if empty
	}

	return Result{
		Name:    "Runtime",
		Status:  "PASS",
		Message: runtime,
	}
}
