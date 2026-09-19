package doctor

import (
	"k8s.io/minikube/pkg/minikube/config"
)

func DriverCheck() Result {

	profiles, err := config.ListValidProfiles()

	if err != nil {
		return Result{
			Name:    "Driver",
			Status:  "FAIL",
			Message: err.Error(),
		}
	}

	if len(profiles) == 0 {
		return Result{
			Name:    "Driver",
			Status:  "FAIL",
			Message: "No Minikube profiles found",
		}
	}

	driver := profiles[0].Config.Driver

	return Result{
		Name:    "Driver",
		Status:  "PASS",
		Message: driver,
	}
}
