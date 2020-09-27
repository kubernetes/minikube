/*
Copyright 2020 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package machine

import (
	"runtime"

	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/style"
)

// MaybeDisplayAdvice will provide advice without exiting, so minikube has a chance to try the failover
func MaybeDisplayAdvice(err error, driver string) {
	if errors.Is(err, oci.ErrDaemonInfo) {
		out.ErrLn("")
		out.ErrT(style.Conflict, "{{.driver_name}} couldn't proceed because {{.driver_name}} service is not healthy.", out.V{"driver_name": driver})
	}

	if errors.Is(err, oci.ErrExitedUnexpectedly) {
		out.ErrLn("")
		out.ErrT(style.Conflict, "The minikube {{.driver_name}} container exited unexpectedly.", out.V{"driver_name": driver})
	}

	if errors.Is(err, oci.ErrExitedUnexpectedly) || errors.Is(err, oci.ErrDaemonInfo) {
		out.T(style.Tip, "If you are still interested to make {{.driver_name}} driver work. The following suggestions might help you get passed this issue:", out.V{"driver_name": driver})
		if driver == oci.Docker || driver == oci.Podman {
			out.String("\n\t")
			out.T(style.Empty, `- Prune unused {{.driver_name}} images, volumes, networks and abandoned containers.

				{{.driver_name}} system prune --volumes`, out.V{"driver_name": driver})
		}
		out.String("\n\t")
		out.T(style.Empty, `- Restart your {{.driver_name}} service`, out.V{"driver_name": driver})
		if runtime.GOOS != "linux" {
			out.String("\n\t")
			out.T(style.Empty, `- Ensure your {{.driver_name}} daemon has access to enough CPU/memory resources.`, out.V{"driver_name": driver})
			if runtime.GOOS == "darwin" && driver == oci.Docker {
				out.String("\n\t")
				out.T(style.Empty, `- Docs https://docs.docker.com/docker-for-mac/#resources`, out.V{"driver_name": driver})
			}
			if runtime.GOOS == "windows" && driver == oci.Docker {
				out.String("\n\t")
				out.T(style.Empty, `- Docs https://docs.docker.com/docker-for-windows/#resources`, out.V{"driver_name": driver})
			}
		}
		out.String("\n\t")
		out.T(style.Empty, `- Delete and recreate minikube cluster
		minikube delete
		minikube start --driver={{.driver_name}}`, out.V{"driver_name": driver})
		// TODO #8348: maybe advice user if to set the --force-systemd https://github.com/kubernetes/minikube/issues/8348
	}
}
