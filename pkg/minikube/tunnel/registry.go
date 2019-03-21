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

// ID represents a registry ID
type ID struct {
	//Route is the key
	Route *Route
	//the rest is metadata
	MachineName string
	Pid         int
}

// Equal checks if two ID are equal
func (t *ID) Equal(other *ID) bool {
	return t.Route.Equal(other.Route) &&
		t.MachineName == other.MachineName &&
		t.Pid == other.Pid
}

func (t *ID) String() string {
	return fmt.Sprintf("ID { Route: %v, machineName: %s, Pid: %d }", t.Route, t.MachineName, t.Pid)
}

type persistentRegistry struct {
	path string
}

func (r *persistentRegistry) IsAlreadyDefinedAndRunning(tunnel *ID) (*ID, error) {
	tunnels, err := r.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list: %s", err)
	}

	for _, t := range tunnels {
		if t.Route.Equal(tunnel.Route) {
			isRunning, err := checkIfRunning(t.Pid)
			if err != nil {
				return nil, fmt.Errorf("error checking whether conflicting tunnel (%v) is running: %s", t, err)
			}
			if isRunning {
				return t, nil
			}
		}
	}
	return nil, nil
}

func (r *persistentRegistry) Register(tunnel *ID) (rerr error) {
	glog.V(3).Infof("registering tunnel: %s", tunnel)
	if tunnel.Route == nil {
		return errors.New("tunnel.Route should not be nil")
	}

	tunnels, err := r.List()
	if err != nil {
		return fmt.Errorf("failed to list: %s", err)
	}

	alreadyExists := false
	for i, t := range tunnels {
		if t.Route.Equal(tunnel.Route) {
			isRunning, err := checkIfRunning(t.Pid)
			if err != nil {
				return fmt.Errorf("error checking whether conflicting tunnel (%v) is running: %s", t, err)
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

	bytes, err := json.Marshal(tunnels)
	if err != nil {
		return fmt.Errorf("error marshalling json %s", err)
	}

	glog.V(5).Infof("json marshalled: %v, %s\n", tunnels, bytes)

	f, err := os.OpenFile(r.path, os.O_RDWR|os.O_TRUNC, 0600)
	if err != nil {
		if os.IsNotExist(err) {
			f, err = os.Create(r.path)
			if err != nil {
				return fmt.Errorf("error creating registry file (%s): %s", r.path, err)
			}
		} else {
			return err
		}
	}
	defer func() {
		err := f.Close()
		if err != nil {
			rerr = fmt.Errorf("error closing registry file: %s", err)
		}
	}()

	n, err := f.Write(bytes)
	if n < len(bytes) || err != nil {
		return fmt.Errorf("error registering tunnel while writing tunnels file: %s", err)
	}

	return nil
}

func (r *persistentRegistry) Remove(route *Route) (rerr error) {
	glog.V(3).Infof("removing tunnel from registry: %s", route)
	tunnels, err := r.List()
	if err != nil {
		return err
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
	f, err := os.OpenFile(r.path, os.O_RDWR|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("error removing tunnel %s", err)
	}
	defer func() {
		err := f.Close()
		if err != nil {
			rerr = fmt.Errorf("error closing tunnel registry file: %s", err)
		}
	}()

	var bytes []byte
	bytes, err = json.Marshal(tunnels)
	if err != nil {
		return fmt.Errorf("error removing tunnel %s", err)
	}
	n, err := f.Write(bytes)
	if n < len(bytes) || err != nil {
		return fmt.Errorf("error removing tunnel %s", err)
	}

	return nil
}
func (r *persistentRegistry) List() ([]*ID, error) {
	f, err := os.Open(r.path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		return []*ID{}, nil
	}
	byteValue, _ := ioutil.ReadAll(f)
	var tunnels []*ID
	if len(byteValue) == 0 {
		return tunnels, nil
	}
	if err = json.Unmarshal(byteValue, &tunnels); err != nil {
		return nil, err
	}

	return tunnels, nil
}
