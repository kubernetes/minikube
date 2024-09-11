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
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/client"
	"github.com/docker/go-units"
	"github.com/docker/machine/libmachine/state"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v2"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/detect"
	"k8s.io/minikube/pkg/minikube/image"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/vmpath"
)

// loadRoot is where images should be loaded from within the guest VM
var loadRoot = path.Join(vmpath.GuestPersistentDir, "images")

// loadImageLock is used to serialize image loads to avoid overloading the guest VM
var loadImageLock sync.Mutex

// saveRoot is where images should be saved from within the guest VM
var saveRoot = path.Join(vmpath.GuestPersistentDir, "images")

// CacheImagesForBootstrapper will cache images for a bootstrapper
func CacheImagesForBootstrapper(imageRepository, version string) error {
	images, err := bootstrapper.GetCachedImageList(imageRepository, version)
	if err != nil {
		return errors.Wrap(err, "cached images list")
	}

	if err := image.SaveToDir(images, detect.ImageCacheDir(), false); err != nil {
		return errors.Wrap(err, "Caching images")
	}

	return nil
}

// LoadCachedImages loads previously cached images into the container runtime
func LoadCachedImages(cc *config.ClusterConfig, runner command.Runner, images []string, cacheDir string, overwrite bool) error {
	cr, err := cruntime.New(cruntime.Config{Type: cc.KubernetesConfig.ContainerRuntime, Runner: runner})
	if err != nil {
		return errors.Wrap(err, "runtime")
	}

	// Skip loading images if images already exist
	if !overwrite && cr.ImagesPreloaded(images) {
		klog.Infof("Images are preloaded, skipping loading")
		return nil
	}

	klog.Infof("LoadCachedImages start: %s", images)
	start := time.Now()

	defer func() {
		klog.Infof("duration metric: took %s to LoadCachedImages", time.Since(start))
	}()

	var g errgroup.Group

	var imgClient *client.Client
	if cr.Name() == "Docker" {
		imgClient, err = client.NewClientWithOpts(client.FromEnv) // image client
		if err != nil {
			klog.Infof("couldn't get a local image daemon which might be ok: %v", err)
		}
	}

	for _, image := range images {
		image := image
		g.Go(func() error {
			// Put a ten second limit on deciding if an image needs transfer
			// because it takes much less than that time to just transfer the image.
			// This is needed because if running in offline mode, we can spend minutes here
			// waiting for i/o timeout.
			err := timedNeedsTransfer(imgClient, image, cr, 10*time.Second)
			if err == nil {
				return nil
			}
			klog.Infof("%q needs transfer: %v", image, err)
			return transferAndLoadCachedImage(runner, cc.KubernetesConfig, image, cacheDir)
		})
	}
	if err := g.Wait(); err != nil {
		return errors.Wrap(err, "LoadCachedImages")
	}
	klog.Infoln("Successfully loaded all cached images")
	return nil
}

func timedNeedsTransfer(imgClient *client.Client, imgName string, cr cruntime.Manager, t time.Duration) error {
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(t)
		timeout <- true
	}()

	transferFinished := make(chan bool, 1)
	var err error
	go func() {
		err = needsTransfer(imgClient, imgName, cr)
		transferFinished <- true
	}()

	select {
	case <-transferFinished:
		return err
	case <-timeout:
		return fmt.Errorf("needs transfer timed out in %f seconds", t.Seconds())
	}
}

// needsTransfer returns an error if an image needs to be retransferred
func needsTransfer(imgClient *client.Client, imgName string, cr cruntime.Manager) error {
	imgDgst := ""         // for instance sha256:7c92a2c6bbcb6b6beff92d0a940779769c2477b807c202954c537e2e0deb9bed
	if imgClient != nil { // if possible try to get img digest from Client lib which is 4s faster.
		imgDgst = image.DigestByDockerLib(imgClient, imgName)
		if imgDgst != "" {
			if !cr.ImageExists(imgName, imgDgst) {
				return fmt.Errorf("%q does not exist at hash %q in container runtime", imgName, imgDgst)
			}
			return nil
		}
	}
	// if not found with method above try go-container lib (which is 4s slower)
	imgDgst = image.DigestByGoLib(imgName)
	if imgDgst == "" {
		return fmt.Errorf("got empty img digest %q for %s", imgDgst, imgName)
	}
	if !cr.ImageExists(imgName, imgDgst) {
		return fmt.Errorf("%q does not exist at hash %q in container runtime", imgName, imgDgst)
	}
	return nil
}

// LoadLocalImages loads images into the container runtime
func LoadLocalImages(cc *config.ClusterConfig, runner command.Runner, images []string) error {
	var g errgroup.Group
	for _, image := range images {
		image := image
		g.Go(func() error {
			return transferAndLoadImage(runner, cc.KubernetesConfig, image, image)
		})
	}
	if err := g.Wait(); err != nil {
		return errors.Wrap(err, "loading images")
	}
	klog.Infoln("Successfully loaded all images")
	return nil
}

// CacheAndLoadImages caches and loads images to all profiles
func CacheAndLoadImages(images []string, profiles []*config.Profile, overwrite bool) error {
	if len(images) == 0 {
		return nil
	}

	// This is the most important thing
	if err := image.SaveToDir(images, detect.ImageCacheDir(), overwrite); err != nil {
		return errors.Wrap(err, "save to dir")
	}

	return DoLoadImages(images, profiles, detect.ImageCacheDir(), overwrite)
}

// DoLoadImages loads images to all profiles
func DoLoadImages(images []string, profiles []*config.Profile, cacheDir string, overwrite bool) error {
	api, err := NewAPIClient()
	if err != nil {
		return errors.Wrap(err, "api")
	}
	defer api.Close()

	succeeded := []string{}
	failed := []string{}

	for _, p := range profiles { // loading images to all running profiles
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

			if status == state.Running.String() { // the not running hosts will load on next start
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
				if cacheDir != "" {
					// loading image names, from cache
					err = LoadCachedImages(c, cr, images, cacheDir, overwrite)
				} else {
					// loading image files
					err = LoadLocalImages(c, cr, images)
				}
				if err != nil {
					failed = append(failed, m)
					klog.Warningf("Failed to load cached images for %q: %v", pName, err)
					continue
				}
				succeeded = append(succeeded, m)
			}
		}
	}

	if len(succeeded) > 0 {
		klog.Infof("succeeded pushing to: %s", strings.Join(succeeded, " "))
	}
	if len(failed) > 0 {
		klog.Infof("failed pushing to: %s", strings.Join(failed, " "))
	}
	// Live pushes are not considered a failure
	return nil
}

// transferAndLoadCachedImage transfers and loads a single image from the cache
func transferAndLoadCachedImage(cr command.Runner, k8s config.KubernetesConfig, imgName string, cacheDir string) error {
	src := filepath.Join(cacheDir, imgName)
	src = localpath.SanitizeCacheDir(src)
	return transferAndLoadImage(cr, k8s, src, imgName)
}

// transferAndLoadImage transfers and loads a single image
func transferAndLoadImage(cr command.Runner, k8s config.KubernetesConfig, src string, imgName string) error {
	r, err := cruntime.New(cruntime.Config{Type: k8s.ContainerRuntime, Runner: cr})
	if err != nil {
		return errors.Wrap(err, "runtime")
	}

	if err := removeExistingImage(r, src, imgName); err != nil {
		return err
	}

	klog.Infof("Loading image from: %s", src)
	filename := filepath.Base(src)
	if _, err := os.Stat(src); err != nil {
		return err
	}

	dst := path.Join(loadRoot, filename)
	f, err := assets.NewFileAsset(src, loadRoot, filename, "0644")
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

	loadImageLock.Lock()
	defer loadImageLock.Unlock()

	err = r.LoadImage(dst)
	if err != nil {
		if strings.Contains(err.Error(), "ctr: image might be filtered out") {
			out.WarningT("The image '{{.imageName}}' does not match arch of the container runtime, use a multi-arch image instead", out.V{"imageName": imgName})
		}
		return errors.Wrapf(err, "%s load %s", r.Name(), dst)
	}

	klog.Infof("Transferred and loaded %s from cache", src)
	return nil
}

func removeExistingImage(r cruntime.Manager, src string, imgName string) error {
	// if loading an image from tar, skip deleting as we don't have the actual image name
	// ie. imgName = "C:\this_is_a_dir\image.tar.gz"
	if src == imgName {
		return nil
	}

	err := r.RemoveImage(imgName)
	if err == nil {
		return nil
	}

	errStr := strings.ToLower(err.Error())
	if !strings.Contains(errStr, "no such image") && !strings.Contains(errStr, "unable to remove the image") {
		return errors.Wrap(err, "removing image")
	}

	return nil
}

// SaveCachedImages saves from the container runtime to the cache
func SaveCachedImages(cc *config.ClusterConfig, runner command.Runner, images []string, cacheDir string) error {
	klog.Infof("SaveCachedImages start: %s", images)
	start := time.Now()

	defer func() {
		klog.Infof("duration metric: took %s to SaveCachedImages", time.Since(start))
	}()

	var g errgroup.Group

	for _, image := range images {
		image := image
		g.Go(func() error {
			return transferAndSaveCachedImage(runner, cc.KubernetesConfig, image, cacheDir)
		})
	}
	if err := g.Wait(); err != nil {
		return errors.Wrap(err, "saving cached images")
	}
	klog.Infoln("Successfully saved all cached images")
	return nil
}

// SaveLocalImages saves images from the container runtime
func SaveLocalImages(cc *config.ClusterConfig, runner command.Runner, images []string, output string) error {
	var g errgroup.Group
	for _, image := range images {
		image := image
		g.Go(func() error {
			return transferAndSaveImage(runner, cc.KubernetesConfig, output, image)
		})
	}
	if err := g.Wait(); err != nil {
		return errors.Wrap(err, "saving images")
	}
	klog.Infoln("Successfully saved all images")
	return nil
}

// SaveAndCacheImages saves images from all profiles into the cache
func SaveAndCacheImages(images []string, profiles []*config.Profile) error {
	if len(images) == 0 {
		return nil
	}

	return DoSaveImages(images, "", profiles, detect.ImageCacheDir())
}

// DoSaveImages saves images from all profiles
func DoSaveImages(images []string, output string, profiles []*config.Profile, cacheDir string) error {
	api, err := NewAPIClient()
	if err != nil {
		return errors.Wrap(err, "api")
	}
	defer api.Close()

	klog.Infof("Save images: %q", images)

	succeeded := []string{}
	failed := []string{}

	for _, p := range profiles { // loading images to all running profiles
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

			if status == state.Running.String() { // the not running hosts will load on next start
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
				if cacheDir != "" {
					// saving image names, to cache
					err = SaveCachedImages(c, cr, images, cacheDir)
				} else {
					// saving mage files
					err = SaveLocalImages(c, cr, images, output)
				}
				if err != nil {
					failed = append(failed, m)
					klog.Warningf("Failed to load cached images for profile %s. make sure the profile is running. %v", pName, err)
					continue
				}
				succeeded = append(succeeded, m)
			}
		}
	}

	klog.Infof("succeeded pulling from : %s", strings.Join(succeeded, " "))
	klog.Infof("failed pulling from : %s", strings.Join(failed, " "))
	// Live pushes are not considered a failure
	return nil
}

// transferAndSaveCachedImage transfers and loads a single image from the cache
func transferAndSaveCachedImage(cr command.Runner, k8s config.KubernetesConfig, imgName string, cacheDir string) error {
	dst := filepath.Join(cacheDir, imgName)
	dst = localpath.SanitizeCacheDir(dst)
	return transferAndSaveImage(cr, k8s, dst, imgName)
}

// transferAndSaveImage transfers and loads a single image
func transferAndSaveImage(cr command.Runner, k8s config.KubernetesConfig, dst string, imgName string) error {
	r, err := cruntime.New(cruntime.Config{Type: k8s.ContainerRuntime, Runner: cr})
	if err != nil {
		return errors.Wrap(err, "runtime")
	}
	found := false
	// the reason why we are doing this is that
	// unlike other tool, podman assume the image has a localhost registry (not docker.io)
	// if the image is loaded with a tarball without a registry specified in tag
	// see https://github.com/containers/podman/issues/15974
	tryImageExist := []string{imgName, cruntime.AddDockerIO(imgName), cruntime.AddLocalhostPrefix(imgName)}
	for _, imgName = range tryImageExist {
		if r.ImageExists(imgName, "") {
			found = true
			break
		}
	}

	if !found {
		// if all of this failed, return the original error
		return fmt.Errorf("image not found %s", imgName)
	}

	klog.Infof("Saving image to: %s", dst)
	filename := filepath.Base(dst)

	_, err = os.OpenFile(dst, os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		return err
	}

	f, err := assets.NewFileAsset(dst, saveRoot, filename, "0644")
	if err != nil {
		return errors.Wrapf(err, "creating copyable file asset: %s", filename)
	}
	defer func() {
		if err := f.Close(); err != nil {
			klog.Warningf("error closing the file %s: %v", f.GetSourcePath(), err)
		}
	}()

	src := path.Join(saveRoot, filename)
	args := append([]string{"rm", "-f"}, src)
	if _, err := cr.RunCmd(exec.Command("sudo", args...)); err != nil {
		return err
	}
	err = r.SaveImage(imgName, src)
	if err != nil {
		return errors.Wrapf(err, "%s save %s", r.Name(), src)
	}

	if err := cr.CopyFrom(f); err != nil {
		return errors.Wrap(err, "transferring cached image")
	}

	klog.Infof("Transferred and saved %s to cache", dst)
	return nil
}

// pullImages pulls images to the container run time
func pullImages(cruntime cruntime.Manager, images []string) error {
	klog.Infof("pullImages start: %s", images)
	start := time.Now()

	defer func() {
		klog.Infof("duration metric: took %s to pullImages", time.Since(start))
	}()

	var g errgroup.Group

	for _, image := range images {
		image := image
		g.Go(func() error {
			return cruntime.PullImage(image)
		})
	}
	if err := g.Wait(); err != nil {
		return errors.Wrap(err, "error pulling images")
	}
	klog.Infoln("Successfully pulled images")
	return nil
}

// PullImages pulls images to all nodes in profile
func PullImages(images []string, profile *config.Profile) error {
	api, err := NewAPIClient()
	if err != nil {
		return errors.Wrap(err, "error creating api client")
	}
	defer api.Close()

	succeeded := []string{}
	failed := []string{}

	pName := profile.Name

	c, err := config.Load(pName)
	if err != nil {
		klog.Errorf("Failed to load profile %q: %v", pName, err)
		return errors.Wrapf(err, "error loading config for profile :%v", pName)
	}

	for _, n := range c.Nodes {
		m := config.MachineName(*c, n)

		status, err := Status(api, m)
		if err != nil {
			klog.Warningf("error getting status for %s: %v", m, err)
			continue
		}

		if status == state.Running.String() {
			h, err := api.Load(m)
			if err != nil {
				klog.Warningf("Failed to load machine %q: %v", m, err)
				continue
			}
			runner, err := CommandRunner(h)
			if err != nil {
				return err
			}
			cruntime, err := cruntime.New(cruntime.Config{Type: c.KubernetesConfig.ContainerRuntime, Runner: runner})
			if err != nil {
				return errors.Wrap(err, "error creating container runtime")
			}
			err = pullImages(cruntime, images)
			if err != nil {
				failed = append(failed, m)
				klog.Warningf("Failed to pull images for profile %s %v", pName, err.Error())
				continue
			}
			succeeded = append(succeeded, m)
		}
	}

	klog.Infof("succeeded pulling to: %s", strings.Join(succeeded, " "))
	klog.Infof("failed pulling to: %s", strings.Join(failed, " "))
	return nil
}

// removeImages removes images from the container run time
func removeImages(cruntime cruntime.Manager, images []string) error {
	klog.Infof("removeImages start: %s", images)
	start := time.Now()

	defer func() {
		klog.Infof("duration metric: took %s to removeImages", time.Since(start))
	}()

	var g errgroup.Group

	for _, image := range images {
		image := image
		g.Go(func() error {
			return cruntime.RemoveImage(image)
		})
	}
	if err := g.Wait(); err != nil {
		return errors.Wrap(err, "error removing images")
	}
	klog.Infoln("Successfully removed images")
	return nil
}

// RemoveImages removes images from all nodes in profile
func RemoveImages(images []string, profile *config.Profile) error {
	api, err := NewAPIClient()
	if err != nil {
		return errors.Wrap(err, "error creating api client")
	}
	defer api.Close()

	succeeded := []string{}
	failed := []string{}

	pName := profile.Name

	c, err := config.Load(pName)
	if err != nil {
		klog.Errorf("Failed to load profile %q: %v", pName, err)
		return errors.Wrapf(err, "error loading config for profile :%v", pName)
	}

	for _, n := range c.Nodes {
		m := config.MachineName(*c, n)

		status, err := Status(api, m)
		if err != nil {
			klog.Warningf("error getting status for %s: %v", m, err)
			continue
		}

		if status == state.Running.String() {
			h, err := api.Load(m)
			if err != nil {
				klog.Warningf("Failed to load machine %q: %v", m, err)
				continue
			}
			runner, err := CommandRunner(h)
			if err != nil {
				return err
			}
			cruntime, err := cruntime.New(cruntime.Config{Type: c.KubernetesConfig.ContainerRuntime, Runner: runner})
			if err != nil {
				return errors.Wrap(err, "error creating container runtime")
			}
			err = removeImages(cruntime, images)
			if err != nil {
				failed = append(failed, m)
				klog.Warningf("Failed to remove images for profile %s %v", pName, err.Error())
				out.WarningT("Failed to remove images for profile {{.pName}} {{.error}}", out.V{"pName": pName, "error": err.Error()})
				continue
			}
			succeeded = append(succeeded, m)
		}
	}

	klog.Infof("succeeded removing from: %s", strings.Join(succeeded, " "))
	klog.Infof("failed removing from: %s", strings.Join(failed, " "))
	return nil
}

// ListImages lists images on all nodes in profile
func ListImages(profile *config.Profile, format string) error {
	api, err := NewAPIClient()
	if err != nil {
		return errors.Wrap(err, "error creating api client")
	}
	defer api.Close()

	pName := profile.Name

	c, err := config.Load(pName)
	if err != nil {
		klog.Errorf("Failed to load profile %q: %v", pName, err)
		return errors.Wrapf(err, "error loading config for profile :%v", pName)
	}

	imageListsFromNodes := [][]cruntime.ListImage{}
	for _, n := range c.Nodes {
		m := config.MachineName(*c, n)

		status, err := Status(api, m)
		if err != nil {
			klog.Warningf("error getting status for %s: %v", m, err)
			continue
		}

		if status == state.Running.String() {
			h, err := api.Load(m)
			if err != nil {
				klog.Warningf("Failed to load machine %q: %v", m, err)
				continue
			}
			runner, err := CommandRunner(h)
			if err != nil {
				return err
			}
			cr, err := cruntime.New(cruntime.Config{Type: c.KubernetesConfig.ContainerRuntime, Runner: runner})
			if err != nil {
				return errors.Wrap(err, "error creating container runtime")
			}
			list, err := cr.ListImages(cruntime.ListImagesOptions{})
			if err != nil {
				klog.Warningf("Failed to list images for profile %s %v", pName, err.Error())
				continue
			}
			imageListsFromNodes = append(imageListsFromNodes, list)

		}
	}

	uniqueImages := mergeImageLists(imageListsFromNodes)

	switch format {
	case "table":
		var data [][]string
		for _, item := range uniqueImages {
			imageSize := humanImageSize(item.Size)
			id := parseImageID(item.ID)
			for _, img := range item.RepoTags {
				imageName, tag := parseRepoTag(img)
				if imageName == "" {
					continue
				}
				data = append(data, []string{imageName, tag, id, imageSize})
			}
		}
		renderImagesTable(data)
	case "json":
		json, err := json.Marshal(uniqueImages)
		if err != nil {
			klog.Warningf("Error marshalling images list: %v", err.Error())
			return nil
		}
		fmt.Printf("%s\n", json)
	case "yaml":
		yaml, err := yaml.Marshal(uniqueImages)
		if err != nil {
			klog.Warningf("Error marshalling images list: %v", err.Error())
			return nil
		}
		fmt.Printf("%s\n", yaml)
	default:
		res := []string{}
		for _, item := range uniqueImages {
			res = append(res, item.RepoTags...)
		}
		sort.Sort(sort.Reverse(sort.StringSlice(res)))
		fmt.Printf("%s\n", strings.Join(res, "\n"))
	}

	return nil
}

// mergeImageLists merges image lists from different nodes into a single list
// all the repo tags of the same image will be preserved and grouped in one image item
func mergeImageLists(lists [][]cruntime.ListImage) []cruntime.ListImage {
	images := map[string]*cruntime.ListImage{}
	// key of this set is img.ID+img.repoTag
	// we use this set to detect whether a tag of a image is included in images map
	imageTagAndIDSet := map[string]struct{}{}
	for _, list := range lists {
		for _, img := range list {
			img := img
			_, existingImg := images[img.ID]
			if !existingImg {
				images[img.ID] = &img
			}
			for _, repoTag := range img.RepoTags {
				if _, ok := imageTagAndIDSet[img.ID+repoTag]; ok {
					continue
				}
				imageTagAndIDSet[img.ID+repoTag] = struct{}{}
				if existingImg {
					images[img.ID].RepoTags = append(images[img.ID].RepoTags, repoTag)
				}
			}
		}

	}
	uniqueImages := []cruntime.ListImage{}
	for _, img := range images {
		uniqueImages = append(uniqueImages, *img)
	}
	return uniqueImages
}

// parseRepoTag splits input string for two parts: image name and image tag
func parseRepoTag(repoTag string) (string, string) {
	idx := strings.LastIndex(repoTag, ":")
	if idx == -1 {
		return "", ""
	}
	return repoTag[:idx], repoTag[idx+1:]
}

// parseImageID truncates image id
func parseImageID(id string) string {
	maxImageIDLen := 13
	if len(id) > maxImageIDLen {
		return id[:maxImageIDLen]
	}
	return id
}

// humanImageSize prints size of image in human readable format
func humanImageSize(imageSize string) string {
	f, err := strconv.ParseFloat(imageSize, 32)
	if err == nil {
		return units.HumanSizeWithPrecision(f, 3)
	}
	return imageSize
}

// renderImagesTable renders pretty table for images list
func renderImagesTable(images [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Image", "Tag", "Image ID", "Size"})
	table.SetAutoFormatHeaders(false)
	table.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: true})
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("|")
	table.AppendBulk(images)
	table.Render()
}

// TagImage tags image in all nodes in profile
func TagImage(profile *config.Profile, source string, target string) error {
	api, err := NewAPIClient()
	if err != nil {
		return errors.Wrap(err, "error creating api client")
	}
	defer api.Close()

	succeeded := []string{}
	failed := []string{}

	pName := profile.Name

	c, err := config.Load(pName)
	if err != nil {
		klog.Errorf("Failed to load profile %q: %v", pName, err)
		return errors.Wrapf(err, "error loading config for profile :%v", pName)
	}

	for _, n := range c.Nodes {
		m := config.MachineName(*c, n)

		status, err := Status(api, m)
		if err != nil {
			klog.Warningf("error getting status for %s: %v", m, err)
			continue
		}

		if status == state.Running.String() {
			h, err := api.Load(m)
			if err != nil {
				klog.Warningf("Failed to load machine %q: %v", m, err)
				continue
			}
			runner, err := CommandRunner(h)
			if err != nil {
				return err
			}
			cruntime, err := cruntime.New(cruntime.Config{Type: c.KubernetesConfig.ContainerRuntime, Runner: runner})
			if err != nil {
				return errors.Wrap(err, "error creating container runtime")
			}
			err = cruntime.TagImage(source, target)
			if err != nil {
				failed = append(failed, m)
				klog.Warningf("Failed to tag image for profile %s %v", pName, err.Error())
				continue
			}
			succeeded = append(succeeded, m)
		}
	}

	klog.Infof("succeeded tagging in: %s", strings.Join(succeeded, " "))
	klog.Infof("failed tagging in: %s", strings.Join(failed, " "))
	return nil
}

// pushImages pushes images from the container run time
func pushImages(cruntime cruntime.Manager, images []string) error {
	klog.Infof("pushImages start: %s", images)
	start := time.Now()

	defer func() {
		klog.Infof("duration metric: took %s to pushImages", time.Since(start))
	}()

	var g errgroup.Group

	for _, image := range images {
		image := image
		g.Go(func() error {
			return cruntime.PushImage(image)
		})
	}
	if err := g.Wait(); err != nil {
		return errors.Wrap(err, "error pushing images")
	}
	klog.Infoln("Successfully pushed images")
	return nil
}

// PushImages push images on all nodes in profile
func PushImages(images []string, profile *config.Profile) error {
	api, err := NewAPIClient()
	if err != nil {
		return errors.Wrap(err, "error creating api client")
	}
	defer api.Close()

	succeeded := []string{}
	failed := []string{}

	pName := profile.Name

	c, err := config.Load(pName)
	if err != nil {
		klog.Errorf("Failed to load profile %q: %v", pName, err)
		return errors.Wrapf(err, "error loading config for profile :%v", pName)
	}

	for _, n := range c.Nodes {
		m := config.MachineName(*c, n)

		status, err := Status(api, m)
		if err != nil {
			klog.Warningf("error getting status for %s: %v", m, err)
			continue
		}

		if status == state.Running.String() {
			h, err := api.Load(m)
			if err != nil {
				klog.Warningf("Failed to load machine %q: %v", m, err)
				continue
			}
			runner, err := CommandRunner(h)
			if err != nil {
				return err
			}
			cruntime, err := cruntime.New(cruntime.Config{Type: c.KubernetesConfig.ContainerRuntime, Runner: runner})
			if err != nil {
				return errors.Wrap(err, "error creating container runtime")
			}
			err = pushImages(cruntime, images)
			if err != nil {
				failed = append(failed, m)
				klog.Warningf("Failed to push image for profile %s %v", pName, err.Error())
				continue
			}
			succeeded = append(succeeded, m)
		}
	}

	klog.Infof("succeeded pushing in: %s", strings.Join(succeeded, " "))
	klog.Infof("failed pushing in: %s", strings.Join(failed, " "))
	return nil
}
