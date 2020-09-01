/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package gvisor

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/vmpath"
)

const (
	nodeDir                        = "/node"
	containerdConfigTomlPath       = "/etc/containerd/config.toml"
	storedContainerdConfigTomlPath = "/tmp/config.toml"
	gvisorContainerdShimURL        = "https://github.com/google/gvisor-containerd-shim/releases/download/v0.0.3/containerd-shim-runsc-v1.linux-amd64"
	gvisorURL                      = "https://storage.googleapis.com/gvisor/releases/nightly/2020-02-14/runsc"
)

// Enable follows these steps for enabling gvisor in minikube:
//   1. creates necessary directories for storing binaries and runsc logs
//   2. downloads runsc and gvisor-containerd-shim
//   3. copies necessary containerd config files
//   4. restarts containerd
func Enable() error {
	if err := makeGvisorDirs(); err != nil {
		return errors.Wrap(err, "creating directories on node")
	}
	if err := downloadBinaries(); err != nil {
		return errors.Wrap(err, "downloading binaries")
	}
	if err := copyConfigFiles(); err != nil {
		return errors.Wrap(err, "copying config files")
	}
	if err := restartContainerd(); err != nil {
		return errors.Wrap(err, "restarting containerd")
	}
	// When pod is terminated, disable gvisor and exit
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		if err := Disable(); err != nil {
			log.Printf("Error disabling gvisor: %v", err)
			os.Exit(1)
		}
		os.Exit(0)
	}()
	log.Print("gvisor successfully enabled in cluster")
	// sleep for one year so the pod continuously runs
	select {}
}

// makeGvisorDirs creates necessary directories on the node
func makeGvisorDirs() error {
	// Make /run/containerd/runsc to hold logs
	fp := filepath.Join(nodeDir, "run/containerd/runsc")
	if err := os.MkdirAll(fp, 0o755); err != nil {
		return errors.Wrap(err, "creating runsc dir")
	}

	// Make /tmp/runsc to also hold logs
	fp = filepath.Join(nodeDir, "tmp/runsc")
	if err := os.MkdirAll(fp, 0o755); err != nil {
		return errors.Wrap(err, "creating runsc logs dir")
	}

	return nil
}

func downloadBinaries() error {
	if err := runsc(); err != nil {
		return errors.Wrap(err, "downloading runsc")
	}
	if err := gvisorContainerdShim(); err != nil {
		return errors.Wrap(err, "downloading gvisor-containerd-shim")
	}
	return nil
}

// downloads the gvisor-containerd-shim
func gvisorContainerdShim() error {
	dest := filepath.Join(nodeDir, "usr/bin/containerd-shim-runsc-v1")
	return downloadFileToDest(gvisorContainerdShimURL, dest)
}

// downloads the runsc binary and returns a path to the binary
func runsc() error {
	dest := filepath.Join(nodeDir, "usr/bin/runsc")
	return downloadFileToDest(gvisorURL, dest)
}

// downloadFileToDest downloads the given file to the dest
// if something already exists at dest, first remove it
func downloadFileToDest(url, dest string) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return errors.Wrapf(err, "creating request for %s", url)
	}
	req.Header.Set("User-Agent", "minikube")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if _, err := os.Stat(dest); err == nil {
		if err := os.Remove(dest); err != nil {
			return errors.Wrapf(err, "removing %s for overwrite", dest)
		}
	}
	fi, err := os.Create(dest)
	if err != nil {
		return errors.Wrapf(err, "creating %s", dest)
	}
	defer fi.Close()
	if _, err := io.Copy(fi, resp.Body); err != nil {
		return errors.Wrap(err, "copying binary")
	}
	if err := fi.Chmod(0o777); err != nil {
		return errors.Wrap(err, "fixing perms")
	}
	return nil
}

// Must write the following files:
//    1. gvisor-containerd-shim.toml
//    2. gvisor containerd config.toml
// and save the default version of config.toml
func copyConfigFiles() error {
	log.Printf("Storing default config.toml at %s", storedContainerdConfigTomlPath)
	if err := mcnutils.CopyFile(filepath.Join(nodeDir, containerdConfigTomlPath), filepath.Join(nodeDir, storedContainerdConfigTomlPath)); err != nil {
		return errors.Wrap(err, "copying default config.toml")
	}
	log.Printf("Copying %s asset to %s", constants.GvisorConfigTomlTargetName, filepath.Join(nodeDir, containerdConfigTomlPath))
	if err := copyAssetToDest(constants.GvisorConfigTomlTargetName, filepath.Join(nodeDir, containerdConfigTomlPath)); err != nil {
		return errors.Wrap(err, "copying gvisor version of config.toml")
	}
	return nil
}

func copyAssetToDest(targetName, dest string) error {
	var asset *assets.BinAsset
	for _, a := range assets.Addons["gvisor"].Assets {
		if a.GetTargetName() == targetName {
			asset = a
		}
	}
	if asset == nil {
		return fmt.Errorf("no asset matching target %s among %+v", targetName, assets.Addons["gvisor"])
	}

	// Now, copy the data from this asset to dest
	src := filepath.Join(vmpath.GuestGvisorDir, asset.GetTargetName())
	log.Printf("%s asset path: %s", targetName, src)
	contents, err := ioutil.ReadFile(src)
	if err != nil {
		return errors.Wrapf(err, "getting contents of %s", asset.GetSourcePath())
	}
	if _, err := os.Stat(dest); err == nil {
		if err := os.Remove(dest); err != nil {
			return errors.Wrapf(err, "removing %s", dest)
		}
	}

	log.Printf("creating %s", dest)
	f, err := os.Create(dest)
	if err != nil {
		return errors.Wrapf(err, "creating %s", dest)
	}
	if _, err := f.Write(contents); err != nil {
		return errors.Wrapf(err, "writing contents to %s", f.Name())
	}
	return nil
}

func restartContainerd() error {
	log.Print("restartContainerd black magic happening")

	log.Print("Stopping rpc-statd.service...")
	cmd := exec.Command("/usr/sbin/chroot", "/node", "sudo", "systemctl", "stop", "rpc-statd.service")
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Println(string(out))
		return errors.Wrap(err, "stopping rpc-statd.service")
	}

	log.Print("Restarting containerd...")
	cmd = exec.Command("/usr/sbin/chroot", "/node", "sudo", "systemctl", "restart", "containerd")
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Print(string(out))
		return errors.Wrap(err, "restarting containerd")
	}

	log.Print("Starting rpc-statd...")
	cmd = exec.Command("/usr/sbin/chroot", "/node", "sudo", "systemctl", "start", "rpc-statd.service")
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Print(string(out))
		return errors.Wrap(err, "restarting rpc-statd.service")
	}
	log.Print("containerd restart complete")
	return nil
}
