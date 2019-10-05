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
	"github.com/danieljoos/wincred"
	"github.com/docker/machine/drivers/hyperv"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/machine"
	pkghyperv "k8s.io/minikube/pkg/minikube/drivers/hyperv"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/third_party/go9p/ufs"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

// nineP is the value of --type used for the 9p filesystem.
const nineP = "9p"

// cifs is the value of --type used for the CIFS FileSystem
const cifs = "cifs"

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
var shareName = "minikube"

// supportedFilesystems is a map of filesystem types to not warn against.
var supportedFilesystems = map[string]bool{nineP: true, cifs: true}

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
				exit.WithCodeT(exit.NoInput, "Cannot find directory {{.path}} for mount", out.V{"path": hostPath})
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
		api, err := machine.NewAPIClient()
		if err != nil {
			exit.WithError("Error getting client", err)
		}
		defer api.Close()
		host, err := api.Load(config.GetMachineName())

		if err != nil {
			exit.WithError("Error loading api", err)
		}
		if host.Driver.DriverName() == constants.DriverNone {
			exit.UsageT(`'none' driver does not support 'minikube mount' command`)
		}
		var ip net.IP
		if mountIP == "" {
			ip, err = cluster.GetVMHostIP(host)
			if err != nil {
				exit.WithError("Error getting the host IP address to use from within the VM", err)
			}
		} else {
			ip = net.ParseIP(mountIP)
			if ip == nil {
				exit.WithCodeT(exit.Data, "error parsing the input ip address for mount")
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

		out.T(out.Mounting, "Mounting host path {{.sourcePath}} into VM as {{.destinationPath}} ...", out.V{"sourcePath": hostPath, "destinationPath": vmPath})
		out.T(out.Option, "Mount type:   {{.name}}", out.V{"type": cfg.Type})
		out.T(out.Option, "User ID:      {{.userID}}", out.V{"userID": cfg.UID})
		out.T(out.Option, "Group ID:     {{.groupID}}", out.V{"groupID": cfg.GID})
		out.T(out.Option, "Version:      {{.version}}", out.V{"version": cfg.Version})
		out.T(out.Option, "Message Size: {{.size}}", out.V{"size": cfg.MSize})
		out.T(out.Option, "Permissions:  {{.octalMode}} ({{.writtenMode}})", out.V{"octalMode": fmt.Sprintf("%o", cfg.Mode), "writtenMode": cfg.Mode})
		out.T(out.Option, "Options:      {{.options}}", out.V{"options": cfg.Options})

		// An escape valve to allow future hackers to try NFS, VirtFS, or other FS types.
		if !supportedFilesystems[cfg.Type] {
			out.T(out.WarningType, "{{.type}} is not yet a supported filesystem. We will try anyways!", out.V{"type": cfg.Type})
		}
		// Use CommandRunner, as the native docker ssh service dies when Ctrl-C is received.
		runner, err := machine.CommandRunner(host)
		if err != nil {
			exit.WithError("Failed to get command runner", err)
		}
		out.T(out.Unmount, "Unmounting {{.path}} ...", out.V{"path": vmPath})
		err = cluster.Unmount(runner, vmPath)
		if err != nil {
			exit.WithCodeT(exit.Interrupted, "Received {{.error}}", out.V{"error": err})
		}

		var wg sync.WaitGroup
		if cfg.Type == nineP {
			wg.Add(1)
			go func() {
				out.T(out.Fileserver, "Userspace file server: ")
				ufs.StartServer(net.JoinHostPort(ip.String(), strconv.Itoa(port)), debugVal, hostPath)
				out.T(out.Stopped, "Userspace file server is shutdown")
				wg.Done()
			}()

			// Unmount if Ctrl-C or kill request is received.
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt, syscall.SIGTERM)
			go func() {
				for sig := range c {
					out.T(out.Unmount, "Unmounting {{.path}} ...", out.V{"path": vmPath})
					err := cluster.Unmount(runner, vmPath)
					if err != nil {
						out.ErrT(out.FailureType, "Failed unmount: {{.error}}", out.V{"error": err})
					}
					exit.WithCodeT(exit.Interrupted, "Received {{.name}} signal", out.V{"name": sig})
				}
			}()

			err = cluster.Mount(runner, ip.String(), vmPath, cfg)
			if err != nil {
				exit.WithError("mount failed", err)
			}
			out.T(out.SuccessType, "Successfully mounted {{.sourcePath}} to {{.destinationPath}}", out.V{"sourcePath": hostPath, "destinationPath": vmPath})
			out.Ln("")
			out.T(out.Notice, "NOTE: This process must stay alive for the mount to be accessible ...")
			wg.Wait()
		} else if cfg.Type == cifs {
			if host.Driver.DriverName() == constants.DriverHyperv {
				// Use CommandRunner, as the native docker ssh service dies when Ctrl-C is received.
				//runner, err := machine.CommandRunner(host)
				//if err != nil {
				//	exit.WithError("Failed to get command runner", err)
				//}
				//out.T(out.Notice, "CIFS Mount will be configured.")
				//configureCifsOnHost(hostPath)
				//
				//hostname, _ := os.Hostname()
				//user, err := hyperv.GetCurrentWindowsUser()
				//if err != nil {
				//	return
				//}
				//user = strings.Replace(user,(hostname + "\\"), "",1)
				//fmt.Printf("Please Type in the password for the user - %s		-- ", user)
				//password, _ := terminal.ReadPassword(int(os.Stdin.Fd()))
				////fmt.Printf("Password is : %s", password)
				//mountCmd := fmt.Sprintf("sudo mkdir -p %s && sudo mount.cifs //%s/%s %s -o username=%s,password=%s,domain=%s",vmPath, hostname, shareName, vmPath, user, password, hostname)
				//error := cluster.MountCifs(runner, mountCmd)
				//if error != nil {
				//	out.ErrT(out.FailureType, "Failed unmount: {{.error}}", out.V{"error": error})
				//}
				var shareName = "minikube"

				// Unmount the share if it already

				out.T(out.Notice, "Trying to start the mounting.")
				if err := pkghyperv.ConfigureHostMount(shareName,hostPath); err == nil {
					if error := enableCifsShare(shareName,vmPath,runner); error != nil {
						exit.WithError("Mount failed %v", error)
					}
				} else {
					exit.WithError("Mount failed %v", err)
				}
				out.T(out.Notice,"Mounting is complete!")
			} else {
				out.T(out.Embarrassed, "CIFS Mounts are currently only supported on {{.driver}} on Windows.", out.V{"driver": constants.DriverHyperv})
			}
		}
	},
}

func init() {
	mountCmd.Flags().StringVar(&mountIP, "ip", "", "Specify the ip that the mount should be setup on")
	mountCmd.Flags().StringVar(&mountType, "type", nineP, "Specify the mount filesystem type (supported types: 9p)")
	mountCmd.Flags().StringVar(&mountVersion, "9p-version", constants.DefaultMountVersion, "Specify the 9p version that the mount should use")
	mountCmd.Flags().BoolVar(&isKill, "kill", false, "Kill the mount process spawned by minikube start")
	mountCmd.Flags().StringVar(&uid, "uid", "docker", "Default user id used for the mount")
	mountCmd.Flags().StringVar(&gid, "gid", "docker", "Default group id used for the mount")
	mountCmd.Flags().UintVar(&mode, "mode", 0755, "File permissions used for the mount")
	mountCmd.Flags().StringSliceVar(&options, "options", []string{}, "Additional mount options, such as cache=fscache")
	mountCmd.Flags().IntVar(&mSize, "msize", constants.DefaultMsize, "The number of bytes to use for 9p packet payload")
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

func enableCifsShare(hostShareName string, vmDestinationPath string, runner command.Runner) (error) {
	// Ensure that the current user is administrator because creating a SMB Share requires Administrator privileges.
	_ , err := hyperv.IsWindowsAdministrator()
	if err != nil {
		return err
	}

	hostname, _ := os.Hostname()
	user, err := hyperv.GetCurrentWindowsUser()
	if err != nil {
		return err
	}
	user = strings.Replace(user,(hostname + "\\"), "",1)

	// Check if the Credential exists in the credential store.
	var credentialName = "minikube"
	var password = ""
	cred, err := wincred.GetGenericCredential(credentialName)
	if err == nil {
		out.T(out.Notice,"Credential {{.credential}} was found in the Windows Credential Store. Using that...",out.V{"credential":credentialName})
		password = string(cred.CredentialBlob)
	} else {
		out.T(out.Enabling,"Please Type in the password for the user - [{{.username}}]",out.V{"username":user})
		inputPassword, _ := terminal.ReadPassword(int(os.Stdin.Fd()))

		cred := wincred.NewGenericCredential(credentialName)
		cred.CredentialBlob = []byte(inputPassword)
		wincrederr := cred.Write()
		if wincrederr != nil {
			return wincrederr
		}
		password = string(inputPassword)
	}
	mountCmd := fmt.Sprintf("sudo mkdir -p %s && sudo mount.cifs //%s/%s %s -o username=%s,password=%s,domain=%s",vmDestinationPath, hostname, shareName, vmDestinationPath, user, password, hostname)
	error := cluster.MountCifs(runner, mountCmd)
	if error != nil {
		out.ErrT(out.FailureType, "Failed mounting: {{.error}}", out.V{"error": error})
		return error
	}
	return nil
}


//func configureCifsOnHost(sourcePath string) (bool, error) {
//	out.T(out.SuccessType, "Inside Cifs Mounting!")
//	if _, err := hyperv.IsWindowsAdministrator(); err != nil {
//		out.T(out.SuccessType, "Inside Cifs Mounting! {{.error}}", out.V{"error":err})
//		return false, err
//	}
//
//	if _, err := EnableCifsShare(sourcePath); err != nil {
//		out.T(out.SuccessType, "Successfully mounted {{.sourcePath}}", out.V{"sourcePath": sourcePath})
//	} else {
//		out.T(out.Embarrassed, "Mounting has failed because of - [{{.err}}]",out.V{"err": err})
//	}
//	return true, nil
//}