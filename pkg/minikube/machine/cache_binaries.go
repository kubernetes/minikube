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

package machine

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/golang/glog"
	"github.com/jimmidyson/go-download"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/out"
)

// CacheBinariesForBootstrapper will cache binaries for a bootstrapper
func CacheBinariesForBootstrapper(version string, clusterBootstrapper string) error {
	binaries := bootstrapper.GetCachedBinaryList(clusterBootstrapper)

	var g errgroup.Group
	for _, bin := range binaries {
		bin := bin
		g.Go(func() error {
			if _, err := CacheKubernetesBinary(bin, version, "linux", runtime.GOARCH); err != nil {
				return errors.Wrapf(err, "caching image %s", bin)
			}
			return nil
		})
	}
	return g.Wait()
}

// CacheDockerArchive will cache an archive on the host
func CacheDockerArchive(binary, version, osName, archName string) (string, error) {
	targetDir := constants.MakeMiniPath("cache", "docker", version)

	url := constants.GetDockerReleaseURL(binary, version, osName, archName)

	targetFilepath := path.Join(targetDir, path.Base(url))

	if err := CacheBinary(binary, version, url, targetDir, targetFilepath, "", false); err != nil {
		return "", err
	}
	return targetFilepath, nil
}

// CacheKubernetesBinary will cache a binary on the host
func CacheKubernetesBinary(binary, version, osName, archName string) (string, error) {
	targetDir := constants.MakeMiniPath("cache", version)
	targetFilepath := path.Join(targetDir, binary)

	url := constants.GetKubernetesReleaseURL(binary, version, osName, archName)
	sha := constants.GetKubernetesReleaseURLSHA1(binary, version, osName, archName)
	executable := osName == runtime.GOOS && archName == runtime.GOARCH

	if err := CacheBinary(binary, version, url, targetDir, targetFilepath, sha, executable); err != nil {
		return "", err
	}
	return targetFilepath, nil
}

// CacheBinary will cache a binary on the host
func CacheBinary(binary, version, url, targetDir, targetFilepath, sha1 string, executable bool) error {
	_, err := os.Stat(targetFilepath)
	// If it exists, do no verification and continue
	if err == nil {
		glog.Infof("Not caching binary, using %s", url)
		return nil
	}
	if !os.IsNotExist(err) {
		return errors.Wrapf(err, "stat %s", targetFilepath)
	}

	if err = os.MkdirAll(targetDir, 0777); err != nil {
		return errors.Wrapf(err, "mkdir %s", targetDir)
	}

	out.T(out.FileDownload, "Downloading {{.name}} {{.version}}", out.V{"name": binary, "version": version})

	options := download.FileOptions{
		Mkdirs: download.MkdirAll,
	}

	if sha1 != "" {
		options.Checksum = sha1
		options.ChecksumHash = crypto.SHA1
	}

	if err = download.ToFile(url, targetFilepath, options); err != nil {
		return errors.Wrapf(err, "Error downloading %s", url)
	}
	if executable {
		if err = os.Chmod(targetFilepath, 0755); err != nil {
			return errors.Wrapf(err, "chmod +x %s", targetFilepath)
		}
	}
	return nil
}

// CopyBinary copies previously cached binaries into the path
func CopyBinary(cr command.Runner, binary, path string) error {
	f, err := assets.NewFileAsset(path, "/usr/bin", binary, "0755")
	if err != nil {
		return errors.Wrap(err, "new file asset")
	}
	if err := cr.Copy(f); err != nil {
		return errors.Wrapf(err, "copy")
	}
	return nil
}

// ExtractBinary will extract a binary from an zip/tgz archive
func ExtractBinary(archive, path, binary string) error {
	switch ext := filepath.Ext(archive); ext {
	case ".zip":
		return unzip(archive, path, binary)
	case ".tgz":
		return untar(archive, path, binary)
	default:
		return fmt.Errorf("unknown ext %s", ext)
	}
}

// unzip will decompress a file from a zip archive
func unzip(src string, dst string, member string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {

		if f.FileHeader.Name != member {
			continue
		}

		outFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		_, err = io.Copy(outFile, rc)
		if err != nil {
			return err
		}

		outFile.Close()
		rc.Close()
	}

	return nil
}

// untar will decompress a file from a tgz archive
func untar(src string, dst string, member string) error {
	r, err := os.Open(src)
	if err != nil {
		return err
	}
	defer r.Close()

	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()

		// if no more files are found return
		if err == io.EOF {
			break
		}

		// return any other error
		if err != nil {
			return err
		}

		// if the header is nil, just skip it
		if header == nil {
			continue
		}

		if header.Typeflag == tar.TypeReg && header.Name == member {
			f, err := os.OpenFile(dst, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			// copy over contents
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}

			f.Close()
		}
	}

	return nil
}
