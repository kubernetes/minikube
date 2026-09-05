package doctor

import (
	"os/exec"
)

func KubectlCheck() Result {
	_, err := exec.LookPath("kubectl")
	if err != nil {
		return Result{
			Name:           "kubectl",
			Status:         "FAIL",
			Message:        "kubectl is not installed",
			Recommendation: "Install kubectl: https://kubernetes.io/docs/tasks/tools/",
		}
	}

	return Result{
		Name:    "kubectl",
		Status:  "PASS",
		Message: "kubectl is installed",
	}
}
