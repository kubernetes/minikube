package doctor

import (
	"os/exec"
)

func MinikubeBinaryCheck() Result {

	_, err := exec.LookPath("minikube")

	if err != nil {

		return Result{
			Name:    "Minikube Binary",
			Status:  "FAIL",
			Message: "Minikube binary not found",
		}

	}

	return Result{
		Name:    "Minikube Binary",
		Status:  "PASS",
		Message: "Minikube binary found",
	}
}
