// Package hyperkit provides a Go wrapper around the hyperkit
// command. It currently shells out to start hyperkit with the
// provided configuration.
//
// Most of the arguments should be self explanatory, but console
// handling deserves a mention. If the Console is configured with
// ConsoleStdio, the hyperkit is started with stdin, stdout, and
// stderr plumbed through to the VM console. If Console is set to
// ConsoleFile hyperkit the console output is redirected to a file and
// console input is disabled. For this mode StateDir has to be set and
// the interactive console is accessible via a 'tty' file created
// there.
//
// Currently this module has some limitations:
// - Only supports zero or one disk image
// - Only support zero or one network interface connected to VPNKit
// - Only kexec boot
//
// This package is currently implemented by shelling out a hyperkit
// process. In the future we may change this to become a wrapper
// around the hyperkit library.
package hyperkit

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/go-ps"
)

const (
	// ConsoleStdio configures console to use Stdio
	ConsoleStdio = iota
	// ConsoleFile configures console to a tty and output to a file
	ConsoleFile

	defaultVPNKitSock = "Library/Containers/com.docker.docker/Data/s50"

	defaultCPUs   = 1
	defaultMemory = 1024 // 1G

	defaultVSockGuestCID = 3

	jsonFile = "hyperkit.json"
	pidFile  = "hyperkit.pid"
)

var defaultHyperKits = []string{"hyperkit",
	"com.docker.hyperkit",
	"/usr/local/bin/hyperkit",
	"/Applications/Docker.app/Contents/Resources/bin/hyperkit",
	"/Applications/Docker.app/Contents/MacOS/com.docker.hyperkit"}

// Socket9P contains a unix domain socket path and 9p tag
type Socket9P struct {
	Path string `json:"path"`
	Tag  string `json:"tag"`
}

// DiskConfig contains the path to a disk image and an optional size if the image needs to be created.
type DiskConfig struct {
	Path   string `json:"path"`
	Size   int    `json:"size"`
	Format string `json:"format"`
	Driver string `json:"driver"`
}

// HyperKit contains the configuration of the hyperkit VM
type HyperKit struct {
	// HyperKit is the path to the hyperkit binary
	HyperKit string `json:"hyperkit"`
	// StateDir is the directory where runtime state is kept. If left empty, no state will be kept.
	StateDir string `json:"state_dir"`
	// VPNKitSock is the location of the VPNKit socket used for networking.
	VPNKitSock string `json:"vpnkit_sock"`
	// VPNKitKey is a string containing a UUID, it can be used in conjunction with VPNKit to get consistent IP address.
	VPNKitKey string `json:"vpnkit_key"`
	// UUID is a string containing a UUID, it sets BIOS DMI UUID for the VM (as found in /sys/class/dmi/id/product_uuid on Linux).
	UUID string `json:"uuid"`
	// Disks contains disk images to use/create.
	Disks []DiskConfig `json:"disks"`
	// ISOImage is the (optional) path to a ISO image to attach
	ISOImage string `json:"iso"`
	// VSock enables the virtio-socket device and exposes it on the host
	VSock bool `json:"vsock"`
	// VSockPorts is a list of guest VSock ports that should be exposed as sockets on the host
	VSockPorts []int `json:"vsock_ports"`
	// VSock guest CID
	VSockGuestCID int `json:"vsock_guest_cid"`

	// VMNet whether to create vmnet network
	VMNet bool `json:"vmnet"`

	// 9P sockets
	Sockets9P []Socket9P `json:"9p_sockets"`

	// Kernel is the path to the kernel image to boot
	Kernel string `json:"kernel"`
	// Initrd is the path to the initial ramdisk to boot off
	Initrd string `json:"initrd"`

	// CPUs is the number CPUs to configure
	CPUs int `json:"cpus"`
	// Memory is the amount of megabytes of memory for the VM
	Memory int `json:"memory"`

	// Console defines where the console of the VM should be
	// connected to. ConsoleStdio and ConsoleFile are supported.
	Console int `json:"console"`

	// Below here are internal members, but they are exported so
	// that they are written to the state json file, if configured.

	// Pid of the hyperkit process
	Pid int `json:"pid"`
	// Arguments used to execute the hyperkit process
	Arguments []string `json:"arguments"`
	// CmdLine is a single string of the command line
	CmdLine string `json:"cmdline"`

	process    *os.Process
	background bool
	log        *log.Logger
}

// New creates a template config structure.
// - If hyperkit can't be found an error is returned.
// - If vpnkitsock is empty no networking is configured. If it is set
//   to "auto" it tries to re-use the Docker for Mac VPNKit connection.
// - If statedir is "" no state is written to disk.
func New(hyperkit, vpnkitsock, statedir string) (*HyperKit, error) {
	h := HyperKit{}
	var err error

	h.HyperKit, err = checkHyperKit(hyperkit)
	if err != nil {
		return nil, err
	}
	h.StateDir = statedir
	h.VPNKitSock, err = checkVPNKitSock(vpnkitsock)
	if err != nil {
		return nil, err
	}

	h.CPUs = defaultCPUs
	h.Memory = defaultMemory

	h.VSockGuestCID = defaultVSockGuestCID

	h.Console = ConsoleStdio

	return &h, nil
}

// FromState reads a json file from statedir and populates a HyperKit structure.
func FromState(statedir string) (*HyperKit, error) {
	b, err := ioutil.ReadFile(filepath.Join(statedir, jsonFile))
	if err != nil {
		return nil, fmt.Errorf("Can't read json file: %s", err)
	}
	h := &HyperKit{}
	err = json.Unmarshal(b, h)
	if err != nil {
		return nil, fmt.Errorf("Can't parse json file: %s", err)
	}

	// Make sure the pid written by hyperkit is the same as in the json
	d, err := ioutil.ReadFile(filepath.Join(statedir, pidFile))
	if err != nil {
		return nil, err
	}
	pid, err := strconv.Atoi(string(d[:]))
	if err != nil {
		return nil, err
	}
	if h.Pid != pid {
		return nil, fmt.Errorf("pids do not match %d != %d", h.Pid, pid)
	}

	h.process, err = os.FindProcess(h.Pid)
	if err != nil {
		return nil, err
	}

	return h, nil
}

// SetLogger sets the log instance to use for the output of the hyperkit process itself (not the console of the VM).
// This is only relevant when Console is set to ConsoleFile
func (h *HyperKit) SetLogger(logger *log.Logger) {
	h.log = logger
}

// Run the VM with a given command line until it exits
func (h *HyperKit) Run(cmdline string) error {
	h.background = false
	return h.execute(cmdline)
}

// Start the VM with a given command line in the background
func (h *HyperKit) Start(cmdline string) error {
	h.background = true
	return h.execute(cmdline)
}

func (h *HyperKit) execute(cmdline string) error {
	var err error
	// Sanity checks on configuration
	if h.Console == ConsoleFile && h.StateDir == "" {
		return fmt.Errorf("If ConsoleFile is set, StateDir must be specified")
	}
	if h.Console == ConsoleStdio && !isTerminal(os.Stdout) && h.StateDir == "" {
		return fmt.Errorf("If ConsoleStdio is set but stdio is not a terminal, StateDir must be specified")
	}
	if h.ISOImage != "" {
		if _, err = os.Stat(h.ISOImage); os.IsNotExist(err) {
			return fmt.Errorf("ISO %s does not exist", h.ISOImage)
		}
	}
	if h.VSock && h.StateDir == "" {
		return fmt.Errorf("If virtio-sockets are enabled, StateDir must be specified")
	}
	if !h.VSock && len(h.VSockPorts) > 0 {
		return fmt.Errorf("To forward vsock ports vsock must be enabled")
	}
	if _, err = os.Stat(h.Kernel); os.IsNotExist(err) {
		return fmt.Errorf("Kernel %s does not exist", h.Kernel)
	}
	if _, err = os.Stat(h.Initrd); os.IsNotExist(err) {
		return fmt.Errorf("initrd %s does not exist", h.Initrd)
	}

	// Create files
	if h.StateDir != "" {
		err = os.MkdirAll(h.StateDir, 0755)
		if err != nil {
			return err
		}
	}

	for idx, config := range h.Disks {
		if config.Path == "" {
			if h.StateDir == "" {
				return fmt.Errorf("Unable to create disk image when neither path nor state dir is set")
			}
			if config.Size <= 0 {
				return fmt.Errorf("Unable to create disk image when size is 0 or not set")
			}
			config.Path = fmt.Sprintf(filepath.Clean(filepath.Join(h.StateDir, "disk%02d.img")), idx)
			h.Disks[idx] = config
		}
		if _, err = os.Stat(config.Path); os.IsNotExist(err) {
			if config.Size != 0 {
				err = CreateDiskImage(config.Path, config.Size)
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("Disk image %s not found and unable to create it as size is not specified", config.Path)
			}
		}
	}

	// Run
	h.buildArgs(cmdline)
	err = h.execHyperKit()
	if err != nil {
		return err
	}

	return nil
}

// Stop the running VM
func (h *HyperKit) Stop() error {
	if h.process == nil {
		return fmt.Errorf("hyperkit process not known")
	}
	if !h.IsRunning() {
		return nil
	}
	err := h.process.Kill()
	if err != nil {
		return err
	}

	return nil
}

// IsRunning returns true if the hyperkit process is running.
func (h *HyperKit) IsRunning() bool {
	// os.FindProcess on Unix always returns a process object even
	// if the process does not exists. There does not seem to be
	// a call to find out if the process is running either, so we
	// use another package to find out.
	proc, err := ps.FindProcess(h.Pid)
	if err != nil {
		return false
	}
	if proc == nil {
		return false
	}
	return true
}

// isDisk checks if the specified path is used as a disk image
func (h *HyperKit) isDisk(path string) bool {
	for _, config := range h.Disks {
		if filepath.Clean(path) == filepath.Clean(config.Path) {
			return true
		}
	}
	return false
}

// Remove deletes all statefiles if present.
// This also removes the StateDir if empty.
// If keepDisk is set, the disks will not get removed.
func (h *HyperKit) Remove(keepDisk bool) error {
	if h.IsRunning() {
		return fmt.Errorf("Can't remove state as process is running")
	}
	if h.StateDir == "" {
		// If there is not state directory we don't mess with the disk
		return nil
	}

	if !keepDisk {
		return os.RemoveAll(h.StateDir)
	}

	files, _ := ioutil.ReadDir(h.StateDir)
	for _, f := range files {
		fn := filepath.Clean(filepath.Join(h.StateDir, f.Name()))
		if h.isDisk(fn) {
			continue
		}
		err := os.Remove(fn)
		if err != nil {
			return err
		}
	}
	return nil
}

// Convert to json string
func (h *HyperKit) String() string {
	s, err := json.Marshal(h)
	if err != nil {
		return err.Error()
	}
	return string(s)
}

// CreateDiskImage creates a empty file suitable for use as a disk image for a hyperkit VM.
func CreateDiskImage(location string, sizeMB int) error {
	diskDir := path.Dir(location)
	if diskDir != "." {
		if err := os.MkdirAll(diskDir, 0755); err != nil {
			return err
		}
	}

	f, err := os.Create(location)
	if err != nil {
		return err
	}
	defer f.Close()

	buf := make([]byte, 1048676)
	for i := 0; i < sizeMB; i++ {
		f.Write(buf)
	}
	return nil
}

func intArrayToString(i []int, sep string) string {
	if len(i) == 0 {
		return ""
	}
	s := make([]string, len(i))
	for idx := range i {
		s[idx] = strconv.Itoa(i[idx])
	}
	return strings.Join(s, sep)
}

func (h *HyperKit) buildArgs(cmdline string) {
	a := []string{"-A", "-u"}
	if h.StateDir != "" {
		a = append(a, "-F", filepath.Join(h.StateDir, pidFile))
	}

	a = append(a, "-c", fmt.Sprintf("%d", h.CPUs))
	a = append(a, "-m", fmt.Sprintf("%dM", h.Memory))

	a = append(a, "-s", "0:0,hostbridge")
	a = append(a, "-s", "31,lpc")

	nextSlot := 1

	if h.VPNKitSock != "" {
		if h.VPNKitKey == "" {
			a = append(a, "-s", fmt.Sprintf("%d:0,virtio-vpnkit,path=%s", nextSlot, h.VPNKitSock))
		} else {
			a = append(a, "-s", fmt.Sprintf("%d:0,virtio-vpnkit,path=%s,uuid=%s", nextSlot, h.VPNKitSock, h.VPNKitKey))
		}
		nextSlot++
	}

	if h.VMNet {
		a = append(a, "-s", fmt.Sprintf("%d:0,virtio-net", nextSlot))
		nextSlot++
	}

	if h.UUID != "" {
		a = append(a, "-U", h.UUID)
	}

	for _, p := range h.Disks {
		// Default the driver to virtio-blk
		driver := "virtio-blk"
		if p.Driver != "" {
			driver = p.Driver
		}
		arg := fmt.Sprintf("%d:0,%s,%s", nextSlot, driver, p.Path)

		// Add on a format instruction if specified.
		if p.Format != "" {
			arg += ",format=" + p.Format
		}
		a = append(a, "-s", arg)
		nextSlot++
	}

	if h.VSock {
		l := fmt.Sprintf("%d,virtio-sock,guest_cid=%d,path=%s", nextSlot, h.VSockGuestCID, h.StateDir)
		if len(h.VSockPorts) > 0 {
			l = fmt.Sprintf("%s,guest_forwards=%s", l, intArrayToString(h.VSockPorts, ";"))
		}
		a = append(a, "-s", l)
		nextSlot++
	}

	if h.ISOImage != "" {
		a = append(a, "-s", fmt.Sprintf("%d,ahci-cd,%s", nextSlot, h.ISOImage))
		nextSlot++
	}

	a = append(a, "-s", fmt.Sprintf("%d,virtio-rnd", nextSlot))
	nextSlot++

	for _, p := range h.Sockets9P {
		a = append(a, "-s", fmt.Sprintf("%d,virtio-9p,path=%s,tag=%s", nextSlot, p.Path, p.Tag))
		nextSlot++
	}

	if h.Console == ConsoleStdio && isTerminal(os.Stdout) {
		a = append(a, "-l", "com1,stdio")
	} else if h.StateDir != "" {
		a = append(a, "-l", fmt.Sprintf("com1,autopty=%s/tty,log=%s/console-ring", h.StateDir, h.StateDir))
	}

	kernArgs := fmt.Sprintf("kexec,%s,%s,earlyprintk=serial %s", h.Kernel, h.Initrd, cmdline)
	a = append(a, "-f", kernArgs)

	h.Arguments = a
	h.CmdLine = h.HyperKit + " " + strings.Join(a, " ")
}

// Execute hyperkit and plumb stdin/stdout/stderr.
func (h *HyperKit) execHyperKit() error {

	cmd := exec.Command(h.HyperKit, h.Arguments...)
	cmd.Env = os.Environ()

	// Plumb in stdin/stdout/stderr.
	//
	// If ConsoleStdio is configured and we are on a terminal,
	// just plugin stdio. If we are not on a terminal we have a
	// StateDir (as per checks above) and have configured HyperKit
	// to use a PTY in the statedir. In this case, we just open
	// the PTY slave and copy it to stdout (and ignore
	// stdin). This allows for redirecting of the VM output, used
	// for testing.
	//
	// If a logger is specified, use it for stdout/stderr
	// logging. Otherwise use the default /dev/null.
	if h.Console == ConsoleStdio {
		if isTerminal(os.Stdout) {
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		} else {
			go func() {
				ttyPath := fmt.Sprintf("%s/tty", h.StateDir)
				var tty *os.File
				var err error
				for {
					tty, err = os.OpenFile(ttyPath, os.O_RDONLY, 0)
					if err != nil {
						time.Sleep(10 * 1000 * 1000 * time.Nanosecond)
						continue
					} else {
						break
					}
				}
				saneTerminal(tty)
				setRaw(tty)
				io.Copy(os.Stdout, tty)
				tty.Close()
			}()
		}
	} else if h.log != nil {
		stdoutChan := make(chan string)
		stderrChan := make(chan string)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return err
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			return err
		}
		stream(stdout, stdoutChan)
		stream(stderr, stderrChan)

		done := make(chan struct{})
		go func() {
			for {
				select {
				case stderrl := <-stderrChan:
					log.Printf("%s", stderrl)
				case stdoutl := <-stdoutChan:
					log.Printf("%s", stdoutl)
				case <-done:
					return
				}
			}
		}()
	}

	err := cmd.Start()
	if err != nil {
		return err
	}
	h.Pid = cmd.Process.Pid
	h.process = cmd.Process
	err = h.writeState()
	if err != nil {
		h.process.Kill()
		return err
	}
	if !h.background {
		err = cmd.Wait()
		if err != nil {
			return err
		}
	} else {
		// Make sure we reap the child when it exits
		go cmd.Wait()
	}
	return nil
}

// writeState write the state to a JSON file
func (h *HyperKit) writeState() error {
	if h.StateDir == "" {
		// This is not an error
		return nil
	}

	s, err := json.Marshal(h)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(h.StateDir, jsonFile), []byte(s), 0644)
}

func stream(r io.ReadCloser, dest chan<- string) {
	go func() {
		defer r.Close()
		reader := bufio.NewReader(r)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			dest <- line
		}
	}()
}

// checkHyperKit tries to find and/or validate the path of hyperkit
func checkHyperKit(hyperkit string) (string, error) {
	if hyperkit != "" {
		p, err := exec.LookPath(hyperkit)
		if err != nil {
			return "", fmt.Errorf("Could not find hyperkit executable %s: %s", hyperkit, err)
		}
		return p, nil
	}

	// Look in a number of default locations
	for _, hyperkit := range defaultHyperKits {
		p, err := exec.LookPath(hyperkit)
		if err == nil {
			return p, nil
		}
	}

	return "", fmt.Errorf("Could not find hyperkit executable")
}

// checkVPNKitSock tries to find and/or validate the path of the VPNKit socket
func checkVPNKitSock(vpnkitsock string) (string, error) {
	if vpnkitsock == "auto" {
		vpnkitsock = filepath.Join(getHome(), defaultVPNKitSock)
	}
	if vpnkitsock == "" {
		return "", nil
	}

	vpnkitsock = filepath.Clean(vpnkitsock)
	_, err := os.Stat(vpnkitsock)
	if err != nil {
		return "", err
	}
	return vpnkitsock, nil
}

func getHome() string {
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return os.Getenv("HOME")
}
