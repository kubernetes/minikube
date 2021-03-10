/*
Copyright 2017 The Kubernetes Authors All rights reserved.

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

package cmd

import (
	"io"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/reason"
	docker "k8s.io/minikube/third_party/go-dockerclient"
)

// imageCmd represents the image command
var imageCmd = &cobra.Command{
	Use:   "image",
	Short: "Work with images in minikube",
	Long:  "Work with images in minikube",
}

var (
	tag        string
	push       bool
	dockerFile string
)

// loadImageCmd represents the image load command
var loadImageCmd = &cobra.Command{
	Use:   "load",
	Short: "Load a local image into minikube",
	Long:  "Load a local image into minikube",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			exit.Message(reason.Usage, "Please provide an image in your local daemon to load into minikube via <minikube image load IMAGE_NAME>")
		}
		// Cache and load images into docker daemon
		profile, err := config.LoadProfile(viper.GetString(config.ProfileName))
		if err != nil {
			exit.Error(reason.Usage, "loading profile", err)
		}
		img := args[0]
		if err := machine.CacheAndLoadImages([]string{img}, []*config.Profile{profile}); err != nil {
			exit.Error(reason.GuestImageLoad, "Failed to load image", err)
		}
	},
}

func createTar(dir string) (string, error) {
	tmp, err := ioutil.TempFile("", "build.*.tar")
	if err != nil {
		return "", err
	}
	tar, err := docker.CreateTarStream(dir, dockerFile)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(tmp, tar)
	if err != nil {
		return "", err
	}
	err = tmp.Close()
	if err != nil {
		return "", err
	}

	return tmp.Name(), nil
}

// buildImageCmd represents the image build command
var buildImageCmd = &cobra.Command{
	Use:     "build",
	Short:   "Build a container image in minikube",
	Long:    "Build a container image, using the container runtime.",
	Example: `minikube image build .`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			exit.Message(reason.Usage, "Please provide a path to build")
		}
		// Cache and load images into docker daemon
		profile, err := config.LoadProfile(viper.GetString(config.ProfileName))
		if err != nil {
			exit.Error(reason.Usage, "loading profile", err)
		}
		img := args[0]
		var tmp string
		info, err := os.Stat(img)
		if err == nil && info.IsDir() {
			tmp, err := createTar(img)
			if err != nil {
				exit.Error(reason.GuestImageBuild, "Failed to build image", err)
			}
			img = tmp
		}
		if err := machine.BuildImage(img, dockerFile, tag, push, []*config.Profile{profile}); err != nil {
			exit.Error(reason.GuestImageBuild, "Failed to build image", err)
		}
		if tmp != "" {
			os.Remove(tmp)
		}
	},
}

func init() {
	imageCmd.AddCommand(loadImageCmd)
	buildImageCmd.Flags().StringVarP(&tag, "tag", "t", "", "Tag to apply to the new image (optional)")
	buildImageCmd.Flags().BoolVarP(&push, "push", "", false, "Push the new image (requires tag)")
	buildImageCmd.Flags().StringVarP(&dockerFile, "file", "f", "", "Path to the Dockerfile to use (optional)")
	imageCmd.AddCommand(buildImageCmd)
}
