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

package localpath

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/golang/glog"
	"github.com/otiai10/copy"
	"k8s.io/client-go/util/homedir"
	"k8s.io/minikube/pkg/minikube/kubeconfig"
)

// migrateLegacyPaths converts legacy (pre-v1.5.0) paths to modern race-free paths
func migrateLegacyPaths(oldHome string) (map[string]string, error) {
	summary := map[string]string{}
	if oldHome == "" {
		glog.Infof("No existing home directory to migrate")
		return summary, nil
	}

	plans, err := migrationPlan(oldHome)
	if err != nil {
		return summary, err
	}
	for src, dst := range plans {
		if src == dst {
			return summary, fmt.Errorf("src == dst: %s", src)
		}
		_, err := os.Stat(src)
		if os.IsNotExist(err) {
			glog.Warningf("%s does not exist", src)
			continue
		}

		glog.Infof("copying %s -> %s", src, dst)
		err = copy.Copy(src, dst)
		if err != nil {
			return summary, err
		}
	}

	// A simplified version of the plan, to communicate to users.
	summary = map[string]string{
		filepath.Join(oldHome, "cache"):  cacheDir(),
		filepath.Join(oldHome, "config"): configDir(),
		oldHome:                          dataDir(),
	}

	// At this point, we should be confident that every file has been copied. Do a sanity check, though.
	if err := validateCopies(oldHome, []string{cacheDir(), configDir(), dataDir()}); err != nil {
		return summary, err
	}

	if err := kubeconfig.EmbedMinikubeCerts(); err != nil {
		return summary, err
	}

	return summary, os.RemoveAll(oldHome)
}

// legacyHome calculates the default root directory exactly like old versions of minikube
func legacyHome() string {
	return filepath.Join(homedir.HomeDir(), ".minikube")
}

// existingHome returns an existing root directory
func existingHome() string {
	for _, dir := range []string{overrideHome(), legacyHome()} {
		if dir != "" {
			if _, err := os.Stat(dir); err == nil {
				return dir
			}
		}
	}
	return ""
}

// migrationPlan return a plan of events to migrate files from old locations
func migrationPlan(root string) (map[string]string, error) {
	toCopy := map[string]string{
		filepath.Join(root, "cache"):    cacheDir(),
		filepath.Join(root, "config"):   configDir(),
		filepath.Join(root, "profiles"): Profiles(),
	}

	ms, err := legacyMachineList(root)
	if err != nil {
		return toCopy, err
	}

	// Move each machine and set of certificates into its own store
	for _, m := range ms {
		toCopy[filepath.Join(root, "machines", m)] = Machine(m)
		toCopy[filepath.Join(root, "certs")] = filepath.Join(Store(m), "certs")
	}

	// Copy common machine files into each store to make them race-proof
	mf, err := filepath.Glob(filepath.Join(root, "machines/*.*"))
	if err != nil {
		return toCopy, err
	}
	for _, f := range mf {
		rp, err := filepath.Rel(root, f)
		if err != nil {
			return toCopy, err
		}
		for _, m := range ms {
			toCopy[f] = filepath.Join(Store(m), rp)
		}
	}

	// Move everything else in the root directory
	of, err := filepath.Glob(filepath.Join(root, "*"))
	if err != nil {
		return toCopy, err
	}
	for _, f := range of {
		rp, err := filepath.Rel(root, f)
		if err != nil {
			return toCopy, err
		}

		if toCopy[f] != "" {
			continue
		}
		ext := filepath.Ext(f)
		if ext == ".key" || ext == ".crt" || ext == ".pem" {
			for _, m := range ms {
				toCopy[f] = filepath.Join(KubernetesCerts(m), rp)
			}
			continue
		}
		toCopy[f] = filepath.Join(dataDir(), rp)
	}
	return toCopy, nil
}

func legacyMachineList(root string) ([]string, error) {
	result := []string{}
	lp := filepath.Join(root, "machines")
	_, err := os.Stat(lp)
	if os.IsNotExist(err) {
		return result, nil
	}

	m, err := ioutil.ReadDir(lp)
	if err != nil {
		return result, err
	}
	for _, fi := range m {
		if fi.IsDir() {
			result = append(result, fi.Name())
		}
	}
	return result, nil
}

type foundFile struct {
	Path string
	Info os.FileInfo
}

// validateCopies validates that all files in src exist in dests
func validateCopies(src string, dests []string) error {
	// Record everything found in the destination folders
	found := map[string][]foundFile{}
	for _, d := range dests {
		err := filepath.Walk(d, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			f := foundFile{Path: path, Info: info}
			base := filepath.Base(path)
			_, ok := found[base]
			if !ok {
				found[base] = []foundFile{f}
			} else {
				found[base] = append(found[base], f)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	// Compare the source folder against the destinations
	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		base := filepath.Base(path)
		matches, ok := found[base]
		if !ok {
			return fmt.Errorf("%s: No file named %s exists in %v", path, base, dests)
		}
		for _, m := range matches {
			if m.Info.Size() == info.Size() && m.Info.Mode() == info.Mode() {
				glog.Infof("%s was copied to %s", path, m.Path)
				return nil
			} else {
				glog.Infof("%s size=%d, mode=%v is different than %s size=%d, mode=%v", path, info.Size(), info.Mode(), m.Path, m.Info.Size(), m.Info.Mode())
			}
		}
		return fmt.Errorf("%s: No file named %s has a length of %d and mode of %v in %v", path, base, info.Size(), info.Mode(), dests)
	})

	return err
}
