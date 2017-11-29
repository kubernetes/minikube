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

package localkube

import (
	"fmt"
	"strings"

	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/sshutil"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/state"
	"github.com/pkg/errors"
)

type LocalkubeBootstrapper struct {
	cmd bootstrapper.CommandRunner
}

func NewLocalkubeBootstrapper(api libmachine.API) (*LocalkubeBootstrapper, error) {
	h, err := api.Load(config.GetMachineName())
	if err != nil {
		return nil, errors.Wrap(err, "getting api client")
	}
	var cmd bootstrapper.CommandRunner
	// The none driver executes commands directly on the host
	if h.Driver.DriverName() == constants.DriverNone {
		cmd = &bootstrapper.ExecRunner{}
	} else {
		client, err := sshutil.NewSSHClient(h.Driver)
		if err != nil {
			return nil, errors.Wrap(err, "getting ssh client")
		}
		cmd = bootstrapper.NewSSHRunner(client)
	}
	return &LocalkubeBootstrapper{
		cmd: cmd,
	}, nil
}

// GetClusterLogs If follow is specified, it will tail the logs
func (lk *LocalkubeBootstrapper) GetClusterLogs(follow bool) (string, error) {
	logsCommand, err := GetLogsCommand(follow)
	if err != nil {
		return "", errors.Wrap(err, "Error getting logs command")
	}

	logs, err := lk.cmd.CombinedOutput(logsCommand)
	if err != nil {
		return "", errors.Wrap(err, "getting cluster logs")
	}

	return logs, nil
}

// GetClusterStatus gets the status of localkube from the host VM.
func (lk *LocalkubeBootstrapper) GetClusterStatus() (string, error) {
	s, err := lk.cmd.CombinedOutput(localkubeStatusCommand)
	if err != nil {
		return "", err
	}
	s = strings.TrimSpace(s)
	if state.Running.String() == s {
		return state.Running.String(), nil
	} else if state.Stopped.String() == s {
		return state.Stopped.String(), nil
	} else {
		return "", fmt.Errorf("Error: Unrecognize output from GetLocalkubeStatus: %s", s)
	}
}

// StartCluster starts a k8s cluster on the specified Host.
func (lk *LocalkubeBootstrapper) StartCluster(kubernetesConfig bootstrapper.KubernetesConfig) error {
	startCommand, err := GetStartCommand(kubernetesConfig)
	if err != nil {
		return errors.Wrapf(err, "Error generating start command: %s", err)
	}
	err = lk.cmd.Run(startCommand) //needs to be sudo for none driver
	if err != nil {
		return errors.Wrapf(err, "Error running ssh command: %s", startCommand)
	}
	return nil
}

func (lk *LocalkubeBootstrapper) RestartCluster(kubernetesConfig bootstrapper.KubernetesConfig) error {
	return lk.StartCluster(kubernetesConfig)
}

func (lk *LocalkubeBootstrapper) UpdateCluster(config bootstrapper.KubernetesConfig) error {
	if config.ShouldLoadCachedImages {
		// Make best effort to load any cached images
		go machine.LoadImages(lk.cmd, constants.LocalkubeCachedImages, constants.ImageCacheDir)
	}

	copyableFiles := []assets.CopyableFile{}
	var localkubeFile assets.CopyableFile
	var err error

	//add url/file/bundled localkube to file list
	lCacher := localkubeCacher{config}
	localkubeFile, err = lCacher.fetchLocalkubeFromURI()
	if err != nil {
		return errors.Wrap(err, "Error updating localkube from uri")
	}
	copyableFiles = append(copyableFiles, localkubeFile)

	// custom addons
	if err := assets.AddMinikubeDirAssets(&copyableFiles); err != nil {
		return errors.Wrap(err, "adding minikube dir assets")
	}
	// bundled addons
	for _, addonBundle := range assets.Addons {
		if isEnabled, err := addonBundle.IsEnabled(); err == nil && isEnabled {
			for _, addon := range addonBundle.Assets {
				copyableFiles = append(copyableFiles, addon)
			}
		} else if err != nil {
			return err
		}
	}

	for _, f := range copyableFiles {
		if err := lk.cmd.Copy(f); err != nil {
			return err
		}
	}
	return nil
}

func (lk *LocalkubeBootstrapper) SetupCerts(k8s bootstrapper.KubernetesConfig) error {
	return bootstrapper.SetupCerts(lk.cmd, k8s)
}
