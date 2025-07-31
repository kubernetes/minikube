package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func getVersion(component string) (string, error) {
	cmd := exec.Command("go", "run", "update/get_version/get_version.go")
	cmd.Env = append(os.Environ(), fmt.Sprintf("DEP=%s", component))
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Printf("failed to get version for %s: %v", component, err)
		log.Printf("command output: %s", out.String())
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}

func main() {
	updateDir := "update"
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	log.Println("starting the update all the current PWD is: ", path)

	dirs, err := os.ReadDir(updateDir)
	if err != nil {
		log.Fatalf("failed to read directory %s: %v", updateDir, err)
	}

	var changes []string
	i := 0
	max := 100
	for _, d := range dirs {
		i++
		if i > max {
			log.Println("debug - skipping more than 100 components")
			break
		}
		if !d.IsDir() {
			continue
		}

		component := d.Name()
		blackList := map[string]bool{
			"get_version":                   true,
			"update_all":                    true,
			"k8s-lib":                       true,
			"amd_device_gpu_plugin_version": true, // sem vers issue https://github.com/ROCm/k8s-device-plugin/issues/144
			"docsy_version":                 true, // this one does not supprt get-dependency-verison
			"istio_operator_version":        true, // till this is fixed https://github.com/istio/istio/issues/57185
			"kicbase_version":               true, // This one is not related to auto updating, this is a tool used by kicbae_auto_build
		}
		if blackList[component] {
			continue
		}

		fmt.Printf("Processing %s...\n", component)

		oldVersion, err := getVersion(component)
		if err != nil {
			log.Fatalf("Could not get old version for %s: %v", component, err)
		}

		updateCmd := exec.Command("go", "run", filepath.Join(updateDir, component, fmt.Sprintf("%s.go", component)))
		updateCmd.Stdout = os.Stdout
		updateCmd.Stderr = os.Stderr

		if err := updateCmd.Run(); err != nil {
			log.Fatalf("Failed to update %s: %v", component, err)
		}

		newVersion, err := getVersion(component)
		if err != nil {
			log.Printf("Could not get new version for %s: %v", component, err)
			continue
		}

		if oldVersion != newVersion {
			change := fmt.Sprintf("- **%s:** `%s` -> `%s`", component, oldVersion, newVersion)
			changes = append(changes, change)
			fmt.Println(change)
		} else {
			fmt.Printf("No change for %s.\n", component)
		}
		fmt.Println()
	}

	fmt.Println("---")
	fmt.Printf("Updated components summary:\n%s\n", strings.Join(changes, "\n"))

	// Print a machine-readable summary for GitHub Actions
	outputChanges := "updates<<EOF\n" + strings.Join(changes, "\n") + "\nEOF"
	fmt.Println(outputChanges)
}
