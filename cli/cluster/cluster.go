/*
Copyright 2015 The Kubernetes Authors All rights reserved.
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

package cluster

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/machine/drivers/virtualbox"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/host"
	"github.com/kubernetes/minikube/cli/constants"
	"rsprd.com/localkube/pkg/localkubectl"
)

// StartHost starts a host VM.
func StartHost(api libmachine.API) (*host.Host, error) {
	setupDirs()

	if exists, err := api.Exists(constants.MachineName); err != nil {
		return nil, fmt.Errorf("Error checking if host exists: %s", err)
	} else if exists {
		log.Println("Machine exists!")
		h, err := api.Load(constants.MachineName)
		if err != nil {
			return nil, fmt.Errorf("Error loading existing host.")
		}
		return h, nil
	} else {
		return createHost(api)
	}
}

// StartCluster starts as k8s cluster on the specified Host.
func StartCluster(h *host.Host) (string, error) {
	host, err := h.Driver.GetURL()
	if err != nil {
		return "", err
	}
	kubeHost := strings.Replace(host, "tcp://", "http://", -1)
	kubeHost = strings.Replace(kubeHost, ":2376", ":8080", -1)

	os.Setenv("DOCKER_HOST", host)
	os.Setenv("DOCKER_CERT_PATH", constants.MakeMiniPath("certs"))
	os.Setenv("DOCKER_TLS_VERIFY", "1")
	ctlr, err := localkubectl.NewControllerFromEnv(os.Stdout)
	if err != nil {
		log.Panicf("Error creating controller: %s", err)
	}

	// Look for an existing container
	ctrID, running, err := ctlr.OnlyLocalkubeCtr()
	if running {
		log.Println("Localkube is already running")
		return kubeHost, nil
	}
	if err == localkubectl.ErrNoContainer {
		// If container doesn't exist, create
		ctrID, running, err = ctlr.CreateCtr(localkubectl.LocalkubeContainerName, "latest")
		if err != nil {
			return "", err
		}
		return kubeHost, nil
	}
	// Start container.
	err = ctlr.StartCtr(ctrID, "")
	if err != nil {
		return "", err
	}
	return kubeHost, nil
}

func createHost(api libmachine.API) (*host.Host, error) {
	driver := virtualbox.NewDriver(constants.MachineName, constants.Minipath)
	driver.Boot2DockerURL = "https://storage.googleapis.com/tinykube/boot2docker.iso"
	data, err := json.Marshal(driver)
	if err != nil {
		return nil, err
	}

	driverName := "virtualbox"
	h, err := api.NewHost(driverName, data)
	if err != nil {
		return nil, fmt.Errorf("Error creating new host: %s", err)
	}

	h.HostOptions.AuthOptions.CertDir = constants.Minipath
	h.HostOptions.AuthOptions.StorePath = constants.Minipath

	if err := api.Create(h); err != nil {
		// Wait for all the logs to reach the client
		time.Sleep(2 * time.Second)
		return nil, fmt.Errorf("Error creating. %s", err)
	}

	if err := api.Save(h); err != nil {
		return nil, fmt.Errorf("Error attempting to save store: %s", err)
	}
	return h, nil
}

func setupDirs() error {
	dirs := [...]string{
		constants.Minipath,
		constants.MakeMiniPath("certs"),
		constants.MakeMiniPath("machines")}

	for _, path := range dirs {
		if err := os.MkdirAll(path, 0777); err != nil {
			return fmt.Errorf("Error creating minikube directory: %s", err)
		}
	}
	return nil
}

func certPath(fileName string) string {
	return filepath.Join(constants.Minipath, "certs", fileName)
}
