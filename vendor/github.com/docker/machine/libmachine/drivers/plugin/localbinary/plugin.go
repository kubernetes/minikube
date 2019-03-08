package localbinary

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/docker/machine/libmachine/log"
)

var (
	// Timeout where we will bail if we're not able to properly contact the
	// plugin server.
	defaultTimeout               = 10 * time.Second
	CurrentBinaryIsDockerMachine = false
	CoreDrivers                  = []string{"amazonec2", "azure", "digitalocean",
		"exoscale", "generic", "google", "hyperv", "none", "openstack",
		"rackspace", "softlayer", "virtualbox", "vmwarefusion",
		"vmwarevcloudair", "vmwarevsphere"}
)

const (
	pluginOut           = "(%s) %s"
	pluginErr           = "(%s) DBG | %s"
	PluginEnvKey        = "MACHINE_PLUGIN_TOKEN"
	PluginEnvVal        = "42"
	PluginEnvDriverName = "MACHINE_PLUGIN_DRIVER_NAME"
)

type PluginStreamer interface {
	// Return a channel for receiving the output of the stream line by
	// line.
	//
	// It happens to be the case that we do this all inside of the main
	// plugin struct today, but that may not be the case forever.
	AttachStream(*bufio.Scanner) <-chan string
}

type PluginServer interface {
	// Get the address where the plugin server is listening.
	Address() (string, error)

	// Serve kicks off the plugin server.
	Serve() error

	// Close shuts down the initialized server.
	Close() error
}

type McnBinaryExecutor interface {
	// Execute the driver plugin.  Returns scanners for plugin binary
	// stdout and stderr.
	Start() (*bufio.Scanner, *bufio.Scanner, error)

	// Stop reading from the plugins in question.
	Close() error
}

// DriverPlugin interface wraps the underlying mechanics of starting a driver
// plugin server and then figuring out where it can be dialed.
type DriverPlugin interface {
	PluginServer
	PluginStreamer
}

type Plugin struct {
	Executor    McnBinaryExecutor
	Addr        string
	MachineName string
	addrCh      chan string
	stopCh      chan struct{}
	timeout     time.Duration
}

type Executor struct {
	pluginStdout, pluginStderr io.ReadCloser
	DriverName                 string
	cmd                        *exec.Cmd
	binaryPath                 string
}

type ErrPluginBinaryNotFound struct {
	driverName string
	driverPath string
}

func (e ErrPluginBinaryNotFound) Error() string {
	return fmt.Sprintf("Driver %q not found. Do you have the plugin binary %q accessible in your PATH?", e.driverName, e.driverPath)
}

// driverPath finds the path of a driver binary by its name.
//  + If the driver is a core driver, there is no separate driver binary. We reuse current binary if it's `docker-machine`
// or we assume `docker-machine` is in the PATH.
//  + If the driver is NOT a core driver, then the separate binary must be in the PATH and it's name must be
// `docker-machine-driver-driverName`
func driverPath(driverName string) string {
	for _, coreDriver := range CoreDrivers {
		if coreDriver == driverName {
			if CurrentBinaryIsDockerMachine {
				return os.Args[0]
			}

			return "docker-machine"
		}
	}

	return fmt.Sprintf("docker-machine-driver-%s", driverName)
}

func NewPlugin(driverName string) (*Plugin, error) {
	driverPath := driverPath(driverName)
	binaryPath, err := exec.LookPath(driverPath)
	if err != nil {
		return nil, ErrPluginBinaryNotFound{driverName, driverPath}
	}

	log.Debugf("Found binary path at %s", binaryPath)

	return &Plugin{
		stopCh: make(chan struct{}),
		addrCh: make(chan string, 1),
		Executor: &Executor{
			DriverName: driverName,
			binaryPath: binaryPath,
		},
	}, nil
}

func (lbe *Executor) Start() (*bufio.Scanner, *bufio.Scanner, error) {
	var err error

	log.Debugf("Launching plugin server for driver %s", lbe.DriverName)

	lbe.cmd = exec.Command(lbe.binaryPath)

	lbe.pluginStdout, err = lbe.cmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("Error getting cmd stdout pipe: %s", err)
	}

	lbe.pluginStderr, err = lbe.cmd.StderrPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("Error getting cmd stderr pipe: %s", err)
	}

	outScanner := bufio.NewScanner(lbe.pluginStdout)
	errScanner := bufio.NewScanner(lbe.pluginStderr)

	os.Setenv(PluginEnvKey, PluginEnvVal)
	os.Setenv(PluginEnvDriverName, lbe.DriverName)

	if err := lbe.cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("Error starting plugin binary: %s", err)
	}

	return outScanner, errScanner, nil
}

func (lbe *Executor) Close() error {
	if err := lbe.cmd.Wait(); err != nil {
		return fmt.Errorf("Error waiting for binary close: %s", err)
	}

	return nil
}

func stream(scanner *bufio.Scanner, streamOutCh chan<- string, stopCh <-chan struct{}) {
	for scanner.Scan() {
		line := scanner.Text()
		if err := scanner.Err(); err != nil {
			log.Warnf("Scanning stream: %s", err)
		}
		select {
		case streamOutCh <- strings.Trim(line, "\n"):
		case <-stopCh:
			return
		}
	}
}

func (lbp *Plugin) AttachStream(scanner *bufio.Scanner) <-chan string {
	streamOutCh := make(chan string)
	go stream(scanner, streamOutCh, lbp.stopCh)
	return streamOutCh
}

func (lbp *Plugin) execServer() error {
	outScanner, errScanner, err := lbp.Executor.Start()
	if err != nil {
		return err
	}

	// Scan just one line to get the address, then send it to the relevant
	// channel.
	outScanner.Scan()
	addr := outScanner.Text()
	if err := outScanner.Err(); err != nil {
		return fmt.Errorf("Reading plugin address failed: %s", err)
	}

	lbp.addrCh <- strings.TrimSpace(addr)

	stdOutCh := lbp.AttachStream(outScanner)
	stdErrCh := lbp.AttachStream(errScanner)

	for {
		select {
		case out := <-stdOutCh:
			log.Infof(pluginOut, lbp.MachineName, out)
		case err := <-stdErrCh:
			log.Debugf(pluginErr, lbp.MachineName, err)
		case <-lbp.stopCh:
			if err := lbp.Executor.Close(); err != nil {
				return fmt.Errorf("Error closing local plugin binary: %s", err)
			}
			return nil
		}
	}
}

func (lbp *Plugin) Serve() error {
	return lbp.execServer()
}

func (lbp *Plugin) Address() (string, error) {
	if lbp.Addr == "" {
		if lbp.timeout == 0 {
			lbp.timeout = defaultTimeout
		}

		select {
		case lbp.Addr = <-lbp.addrCh:
			log.Debugf("Plugin server listening at address %s", lbp.Addr)
			close(lbp.addrCh)
			return lbp.Addr, nil
		case <-time.After(lbp.timeout):
			return "", fmt.Errorf("Failed to dial the plugin server in %s", lbp.timeout)
		}
	}
	return lbp.Addr, nil
}

func (lbp *Plugin) Close() error {
	close(lbp.stopCh)
	return nil
}
