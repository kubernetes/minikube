/*
Copyright 2018 The Kubernetes Authors All rights reserved.

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

package tunnel

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

// There is one tunnel registry per user, shared across multiple vms.
// It can register, list and check for existing and running tunnels
type ID struct {
	//Route is the key
	Route *Route
	//the rest is metadata
	MachineName string
	Pid         int
}

func (t *ID) Equal(other *ID) bool {
	return t.Route.Equal(other.Route) &&
		t.MachineName == other.MachineName &&
		t.Pid == other.Pid
}

func (t *ID) String() string {
	return fmt.Sprintf("ID { Route: %v, machineName: %s, Pid: %d }", t.Route, t.MachineName, t.Pid)
}

type persistentRegistry struct {
	fileName string
}

func (r *persistentRegistry) IsAlreadyDefinedAndRunning(tunnel *ID) (*ID, error) {
	tunnels, e := r.List()
	if e != nil {
		return nil, fmt.Errorf("failed to list: %s", e)
	}

	for _, t := range tunnels {
		if t.Route.Equal(tunnel.Route) {
			isRunning, e := checkIfRunning(t.Pid)
			if e != nil {
				return nil, fmt.Errorf("error checking whether conflicting tunnel (%v) is running: %s", t, e)
			}
			if isRunning {
				return t, nil
			}
		}
	}
	return nil, nil
}

func (r *persistentRegistry) Register(tunnel *ID) error {
	glog.V(3).Infof("registering tunnel: %s", tunnel)
	if tunnel.Route == nil {
		return errors.New("tunnel.Route should not be nil")
	}

	tunnels, e := r.List()
	if e != nil {
		return fmt.Errorf("failed to list: %s", e)
	}

	alreadyExists := false
	for i, t := range tunnels {
		if t.Route.Equal(tunnel.Route) {
			isRunning, e := checkIfRunning(t.Pid)
			if e != nil {
				return fmt.Errorf("error checking whether conflicting tunnel (%v) is running: %s", t, e)
			}
			if isRunning {
				return errorTunnelAlreadyExists(t)
			}
			tunnels[i] = tunnel
			alreadyExists = true
		}
	}

	if !alreadyExists {
		tunnels = append(tunnels, tunnel)
	}

	bytes, e := json.Marshal(tunnels)
	if e != nil {
		return fmt.Errorf("error marshalling json %s", e)
	}

	glog.V(5).Infof("json marshalled: %v, %s\n", tunnels, bytes)

	f, e := os.OpenFile(r.fileName, os.O_RDWR|os.O_TRUNC, 0666)
	if e != nil {
		if os.IsNotExist(e) {
			f, e = os.Create(r.fileName)
			if e != nil {
				return fmt.Errorf("error creating registry file (%s): %s", r.fileName, e)
			}
		} else {
			return e
		}
	}
	defer f.Close()

	n, err := f.Write(bytes)
	if n < len(bytes) || err != nil {
		return fmt.Errorf("error registering tunnel while writing tunnels file: %s", err)
	}

	return nil
}

func (r *persistentRegistry) Remove(route *Route) error {
	glog.V(3).Infof("removing tunnel from registry: %s", route)
	tunnels, e := r.List()
	if e != nil {
		return e
	}
	idx := -1
	for i := range tunnels {
		if tunnels[i].Route.Equal(route) {
			idx = i
			break
		}
	}
	if idx == -1 {
		return fmt.Errorf("can't remove route: %s not found in tunnel registry", route)
	}
	tunnels = append(tunnels[:idx], tunnels[idx+1:]...)
	glog.V(4).Infof("tunnels after remove: %s", tunnels)
	f, e := os.OpenFile(r.fileName, os.O_RDWR|os.O_TRUNC, 0666)
	if e != nil {
		return fmt.Errorf("error removing tunnel %s", e)
	}
	defer f.Close()

	var bytes []byte
	bytes, e = json.Marshal(tunnels)
	if e != nil {
		return fmt.Errorf("error removing tunnel %s", e)
	}
	n, e := f.Write(bytes)
	if n < len(bytes) || e != nil {
		return fmt.Errorf("error removing tunnel %s", e)
	}

	return nil
}
func (r *persistentRegistry) List() ([]*ID, error) {
	f, e := os.Open(r.fileName)
	if e != nil {
		if !os.IsNotExist(e) {
			return nil, e
		}
		return []*ID{}, nil
	}
	byteValue, _ := ioutil.ReadAll(f)
	var tunnels []*ID
	if len(byteValue) == 0 {
		return tunnels, nil
	}
	if e = json.Unmarshal(byteValue, &tunnels); e != nil {
		return nil, e
	}

	return tunnels, nil
}
