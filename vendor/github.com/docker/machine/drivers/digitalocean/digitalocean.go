package digitalocean

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"time"

	"github.com/digitalocean/godo"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
	"golang.org/x/oauth2"
)

type Driver struct {
	*drivers.BaseDriver
	AccessToken       string
	DropletID         int
	DropletName       string
	Image             string
	Region            string
	SSHKeyID          int
	SSHKeyFingerprint string
	Size              string
	IPv6              bool
	Backups           bool
	PrivateNetworking bool
	UserDataFile      string
}

const (
	defaultSSHPort = 22
	defaultSSHUser = "root"
	defaultImage   = "ubuntu-15-10-x64"
	defaultRegion  = "nyc3"
	defaultSize    = "512mb"
)

// GetCreateFlags registers the flags this driver adds to
// "docker hosts create"
func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			EnvVar: "DIGITALOCEAN_ACCESS_TOKEN",
			Name:   "digitalocean-access-token",
			Usage:  "Digital Ocean access token",
		},
		mcnflag.StringFlag{
			EnvVar: "DIGITALOCEAN_SSH_USER",
			Name:   "digitalocean-ssh-user",
			Usage:  "SSH username",
			Value:  defaultSSHUser,
		},
		mcnflag.StringFlag{
			EnvVar: "DIGITALOCEAN_SSH_KEY_FINGERPRINT",
			Name:   "digitalocean-ssh-key-fingerprint",
			Usage:  "SSH key fingerprint",
		},
		mcnflag.IntFlag{
			EnvVar: "DIGITALOCEAN_SSH_PORT",
			Name:   "digitalocean-ssh-port",
			Usage:  "SSH port",
			Value:  defaultSSHPort,
		},
		mcnflag.StringFlag{
			EnvVar: "DIGITALOCEAN_IMAGE",
			Name:   "digitalocean-image",
			Usage:  "Digital Ocean Image",
			Value:  defaultImage,
		},
		mcnflag.StringFlag{
			EnvVar: "DIGITALOCEAN_REGION",
			Name:   "digitalocean-region",
			Usage:  "Digital Ocean region",
			Value:  defaultRegion,
		},
		mcnflag.StringFlag{
			EnvVar: "DIGITALOCEAN_SIZE",
			Name:   "digitalocean-size",
			Usage:  "Digital Ocean size",
			Value:  defaultSize,
		},
		mcnflag.BoolFlag{
			EnvVar: "DIGITALOCEAN_IPV6",
			Name:   "digitalocean-ipv6",
			Usage:  "enable ipv6 for droplet",
		},
		mcnflag.BoolFlag{
			EnvVar: "DIGITALOCEAN_PRIVATE_NETWORKING",
			Name:   "digitalocean-private-networking",
			Usage:  "enable private networking for droplet",
		},
		mcnflag.BoolFlag{
			EnvVar: "DIGITALOCEAN_BACKUPS",
			Name:   "digitalocean-backups",
			Usage:  "enable backups for droplet",
		},
		mcnflag.StringFlag{
			EnvVar: "DIGITALOCEAN_USERDATA",
			Name:   "digitalocean-userdata",
			Usage:  "path to file with cloud-init user-data",
		},
	}
}

func NewDriver(hostName, storePath string) *Driver {
	return &Driver{
		Image:  defaultImage,
		Size:   defaultSize,
		Region: defaultRegion,
		BaseDriver: &drivers.BaseDriver{
			MachineName: hostName,
			StorePath:   storePath,
		},
	}
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return "digitalocean"
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.AccessToken = flags.String("digitalocean-access-token")
	d.Image = flags.String("digitalocean-image")
	d.Region = flags.String("digitalocean-region")
	d.Size = flags.String("digitalocean-size")
	d.IPv6 = flags.Bool("digitalocean-ipv6")
	d.PrivateNetworking = flags.Bool("digitalocean-private-networking")
	d.Backups = flags.Bool("digitalocean-backups")
	d.UserDataFile = flags.String("digitalocean-userdata")
	d.SSHUser = flags.String("digitalocean-ssh-user")
	d.SSHPort = flags.Int("digitalocean-ssh-port")
	d.SSHKeyFingerprint = flags.String("digitalocean-ssh-key-fingerprint")
	d.SetSwarmConfigFromFlags(flags)

	if d.AccessToken == "" {
		return fmt.Errorf("digitalocean driver requires the --digitalocean-access-token option")
	}

	return nil
}

func (d *Driver) PreCreateCheck() error {
	if d.UserDataFile != "" {
		if _, err := os.Stat(d.UserDataFile); os.IsNotExist(err) {
			return fmt.Errorf("user-data file %s could not be found", d.UserDataFile)
		}
	}

	client := d.getClient()
	regions, _, err := client.Regions.List(nil)
	if err != nil {
		return err
	}
	for _, region := range regions {
		if region.Slug == d.Region {
			return nil
		}
	}

	return fmt.Errorf("digitalocean requires a valid region")
}

func (d *Driver) Create() error {
	var userdata string
	if d.UserDataFile != "" {
		buf, err := ioutil.ReadFile(d.UserDataFile)
		if err != nil {
			return err
		}
		userdata = string(buf)
	}

	log.Infof("Creating SSH key...")

	key, err := d.createSSHKey()
	if err != nil {
		return err
	}

	d.SSHKeyID = key.ID

	log.Infof("Creating Digital Ocean droplet...")

	client := d.getClient()

	createRequest := &godo.DropletCreateRequest{
		Image:             godo.DropletCreateImage{Slug: d.Image},
		Name:              d.MachineName,
		Region:            d.Region,
		Size:              d.Size,
		IPv6:              d.IPv6,
		PrivateNetworking: d.PrivateNetworking,
		Backups:           d.Backups,
		UserData:          userdata,
		SSHKeys:           []godo.DropletCreateSSHKey{{ID: d.SSHKeyID}},
	}

	newDroplet, _, err := client.Droplets.Create(createRequest)
	if err != nil {
		return err
	}

	d.DropletID = newDroplet.ID

	log.Info("Waiting for IP address to be assigned to the Droplet...")
	for {
		newDroplet, _, err = client.Droplets.Get(d.DropletID)
		if err != nil {
			return err
		}
		for _, network := range newDroplet.Networks.V4 {
			if network.Type == "public" {
				d.IPAddress = network.IPAddress
			}
		}

		if d.IPAddress != "" {
			break
		}

		time.Sleep(1 * time.Second)
	}

	log.Debugf("Created droplet ID %d, IP address %s",
		newDroplet.ID,
		d.IPAddress)

	return nil
}

func (d *Driver) createSSHKey() (*godo.Key, error) {
	if d.SSHKeyFingerprint != "" {
		key, resp, err := d.getClient().Keys.GetByFingerprint(d.SSHKeyFingerprint)
		if err != nil && resp.StatusCode == 404 {
			return nil, fmt.Errorf("Digital Ocean SSH key with fingerprint %s doesn't exist", d.SSHKeyFingerprint)
		}
		return key, err
	}

	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return nil, err
	}

	publicKey, err := ioutil.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return nil, err
	}

	createRequest := &godo.KeyCreateRequest{
		Name:      d.MachineName,
		PublicKey: string(publicKey),
	}

	key, _, err := d.getClient().Keys.Create(createRequest)
	if err != nil {
		return key, err
	}

	return key, nil
}

func (d *Driver) GetURL() (string, error) {
	if err := drivers.MustBeRunning(d); err != nil {
		return "", err
	}

	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("tcp://%s", net.JoinHostPort(ip, "2376")), nil
}

func (d *Driver) GetState() (state.State, error) {
	droplet, _, err := d.getClient().Droplets.Get(d.DropletID)
	if err != nil {
		return state.Error, err
	}
	switch droplet.Status {
	case "new":
		return state.Starting, nil
	case "active":
		return state.Running, nil
	case "off":
		return state.Stopped, nil
	}
	return state.None, nil
}

func (d *Driver) Start() error {
	_, _, err := d.getClient().DropletActions.PowerOn(d.DropletID)
	return err
}

func (d *Driver) Stop() error {
	_, _, err := d.getClient().DropletActions.Shutdown(d.DropletID)
	return err
}

func (d *Driver) Restart() error {
	_, _, err := d.getClient().DropletActions.Reboot(d.DropletID)
	return err
}

func (d *Driver) Kill() error {
	_, _, err := d.getClient().DropletActions.PowerOff(d.DropletID)
	return err
}

func (d *Driver) Remove() error {
	client := d.getClient()
	if d.SSHKeyFingerprint == "" {
		if resp, err := client.Keys.DeleteByID(d.SSHKeyID); err != nil {
			if resp.StatusCode == 404 {
				log.Infof("Digital Ocean SSH key doesn't exist, assuming it is already deleted")
			} else {
				return err
			}
		}
	}
	if resp, err := client.Droplets.Delete(d.DropletID); err != nil {
		if resp.StatusCode == 404 {
			log.Infof("Digital Ocean droplet doesn't exist, assuming it is already deleted")
		} else {
			return err
		}
	}
	return nil
}

func (d *Driver) getClient() *godo.Client {
	token := &oauth2.Token{AccessToken: d.AccessToken}
	tokenSource := oauth2.StaticTokenSource(token)
	client := oauth2.NewClient(oauth2.NoContext, tokenSource)

	return godo.NewClient(client)
}

func (d *Driver) GetSSHKeyPath() string {
	// don't set SSHKeyPath when using an existing key fingerprint
	if d.SSHKeyPath == "" && d.SSHKeyFingerprint == "" {
		d.SSHKeyPath = d.ResolveStorePath("id_rsa")
	}
	return d.SSHKeyPath
}

func (d *Driver) publicSSHKeyPath() string {
	return d.GetSSHKeyPath() + ".pub"
}
