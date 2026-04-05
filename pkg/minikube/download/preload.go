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

package download

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"

	"github.com/spf13/viper"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/download/gh"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/style"
)

const (
	// PreloadVersion is the current version of the preloaded tarball
	//
	// NOTE: You may need to bump this version up when upgrading auxiliary docker images
	PreloadVersion = "v18"
	// PreloadBucket is the name of the GCS bucket where preloaded volume tarballs exist
	PreloadBucket     = "minikube-preloaded-volume-tarballs"
	PreloadGitHubOrg  = "kubernetes-sigs"
	PreloadGitHubRepo = "minikube-preloads"
)

type preloadSource string

const (
	preloadSourceNone   preloadSource = ""
	preloadSourceLocal  preloadSource = "local"
	preloadSourceGCS    preloadSource = "gcs"
	preloadSourceGitHub preloadSource = "github"
)

type preloadState struct {
	exists bool
	source preloadSource
}

var (
	preloadStates = make(map[string]map[string]preloadState)
)

// TarballName returns name of the tarball
func TarballName(k8sVersion, containerRuntime string) string {
	if containerRuntime == "crio" {
		containerRuntime = "cri-o"
	}
	var storageDriver string
	if containerRuntime == "cri-o" {
		storageDriver = "overlay"
	} else {
		storageDriver = "overlay2"
	}
	return fmt.Sprintf("preloaded-images-k8s-%s-%s-%s-%s-%s.tar.lz4", PreloadVersion, k8sVersion, containerRuntime, storageDriver, runtime.GOARCH)
}

// returns target dir for all cached items related to preloading
func targetDir() string {
	return localpath.MakeMiniPath("cache", "preloaded-tarball")
}

// TarballPath returns the local path to the cached preload tarball
func TarballPath(k8sVersion, containerRuntime string) string {
	return filepath.Join(targetDir(), TarballName(k8sVersion, containerRuntime))
}

// remoteTarballURLGCS returns the URL for the remote tarball in GCS
func remoteTarballURLGCS(k8sVersion, containerRuntime string) string {
	return fmt.Sprintf("https://%s/%s/%s/%s/%s", downloadHost, PreloadBucket, PreloadVersion, k8sVersion, TarballName(k8sVersion, containerRuntime))
}

// remoteTarballURLGitHub returns the URL for the remote tarball hosted on GitHub releases
func remoteTarballURLGitHub(k8sVersion, containerRuntime string) string {
	return fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s", PreloadGitHubOrg, PreloadGitHubRepo, PreloadVersion, TarballName(k8sVersion, containerRuntime))
}

func remoteTarballURL(k8sVersion, containerRuntime string, source preloadSource) string {
	switch source {
	case preloadSourceGitHub:
		return remoteTarballURLGitHub(k8sVersion, containerRuntime)
	case preloadSourceGCS:
		return remoteTarballURLGCS(k8sVersion, containerRuntime)
	default:
		return string(preloadSourceNone)
	}
}

func setPreloadState(k8sVersion, containerRuntime string, state preloadState) {
	cRuntimes, ok := preloadStates[k8sVersion]
	if !ok {
		cRuntimes = make(map[string]preloadState)
		preloadStates[k8sVersion] = cRuntimes
	}
	cRuntimes[containerRuntime] = state
}

func getPreloadState(k8sVersion, containerRuntime string) (preloadState, bool) {
	if cRuntimes, ok := preloadStates[k8sVersion]; ok {
		if state, ok := cRuntimes[containerRuntime]; ok {
			return state, true
		}
	}
	return preloadState{}, false
}

func remotePreloadExists(url string) bool {
	resp, err := http.Head(url)
	if err != nil {
		klog.Warningf("%s fetch error: %v", url, err)
		return false
	}

	// note: err won't be set if it's a 404
	if resp.StatusCode != http.StatusOK {
		klog.Warningf("%s status code: %d", url, resp.StatusCode)
		return false
	}

	klog.Infof("Found remote preload: %s", url)
	return true
}

// this is a function variable so it can be overridden in tests
var checkRemotePreloadExistsGCS = func(k8sVersion, containerRuntime string) bool {
	url := remoteTarballURLGCS(k8sVersion, containerRuntime)
	return remotePreloadExists(url)
}

// this is a function variable so it can be overridden in tests
var checkRemotePreloadExistsGitHub = func(k8sVersion, containerRuntime string) bool {
	url := remoteTarballURLGitHub(k8sVersion, containerRuntime)
	return remotePreloadExists(url)
}

// PreloadExistsGCS returns true if there is a preloaded tarball in GCS that can be used
func PreloadExistsGCS(k8sVersion, containerRuntime string) bool {
	return checkRemotePreloadExistsGCS(k8sVersion, containerRuntime)
}

// PreloadExistsGH returns true if there is a preloaded tarball in GitHub releases that can be used
func PreloadExistsGH(k8sVersion, containerRuntime string) bool {
	return checkRemotePreloadExistsGitHub(k8sVersion, containerRuntime)
}

// PreloadExists returns true if there is a preloaded tarball that can be used
func PreloadExists(k8sVersion, containerRuntime, driverName string, forcePreload ...bool) bool {
	// Prevent preload logic in --no-kubernetes mode
	if viper.GetBool("no-kubernetes") {
		klog.Infof("Skipping preload logic due to --no-kubernetes flag")
		return false
	}
	// TODO (#8166): Get rid of the need for this and viper at all
	force := false
	if len(forcePreload) > 0 {
		force = forcePreload[0]
	}

	// TODO: debug why this func is being called two times
	klog.Infof("Checking if preload exists for k8s version %s and runtime %s", k8sVersion, containerRuntime)
	// If `driverName` is BareMetal, there is no preload. Note: some uses of
	// `PreloadExists` assume that the driver is irrelevant unless BareMetal.
	if !driver.AllowsPreload(driverName) || !viper.GetBool("preload") && !force {
		return false
	}

	// If the preload existence is cached, just return that value.
	if state, ok := getPreloadState(k8sVersion, containerRuntime); ok {
		return state.exists
	}

	// Omit remote check if tarball exists locally
	targetPath := TarballPath(k8sVersion, containerRuntime)
	if f, err := checkCache(targetPath); err == nil && f.Size() != 0 {
		klog.Infof("Found local preload: %s", targetPath)
		setPreloadState(k8sVersion, containerRuntime, preloadState{exists: true, source: preloadSourceLocal})
		return true
	}

	switch viper.GetString("preload-source") {
	case "github":
		if PreloadExistsGH(k8sVersion, containerRuntime) {
			setPreloadState(k8sVersion, containerRuntime, preloadState{exists: true, source: preloadSourceGitHub})
			return true
		}
	case "gcs":
		if PreloadExistsGCS(k8sVersion, containerRuntime) {
			setPreloadState(k8sVersion, containerRuntime, preloadState{exists: true, source: preloadSourceGCS})
			return true
		}
	default:
		// auto or unknown - try both
		if PreloadExistsGCS(k8sVersion, containerRuntime) {
			setPreloadState(k8sVersion, containerRuntime, preloadState{exists: true, source: preloadSourceGCS})
			return true
		}
		if PreloadExistsGH(k8sVersion, containerRuntime) {
			setPreloadState(k8sVersion, containerRuntime, preloadState{exists: true, source: preloadSourceGitHub})
			return true
		}
	}

	setPreloadState(k8sVersion, containerRuntime, preloadState{exists: false, source: preloadSourceNone})
	return false
}

var checkPreloadExists = PreloadExists

// Preload caches the preloaded images tarball on the host machine
func Preload(k8sVersion, containerRuntime, driverName string) error {
	targetPath := TarballPath(k8sVersion, containerRuntime)
	targetLock := targetPath + ".lock"

	releaser, err := lockDownload(targetLock)
	if releaser != nil {
		defer releaser.Release()
	}
	if err != nil {
		return err
	}

	if f, err := checkCache(targetPath); err == nil && f.Size() != 0 {
		klog.Infof("Found %s in cache, skipping download", targetPath)
		return nil
	}

	// Make sure we support this k8s version
	if !checkPreloadExists(k8sVersion, containerRuntime, driverName) {
		klog.Infof("Preloaded tarball for k8s version %s does not exist", k8sVersion)
		return nil
	}

	out.Step(style.FileDownload, "Downloading Kubernetes {{.version}} preload ...", out.V{"version": k8sVersion})
	state, ok := getPreloadState(k8sVersion, containerRuntime)
	source := preloadSourceNone
	// check if one of the sources are set to be used
	if ok && state.source != preloadSourceNone {
		source = state.source
	}
	url := remoteTarballURL(k8sVersion, containerRuntime, source)
	klog.Infof("Downloading preload from %s", url)
	var checksum []byte
	var chksErr error
	checksum, chksErr = getChecksum(source, k8sVersion, containerRuntime)

	var realPath string
	if chksErr != nil {
		klog.Warningf("No checksum for preloaded tarball for k8s version %s: %v", k8sVersion, chksErr)
		realPath = targetPath
		tmp, err := os.CreateTemp(targetDir(), TarballName(k8sVersion, containerRuntime)+".*")
		if err != nil {
			return fmt.Errorf("tempfile: %w", err)
		}
		targetPath = tmp.Name()
		if err := tmp.Close(); err != nil {
			return fmt.Errorf("tempfile close: %w", err)
		}
	} else if checksum != nil { // add URL parameter for go-getter to automatically verify the checksum
		url = addChecksumToURL(url, source, checksum)
	}

	if err := download(url, targetPath); err != nil {
		return fmt.Errorf("download failed: %s: %w", url, err)
	}

	//  to avoid partial/corrupt files in final dest. only rename tmp if download didn't error out.
	if realPath != "" {
		klog.Infof("renaming tempfile to %s ...", TarballName(k8sVersion, containerRuntime))
		err := os.Rename(targetPath, realPath)
		if err != nil {
			return fmt.Errorf("rename: %w", err)
		}
	}

	// If the download was successful, mark off that the preload exists in the cache.
	setPreloadState(k8sVersion, containerRuntime, preloadState{exists: true, source: source})
	return nil
}

// addChecksumToURL appends the checksum query parameter to the URL for go-getter (so it can verify before/after download)
func addChecksumToURL(url string, ps preloadSource, checksum []byte) string {
	switch ps {
	case preloadSourceGCS: // GCS API gives us MD5 checksums only
		url += fmt.Sprintf("?checksum=md5:%s", hex.EncodeToString(checksum))
		klog.Infof("Got checksum from GCS API %q", hex.EncodeToString(checksum))
	case preloadSourceGitHub: // GCS API gives us sha256
		url += fmt.Sprintf("?checksum=sha256:%s", checksum)
		klog.Infof("Got checksum from Github API %q", checksum)
	}
	return url
}

func getStorageAttrs(name string) (*storage.ObjectAttrs, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithoutAuthentication())
	if err != nil {
		return nil, fmt.Errorf("getting storage client: %w", err)
	}
	attrs, err := client.Bucket(PreloadBucket).Object(name).Attrs(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting storage object: %w", err)
	}
	return attrs, nil
}

// getChecksumGCS returns the MD5 checksum of the preload tarball
var getChecksumGCS = func(k8sVersion, containerRuntime string) ([]byte, error) {
	klog.Infof("getting checksum for %s from gcs api...", TarballName(k8sVersion, containerRuntime))
	filename := fmt.Sprintf("%s/%s/%s", PreloadVersion, k8sVersion, TarballName(k8sVersion, containerRuntime))
	attrs, err := getStorageAttrs(filename)
	if err != nil {
		return nil, err
	}
	return attrs.MD5, nil
}

// getChecksumGithub returns the SHA256 checksum of the preload tarball
var getChecksumGithub = func(k8sVersion, containerRuntime string) ([]byte, error) {
	klog.Infof("getting checksum for %s from github api...", TarballName(k8sVersion, containerRuntime))
	assets, err := gh.ReleaseAssets(PreloadGitHubOrg, PreloadGitHubRepo, PreloadVersion)
	if err != nil { // could not find release or rate limited
		return nil, err
	}
	return gh.AssetSHA256(TarballName(k8sVersion, containerRuntime), assets)
}

func getChecksum(ps preloadSource, k8sVersion, containerRuntime string) ([]byte, error) {
	switch ps {
	case preloadSourceGCS:
		return getChecksumGCS(k8sVersion, containerRuntime)
	case preloadSourceGitHub:
		return getChecksumGithub(k8sVersion, containerRuntime)
	default:
		return nil, fmt.Errorf("unknown preload source: %s", ps)
	}
}

// CleanUpOlderPreloads deletes preload files belonging to older minikube versions
// checks the current preload version and then if the saved tar file is belongs to older minikube it will delete it
// in case of failure only logs to the user
func CleanUpOlderPreloads() {
	files, err := os.ReadDir(targetDir())
	if err != nil {
		klog.Warningf("Failed to list preload files: %v", err)
	}

	for _, file := range files {
		split := strings.Split(file.Name(), "-")
		if len(split) < 4 {
			continue
		}
		ver := split[3]
		if ver != PreloadVersion {
			fn := path.Join(targetDir(), file.Name())
			klog.Infof("deleting older generation preload %s", fn)
			err := os.Remove(fn)
			if err != nil {
				klog.Warningf("Failed to clean up older preload files, consider running `minikube delete --all --purge`")
			}
		}
	}
}
