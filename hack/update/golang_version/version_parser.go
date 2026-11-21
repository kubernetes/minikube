package main

import (
	"fmt"
	"strings"
)

func parseGoBuildImageVersion(raw string) (stable, k8sVersion string, err error) {
	// Example: v1.23.0-go1.17-buster.0.
	version := strings.TrimSpace(raw)
	k8sVersion, rest, ok := strings.Cut(version, "-")
	if !ok || k8sVersion == "" {
		return "", "", fmt.Errorf("invalid Go build image version %q", version)
	}
	goPart, _, _ := strings.Cut(rest, "-")
	if !strings.HasPrefix(goPart, "go") || len(goPart) <= len("go") {
		return "", "", fmt.Errorf("invalid Go build image version %q", version)
	}
	return strings.TrimPrefix(goPart, "go"), k8sVersion, nil
}
