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
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"errors"

	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/detect"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/style"
	"k8s.io/minikube/pkg/util/lock"
	"k8s.io/minikube/pkg/version"
)

var isWindowsISO bool

const fileScheme = "file"

// DefaultISOURLs returns a list of ISO URL's to consult by default, in priority order
func DefaultISOURLs() []string {
	v := version.GetISOVersion()
	isoBucket := "minikube-builds/iso/22436"

	return []string{
		fmt.Sprintf("https://storage.googleapis.com/%s/minikube-%s-%s.iso", isoBucket, v, runtime.GOARCH),
		fmt.Sprintf("https://github.com/kubernetes/minikube/releases/download/%s/minikube-%s-%s.iso", v, v, runtime.GOARCH),
		fmt.Sprintf("https://kubernetes.oss-cn-hangzhou.aliyuncs.com/minikube/iso/minikube-%s-%s.iso", v, runtime.GOARCH),
	}
}

// WindowsISOURL retrieves the ISO URL for the Windows version specified
func WindowsISOURL(version string) string {
	versionToIsoURL := map[string]string{
		"2022": constants.DefaultWindowsServerIsoURL,
		// Add more versions here when we support them
	}

	url, exists := versionToIsoURL[version]
	if !exists {
		klog.Warningf("Windows version %s is not supported. Using default Windows Server ISO URL", version)
		return constants.DefaultWindowsServerIsoURL
	}

	return url
}

// LocalISOResource returns a local file:// URI equivalent for a local or remote ISO path
func LocalISOResource(isoURL string) string {
	u, err := url.Parse(isoURL)
	if err != nil {
		fake := "file://" + filepath.ToSlash(isoURL)
		klog.Errorf("%s is not a URL! Returning %s", isoURL, fake)
		return fake
	}

	if u.Scheme == fileScheme {
		return isoURL
	}

	return fileURI(localISOPath(u))
}

// fileURI returns a file:// URI for a path
func fileURI(filePath string) string {
	return "file://" + filepath.ToSlash(filePath)
}

// localISOPath returns where an ISO should be stored locally
func localISOPath(u *url.URL) string {
	if u.Scheme == fileScheme {
		return u.String()
	}

	return filepath.Join(detect.ISOCacheDir(), path.Base(u.Path))
}

// ISO downloads and returns the path to the downloaded ISO
func ISO(urls []string, skipChecksum bool) (string, error) {
	errs := map[string]string{}

	for _, url := range urls {
		err := downloadISO(url, skipChecksum)
		if err != nil {
			klog.Errorf("Unable to download %s: %v", url, err)
			errs[url] = err.Error()
			continue
		}
		return url, nil
	}

	var msg strings.Builder
	msg.WriteString("unable to cache ISO: \n")
	for u, err := range errs {
		msg.WriteString(fmt.Sprintf("  %s: %s\n", u, err))
	}

	return "", errors.New(msg.String())
}

func WindowsISO(windowsVersion string) error {
	isWindowsISO = true
	isoURL := WindowsISOURL(windowsVersion)
	return downloadISO(isoURL, false)
}

// downloadISO downloads an ISO URL
func downloadISO(isoURL string, skipChecksum bool) error {
	u, err := url.Parse(isoURL)
	if err != nil {
		return fmt.Errorf("url.parse %q: %w", isoURL, err)
	}

	// It's already downloaded
	if u.Scheme == fileScheme {
		return nil
	}

	// Lock before we check for existence to avoid thundering herd issues
	dst := localISOPath(u)
	if isWindowsISO {
		resp, err := http.Head(isoURL)
		if err != nil {
			return errors.Wrapf(err, "HEAD %s", isoURL)
		}

		_, params, err := mime.ParseMediaType(resp.Header.Get("Content-Disposition"))
		if err != nil {
			return errors.Wrapf(err, "ParseMediaType %s", resp.Header.Get("Content-Disposition"))
		}

		dst = filepath.Join(detect.ISOCacheDir(), params["filename"])

		isWindowsISO = false
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0777); err != nil {
		return fmt.Errorf("making cache image directory: %s: %w", dst, err)
	}
	spec := lock.PathMutexSpec(dst)
	spec.Timeout = 10 * time.Minute
	klog.Infof("acquiring lock: %+v", spec)
	releaser, err := lock.Acquire(spec)
	if err != nil {
		return fmt.Errorf("unable to acquire lock for %+v: %w", spec, err)
	}
	defer releaser.Release()

	if _, err := os.Stat(dst); err == nil {
		return nil
	}

	out.Step(style.ISODownload, "Downloading VM boot image ...")

	urlWithChecksum := isoURL + "?checksum=file:" + isoURL + ".sha256"
	if skipChecksum {
		urlWithChecksum = isoURL
	}

	return download(urlWithChecksum, dst)
}
