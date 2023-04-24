package provision

import (
	"k8s.io/minikube/pkg/libmachine/auth"
	"k8s.io/minikube/pkg/libmachine/cruntime"
)

type CRuntimeConfigContext struct {
	Port          int
	AuthOptions   auth.Options
	EngineOptions cruntime.Options
	OptionsDir    string
}
