/*
Copyright 2024 The Kubernetes Authors All rights reserved.

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
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/state"
	"github.com/pkg/errors"

	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/bootstrapper/bsutil/kverify"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/kubeconfig"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/out/register"
	"k8s.io/minikube/pkg/version"
)

// Additional legacy states
const (
	// Configured means configured
	Configured = "Configured" // ~state.Saved
	// Misconfigured means misconfigured
	Misconfigured = "Misconfigured" // ~state.Error
	// Nonexistent means the resource does not exist
	Nonexistent = "Nonexistent" // ~state.None
	// Irrelevant is used for statuses that aren't meaningful for worker nodes
	Irrelevant = "Irrelevant"
)

// New status modes, based roughly on HTTP/SMTP standards
const (

	// 1xx signifies a transitional state. If retried, it will soon return a 2xx, 4xx, or 5xx

	Starting  = 100
	Pausing   = 101
	Unpausing = 102
	Stopping  = 110
	Deleting  = 120

	// 2xx signifies that the API Server is able to service requests

	OK       = 200
	HAppy    = 201
	Warning  = 203
	Degraded = 204

	// 4xx signifies an error that requires help from the client to resolve

	NotFound = 404
	Stopped  = 405
	Paused   = 418 // I'm a teapot!

	// 5xx signifies a server-side error (that may be retryable)

	Error               = 500
	InsufficientStorage = 507
	Unknown             = 520
)

var (
	exitCodeToHTTPCode = map[int]int{
		// exit code 26 corresponds to insufficient storage
		26: 507,
	}

	codeNames = map[int]string{
		100: "Starting",
		101: "Pausing",
		102: "Unpausing",
		110: "Stopping",
		103: "Deleting",

		200: "OK",
		201: "HAppy",
		203: "Warning",
		204: "Degraded",

		404: "NotFound",
		405: "Stopped",
		418: "Paused",

		500: "Error",
		507: "InsufficientStorage",
		520: "Unknown",
	}

	codeDetails = map[int]string{
		507: "/var is almost out of disk space",
	}
)

// Status holds string representations of component states
type Status struct {
	Name       string
	Host       string
	Kubelet    string
	APIServer  string
	Kubeconfig string
	Worker     bool
	TimeToStop string `json:",omitempty"`
	DockerEnv  string `json:",omitempty"`
	PodManEnv  string `json:",omitempty"`
}

// State holds a cluster state representation
//
//nolint:revive
type State struct {
	BaseState

	BinaryVersion string
	TimeToStop    string `json:",omitempty"`
	Components    map[string]BaseState
	Nodes         []NodeState
}

// NodeState holds a node state representation
type NodeState struct {
	BaseState
	Components map[string]BaseState `json:",omitempty"`
}

// BaseState holds a component state representation, such as "apiserver" or "kubeconfig"
type BaseState struct {
	// Name is the name of the object
	Name string

	// StatusCode is an HTTP-like status code for this object
	StatusCode int
	// Name is a human-readable name for the status code
	StatusName string
	// StatusDetail is long human-readable string describing why this particular status code was chosen
	StatusDetail string `json:",omitempty"` // Not yet implemented

	// Step is which workflow step the object is at.
	Step string `json:",omitempty"`
	// StepDetail is a long human-readable string describing the step
	StepDetail string `json:",omitempty"`
}

// GetStatus returns the statuses of each node
func GetStatus(api libmachine.API, cc *config.ClusterConfig) ([]*Status, error) {
	var statuses []*Status
	for _, n := range cc.Nodes {
		machineName := config.MachineName(*cc, n)
		klog.Infof("checking status of %s ...", machineName)
		st, err := NodeStatus(api, *cc, n)
		klog.Infof("%s status: %+v", machineName, st)

		if err != nil {
			klog.Errorf("status error: %v", err)
			return nil, err
		}
		if st.Host == Nonexistent {
			err := fmt.Errorf("the %q host does not exist", machineName)
			klog.Error(err)
			return nil, err
		}
		statuses = append(statuses, st)
	}
	return statuses, nil
}

// GetState converts Status structs into a State struct
//
//nolint:gocyclo
func GetState(sts []*Status, profile string, cc *config.ClusterConfig) State {
	statusName := ""
	if len(sts) > 0 {
		statusName = sts[0].APIServer
		if sts[0].Host == codeNames[InsufficientStorage] {
			statusName = sts[0].Host
		}
	}
	sc := statusCode(statusName)

	cs := State{
		BinaryVersion: version.GetVersion(),

		BaseState: BaseState{
			Name:         profile,
			StatusCode:   sc,
			StatusName:   statusName,
			StatusDetail: codeDetails[sc],
		},

		TimeToStop: sts[0].TimeToStop,

		Components: map[string]BaseState{
			"kubeconfig": {Name: "kubeconfig", StatusCode: statusCode(sts[0].Kubeconfig), StatusName: codeNames[statusCode(sts[0].Kubeconfig)]},
		},
	}
	healthyCPs := 0
	for _, st := range sts {
		ns := NodeState{
			BaseState: BaseState{
				Name:       st.Name,
				StatusCode: statusCode(st.Host),
			},
			Components: map[string]BaseState{
				"kubelet": {Name: "kubelet", StatusCode: statusCode(st.Kubelet)},
			},
		}

		if st.APIServer != Irrelevant {
			ns.Components["apiserver"] = BaseState{Name: "apiserver", StatusCode: statusCode(st.APIServer)}
		}

		// Convert status codes to status names
		ns.StatusName = codeNames[ns.StatusCode]
		for k, v := range ns.Components {
			v.StatusName = codeNames[v.StatusCode]
			ns.Components[k] = v
		}

		cs.Nodes = append(cs.Nodes, ns)

		// we also need to calculate how many control plane node is healthy
		if !st.Worker &&
			st.Host == state.Running.String() &&
			st.Kubeconfig == Configured &&
			st.Kubelet == state.Running.String() &&
			st.APIServer == state.Running.String() {
			healthyCPs++
		}
	}

	evs, mtime, err := readEventLog(sts[0].Name)
	if err != nil {
		klog.Errorf("unable to read event log: %v", err)
		return cs
	}

	transientCode := 0
	started := false
	var finalStep map[string]string

	for _, ev := range evs {
		if ev.Type() == "io.k8s.sigs.minikube.step" {
			var data map[string]string
			err := ev.DataAs(&data)
			if err != nil {
				klog.Errorf("unable to parse data: %v\nraw data: %s", err, ev.Data())
				continue
			}

			switch data["name"] {
			case string(register.InitialSetup):
				transientCode = Starting
			case string(register.Done):
				transientCode = 0
				started = true
			case string(register.Stopping):
				klog.Infof("%q == %q", data["name"], register.Stopping)
				transientCode = Stopping
			case string(register.Deleting):
				transientCode = Deleting
			case string(register.Pausing):
				transientCode = Pausing
			case string(register.Unpausing):
				transientCode = Unpausing
			}

			finalStep = data
			klog.Infof("transient code %d (%q) for step: %+v", transientCode, codeNames[transientCode], data)
		}

		if ev.Type() == "io.k8s.sigs.minikube.error" {
			var data map[string]string
			err := ev.DataAs(&data)
			if err != nil {
				klog.Errorf("unable to parse data: %v\nraw data: %s", err, ev.Data())
				continue
			}
			// process exit code, if present
			if ec, ok := data["exitcode"]; ok && ec != "" {
				exitCode, err := strconv.Atoi(ec)
				if err != nil {
					klog.Errorf("exit code not found: %v", err)
					continue
				}

				if val, ok := exitCodeToHTTPCode[exitCode]; ok {
					exitCode = val
				}

				transientCode = exitCode
			}

			for _, n := range cs.Nodes {
				n.StatusCode = transientCode
				n.StatusName = codeNames[n.StatusCode]
			}

			klog.Infof("transient code %d (%q) for step: %+v", transientCode, codeNames[transientCode], data)
		}
	}

	if finalStep != nil {
		if mtime.Before(time.Now().Add(-10 * time.Minute)) {
			klog.Warningf("event stream is too old (%s) to be considered a transient state", mtime)
		} else {
			cs.Step = strings.TrimSpace(finalStep["name"])
			cs.StepDetail = strings.TrimSpace(finalStep["message"])
			if transientCode != 0 {
				cs.StatusCode = transientCode
			}
		}
	}

	if config.IsHA(*cc) && started {
		switch {
		case healthyCPs < 2:
			cs.StatusCode = Stopped
		case healthyCPs == 2:
			cs.StatusCode = Degraded
		default:
			cs.StatusCode = HAppy
		}
	}

	cs.StatusName = codeNames[cs.StatusCode]
	cs.StatusDetail = codeDetails[cs.StatusCode]

	return cs
}

// isTunnelWorking checks if the SSH tunnel endpoint is accessible
func isTunnelWorking(hostname string, port int) bool {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(hostname, fmt.Sprintf("%d", port)), 2*time.Second)
	if err != nil {
		klog.V(3).Infof("Tunnel connectivity check failed: %v", err)
		return false
	}
	conn.Close()
	klog.V(3).Infof("Tunnel connectivity check passed for %s:%d", hostname, port)
	return true
}

// NodeStatus looks up the status of a node
func NodeStatus(api libmachine.API, cc config.ClusterConfig, n config.Node) (*Status, error) {
	controlPlane := n.ControlPlane
	name := config.MachineName(cc, n)

	st := &Status{
		Name:       name,
		Host:       Nonexistent,
		APIServer:  Nonexistent,
		Kubelet:    Nonexistent,
		Kubeconfig: Nonexistent,
		Worker:     !controlPlane,
	}

	hs, err := machine.Status(api, name)
	klog.Infof("%s host status = %q (err=%v)", name, hs, err)
	if err != nil {
		return st, errors.Wrap(err, "host")
	}

	// We have no record of this host. Return nonexistent struct
	if hs == state.None.String() {
		return st, nil
	}
	st.Host = hs

	// If it's not running, quickly bail out rather than delivering conflicting messages
	if st.Host != state.Running.String() {
		klog.Infof("host is not running, skipping remaining checks")
		st.APIServer = st.Host
		st.Kubelet = st.Host
		st.Kubeconfig = st.Host
		return st, nil
	}

	// We have a fully operational host, now we can check for details
	if _, err := DriverIP(api, name); err != nil {
		klog.Errorf("failed to get driver ip: %v", err)
		st.Host = state.Error.String()
		return st, err
	}

	st.Kubeconfig = Configured
	if !controlPlane {
		st.Kubeconfig = Irrelevant
		st.APIServer = Irrelevant
	}

	host, err := machine.LoadHost(api, name)
	if err != nil {
		return st, err
	}

	cr, err := machine.CommandRunner(host)
	if err != nil {
		return st, err
	}

	// Run parallel checks for better performance
	type checkResult struct {
		diskUsage int
		diskErr   error
		kubelet   state.State
	}

	resultCh := make(chan checkResult, 1)

	// Run storage and kubelet checks in parallel
	go func() {
		var result checkResult

		// Check storage concurrently
		result.diskUsage, result.diskErr = machine.DiskUsed(cr, "/var")

		// Check kubelet status concurrently
		result.kubelet = kverify.ServiceStatus(cr, "kubelet")

		resultCh <- result
	}()

	// Get results from parallel checks
	result := <-resultCh

	if result.diskErr != nil {
		klog.Errorf("failed to get storage capacity of /var: %v", result.diskErr)
		st.Host = state.Error.String()
		return st, result.diskErr
	}
	if result.diskUsage >= 99 {
		st.Host = codeNames[InsufficientStorage]
	}

	st.Kubelet = result.kubelet.String()
	if cc.ScheduledStop != nil {
		initiationTime := time.Unix(cc.ScheduledStop.InitiationTime, 0)
		st.TimeToStop = time.Until(initiationTime.Add(cc.ScheduledStop.Duration)).String()
	}
	if os.Getenv(constants.MinikubeActiveDockerdEnv) != "" {
		st.DockerEnv = "in-use"
	}
	if os.Getenv(constants.MinikubeActivePodmanEnv) != "" {
		st.PodManEnv = "in-use"
	}
	// Early exit for worker nodes
	if !controlPlane {
		return st, nil
	}

	var hostname string
	var port int
	if cc.Addons["auto-pause"] {
		hostname, _, port, err = driver.AutoPauseProxyEndpoint(&cc, &n, host.DriverName)
	} else {
		hostname = cc.KubernetesConfig.APIServerHAVIP
		port = cc.APIServerPort
		if !config.IsHA(cc) || driver.NeedsPortForward(cc.Driver) {
			hostname, _, port, err = driver.ControlPlaneEndpoint(&cc, &n, host.DriverName)
		}
	}

	if err != nil {
		klog.Errorf("forwarded endpoint: %v", err)
		st.Kubeconfig = Misconfigured
	} else if err := kubeconfig.VerifyEndpoint(cc.Name, hostname, port, ""); err != nil && st.Host != state.Starting.String() {
		klog.Errorf("kubeconfig endpoint: %v", err)
		st.Kubeconfig = Misconfigured
	}

	// For remote Docker contexts, adjust endpoint based on context type
	if driver.IsKIC(cc.Driver) && oci.IsRemoteDockerContext() {
		klog.Infof("Remote Docker context detected: SSH=%v", oci.IsSSHDockerContext())

		// Get kubeconfig endpoint for both SSH and TLS contexts
		kcEndpoint, kcPort, err := kubeconfig.Endpoint(cc.Name, kubeconfig.PathFromEnv())

		if oci.IsSSHDockerContext() {
			// SSH context: ensure tunnel is up before proceeding
			if err == nil && (kcEndpoint == "localhost" || kcEndpoint == "127.0.0.1" || kcEndpoint == "::1") {
				// Verify tunnel is actually working before using it
				if !isTunnelWorking(kcEndpoint, kcPort) {
					klog.Warningf("SSH tunnel endpoint %s:%d not accessible, attempting to establish tunnel", kcEndpoint, kcPort)
					tunnelEndpoint, cleanup, setupErr := oci.SetupAPIServerTunnel(cc.APIServerPort)
					if setupErr != nil {
						klog.Errorf("Failed to setup API server tunnel: %v", setupErr)
						// Fall back to container check
						sta, err := kverify.APIServerStatusFromContainer(cr, hostname, port)
						if err != nil {
							klog.Errorln("Error checking API server from container:", err)
							st.APIServer = state.Error.String()
						} else {
							klog.Infof("API server status from container: %s", sta.String())
							st.APIServer = sta.String()
						}
						return st, nil
					}
					if tunnelEndpoint != "" && cleanup != nil {
						// Note: We don't defer cleanup here since status is a quick check
						// The tunnel will be managed by the tunnel manager's lifecycle
						klog.Infof("Successfully established SSH tunnel: %s", tunnelEndpoint)
						// Re-get endpoint after tunnel setup
						kcEndpoint, kcPort, err = kubeconfig.Endpoint(cc.Name, kubeconfig.PathFromEnv())
						if err != nil {
							klog.Errorf("Failed to get kubeconfig endpoint after tunnel setup: %v", err)
							st.APIServer = state.Error.String()
							return st, nil
						}
					}
				}
				hostname = kcEndpoint
				port = kcPort
				klog.Infof("Using SSH tunnel endpoint: %s:%d", hostname, port)
			} else {
				klog.Infof("No tunnel endpoint found (got %s:%d), checking API server from container", hostname, port)
				sta, err := kverify.APIServerStatusFromContainer(cr, hostname, port)
				if err != nil {
					klog.Errorln("Error checking API server from container:", err)
					st.APIServer = state.Error.String()
				} else {
					klog.Infof("API server status from container: %s", sta.String())
					st.APIServer = sta.String()
				}
				return st, nil
			}
		} else {
			// TLS context: use remote endpoint from kubeconfig
			if err == nil {
				hostname = kcEndpoint
				port = kcPort
				klog.Infof("Using TLS remote endpoint: %s:%d", hostname, port)
			} else {
				klog.Warningf("Failed to get kubeconfig endpoint for TLS context: %v", err)
			}
		}
	}

	sta, err := kverify.APIServerStatus(cr, hostname, port)
	klog.Infof("%s apiserver status = %s (err=%v)", name, sta, err)

	if err != nil {
		klog.Errorln("Error apiserver status:", err)
		st.APIServer = state.Error.String()
	} else {
		st.APIServer = sta.String()
	}

	return st, nil
}

// readEventLog reads cloudevent logs from $MINIKUBE_HOME/profiles/<name>/events.json
func readEventLog(name string) ([]cloudevents.Event, time.Time, error) {
	path := localpath.EventLog(name)

	st, err := os.Stat(path)
	if err != nil {
		return nil, time.Time{}, errors.Wrap(err, "stat")
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, st.ModTime(), errors.Wrap(err, "open")
	}
	var events []cloudevents.Event

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		ev := cloudevents.NewEvent()
		if err = json.Unmarshal(scanner.Bytes(), &ev); err != nil {
			return events, st.ModTime(), err
		}
		events = append(events, ev)
	}

	return events, st.ModTime(), scanner.Err()
}

// statusCode returns a status code number given a name
func statusCode(st string) int {
	// legacy names
	switch st {
	case "Running", "Configured":
		return OK
	case "Misconfigured":
		return Error
	}

	// new names
	for code, name := range codeNames {
		if name == st {
			return code
		}
	}

	return Unknown
}
