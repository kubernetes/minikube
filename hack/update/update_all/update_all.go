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
		if component == "get_version" || component == "update_all" || component == "k8s-lib" || component == "amd_device_gpu_plugin_version" {
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
