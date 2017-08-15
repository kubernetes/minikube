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

// GetHostLogs gets the localkube logs of the host VM.
import (
	"fmt"
	"strings"

	"golang.org/x/crypto/ssh"

	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/sshutil"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/state"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type LocalkubeBootstrapper struct {
	c      *ssh.Client
	driver string // TODO(r2d4): get rid of this dependency
}

func NewLocalkubeBootstrapper(api libmachine.API) (*LocalkubeBootstrapper, error) {
	h, err := api.Load(config.GetMachineName())
	if err != nil {
		return nil, errors.Wrap(err, "getting api client")
	}
	client, err := sshutil.NewSSHClient(h.Driver)
	if err != nil {
		return nil, errors.Wrap(err, "getting ssh client")
	}
	return &LocalkubeBootstrapper{
		c:      client,
		driver: h.Driver.DriverName(),
	}, nil
}

// GetClusterLogs If follow is specified, it will tail the logs
func (lk *LocalkubeBootstrapper) GetClusterLogs(follow bool) (string, error) {
	logsCommand, err := GetLogsCommand(follow)
	if err != nil {
		return "", errors.Wrap(err, "Error getting logs command")
	}
	sess, err := lk.c.NewSession()
	if err != nil {
		return "", errors.Wrap(err, "getting ssh session")
	}
	defer sess.Close()
	if follow {
		err := sshutil.GetShell(sess, logsCommand)
		return "", errors.Wrap(err, "error getting shell")
	}
	fmt.Println("running cmd")
	s, err := cluster.RunCommand(lk.c, lk.driver, logsCommand, false)
	fmt.Println("ended running cmd")
	if err != nil {
		return "", errors.Wrap(err, "error running logs command")
	}
	return s, nil
}

// GetClusterStatus gets the status of localkube from the host VM.
func (lk *LocalkubeBootstrapper) GetClusterStatus() (string, error) {
	s, err := cluster.RunCommand(lk.c, lk.driver, localkubeStatusCommand, false)
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
	glog.Infoln(startCommand)
	output, err := cluster.RunCommand(lk.c, lk.driver, startCommand, true)
	glog.Infoln(output)
	if err != nil {
		return errors.Wrapf(err, "Error running ssh command: %s", startCommand)
	}
	return nil
}

func (lk *LocalkubeBootstrapper) RestartCluster(kubernetesConfig bootstrapper.KubernetesConfig) error {
	return lk.StartCluster(kubernetesConfig)
}

func (lk *LocalkubeBootstrapper) UpdateCluster(config bootstrapper.KubernetesConfig) error {
	copyableFiles := []assets.CopyableFile{}
	var localkubeFile assets.CopyableFile
	var err error

	//add url/file/bundled localkube to file list
	if localkubeURIWasSpecified(config) && config.KubernetesVersion != constants.DefaultKubernetesVersion {
		lCacher := localkubeCacher{config}
		localkubeFile, err = lCacher.fetchLocalkubeFromURI()
		if err != nil {
			return errors.Wrap(err, "Error updating localkube from uri")
		}
	} else {
		localkubeFile = assets.NewBinDataAsset("out/localkube", "/usr/local/bin", "localkube", "0777")
	}
	copyableFiles = append(copyableFiles, localkubeFile)

	// add addons to file list
	// custom addons
	assets.AddMinikubeAddonsDirToAssets(&copyableFiles)
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

	if lk.driver == constants.DriverNone {
		// transfer files to correct place on filesystem
		for _, f := range copyableFiles {
			if err := assets.CopyFileLocal(f); err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range copyableFiles {
		if err := sshutil.TransferFile(f, lk.c); err != nil {
			return err
		}
	}
	return nil
}
