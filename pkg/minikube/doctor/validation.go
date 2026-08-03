package doctor

import (
	"fmt"

	"k8s.io/minikube/pkg/minikube/config"
)

func ProfileValidationCheck() Result {
	_, invalidProfiles, err := config.ListProfiles()
	if err != nil {
		return Result{
			Name:    "Config Validation",
			Status:  "FAIL",
			Message: err.Error(),
		}
	}

	if len(invalidProfiles) > 0 {
		var names []string
		for _, p := range invalidProfiles {
			names = append(names, p.Name)
		}
		return Result{
			Name:           "Config Validation",
			Status:         "FAIL",
			Message:        fmt.Sprintf("Found %d corrupted profiles: %v", len(invalidProfiles), names),
			Recommendation: "Remove corrupted profiles by running: minikube delete -p <profile>",
		}
	}

	return Result{
		Name:    "Config Validation",
		Status:  "PASS",
		Message: "All profile configurations are valid",
	}
}
