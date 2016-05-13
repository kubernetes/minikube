package generic

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/state"
)

type Driver struct {
	*drivers.BaseDriver
	EnginePort int
	SSHKey     string
}

const (
	defaultTimeout = 1 * time.Second
)

// GetCreateFlags registers the flags this driver adds to
// "docker hosts create"
func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.IntFlag{
			Name:   "generic-engine-port",
			Usage:  "Docker engine port",
			Value:  engine.DefaultPort,
			EnvVar: "GENERIC_ENGINE_PORT",
		},
		mcnflag.StringFlag{
			Name:   "generic-ip-address",
			Usage:  "IP Address of machine",
			EnvVar: "GENERIC_IP_ADDRESS",
		},
		mcnflag.StringFlag{
			Name:   "generic-ssh-user",
			Usage:  "SSH user",
			Value:  drivers.DefaultSSHUser,
			EnvVar: "GENERIC_SSH_USER",
		},
		mcnflag.StringFlag{
			Name:   "generic-ssh-key",
			Usage:  "SSH private key path (if not provided, identities in ssh-agent will be used)",
			Value:  "",
			EnvVar: "GENERIC_SSH_KEY",
		},
		mcnflag.IntFlag{
			Name:   "generic-ssh-port",
			Usage:  "SSH port",
			Value:  drivers.DefaultSSHPort,
			EnvVar: "GENERIC_SSH_PORT",
		},
	}
}

// NewDriver creates and returns a new instance of the driver
func NewDriver(hostName, storePath string) drivers.Driver {
	return &Driver{
		EnginePort: engine.DefaultPort,
		BaseDriver: &drivers.BaseDriver{
			MachineName: hostName,
			StorePath:   storePath,
		},
	}
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return "generic"
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

func (d *Driver) GetSSHUsername() string {
	return d.SSHUser
}

func (d *Driver) GetSSHKeyPath() string {
	if d.SSHKey == "" {
		return ""
	}

	if d.SSHKeyPath == "" {
		d.SSHKeyPath = d.ResolveStorePath("id_rsa")
	}
	return d.SSHKeyPath
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.EnginePort = flags.Int("generic-engine-port")
	d.IPAddress = flags.String("generic-ip-address")
	d.SSHUser = flags.String("generic-ssh-user")
	d.SSHKey = flags.String("generic-ssh-key")
	d.SSHPort = flags.Int("generic-ssh-port")

	if d.IPAddress == "" {
		return errors.New("generic driver requires the --generic-ip-address option")
	}

	return nil
}

func (d *Driver) PreCreateCheck() error {
	if d.SSHKey != "" {
		if _, err := os.Stat(d.SSHKey); os.IsNotExist(err) {
			return fmt.Errorf("Ssh key does not exist: %q", d.SSHKey)
		}
	}

	return nil
}

func (d *Driver) Create() error {
	if d.SSHKey == "" {
		log.Info("No SSH key specified. Connecting to this machine now and in the" +
			" future will require the ssh agent to contain the appropriate key.")
	} else {
		log.Info("Importing SSH key...")
		// TODO: validate the key is a valid key
		if err := mcnutils.CopyFile(d.SSHKey, d.GetSSHKeyPath()); err != nil {
			return fmt.Errorf("unable to copy ssh key: %s", err)
		}

		if err := os.Chmod(d.GetSSHKeyPath(), 0600); err != nil {
			return fmt.Errorf("unable to set permissions on the ssh key: %s", err)
		}
	}

	log.Debugf("IP: %s", d.IPAddress)

	return nil
}

func (d *Driver) GetURL() (string, error) {
	if err := drivers.MustBeRunning(d); err != nil {
		return "", err
	}

	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("tcp://%s", net.JoinHostPort(ip, strconv.Itoa(d.EnginePort))), nil
}

func (d *Driver) GetState() (state.State, error) {
	address := net.JoinHostPort(d.IPAddress, strconv.Itoa(d.SSHPort))

	_, err := net.DialTimeout("tcp", address, defaultTimeout)
	if err != nil {
		return state.Stopped, nil
	}

	return state.Running, nil
}

func (d *Driver) Start() error {
	return errors.New("generic driver does not support start")
}

func (d *Driver) Stop() error {
	return errors.New("generic driver does not support stop")
}

func (d *Driver) Restart() error {
	_, err := drivers.RunSSHCommandFromDriver(d, "sudo shutdown -r now")
	return err
}

func (d *Driver) Kill() error {
	return errors.New("generic driver does not support kill")
}

func (d *Driver) Remove() error {
	return nil
}
