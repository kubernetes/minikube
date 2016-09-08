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
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/machine/drivers/virtualbox"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/state"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	kubeApi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/sshutil"
	"k8s.io/minikube/pkg/util"
)

var (
	certs = []string{"ca.crt", "ca.key", "apiserver.crt", "apiserver.key"}
)

const fileScheme = "file"

//This init function is used to set the logtostderr variable to false so that INFO level log info does not clutter the CLI
//INFO lvl logging is displayed due to the kubernetes api calling flag.Set("logtostderr", "true") in its init()
//see: https://github.com/kubernetes/kubernetes/blob/master/pkg/util/logs.go#L32-34
func init() {
	flag.Set("logtostderr", "false")
}

// StartHost starts a host VM.
func StartHost(api libmachine.API, config MachineConfig) (*host.Host, error) {
	exists, err := api.Exists(constants.MachineName)
	if err != nil {
		return nil, errors.Wrapf(err, "Error checking if host exists: %s", constants.MachineName)
	}
	if !exists {
		return createHost(api, config)
	}

	glog.Infoln("Machine exists!")
	h, err := api.Load(constants.MachineName)
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
			return nil, errors.Wrapf(err, "Error starting stopped host")
		}
		if err := api.Save(h); err != nil {
			return nil, errors.Wrapf(err, "Error saving started host")
		}
	}

	if err := h.ConfigureAuth(); err != nil {
		return nil, errors.Wrap(err, "Error configuring auth on host: %s")
	}
	return h, nil
}

// StopHost stops the host VM.
func StopHost(api libmachine.API) error {
	host, err := api.Load(constants.MachineName)
	if err != nil {
		return errors.Wrapf(err, "Error loading host: %s", constants.MachineName)
	}
	if err := host.Stop(); err != nil {
		return errors.Wrapf(err, "Error stopping host: %s", constants.MachineName)
	}
	return nil
}

// DeleteHost deletes the host VM.
func DeleteHost(api libmachine.API) error {
	host, err := api.Load(constants.MachineName)
	if err != nil {
		return errors.Wrapf(err, "Error deleting host: %s", constants.MachineName)
	}
	m := util.MultiError{}
	m.Collect(host.Driver.Remove())
	m.Collect(api.Remove(constants.MachineName))
	return m.ToError()
}

// GetHostStatus gets the status of the host VM.
func GetHostStatus(api libmachine.API) (string, error) {
	dne := "Does Not Exist"
	exists, err := api.Exists(constants.MachineName)
	if err != nil {
		return "", errors.Wrapf(err, "Error checking that api exists for: ", constants.MachineName)
	}
	if !exists {
		return dne, nil
	}

	host, err := api.Load(constants.MachineName)
	if err != nil {
		return "", errors.Wrapf(err, "Error loading api for: ", constants.MachineName)
	}

	s, err := host.Driver.GetState()
	if s.String() == "" {
		return dne, nil
	}
	if err != nil {
		return "", errors.Wrap(err, "Error getting host state")
	}
	return s.String(), nil
}

// GetLocalkubeStatus gets the status of localkube from the host VM.
func GetLocalkubeStatus(api libmachine.API) (string, error) {
	host, err := checkIfApiExistsAndLoad(api)
	if err != nil {
		return "", err
	}
	s, err := host.RunSSHCommand(localkubeStatusCommand)
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

type sshAble interface {
	RunSSHCommand(string) (string, error)
}

// MachineConfig contains the parameters used to start a cluster.
type MachineConfig struct {
	MinikubeISO      string
	Memory           int
	CPUs             int
	DiskSize         int
	VMDriver         string
	DockerEnv        []string // Each entry is formatted as KEY=VALUE.
	InsecureRegistry []string
	RegistryMirror   []string
	HostOnlyCIDR     string // Only used by the virtualbox driver
}

// KubernetesConfig contains the parameters used to configure the VM Kubernetes.
type KubernetesConfig struct {
	KubernetesVersion string
	NodeIP            string
}

// StartCluster starts a k8s cluster on the specified Host.
func StartCluster(h sshAble, kubernetesConfig KubernetesConfig) error {
	commands := []string{stopCommand, GetStartCommand(kubernetesConfig)}

	for _, cmd := range commands {
		glog.Infoln(cmd)
		output, err := h.RunSSHCommand(cmd)
		glog.Infoln(output)
		if err != nil {
			return errors.Wrapf(err, "Error running ssh command: %s", cmd)
		}
	}

	return nil
}

type CopyableFile interface {
	io.Reader
	GetLength() int
	GetAssetName() string
	GetTargetDir() string
	GetTargetName() string
	GetPermissions() string
}

type BaseAsset struct {
	data        []byte
	reader      io.Reader
	Length      int
	AssetName   string
	TargetDir   string
	TargetName  string
	Permissions string
}

func (b *BaseAsset) GetAssetName() string {
	return b.AssetName
}

func (b *BaseAsset) GetTargetDir() string {
	return b.TargetDir
}

func (b *BaseAsset) GetTargetName() string {
	return b.TargetName
}

func (b *BaseAsset) GetPermissions() string {
	return b.Permissions
}

type FileAsset struct {
	BaseAsset
}

func NewFileAsset(assetName, targetDir, targetName, permissions string) (*FileAsset, error) {
	f := &FileAsset{
		BaseAsset{
			AssetName:   assetName,
			TargetDir:   targetDir,
			TargetName:  targetName,
			Permissions: permissions,
		},
	}
	file, err := os.Open(f.AssetName)
	if err != nil {
		return nil, errors.Wrapf(err, "Error opening file asset: %s", f.AssetName)
	}
	f.reader = file
	return f, nil
}

func (f *FileAsset) GetLength() int {
	file, err := os.Open(f.AssetName)
	defer file.Close()
	if err != nil {
		return 0
	}
	fi, err := file.Stat()
	if err != nil {
		return 0
	}
	return int(fi.Size())
}

func (f *FileAsset) Read(p []byte) (int, error) {
	if f.reader == nil {
		return 0, errors.New("Error attempting FileAsset.Read, FileAsset.reader uninitialized")
	}
	return f.reader.Read(p)
}

type MemoryAsset struct {
	BaseAsset
}

func NewMemoryAsset(assetName, targetDir, targetName, permissions string) *MemoryAsset {
	m := &MemoryAsset{
		BaseAsset{
			AssetName:   assetName,
			TargetDir:   targetDir,
			TargetName:  targetName,
			Permissions: permissions,
		},
	}
	m.loadData()
	return m
}

func (m *MemoryAsset) loadData() error {
	contents, err := Asset(m.AssetName)
	if err != nil {
		return err
	}
	m.data = contents
	m.Length = len(contents)
	m.reader = bytes.NewReader(m.data)
	return nil
}

func (m *MemoryAsset) GetLength() int {
	return m.Length
}

func (m *MemoryAsset) Read(p []byte) (int, error) {
	return m.reader.Read(p)
}

var memoryAssets = []CopyableFile{
	NewMemoryAsset(
		"deploy/iso/addon-manager.yaml",
		"/etc/kubernetes/manifests/",
		"addon-manager.yaml",
		"0640"),
	NewMemoryAsset(
		"deploy/addons/dashboard-rc.yaml",
		"/etc/kubernetes/addons/",
		"dashboard-rc.yaml",
		"0640"),
	NewMemoryAsset(
		"deploy/addons/dashboard-svc.yaml",
		"/etc/kubernetes/addons/",
		"dashboard-svc.yaml",
		"0640"),
}

func UpdateCluster(h sshAble, d drivers.Driver, config KubernetesConfig) error {
	client, err := sshutil.NewSSHClient(d)
	if err != nil {
		return errors.Wrap(err, "Error creating new ssh client")
	}

	// transfer localkube from cache/asset to vm
	if localkubeURIWasSpecified(config) {
		lCacher := localkubeCacher{config}
		if err = lCacher.updateLocalkubeFromURI(client); err != nil {
			return errors.Wrap(err, "Error updating localkube from uri")
		}
	} else {
		if err = updateLocalkubeFromAsset(client); err != nil {
			return errors.Wrap(err, "Error updating localkube from asset")
		}
	}
	fileAssets := []CopyableFile{}
	addMinikubeAddonsDirToAssets(&fileAssets)
	// merge files to copy
	var copyableFiles []CopyableFile
	copyableFiles = append(copyableFiles, memoryAssets...)
	copyableFiles = append(copyableFiles, fileAssets...)
	// transfer files to vm
	for _, copyableFile := range copyableFiles {
		if err := sshutil.Transfer(copyableFile, copyableFile.GetLength(),
			copyableFile.GetTargetDir(), copyableFile.GetTargetName(),
			copyableFile.GetPermissions(), client); err != nil {
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
func SetupCerts(d drivers.Driver) error {
	localPath := constants.Minipath
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
	if err := GenerateCerts(caCert, caKey, publicPath, privatePath, ip); err != nil {
		return errors.Wrap(err, "Error generating certs")
	}

	client, err := sshutil.NewSSHClient(d)
	if err != nil {
		return errors.Wrap(err, "Error creating new ssh client")
	}

	for _, cert := range certs {
		p := filepath.Join(localPath, cert)
		data, err := ioutil.ReadFile(p)
		if err != nil {
			return errors.Wrapf(err, "Error reading file: %s", p)
		}
		perms := "0644"
		if strings.HasSuffix(cert, ".key") {
			perms = "0600"
		}
		if err := sshutil.Transfer(bytes.NewReader(data), len(data), util.DefaultCertPath, cert, perms, client); err != nil {
			return errors.Wrapf(err, "Error transferring data: %s", string(data))
		}
	}
	return nil
}

func engineOptions(config MachineConfig) *engine.Options {

	o := engine.Options{
		Env:              config.DockerEnv,
		InsecureRegistry: config.InsecureRegistry,
		RegistryMirror:   config.RegistryMirror,
	}
	return &o
}

func createVirtualboxHost(config MachineConfig) drivers.Driver {
	d := virtualbox.NewDriver(constants.MachineName, constants.Minipath)
	d.Boot2DockerURL = config.GetISOFileURI()
	d.Memory = config.Memory
	d.CPU = config.CPUs
	d.DiskSize = int(config.DiskSize)
	d.HostOnlyCIDR = config.HostOnlyCIDR
	return d
}

func isIsoChecksumValid(isoData *[]byte, shaURL string) bool {
	r, err := http.Get(shaURL)
	if err != nil {
		glog.Errorf("Error downloading ISO checksum: %s", err)
		return false
	} else if r.StatusCode != http.StatusOK {
		glog.Errorf("Error downloading ISO checksum. Got HTTP Error: %s", r.Status)
		return false
	}

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		glog.Errorf("Error reading ISO checksum: %s", err)
		return false
	}

	expectedSum := strings.Trim(string(body), "\n")

	b := sha256.Sum256(*isoData)
	actualSum := hex.EncodeToString(b[:])
	if string(expectedSum) != actualSum {
		glog.Errorf("Downloaded ISO checksum does not match expected value. Actual: %s. Expected: %s", actualSum, expectedSum)
		return false
	}
	return true
}

func (m *MachineConfig) CacheMinikubeISOFromURL() error {
	// store the miniube-iso inside the .minikube dir
	response, err := http.Get(m.MinikubeISO)
	if err != nil {
		return errors.Wrapf(err, "Error getting minikube iso at %s via http", m.MinikubeISO)
	}

	defer response.Body.Close()
	isoData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return errors.Wrap(err, "Error reading minikubeISO url response")
	}

	// Validate the ISO if it was the default URL, before writing it to disk.
	if m.MinikubeISO == constants.DefaultIsoUrl {
		if !isIsoChecksumValid(&isoData, constants.DefaultIsoShaUrl) {
			return errors.New("Error validating ISO checksum.")
		}
	}

	if response.StatusCode != http.StatusOK {
		return errors.Errorf("Received %d response from %s while trying to download minikube.iso", response.StatusCode, m.MinikubeISO)
	}

	out, err := os.Create(m.GetISOCacheFilepath())
	if err != nil {
		return errors.Wrap(err, "Error creating minikube iso cache filepath")
	}
	defer out.Close()

	if _, err = out.Write(isoData); err != nil {
		return errors.Wrap(err, "Error writing iso data to file")
	}
	return nil
}

func (m *MachineConfig) ShouldCacheMinikubeISO() bool {
	// store the miniube-iso inside the .minikube dir

	urlObj, err := url.Parse(m.MinikubeISO)
	if err != nil {
		return false
	}
	if urlObj.Scheme == fileScheme {
		return false
	}
	if m.IsMinikubeISOCached() {
		return false
	}
	return true
}

func (m *MachineConfig) GetISOCacheFilepath() string {
	return filepath.Join(constants.Minipath, "cache", "iso", filepath.Base(m.MinikubeISO))
}

func (m *MachineConfig) GetISOFileURI() string {
	urlObj, err := url.Parse(m.MinikubeISO)
	if err != nil {
		return m.MinikubeISO
	}
	if urlObj.Scheme == fileScheme {
		return m.MinikubeISO
	}
	isoPath := filepath.Join(constants.Minipath, "cache", "iso", filepath.Base(m.MinikubeISO))
	// As this is a file URL there should be no backslashes regardless of platform running on.
	return "file://" + filepath.ToSlash(isoPath)
}

func (m *MachineConfig) IsMinikubeISOCached() bool {
	if _, err := os.Stat(m.GetISOCacheFilepath()); os.IsNotExist(err) {
		return false
	}
	return true
}

func createHost(api libmachine.API, config MachineConfig) (*host.Host, error) {
	var driver interface{}

	if config.ShouldCacheMinikubeISO() {
		if err := config.CacheMinikubeISOFromURL(); err != nil {
			return nil, errors.Wrap(err, "Error attempting to cache minikube iso from url")
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
	default:
		glog.Exitf("Unsupported driver: %s\n", config.VMDriver)
	}

	data, err := json.Marshal(driver)
	if err != nil {
		return nil, errors.Wrap(err, "Error marshalling json")
	}

	h, err := api.NewHost(config.VMDriver, data)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating new host: %s")
	}

	h.HostOptions.AuthOptions.CertDir = constants.Minipath
	h.HostOptions.AuthOptions.StorePath = constants.Minipath
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
	host, err := checkIfApiExistsAndLoad(api)
	if err != nil {
		return nil, errors.Wrap(err, "Error checking that api exists and loading it")
	}
	ip, err := host.Driver.GetIP()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting ip from host")
	}

	tcpPrefix := "tcp://"
	portDelimiter := ":"
	port := "2376"

	envMap := map[string]string{
		"DOCKER_TLS_VERIFY": "1",
		"DOCKER_HOST":       tcpPrefix + ip + portDelimiter + port,
		"DOCKER_CERT_PATH":  constants.MakeMiniPath("certs"),
	}
	return envMap, nil
}

// GetHostLogs gets the localkube logs of the host VM.
func GetHostLogs(api libmachine.API) (string, error) {
	host, err := checkIfApiExistsAndLoad(api)
	if err != nil {
		return "", errors.Wrap(err, "Error checking that api exists and loading it")
	}
	s, err := host.RunSSHCommand(logsCommand)
	if err != nil {
		return "", err
	}
	return s, nil
}

func checkIfApiExistsAndLoad(api libmachine.API) (*host.Host, error) {
	exists, err := api.Exists(constants.MachineName)
	if err != nil {
		return nil, errors.Wrapf(err, "Error checking that api exists for: ", constants.MachineName)
	}
	if !exists {
		return nil, errors.Errorf("Machine does not exist for api.Exists(%s)", constants.MachineName)
	}

	host, err := api.Load(constants.MachineName)
	if err != nil {
		return nil, errors.Wrapf(err, "Error loading api for: ", constants.MachineName)
	}
	return host, nil
}

func CreateSSHShell(api libmachine.API, args []string) error {
	host, err := checkIfApiExistsAndLoad(api)
	if err != nil {
		return errors.Wrap(err, "Error checking if api exist and loading it")
	}

	currentState, err := host.Driver.GetState()
	if err != nil {
		return errors.Wrap(err, "Error getting state of host")
	}

	if currentState != state.Running {
		return errors.Errorf("Error: Cannot run ssh command: Host %q is not running", constants.MachineName)
	}

	client, err := host.CreateSSHClient()
	if err != nil {
		return errors.Wrap(err, "Error creating ssh client")
	}
	return client.Shell(strings.Join(args, " "))
}

func GetServiceURL(api libmachine.API, namespace, service string) (string, error) {
	host, err := checkIfApiExistsAndLoad(api)
	if err != nil {
		return "", errors.Wrap(err, "Error checking if api exist and loading it")
	}

	ip, err := host.Driver.GetIP()
	if err != nil {
		return "", errors.Wrap(err, "Error getting ip from host")
	}

	port, err := getServicePort(namespace, service)
	if err != nil {
		return "", errors.Wrapf(err, "Error getting service port from %s, %s", namespace, service)
	}

	return fmt.Sprintf("http://%s:%d", ip, port), nil
}

type serviceGetter interface {
	Get(name string) (*kubeApi.Service, error)
}

type endpointGetter interface {
	Get(name string) (*kubeApi.Endpoints, error)
}

func getServicePort(namespace, service string) (int, error) {
	services, err := GetKubernetesServicesWithNamespace(namespace)
	if err != nil {
		return 0, errors.Wrapf(err, "Error getting kubernetes service with namespace", namespace)
	}
	return getServicePortFromServiceGetter(services, service)
}

func getServicePortFromServiceGetter(services serviceGetter, service string) (int, error) {
	svc, err := services.Get(service)
	if err != nil {
		return 0, errors.Wrapf(err, "Error getting %s service: %s", service)
	}
	nodePort := 0
	if len(svc.Spec.Ports) > 0 {
		nodePort = int(svc.Spec.Ports[0].NodePort)
	}
	if nodePort == 0 {
		return 0, errors.Errorf("Service %s does not have a node port. To have one assigned automatically, the service type must be NodePort or LoadBalancer, but this service is of type %s.", service, svc.Spec.Type)
	}
	return nodePort, nil
}

func GetKubernetesClient() (*unversioned.Client, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, errors.Wrap(err, "Error creating kubeConfig: %s")
	}
	client, err := unversioned.New(config)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating new client from kubeConfig.ClientConfig()")
	}
	return client, nil
}

func GetKubernetesServicesWithNamespace(namespace string) (serviceGetter, error) {
	client, err := GetKubernetesClient()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting kubernetes client")
	}
	services := client.Services(namespace)
	return services, nil
}

func GetKubernetesEndpointsWithNamespace(namespace string) (endpointGetter, error) {
	client, err := GetKubernetesClient()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting kubernetes client")
	}
	endpoints := client.Endpoints(namespace)
	return endpoints, nil
}

// EnsureMinikubeRunningOrExit checks that minikube has a status available and that
// that the status is `Running`, otherwise it will exit
func EnsureMinikubeRunningOrExit(api libmachine.API) {
	s, err := GetHostStatus(api)
	if err != nil {
		glog.Errorln("Error getting machine status:", err)
		os.Exit(1)
	}
	if s != state.Running.String() {
		fmt.Fprintln(os.Stdout, "minikube is not currently running so the service cannot be accessed")
		os.Exit(1)
	}
}

func addMinikubeAddonsDirToAssets(assetList *[]CopyableFile) {
	// loop over .minikube/addons and add them to assets
	searchDir := constants.MakeMiniPath("addons")
	err := filepath.Walk(searchDir, func(addonFile string, f os.FileInfo, err error) error {
		isDir, err := util.IsDirectory(addonFile)
		if err == nil && !isDir {
			f, err := NewFileAsset(addonFile, "/etc/kubernetes/addons", filepath.Base(addonFile), "0640")
			if err == nil {
				*assetList = append(*assetList, f)
			}
		} else if err != nil {
			glog.Infoln("Error encountered while walking .minikube/addons: ", err)
		}
		return nil
	})
	if err != nil {
		glog.Infoln("Error encountered while walking .minikube/addons: ", err)
	}
}
