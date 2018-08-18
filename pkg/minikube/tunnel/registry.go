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
	"github.com/pkg/errors"
	"os"
	"fmt"
	"encoding/json"
	"io/ioutil"
	"golang.org/x/sys/unix"
	"github.com/sirupsen/logrus"
)

//registry should be
// - configurable in terms of directory for testing
// - one per user, across multiple vms
// should have a list of tunnels:
// tunnel: Route, Pid, machinename
// when cleanup is called, all the non running tunnels should be checked for removal
// when a new tunnel is created it should register itself with the registry Pid/machinename/Route

type TunnelID struct {
	//Route is the key
	Route *Route
	//the rest is metadata
	MachineName string
	Pid         int
}

func (t *TunnelID) Equal(other *TunnelID) bool {
	return t.Route.Equal(other.Route) &&
		t.MachineName == other.MachineName &&
		t.Pid == other.Pid
}

func (t *TunnelID) String() string {
	return fmt.Sprintf("TunnelID { Route: %v, machineName: %s, Pid: %d }", t.Route, t.MachineName, t.Pid)
}

type persistentRegistry struct {
	fileName string
}

//TODO(balopat): register should check against conflicting/overlapping routes
func (r *persistentRegistry) Register(tunnel *TunnelID) error {
	if tunnel.Route == nil {
		return errors.New("tunnel.Route should not be nil")
	}

	tunnels, e := r.List()
	if e != nil {
		return fmt.Errorf("failed to list: %s", e)
	}
	tunnels = append(tunnels, tunnel)

	bytes, e := json.Marshal(tunnels)
	if e != nil {
		return fmt.Errorf("error marshalling json %s", e)
	}

	logrus.Infof("json marshalled: %v, %s\n", tunnels, bytes)

	f, e := os.OpenFile(r.fileName, unix.O_RDWR|unix.O_TRUNC, 0666)
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
	logrus.Debugf("tunnels after remove: %s", tunnels)
	f, e := os.OpenFile(r.fileName, unix.O_RDWR|unix.O_TRUNC, 0666)
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
func (r *persistentRegistry) List() ([]*TunnelID, error) {
	f, e := os.Open(r.fileName)
	if e != nil {
		if !os.IsNotExist(e) {
			return nil, e
		}
		return []*TunnelID{}, nil
	}
	byteValue, _ := ioutil.ReadAll(f)
	var tunnels []*TunnelID
	if len(byteValue) == 0 {
		return tunnels, nil
	}
	if e = json.Unmarshal(byteValue, &tunnels); e != nil {
		return nil, e
	}

	return tunnels, nil
}
