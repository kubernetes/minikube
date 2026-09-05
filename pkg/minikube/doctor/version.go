package doctor

import (
	"k8s.io/minikube/pkg/minikube/config"
)

func KubernetesVersionCheck() Result {
	profiles, err := config.ListValidProfiles()
	if err != nil {
		return Result{
			Name:    "Kubernetes",
			Status:  "FAIL",
			Message: err.Error(),
		}
	}

	if len(profiles) == 0 {
		return Result{
			Name:    "Kubernetes",
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
			Name:    "Kubernetes",
			Status:  "FAIL",
			Message: "Profile configuration is nil",
		}
	}

	version := activeProfile.Config.KubernetesConfig.KubernetesVersion
	if version == "" {
		return Result{
			Name:    "Kubernetes",
			Status:  "FAIL",
			Message: "Kubernetes version not configured",
		}
	}

	return Result{
		Name:    "Kubernetes",
		Status:  "PASS",
		Message: version,
	}
}
