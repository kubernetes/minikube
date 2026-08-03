package doctor

import (
	"os/exec"
	"strings"
)

func DriverBinaryCheck(driver string) Result {
	driverName := strings.ToLower(driver)
	if driverName == "" {
		return Result{
			Name:    "Driver Binary",
			Status:  "FAIL",
			Message: "No driver configured",
		}
	}

	// Map driver alias to binary name if necessary
	binaryName := driverName
	if driverName == "virtualbox" {
		binaryName = "vboxmanage"
	}

	_, err := exec.LookPath(binaryName)
	if err != nil {
		return Result{
			Name:           "Driver Binary",
			Status:         "FAIL",
			Message:        binaryName + " is not installed",
			Recommendation: "Install " + driver + " on your host machine",
		}
	}

	return Result{
		Name:    "Driver Binary",
		Status:  "PASS",
		Message: binaryName + " is installed",
	}
}

func DriverDaemonCheck(driver string) Result {
	driverName := strings.ToLower(driver)
	if driverName == "docker" {
		cmd := exec.Command("docker", "info")
		err := cmd.Run()
		if err != nil {
			return Result{
				Name:           "Driver Daemon",
				Status:         "FAIL",
				Message:        "Docker daemon is not running",
				Recommendation: "Start Docker Desktop or run: systemctl start docker (on Linux)",
			}
		}
		return Result{
			Name:    "Driver Daemon",
			Status:  "PASS",
			Message: "Docker daemon is running",
		}
	} else if driverName == "podman" {
		cmd := exec.Command("podman", "info")
		err := cmd.Run()
		if err != nil {
			return Result{
				Name:           "Driver Daemon",
				Status:         "FAIL",
				Message:        "Podman service is not running",
				Recommendation: "Start podman machine or daemon service",
			}
		}
		return Result{
			Name:    "Driver Daemon",
			Status:  "PASS",
			Message: "Podman service is running",
		}
	}

	return Result{
		Name:    "Driver Daemon",
		Status:  "PASS",
		Message: "Not required for " + driver,
	}
}
