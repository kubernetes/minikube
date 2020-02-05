/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/vmpath"
)

// syncLocalAssets syncs files from MINIKUBE_HOME into the cluster
func syncLocalAssets(cr command.Runner) error {
	fs, err := localAssets()
	if err != nil {
		return err
	}

	for _, f := range fs {
		err := cr.Copy(f)
		if err != nil {
			return err
		}
	}
	return nil
}

// localAssets returns local files and addons from the minikube home directory
func localAssets() ([]assets.CopyableFile, error) {
	fs, err := assetsFromDir(localpath.MakeMiniPath("addons"), vmpath.GuestAddonsDir, true)
	if err != nil {
		return fs, errors.Wrap(err, "addons dir")
	}

	localFiles, err := assetsFromDir(localpath.MakeMiniPath("files"), "/", false)
	if err != nil {
		return fs, errors.Wrap(err, "files dir")
	}

	fs = append(fs, localFiles...)
	return fs, nil
}

// assetsFromDir generates assets from a local filepath, with/without a flattened hierarchy
func assetsFromDir(localRoot string, destRoot string, flatten bool) ([]assets.CopyableFile, error) {
	glog.Infof("Scanning %s for local assets ...", localRoot)
	fs := []assets.CopyableFile{}
	err := filepath.Walk(localRoot, func(localPath string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(localRoot, localPath)
		if err != nil {
			return err
		}

		// The conversion will strip the leading 0 if present, so add it back if necessary
		ps := fmt.Sprintf("%o", fi.Mode().Perm())
		if len(ps) == 3 {
			ps = fmt.Sprintf("0%s", ps)
		}

		dest := path.Join(destRoot, rel)
		targetDir := path.Dir(dest)
		targetName := path.Base(dest)
		if flatten {
			targetDir = destRoot
		}
		glog.Infof("local asset: %s -> %s in %s", localPath, targetName, targetDir)
		f, err := assets.NewFileAsset(localPath, targetDir, targetName, ps)
		if err != nil {
			return errors.Wrapf(err, "creating file asset for %s", localPath)
		}
		fs = append(fs, f)
		return nil
	})
	return fs, err
}
