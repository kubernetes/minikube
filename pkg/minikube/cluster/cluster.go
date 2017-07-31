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

package cluster

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/docker/machine/drivers/virtualbox"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/mcnerror"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/tools/clientcmd/api/latest"

	"k8s.io/minikube/pkg/minikube/assets"
	cfg "k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/sshutil"
	"k8s.io/minikube/pkg/util"
	"k8s.io/minikube/pkg/util/kubeconfig"
)

var (
	certs = []string{"ca.crt", "ca.key", "apiserver.crt", "apiserver.key"}
)

const fileScheme = "file"

//This init function is used to set the logtostderr variable to false so that INFO level log info does not clutter the CLI
//INFO lvl logging is displayed due to the kubernetes api calling flag.Set("logtostderr", "true") in its init()
//see: https://github.com/kubernetes/kubernetes/blob/master/pkg/util/logs/logs.go#L32-34
func init() {
	flag.Set("logtostderr", "false")
	// Setting the default client to native gives much better performance.
	ssh.SetDefaultClient(ssh.Native)
}

// StartHost starts a host VM.
func StartHost(api libmachine.API, config MachineConfig) (*host.Host, error) {
	exists, err := api.Exists(cfg.GetMachineName())
	if err != nil {
		return nil, errors.Wrapf(err, "Error checking if host exists: %s", cfg.GetMachineName())
	}
	if !exists {
		return createHost(api, config)
	}

	glog.Infoln("Machine exists!")
	h, err := api.Load(cfg.GetMachineName())
	if err != nil {
		return nil, errors.Wrap(err, "Error loading existing host. Please try running [minikube delete], then run [minikube start] again.")
	}

	s, err := h.Driver.GetState()
	glog.Infoln("Machine state: ", s)
	if err != nil {
		return nil, errors.Wrap(err, "Error getting state for host")
	}

	if s != state.Running {
		if err := h.Driver.Start(); err != nil {
			return nil, errors.Wrap(err, "Error starting stopped host")
		}
		if err := api.Save(h); err != nil {
			return nil, errors.Wrap(err, "Error saving started host")
		}
	}

	if h.Driver.DriverName() != "none" {
		if err := h.ConfigureAuth(); err != nil {
			return nil, &util.RetriableError{Err: errors.Wrap(err, "Error configuring auth on host")}
		}
	}
	return h, nil
}

// StopHost stops the host VM.
func StopHost(api libmachine.API) error {
	host, err := api.Load(cfg.GetMachineName())
	if err != nil {
		return errors.Wrapf(err, "Error loading host: %s", cfg.GetMachineName())
	}
	if err := host.Stop(); err != nil {
		alreadyInStateError, ok := err.(mcnerror.ErrHostAlreadyInState)
		if ok && alreadyInStateError.State == state.Stopped {
			return nil
		}
		return errors.Wrapf(err, "Error stopping host: %s", cfg.GetMachineName())
	}
	return nil
}

// DeleteHost deletes the host VM.
func DeleteHost(api libmachine.API) error {
	host, err := api.Load(cfg.GetMachineName())
	if err != nil {
		return errors.Wrapf(err, "Error deleting host: %s", cfg.GetMachineName())
	}
	m := util.MultiError{}
	m.Collect(host.Driver.Remove())
	m.Collect(api.Remove(cfg.GetMachineName()))
	return m.ToError()
}

// GetHostStatus gets the status of the host VM.
func GetHostStatus(api libmachine.API) (string, error) {
	exists, err := api.Exists(cfg.GetMachineName())
	if err != nil {
		return "", errors.Wrapf(err, "Error checking that api exists for: %s", cfg.GetMachineName())
	}
	if !exists {
		return state.None.String(), nil
	}

	host, err := api.Load(cfg.GetMachineName())
	if err != nil {
		return "", errors.Wrapf(err, "Error loading api for: %s", cfg.GetMachineName())
	}

	s, err := host.Driver.GetState()
	if err != nil {
		return "", errors.Wrap(err, "Error getting host state")
	}
	return s.String(), nil
}

// GetLocalkubeStatus gets the status of localkube from the host VM.
func GetLocalkubeStatus(api libmachine.API) (string, error) {
	h, err := CheckIfApiExistsAndLoad(api)
	if err != nil {
		return "", err
	}
	s, err := RunCommand(h, localkubeStatusCommand, false)
	if err != nil {
		return "", err
	}
	s = strings.TrimSpace(s)
	if state.Running.String() == s {
		return state.Running.String(), nil
	} else if state.Stopped.String() == s {
		return state.Stopped.String(), nil
	} else {
		return "", fmt.Errorf("Error: Unrecognize output from GetLocalkubeStatus: %s", s)
	}
}

// GetHostDriverIP gets the ip address of the current minikube cluster
func GetHostDriverIP(api libmachine.API) (net.IP, error) {
	host, err := CheckIfApiExistsAndLoad(api)
	if err != nil {
		return nil, err
	}

	ipStr, err := host.Driver.GetIP()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting IP")
	}
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, errors.Wrap(err, "Error parsing IP")
	}
	return ip, nil
}

// StartCluster starts a k8s cluster on the specified Host.
func StartCluster(api libmachine.API, kubernetesConfig KubernetesConfig) error {
	h, err := CheckIfApiExistsAndLoad(api)
	if err != nil {
		return errors.Wrap(err, "Error checking that api exists and loading it")
	}

	startCommand, err := GetStartCommand(kubernetesConfig)
	if err != nil {
		return errors.Wrapf(err, "Error generating start command: %s", err)
	}
	glog.Infoln(startCommand)
	output, err := RunCommand(h, startCommand, true)
	glog.Infoln(output)
	if err != nil {
		return errors.Wrapf(err, "Error running ssh command: %s", startCommand)
	}
	return nil
}

func UpdateCluster(d drivers.Driver, config KubernetesConfig) error {
	copyableFiles := []assets.CopyableFile{}
	var localkubeFile assets.CopyableFile
	var err error

	//add url/file/bundled localkube to file list
	if localkubeURIWasSpecified(config) && config.KubernetesVersion != constants.DefaultKubernetesVersion {
		lCacher := localkubeCacher{config}
		localkubeFile, err = lCacher.fetchLocalkubeFromURI()
		if err != nil {
			return errors.Wrap(err, "Error updating localkube from uri")
		}
	} else {
		localkubeFile = assets.NewBinDataAsset("out/localkube", "/usr/local/bin", "localkube", "0777")
	}
	copyableFiles = append(copyableFiles, localkubeFile)

	// add addons to file list
	// custom addons
	assets.AddMinikubeAddonsDirToAssets(&copyableFiles)
	// bundled addons
	for _, addonBundle := range assets.Addons {
		if isEnabled, err := addonBundle.IsEnabled(); err == nil && isEnabled {
			for _, addon := range addonBundle.Assets {
				copyableFiles = append(copyableFiles, addon)
			}
		} else if err != nil {
			return err
		}
	}

	if d.DriverName() == "none" {
		// transfer files to correct place on filesystem
		for _, f := range copyableFiles {
			if err := assets.CopyFileLocal(f); err != nil {
				return err
			}
		}
		return nil
	}

	// transfer files to vm via SSH
	client, err := sshutil.NewSSHClient(d)
	if err != nil {
		return errors.Wrap(err, "Error creating new ssh client")
	}

	for _, f := range copyableFiles {
		if err := sshutil.TransferFile(f, client); err != nil {
			return err
		}
	}
	return nil
}

func localkubeURIWasSpecified(config KubernetesConfig) bool {
	// see if flag is different than default -> it was passed by user
	return config.KubernetesVersion != constants.DefaultKubernetesVersion
}

// SetupCerts gets the generated credentials required to talk to the APIServer.
func SetupCerts(d drivers.Driver, apiServerName string, clusterDnsDomain string) error {
	localPath := constants.GetMinipath()
	ipStr, err := d.GetIP()
	if err != nil {
		return errors.Wrap(err, "Error getting ip from driver")
	}
	glog.Infoln("Setting up certificates for IP: %s", ipStr)

	ip := net.ParseIP(ipStr)
	caCert := filepath.Join(localPath, "ca.crt")
	caKey := filepath.Join(localPath, "ca.key")
	publicPath := filepath.Join(localPath, "apiserver.crt")
	privatePath := filepath.Join(localPath, "apiserver.key")
	if err := GenerateCerts(caCert, caKey, publicPath, privatePath, ip, apiServerName, clusterDnsDomain); err != nil {
		return errors.Wrap(err, "Error generating certs")
	}

	copyableFiles := []assets.CopyableFile{}

	for _, cert := range certs {
		p := filepath.Join(localPath, cert)
		perms := "0644"
		if strings.HasSuffix(cert, ".key") {
			perms = "0600"
		}
		certFile, err := assets.NewFileAsset(p, util.DefaultCertPath, cert, perms)
		if err != nil {
			return err
		}
		copyableFiles = append(copyableFiles, certFile)
	}

	kubeCfgSetup := &kubeconfig.KubeConfigSetup{
		ClusterName:          cfg.GetMachineName(),
		ClusterServerAddress: "https://localhost:8443",
		ClientCertificate:    filepath.Join(util.DefaultCertPath, "apiserver.crt"),
		ClientKey:            filepath.Join(util.DefaultCertPath, "apiserver.key"),
		CertificateAuthority: filepath.Join(util.DefaultCertPath, "ca.crt"),
		KeepContext:          false,
	}

	kubeCfg := api.NewConfig()
	kubeconfig.PopulateKubeConfig(kubeCfgSetup, kubeCfg)
	data, err := runtime.Encode(latest.Codec, kubeCfg)

	kubeCfgFile := assets.NewMemoryAsset(data,
		util.DefaultLocalkubeDirectory, "kubeconfig", "0644")
	copyableFiles = append(copyableFiles, kubeCfgFile)

	if d.DriverName() == "none" {
		// transfer files to correct place on filesystem
		for _, f := range copyableFiles {
			if err := assets.CopyFileLocal(f); err != nil {
				return err
			}
		}
		return nil
	}

	// transfer files to vm via SSH
	client, err := sshutil.NewSSHClient(d)
	if err != nil {
		return errors.Wrap(err, "Error creating new ssh client")
	}

	for _, f := range copyableFiles {
		if err := sshutil.TransferFile(f, client); err != nil {
			return err
		}
	}
	return nil
}

func engineOptions(config MachineConfig) *engine.Options {

	o := engine.Options{
		Env:              config.DockerEnv,
		InsecureRegistry: config.InsecureRegistry,
		RegistryMirror:   config.RegistryMirror,
		ArbitraryFlags:   config.DockerOpt,
	}
	return &o
}

func createVirtualboxHost(config MachineConfig) drivers.Driver {
	d := virtualbox.NewDriver(cfg.GetMachineName(), constants.GetMinipath())
	d.Boot2DockerURL = config.Downloader.GetISOFileURI(config.MinikubeISO)
	d.Memory = config.Memory
	d.CPU = config.CPUs
	d.DiskSize = int(config.DiskSize)
	d.HostOnlyCIDR = config.HostOnlyCIDR
	d.NoShare = config.DisableDriverMounts
	return d
}

func createHost(api libmachine.API, config MachineConfig) (*host.Host, error) {
	var driver interface{}

	if config.VMDriver != "none" {
		if err := config.Downloader.CacheMinikubeISOFromURL(config.MinikubeISO); err != nil {
			return nil, errors.Wrap(err, "Error attempting to cache minikube ISO from URL")
		}
	}

	switch config.VMDriver {
	case "virtualbox":
		driver = createVirtualboxHost(config)
	case "vmwarefusion":
		driver = createVMwareFusionHost(config)
	case "kvm":
		driver = createKVMHost(config)
	case "xhyve":
		driver = createXhyveHost(config)
	case "hyperv":
		driver = createHypervHost(config)
	case "none":
		driver = createNoneHost(config)
	default:
		glog.Exitf("Unsupported driver: %s\n", config.VMDriver)
	}

	data, err := json.Marshal(driver)
	if err != nil {
		return nil, errors.Wrap(err, "Error marshalling json")
	}

	h, err := api.NewHost(config.VMDriver, data)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating new host")
	}

	h.HostOptions.AuthOptions.CertDir = constants.GetMinipath()
	h.HostOptions.AuthOptions.StorePath = constants.GetMinipath()
	h.HostOptions.EngineOptions = engineOptions(config)

	if err := api.Create(h); err != nil {
		// Wait for all the logs to reach the client
		time.Sleep(2 * time.Second)
		return nil, errors.Wrap(err, "Error creating host")
	}

	if err := api.Save(h); err != nil {
		return nil, errors.Wrap(err, "Error attempting to save")
	}
	return h, nil
}

// GetHostDockerEnv gets the necessary docker env variables to allow the use of docker through minikube's vm
func GetHostDockerEnv(api libmachine.API) (map[string]string, error) {
	host, err := CheckIfApiExistsAndLoad(api)
	if err != nil {
		return nil, errors.Wrap(err, "Error checking that api exists and loading it")
	}
	ip, err := host.Driver.GetIP()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting ip from host")
	}

	tcpPrefix := "tcp://"
	port := "2376"

	envMap := map[string]string{
		"DOCKER_TLS_VERIFY": "1",
		"DOCKER_HOST":       tcpPrefix + net.JoinHostPort(ip, port),
		"DOCKER_CERT_PATH":  constants.MakeMiniPath("certs"),
	}
	return envMap, nil
}

// GetHostLogs gets the localkube logs of the host VM.
// If follow is specified, it will tail the logs
func GetHostLogs(api libmachine.API, follow bool) (string, error) {
	h, err := CheckIfApiExistsAndLoad(api)
	if err != nil {
		return "", errors.Wrap(err, "Error checking that api exists and loading it")
	}
	logsCommand, err := GetLogsCommand(follow)
	if err != nil {
		return "", errors.Wrap(err, "Error getting logs command")
	}
	if follow {
		c, err := h.CreateSSHClient()
		if err != nil {
			return "", errors.Wrap(err, "Error creating ssh client")
		}
		err = c.Shell(logsCommand)
		if err != nil {
			return "", errors.Wrap(err, "error ssh shell")
		}
		return "", err
	}
	s, err := RunCommand(h, logsCommand, false)

	if err != nil {
		return s, err
	}
	return s, nil
}

// MountHost runs the mount command from the 9p client on the VM to the 9p server on the host
func MountHost(api libmachine.API, ip net.IP, path, port, mountVersion string, uid, gid, msize int) error {
	host, err := CheckIfApiExistsAndLoad(api)
	if err != nil {
		return errors.Wrap(err, "Error checking that api exists and loading it")
	}
	if ip == nil {
		ip, err = GetVMHostIP(host)
		if err != nil {
			return errors.Wrap(err, "Error getting the host IP address to use from within the VM")
		}
	}
	host.RunSSHCommand(GetMountCleanupCommand(path))
	mountCmd, err := GetMountCommand(ip, path, port, mountVersion, uid, gid, msize)
	if err != nil {
		return errors.Wrap(err, "Error getting mount command")
	}
	_, err = host.RunSSHCommand(mountCmd)
	if err != nil {
		return err
	}
	return nil
}

// GetVMHostIP gets the ip address to be used for mapping host -> VM and VM -> host
func GetVMHostIP(host *host.Host) (net.IP, error) {
	switch host.DriverName {
	case "kvm":
		return net.ParseIP("192.168.42.1"), nil
	case "hyperv":
		re := regexp.MustCompile(`"VSwitch": "(.*?)",`)
		// TODO(aprindle) Change this to deserialize the driver instead
		hypervVirtualSwitch := re.FindStringSubmatch(string(host.RawDriver))[1]
		ip, err := getIPForInterface(fmt.Sprintf("vEthernet (%s)", hypervVirtualSwitch))
		if err != nil {
			return []byte{}, errors.Wrap(err, "Error getting VM/Host IP address")
		}
		return ip, nil
	case "virtualbox":
		out, err := exec.Command(detectVBoxManageCmd(), "showvminfo", "minikube", "--machinereadable").Output()
		if err != nil {
			return []byte{}, errors.Wrap(err, "Error running vboxmanage command")
		}
		re := regexp.MustCompile(`hostonlyadapter2="(.*?)"`)
		iface := re.FindStringSubmatch(string(out))[1]
		ip, err := getIPForInterface(iface)
		if err != nil {
			return []byte{}, errors.Wrap(err, "Error getting VM/Host IP address")
		}
		return ip, nil
	case "xhyve":
		return net.ParseIP("192.168.64.1"), nil
	default:
		return []byte{}, errors.New("Error, attempted to get host ip address for unsupported driver")
	}
}

// Based on code from http://stackoverflow.com/questions/23529663/how-to-get-all-addresses-and-masks-from-local-interfaces-in-go
func getIPForInterface(name string) (net.IP, error) {
	i, _ := net.InterfaceByName(name)
	addrs, _ := i.Addrs()
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok {
			if ip := ipnet.IP.To4(); ip != nil {
				return ip, nil
			}
		}
	}
	return nil, errors.Errorf("Error finding IPV4 address for %s", name)
}

func CheckIfApiExistsAndLoad(api libmachine.API) (*host.Host, error) {
	exists, err := api.Exists(cfg.GetMachineName())
	if err != nil {
		return nil, errors.Wrapf(err, "Error checking that api exists for: %s", cfg.GetMachineName())
	}
	if !exists {
		return nil, errors.Errorf("Machine does not exist for api.Exists(%s)", cfg.GetMachineName())
	}

	host, err := api.Load(cfg.GetMachineName())
	if err != nil {
		return nil, errors.Wrapf(err, "Error loading api for: %s", cfg.GetMachineName())
	}
	return host, nil
}

func CreateSSHShell(api libmachine.API, args []string) error {
	host, err := CheckIfApiExistsAndLoad(api)
	if err != nil {
		return errors.Wrap(err, "Error checking if api exist and loading it")
	}

	currentState, err := host.Driver.GetState()
	if err != nil {
		return errors.Wrap(err, "Error getting state of host")
	}

	if currentState != state.Running {
		return errors.Errorf("Error: Cannot run ssh command: Host %q is not running", cfg.GetMachineName())
	}

	client, err := host.CreateSSHClient()
	if err != nil {
		return errors.Wrap(err, "Error creating ssh client")
	}
	return client.Shell(args...)
}

// EnsureMinikubeRunningOrExit checks that minikube has a status available and that
// that the status is `Running`, otherwise it will exit
func EnsureMinikubeRunningOrExit(api libmachine.API, exitStatus int) {
	s, err := GetHostStatus(api)
	if err != nil {
		glog.Errorln("Error getting machine status:", err)
		os.Exit(1)
	}
	if s != state.Running.String() {
		fmt.Fprintln(os.Stderr, "minikube is not currently running so the service cannot be accessed")
		os.Exit(exitStatus)
	}
}

// RunCommand executes commands for both the local and driver implementations
func RunCommand(h *host.Host, command string, sudo bool) (string, error) {
	if h.Driver.DriverName() == "none" {
		cmd := exec.Command("/bin/bash", "-c", command)
		if sudo {
			cmd = exec.Command("sudo", "/bin/bash", "-c", command)
		}
		out, err := cmd.CombinedOutput()
		if err != nil {
			return "", errors.Wrap(err, string(out))
		}
		return string(out), err
	}
	out, err := h.RunSSHCommand(command)
	if err != nil {
		return "", errors.Wrap(err, string(out))
	}
	return string(out), err

}
