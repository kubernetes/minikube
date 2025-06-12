package environment

import (
	"io"
	"k8s.io/minikube/pkg/minikube/shell"
)

// EnvConfigurator is the interface that any environment configuration logic must implement.
type EnvConfigurator interface {
	// Vars returns a map of environment variables that need to be set.
	Vars() (map[string]string, error)

	// UnsetVars returns a list of environment variable names that need to be unset.
	UnsetVars() ([]string, error)

	// DisplayScript generates a script that needs to be displayed and is used to set environment variables.
	DisplayScript(sh shell.Config, w io.Writer) error
}