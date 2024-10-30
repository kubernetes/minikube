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
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/juju/mutex/v2"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/detect"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/style"
	"k8s.io/minikube/pkg/util/lock"
	"k8s.io/minikube/pkg/version"
)

const fileScheme = "file"

// DefaultISOURLs returns a list of ISO URL's to consult by default, in priority order
func DefaultISOURLs() []string {
	v := version.GetISOVersion()
	isoBucket := "minikube-builds/iso/19883"

	return []string{
		fmt.Sprintf("https://storage.googleapis.com/%s/minikube-%s-%s.iso", isoBucket, v, runtime.GOARCH),
		fmt.Sprintf("https://github.com/kubernetes/minikube/releases/download/%s/minikube-%s-%s.iso", v, v, runtime.GOARCH),
		fmt.Sprintf("https://kubernetes.oss-cn-hangzhou.aliyuncs.com/minikube/iso/minikube-%s-%s.iso", v, runtime.GOARCH),
	}
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
func fileURI(path string) string {
	return "file://" + filepath.ToSlash(path)
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

// downloadISO downloads an ISO URL
func downloadISO(isoURL string, skipChecksum bool) error {
	u, err := url.Parse(isoURL)
	if err != nil {
		return errors.Wrapf(err, "url.parse %q", isoURL)
	}

	// It's already downloaded
	if u.Scheme == fileScheme {
		return nil
	}

	// Lock before we check for existence to avoid thundering herd issues
	dst := localISOPath(u)
	if err := os.MkdirAll(filepath.Dir(dst), 0777); err != nil {
		return errors.Wrapf(err, "making cache image directory: %s", dst)
	}
	spec := lock.PathMutexSpec(dst)
	spec.Timeout = 10 * time.Minute
	klog.Infof("acquiring lock: %+v", spec)
	releaser, err := mutex.Acquire(spec)
	if err != nil {
		return errors.Wrapf(err, "unable to acquire lock for %+v", spec)
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
