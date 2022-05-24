package integration

import (
	"context"
	"encoding/json"
	"os/exec"
	"testing"

	"k8s.io/minikube/pkg/minikube/config"
)

// profileJSON is the output of `minikube profile list -ojson`
type profileJSON struct {
	Valid   []config.Profile `json:"valid"`
	Invalid []config.Profile `json:"invalid"`
}

func TestMinikubeProfile(t *testing.T) {
	// 1. Setup two minikube cluster profiles
	profiles := []string{UniqueProfileName("first"), UniqueProfileName("second")}
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(40))
	for _, p := range profiles {
		defer CleanupWithLogs(t, p, cancel)
		_, err := Run(t, exec.CommandContext(ctx, Target(), "start", "-p", p))
		if err != nil {
			t.Errorf("test pre-condition failed: minikube start -p " + p + " failed: " + err.Error())
			return
		}
	}
	// 2. Change minikube profile
	for _, p := range profiles {
		_, err := Run(t, exec.CommandContext(ctx, Target(), "profile", p))
		if err != nil {
			t.Errorf("minikube profile %s failed with error: %v\n", p, err.Error())
			return
		}
		r, err := Run(t, exec.CommandContext(ctx, Target(), "profile", "list", "-ojson"))
		var profile profileJSON
		err = json.NewDecoder(r.Stdout).Decode(&profile)
		if err != nil {
			t.Fatalf("minikube profile list -ojson failed with error: %v\n", err.Error())
			return
		}
		// 3. Assert minikube profile is set to the correct profile in JSON
		for _, s := range profile.Valid {
			if s.Name == p && !s.Active {
				t.Fatalf("minikube profile %s is not active\n", p)
				return
			}
		}
	}
}
