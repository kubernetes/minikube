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
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/drivers/plugin"
	"github.com/docker/machine/libmachine/drivers/rpc"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/swarm"
	"rsprd.com/localkube/pkg/localkubectl"
)

const machineName = "mymachine"

// Fix for windows
var minipath = filepath.Join(os.Getenv("HOME"), "minikube")

// StartDriver starts the specified driver.
func StartDriver(driverName string) {
	switch driverName {
	case "virtualbox":
		plugin.RegisterDriver(virtualbox.NewDriver("", ""))
	default:
		fmt.Fprintf(os.Stderr, "Unsupported driver: %s\n", driverName)
		os.Exit(1)
	}
}

// StartHost starts a host VM.
func StartHost() error {
	setupDirs()
	api := libmachine.NewClient(minipath, makeMiniPath("certs"))
	defer api.Close()

	h, err := getOrCreateHost(api)
	if err != nil {
		log.Panicf("Error getting host: ", err)
	}
	host, err := h.Driver.GetURL()
	if err != nil {
		log.Panicf("Error getting URL: ", err)
	}
	os.Setenv("DOCKER_HOST", host)
	os.Setenv("DOCKER_CERT_PATH", makeMiniPath("certs"))
	os.Setenv("DOCKER_TLS_VERIFY", "1")
	ctlr, err := localkubectl.NewControllerFromEnv(os.Stdout)
	if err != nil {
		log.Panicf("Error creating controller: ", err)
	}

	startCluster := func() error {
		ctrID, running, err := ctlr.OnlyLocalkubeCtr()
		if err != nil {
			if err == localkubectl.ErrNoContainer {
				// if container doesn't exist, create
				ctrID, running, err = ctlr.CreateCtr(localkubectl.LocalkubeContainerName, "latest")
				if err != nil {
					return err
				}
			} else {
				// stop for all other errors
				return err
			}
		}

		// start container if not running
		if !running {
			err = ctlr.StartCtr(ctrID, "")
			if err != nil {
				return err
			}
		} else {
			log.Println("Localkube is already running")
		}
		return nil
	}
	if err := startCluster(); err != nil {
		log.Println("Error starting cluster: ", err)
	} else {
		kubeHost := strings.Replace(host, "tcp://", "http://", -1)
		kubeHost = strings.Replace(kubeHost, ":2376", ":8080", -1)
		log.Printf("Kubernetes is available at %s.\n", kubeHost)
		log.Println("Run this command to use the cluster: ")
		log.Printf("kubectl config set-cluster localkube --insecure-skip-tls-verify=true --server=%s\n", kubeHost)
	}
	return nil
}

func getOrCreateHost(api *libmachine.Client) (*host.Host, error) {
	if exists, err := api.Exists(machineName); err != nil {
		return nil, fmt.Errorf("Error checking if host exists: %s", err)
	} else if exists {
		log.Println("Machine exists!")
		h, err := api.Load(machineName)
		if err != nil {
			return nil, fmt.Errorf("Error loading existing host.")
		}
		return h, nil
	} else {
		return createHost(api)
	}
}

func createHost(api *libmachine.Client) (*host.Host, error) {
	rawDriver, err := json.Marshal(&drivers.BaseDriver{
		MachineName: machineName,
		StorePath:   minipath,
	})
	if err != nil {
		return nil, fmt.Errorf("Error attempting to marshal bare driver data: %s", err)
	}

	driverName := "virtualbox"
	h, err := api.NewHost(driverName, rawDriver)
	if err != nil {
		return nil, fmt.Errorf("Error getting new host: %s", err)
	}

	setHostOptions(h)
	if err := setDriverOptions(h); err != nil {
		return nil, fmt.Errorf("Error setting driver options: %s", err)
	}

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
	for _, path := range [...]string{minipath, makeMiniPath("certs"), makeMiniPath("machines")} {
		if err := os.MkdirAll(path, 0777); err != nil {
			return fmt.Errorf("Error creating minikube directory: ", err)
		}
	}
	return nil
}

func certPath(fileName string) string {
	return filepath.Join(minipath, "certs", fileName)
}

func makeMiniPath(fileName string) string {
	return filepath.Join(minipath, fileName)
}

func setHostOptions(h *host.Host) {
	h.HostOptions = &host.Options{
		AuthOptions: &auth.Options{
			CertDir:          minipath,
			CaCertPath:       certPath("ca.pem"),
			CaPrivateKeyPath: certPath("ca-key.pem"),
			ClientCertPath:   certPath("cert.pem"),
			ClientKeyPath:    certPath("key.pem"),
			ServerCertPath:   certPath("server.pem"),
			ServerKeyPath:    certPath("server-key.pem"),
			StorePath:        minipath,
			ServerCertSANs:   []string{},
		},
		EngineOptions: &engine.Options{
			TLSVerify:        true,
			ArbitraryFlags:   []string{},
			Env:              []string{},
			InsecureRegistry: []string{},
			Labels:           []string{},
			RegistryMirror:   []string{},
			StorageDriver:    "",
			InstallURL:       "https://get.docker.com",
		},
		SwarmOptions: &swarm.Options{
			IsSwarm:        false,
			Image:          "",
			Master:         false,
			Discovery:      "",
			Address:        "",
			Host:           "",
			Strategy:       "",
			ArbitraryFlags: []string{},
			IsExperimental: false,
		},
	}
}

func setDriverOptions(h *host.Host) error {
	driverOpts := rpcdriver.RPCFlags{
		Values: make(map[string]interface{}),
	}
	mcnFlags := h.Driver.GetCreateFlags()
	for _, f := range mcnFlags {
		driverOpts.Values[f.String()] = f.Default()
	}
	driverOpts.Values["virtualbox-boot2docker-url"] = "https://storage.googleapis.com/tinykube/boot2docker.iso"
	if err := h.Driver.SetConfigFromFlags(driverOpts); err != nil {
		return fmt.Errorf("Error setting machine configuration from flags provided: %s", err)
	}
	return nil
}
