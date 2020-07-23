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

package node

import (
	"fmt"

	"github.com/docker/machine/libmachine"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/state"
)

func Status(api libmachine.API, cc config.ClusterConfig, n config.Node) (state.Collection, error) {
	name := driver.MachineName(cc, n)

	st := state.Collection{
		Component: state.Component{
			Name: name,
		},
	}

	if n.ControlPlane {
		st.Kind = "control-plane"
	} else {
		st.Kind = "worker"
	}

	hs, err := machine.Status(api, name)
	glog.Infof("%s host status = %q (err=%v)", name, hs, err)
	if err != nil {
		return st, errors.Wrap(err, "host")
	}

	st.Condition = state.LibMachineCondition(hs)
	if st.Condition == state.Nonexistent {
		st.Errors = append(st.Errors, "host does not appear to exist")
		return st, nil
	}

	// If it's not running, quickly bail out rather than delivering conflicting messages
	if st.Condition == state.Unavailable || st.Condition == state.Error {
		glog.Warningf("%s is unavailable, skipping remaining checks", n.Name)
		return st, nil
	}

	host, err := machine.LoadHost(api, name)
	if err != nil {
		st.Condition = state.Error
		st.Errors = append(st.Errors, fmt.Sprintf("unable to load host: %v", err))
	}

	_, err = machine.CommandRunner(host)
	if err != nil {
		st.Condition = state.Error
		st.Errors = append(st.Errors, fmt.Sprintf("unable to get command-runner: %v", err))
	}
	/*
		st.Components["kubelet"] = kverify.KubeletStatus(cr)

		apiserver, err := kverify.APIServerStatus(cr, hostname, port)
		st.Components["apiserver"] = apiserver
	*/
	return st, nil
}
