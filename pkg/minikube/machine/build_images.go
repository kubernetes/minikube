/*
Copyright 2021 The Kubernetes Authors All rights reserved.

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
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/docker/machine/libmachine/state"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/vmpath"
)

// buildRoot is where images should be built from within the guest VM
var buildRoot = path.Join(vmpath.GuestPersistentDir, "build")

// BuildImage builds image to all profiles
func BuildImage(path string, file string, tag string, push bool, env []string, opt []string, profiles []*config.Profile) error {
	api, err := NewAPIClient()
	if err != nil {
		return errors.Wrap(err, "api")
	}
	defer api.Close()

	succeeded := []string{}
	failed := []string{}

	u, err := url.Parse(path)
	if err == nil && u.Scheme == "file" {
		path = u.Path
	}
	remote := err == nil && u.Scheme != ""
	if runtime.GOOS == "windows" && filepath.VolumeName(path) != "" {
		remote = false
	}

	for _, p := range profiles { // building images to all running profiles
		pName := p.Name // capture the loop variable

		c, err := config.Load(pName)
		if err != nil {
			// Non-fatal because it may race with profile deletion
			klog.Errorf("Failed to load profile %q: %v", pName, err)
			failed = append(failed, pName)
			continue
		}

		for _, n := range c.Nodes {
			m := config.MachineName(*c, n)

			status, err := Status(api, m)
			if err != nil {
				klog.Warningf("error getting status for %s: %v", m, err)
				failed = append(failed, m)
				continue
			}

			if status == state.Running.String() {
				h, err := api.Load(m)
				if err != nil {
					klog.Warningf("Failed to load machine %q: %v", m, err)
					failed = append(failed, m)
					continue
				}
				cr, err := CommandRunner(h)
				if err != nil {
					return err
				}
				if remote {
					err = buildImage(cr, c.KubernetesConfig, path, file, tag, push, env, opt)
				} else {
					err = transferAndBuildImage(cr, c.KubernetesConfig, path, file, tag, push, env, opt)
				}
				if err != nil {
					failed = append(failed, m)
					klog.Warningf("Failed to build image for profile %s. make sure the profile is running. %v", pName, err)
					continue
				}
				succeeded = append(succeeded, m)
			}
		}
	}

	klog.Infof("succeeded building to: %s", strings.Join(succeeded, " "))
	klog.Infof("failed building to: %s", strings.Join(failed, " "))
	return nil
}

// buildImage builds a single image
func buildImage(cr command.Runner, k8s config.KubernetesConfig, src string, file string, tag string, push bool, env []string, opt []string) error {
	r, err := cruntime.New(cruntime.Config{Type: k8s.ContainerRuntime, Runner: cr})
	if err != nil {
		return errors.Wrap(err, "runtime")
	}
	klog.Infof("Building image from url: %s", src)

	err = r.BuildImage(src, file, tag, push, env, opt)
	if err != nil {
		return errors.Wrapf(err, "%s build %s", r.Name(), src)
	}

	klog.Infof("Built %s from %s", tag, src)
	return nil
}

// transferAndBuildImage transfers and builds a single image
func transferAndBuildImage(cr command.Runner, k8s config.KubernetesConfig, src string, file string, tag string, push bool, env []string, opt []string) error {
	r, err := cruntime.New(cruntime.Config{Type: k8s.ContainerRuntime, Runner: cr})
	if err != nil {
		return errors.Wrap(err, "runtime")
	}
	klog.Infof("Building image from path: %s", src)

	filename := filepath.Base(src)
	filename = localpath.SanitizeCacheDir(filename)

	if _, err := os.Stat(src); err != nil {
		return err
	}

	args := append([]string{"mkdir", "-p"}, buildRoot)
	if _, err := cr.RunCmd(exec.Command("sudo", args...)); err != nil {
		return err
	}

	dst := path.Join(buildRoot, filename)
	f, err := assets.NewFileAsset(src, buildRoot, filename, "0644")
	if err != nil {
		return errors.Wrapf(err, "creating copyable file asset: %s", filename)
	}
	defer func() {
		if err := f.Close(); err != nil {
			klog.Warningf("error closing the file %s: %v", f.GetSourcePath(), err)
		}
	}()

	if err := cr.Copy(f); err != nil {
		return errors.Wrap(err, "transferring cached image")
	}

	context := path.Join(buildRoot, ".", strings.TrimSuffix(filename, filepath.Ext(filename)))
	args = append([]string{"mkdir", "-p"}, context)
	if _, err := cr.RunCmd(exec.Command("sudo", args...)); err != nil {
		return err
	}
	args = append([]string{"tar", "-C", context, "-xf"}, dst)
	if _, err := cr.RunCmd(exec.Command("sudo", args...)); err != nil {
		return err
	}

	if file != "" && !path.IsAbs(file) {
		file = path.Join(context, file)
	}
	err = r.BuildImage(context, file, tag, push, env, opt)
	if err != nil {
		return errors.Wrapf(err, "%s build %s", r.Name(), dst)
	}

	args = append([]string{"rm", "-rf"}, context)
	if _, err := cr.RunCmd(exec.Command("sudo", args...)); err != nil {
		return err
	}
	args = append([]string{"rm", "-f"}, dst)
	if _, err := cr.RunCmd(exec.Command("sudo", args...)); err != nil {
		return err
	}

	klog.Infof("Built %s from %s", tag, src)
	return nil
}
