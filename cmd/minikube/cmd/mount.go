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

package cmd

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/exitcode"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/third_party/go9p/ufs"
)

const (
	// nineP is the value of --type used for the 9p filesystem.
	nineP               = "9p"
	defaultMountVersion = "9p2000.L"
	defaultMsize        = 262144
)

// placeholders for flag values
var mountIP string
var mountVersion string
var mountType string
var isKill bool
var uid string
var gid string
var mSize int
var options []string
var mode uint

// supportedFilesystems is a map of filesystem types to not warn against.
var supportedFilesystems = map[string]bool{nineP: true}

// mountCmd represents the mount command
var mountCmd = &cobra.Command{
	Use:   "mount [flags] <source directory>:<target directory>",
	Short: "Mounts the specified directory into minikube",
	Long:  `Mounts the specified directory into minikube.`,
	Run: func(cmd *cobra.Command, args []string) {
		if isKill {
			if err := killMountProcess(); err != nil {
				exit.WithError("Error killing mount process", err)
			}
			os.Exit(0)
		}

		if len(args) != 1 {
			exit.UsageT(`Please specify the directory to be mounted: 
	minikube mount <source directory>:<target directory>   (example: "/host-home:/vm-home")`)
		}
		mountString := args[0]
		idx := strings.LastIndex(mountString, ":")
		if idx == -1 { // no ":" was present
			exit.UsageT(`mount argument "{{.value}}" must be in form: <source directory>:<target directory>`, out.V{"value": mountString})
		}
		hostPath := mountString[:idx]
		vmPath := mountString[idx+1:]
		if _, err := os.Stat(hostPath); err != nil {
			if os.IsNotExist(err) {
				exit.WithCodeT(exitcode.HostError, "Cannot find directory {{.path}} for mount", out.V{"path": hostPath})
			} else {
				exit.WithError("stat failed", err)
			}
		}
		if len(vmPath) == 0 || !strings.HasPrefix(vmPath, "/") {
			exit.UsageT("Target directory {{.path}} must be an absolute path", out.V{"path": vmPath})
		}
		var debugVal int
		if glog.V(1) {
			debugVal = 1 // ufs.StartServer takes int debug param
		}

		co := mustload.Running(ClusterFlagValue())
		if co.CP.Host.Driver.DriverName() == driver.None {
			exit.UsageT(`'none' driver does not support 'minikube mount' command`)
		}

		var ip net.IP
		var err error
		if mountIP == "" {
			ip, err = cluster.HostIP(co.CP.Host)
			if err != nil {
				exit.WithError("Error getting the host IP address to use from within the VM", err)
			}
		} else {
			ip = net.ParseIP(mountIP)
			if ip == nil {
				exit.WithCodeT(exitcode.LocalNetworkConfig, "error parsing the input ip address for mount")
			}
		}
		port, err := getPort()
		if err != nil {
			exit.WithError("Error finding port for mount", err)
		}

		cfg := &cluster.MountConfig{
			Type:    mountType,
			UID:     uid,
			GID:     gid,
			Version: mountVersion,
			MSize:   mSize,
			Port:    port,
			Mode:    os.FileMode(mode),
			Options: map[string]string{},
		}

		for _, o := range options {
			if !strings.Contains(o, "=") {
				cfg.Options[o] = ""
				continue
			}
			parts := strings.Split(o, "=")
			cfg.Options[parts[0]] = parts[1]
		}

		// An escape valve to allow future hackers to try NFS, VirtFS, or other FS types.
		if !supportedFilesystems[cfg.Type] {
			out.WarningT("{{.type}} is not yet a supported filesystem. We will try anyways!", out.V{"type": cfg.Type})
		}

		bindIP := ip.String() // the ip to listen on the user's host machine
		if driver.IsKIC(co.CP.Host.Driver.DriverName()) && runtime.GOOS != "linux" {
			bindIP = "127.0.0.1"
		}
		out.T(out.Mounting, "Mounting host path {{.sourcePath}} into VM as {{.destinationPath}} ...", out.V{"sourcePath": hostPath, "destinationPath": vmPath})
		out.Infof("Mount type:   {{.name}}", out.V{"type": cfg.Type})
		out.Infof("User ID:      {{.userID}}", out.V{"userID": cfg.UID})
		out.Infof("Group ID:     {{.groupID}}", out.V{"groupID": cfg.GID})
		out.Infof("Version:      {{.version}}", out.V{"version": cfg.Version})
		out.Infof("Message Size: {{.size}}", out.V{"size": cfg.MSize})
		out.Infof("Permissions:  {{.octalMode}} ({{.writtenMode}})", out.V{"octalMode": fmt.Sprintf("%o", cfg.Mode), "writtenMode": cfg.Mode})
		out.Infof("Options:      {{.options}}", out.V{"options": cfg.Options})
		out.Infof("Bind Address: {{.Address}}", out.V{"Address": net.JoinHostPort(bindIP, fmt.Sprint(port))})

		var wg sync.WaitGroup
		if cfg.Type == nineP {
			wg.Add(1)
			go func() {
				out.T(out.Fileserver, "Userspace file server: ")
				ufs.StartServer(net.JoinHostPort(bindIP, strconv.Itoa(port)), debugVal, hostPath)
				out.T(out.Stopped, "Userspace file server is shutdown")
				wg.Done()
			}()
		}

		// Unmount if Ctrl-C or kill request is received.
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			for sig := range c {
				out.T(out.Unmount, "Unmounting {{.path}} ...", out.V{"path": vmPath})
				err := cluster.Unmount(co.CP.Runner, vmPath)
				if err != nil {
					out.FailureT("Failed unmount: {{.error}}", out.V{"error": err})
				}
				exit.WithCodeT(exitcode.Interrupted, "Received {{.name}} signal", out.V{"name": sig})
			}
		}()

		err = cluster.Mount(co.CP.Runner, ip.String(), vmPath, cfg)
		if err != nil {
			exit.WithError("mount failed", err)
		}
		out.T(out.SuccessType, "Successfully mounted {{.sourcePath}} to {{.destinationPath}}", out.V{"sourcePath": hostPath, "destinationPath": vmPath})
		out.Ln("")
		out.T(out.Notice, "NOTE: This process must stay alive for the mount to be accessible ...")
		wg.Wait()
	},
}

func init() {
	mountCmd.Flags().StringVar(&mountIP, "ip", "", "Specify the ip that the mount should be setup on")
	mountCmd.Flags().StringVar(&mountType, "type", nineP, "Specify the mount filesystem type (supported types: 9p)")
	mountCmd.Flags().StringVar(&mountVersion, "9p-version", defaultMountVersion, "Specify the 9p version that the mount should use")
	mountCmd.Flags().BoolVar(&isKill, "kill", false, "Kill the mount process spawned by minikube start")
	mountCmd.Flags().StringVar(&uid, "uid", "docker", "Default user id used for the mount")
	mountCmd.Flags().StringVar(&gid, "gid", "docker", "Default group id used for the mount")
	mountCmd.Flags().UintVar(&mode, "mode", 0755, "File permissions used for the mount")
	mountCmd.Flags().StringSliceVar(&options, "options", []string{}, "Additional mount options, such as cache=fscache")
	mountCmd.Flags().IntVar(&mSize, "msize", defaultMsize, "The number of bytes to use for 9p packet payload")
}

// getPort asks the kernel for a free open port that is ready to use
func getPort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return -1, errors.Errorf("Error accessing port %d", addr.Port)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
