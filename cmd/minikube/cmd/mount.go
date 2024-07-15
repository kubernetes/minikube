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
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/detect"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/style"
	pkgnetwork "k8s.io/minikube/pkg/network"
	"k8s.io/minikube/pkg/util/lock"
	"k8s.io/minikube/third_party/go9p/ufs"
)

const (
	// nineP is the value of --type used for the 9p filesystem.
	nineP                     = "9p"
	defaultMount9PVersion     = "9p2000.L"
	mount9PVersionDescription = "Specify the 9p version that the mount should use"
	defaultMountGID           = "docker"
	mountGIDDescription       = "Default group id used for the mount"
	defaultMountIP            = ""
	mountIPDescription        = "Specify the ip that the mount should be setup on"
	defaultMountMSize         = 262144
	mountMSizeDescription     = "The number of bytes to use for 9p packet payload"
	mountOptionsDescription   = "Additional mount options, such as cache=fscache"
	defaultMountPort          = 0
	mountPortDescription      = "Specify the port that the mount should be setup on, where 0 means any free port."
	defaultMountType          = nineP
	mountTypeDescription      = "Specify the mount filesystem type (supported types: 9p)"
	defaultMountUID           = "docker"
	mountUIDDescription       = "Default user id used for the mount"
)

func defaultMountOptions() []string {
	return []string{}
}

// placeholders for flag values
var (
	mountIP      string
	mountPort    uint16
	mountVersion string
	mountType    string
	isKill       bool
	uid          string
	gid          string
	mSize        int
	options      []string
)

// supportedFilesystems is a map of filesystem types to not warn against.
var supportedFilesystems = map[string]bool{nineP: true}

// mountCmd represents the mount command
var mountCmd = &cobra.Command{
	Use:   "mount [flags] <source directory>:<target directory>",
	Short: "Mounts the specified directory into minikube",
	Long:  `Mounts the specified directory into minikube.`,
	Run: func(_ *cobra.Command, args []string) {
		if isKill {
			if err := killMountProcess(); err != nil {
				exit.Error(reason.HostKillMountProc, "Error killing mount process", err)
			}
			os.Exit(0)
		}

		if len(args) != 1 {
			exit.Message(reason.Usage, `Please specify the directory to be mounted: 
	minikube mount <source directory>:<target directory>   (example: "/host-home:/vm-home")`)
		}
		mountString := args[0]
		idx := strings.LastIndex(mountString, ":")
		if idx == -1 { // no ":" was present
			exit.Message(reason.Usage, `mount argument "{{.value}}" must be in form: <source directory>:<target directory>`, out.V{"value": mountString})
		}
		hostPath := mountString[:idx]
		vmPath := mountString[idx+1:]
		if _, err := os.Stat(hostPath); err != nil {
			if os.IsNotExist(err) {
				exit.Message(reason.HostPathMissing, "Cannot find directory {{.path}} for mount", out.V{"path": hostPath})
			} else {
				exit.Error(reason.HostPathStat, "stat failed", err)
			}
		}
		if len(vmPath) == 0 || !strings.HasPrefix(vmPath, "/") {
			exit.Message(reason.Usage, "Target directory {{.path}} must be an absolute path", out.V{"path": vmPath})
		}
		var debugVal int
		if klog.V(1).Enabled() {
			debugVal = 1 // ufs.StartServer takes int debug param
		}

		co := mustload.Running(ClusterFlagValue())
		if co.CP.Host.Driver.DriverName() == driver.None {
			exit.Message(reason.Usage, `'none' driver does not support 'minikube mount' command`)
		}
		if driver.IsQEMU(co.Config.Driver) && pkgnetwork.IsBuiltinQEMU(co.Config.Network) {
			msg := "minikube mount is not currently implemented with the builtin network on QEMU"
			if runtime.GOOS == "darwin" {
				msg += ", try starting minikube with '--network=socket_vmnet'"
			}
			exit.Message(reason.Unimplemented, msg)
		}

		var ip net.IP
		var err error
		if mountIP == "" {
			if detect.IsMicrosoftWSL() {
				klog.Infof("Selecting IP for WSL. This may be incorrect...")
				ip, err = func() (net.IP, error) {
					conn, err := net.Dial("udp", "8.8.8.8:80")
					if err != nil {
						return nil, err
					}
					defer conn.Close()
					return conn.LocalAddr().(*net.UDPAddr).IP, nil
				}()
			} else {
				ip, err = cluster.HostIP(co.CP.Host, co.Config.Name)
			}
			if err != nil {
				exit.Error(reason.IfHostIP, "Error getting the host IP address to use from within the VM", err)
			}
		} else {
			ip = net.ParseIP(mountIP)
			if ip == nil {
				exit.Message(reason.IfMountIP, "error parsing the input ip address for mount")
			}
		}
		port, err := getPort()
		if err != nil {
			exit.Error(reason.IfMountPort, "Error finding port for mount", err)
		}

		cfg := &cluster.MountConfig{
			Type:    mountType,
			UID:     uid,
			GID:     gid,
			Version: mountVersion,
			MSize:   mSize,
			Port:    port,
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

		if runtime.GOOS == "linux" && !detect.IsNinePSupported() {
			exit.Message(reason.HostUnsupported, "The host does not support filesystem 9p.")

		}

		// An escape valve to allow future hackers to try NFS, VirtFS, or other FS types.
		if !supportedFilesystems[cfg.Type] {
			out.WarningT("{{.type}} is not yet a supported filesystem. We will try anyways!", out.V{"type": cfg.Type})
		}

		bindIP := ip.String() // the ip to listen on the user's host machine
		if driver.IsKIC(co.CP.Host.Driver.DriverName()) && runtime.GOOS != "linux" {
			bindIP = "127.0.0.1"
		}
		out.Step(style.Mounting, "Mounting host path {{.sourcePath}} into VM as {{.destinationPath}} ...", out.V{"sourcePath": hostPath, "destinationPath": vmPath})
		out.Infof("Mount type:   {{.name}}", out.V{"name": cfg.Type})
		out.Infof("User ID:      {{.userID}}", out.V{"userID": cfg.UID})
		out.Infof("Group ID:     {{.groupID}}", out.V{"groupID": cfg.GID})
		out.Infof("Version:      {{.version}}", out.V{"version": cfg.Version})
		out.Infof("Message Size: {{.size}}", out.V{"size": cfg.MSize})
		out.Infof("Options:      {{.options}}", out.V{"options": cfg.Options})
		out.Infof("Bind Address: {{.Address}}", out.V{"Address": net.JoinHostPort(bindIP, fmt.Sprint(port))})

		var wg sync.WaitGroup
		pidchan := make(chan int)
		if cfg.Type == nineP {
			wg.Add(1)
			go func(pid chan int) {
				pid <- os.Getpid()
				out.Styled(style.Fileserver, "Userspace file server: ")
				ufs.StartServer(net.JoinHostPort(bindIP, strconv.Itoa(port)), debugVal, hostPath)
				out.Step(style.Stopped, "Userspace file server is shutdown")
				wg.Done()
			}(pidchan)
		}
		pid := <-pidchan

		// Unmount if Ctrl-C or kill request is received.
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			for sig := range c {
				out.Step(style.Unmount, "Unmounting {{.path}} ...", out.V{"path": vmPath})
				err := cluster.Unmount(co.CP.Runner, vmPath)
				if err != nil {
					out.FailureT("Failed unmount: {{.error}}", out.V{"error": err})
				}

				err = removePidFromFile(pid)
				if err != nil {
					out.FailureT("Failed removing pid from pidfile: {{.error}}", out.V{"error": err})
				}

				exit.Message(reason.Interrupted, "Received {{.name}} signal", out.V{"name": sig})
			}
		}()

		err = cluster.Mount(co.CP.Runner, ip.String(), vmPath, cfg, pid)
		if err != nil {
			if rtErr, ok := err.(*cluster.MountError); ok && rtErr.ErrorType == cluster.MountErrorConnect {
				exit.Error(reason.GuestMountCouldNotConnect, "mount could not connect", rtErr)
			}
			exit.Error(reason.GuestMount, "mount failed", err)
		}
		out.Step(style.Success, "Successfully mounted {{.sourcePath}} to {{.destinationPath}}", out.V{"sourcePath": hostPath, "destinationPath": vmPath})
		out.Ln("")
		out.Styled(style.Notice, "NOTE: This process must stay alive for the mount to be accessible ...")
		wg.Wait()
	},
}

func init() {
	mountCmd.Flags().StringVar(&mountIP, constants.MountIPFlag, defaultMountIP, mountIPDescription)
	mountCmd.Flags().Uint16Var(&mountPort, constants.MountPortFlag, defaultMountPort, mountPortDescription)
	mountCmd.Flags().StringVar(&mountType, constants.MountTypeFlag, defaultMountType, mountTypeDescription)
	mountCmd.Flags().StringVar(&mountVersion, constants.Mount9PVersionFlag, defaultMount9PVersion, mount9PVersionDescription)
	mountCmd.Flags().BoolVar(&isKill, "kill", false, "Kill the mount process spawned by minikube start")
	mountCmd.Flags().StringVar(&uid, constants.MountUIDFlag, defaultMountUID, mountUIDDescription)
	mountCmd.Flags().StringVar(&gid, constants.MountGIDFlag, defaultMountGID, mountGIDDescription)
	mountCmd.Flags().StringSliceVar(&options, constants.MountOptionsFlag, defaultMountOptions(), mountOptionsDescription)
	mountCmd.Flags().IntVar(&mSize, constants.MountMSizeFlag, defaultMountMSize, mountMSizeDescription)
}

// getPort uses the requested port or asks the kernel for a free open port that is ready to use
func getPort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("localhost:%d", mountPort))
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

// removePidFromFile looks at the default locations for the mount-pids file,
// for the profile in use. If a file is found and its content shows PID, PID gets removed.
func removePidFromFile(pid int) error {
	profile := ClusterFlagValue()
	paths := []string{
		localpath.MiniPath(), // legacy mount-process path for backwards compatibility
		localpath.Profile(profile),
	}

	for _, path := range paths {
		err := removePid(path, strconv.Itoa(pid))
		if err != nil {
			return err
		}
	}

	return nil
}

// removePid reads the file at PATH and tries to remove PID from it if found
func removePid(path string, pid string) error {
	// is it the file we're looking for?
	pidPath := filepath.Join(path, constants.MountProcessFileName)
	if _, err := os.Stat(pidPath); os.IsNotExist(err) {
		return nil
	}

	// we found the correct file
	// we're reading the pids...
	out, err := os.ReadFile(pidPath)
	if err != nil {
		return errors.Wrap(err, "readFile")
	}

	pids := []string{}
	// we're splitting the mount-pids file content into a slice of strings
	// so that we can compare each to the PID we're looking for
	strPids := strings.Fields(string(out))
	for _, p := range strPids {
		// If we find the PID, we don't add it to the slice
		if p == pid {
			continue
		}

		// if p doesn't correspond to PID, we add to a list
		pids = append(pids, p)
	}

	// we write the slice that we obtained back to the mount-pids file
	newPids := strings.Join(pids, " ")
	return lock.WriteFile(pidPath, []byte(newPids), 0o644)
}
