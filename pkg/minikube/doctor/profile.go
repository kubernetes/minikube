package doctor

import (
	"k8s.io/minikube/pkg/minikube/config"
)

func ProfileCheck() Result {
	profiles, err := config.ListValidProfiles()
	if err != nil {
		return Result{
			Name:    "Profile",
			Status:  "FAIL",
			Message: err.Error(),
		}
	}

	if len(profiles) == 0 {
		return Result{
			Name:    "Profile",
			Status:  "FAIL",
			Message: "No profiles found",
		}
	}

	// Find active profile
	activeProfile := profiles[0].Name
	for _, p := range profiles {
		if p.Active {
			activeProfile = p.Name
			break
		}
	}

	return Result{
		Name:    "Profile",
		Status:  "PASS",
		Message: activeProfile,
	}
}
