package localkubectl

import (
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	docker "github.com/fsouza/go-dockerclient"
)

var (
	// LocalkubeLabel is the label that identifies localkube containers
	LocalkubeLabel = "rsprd.com/name=localkube"

	// LocalkubeContainerName is the name of the container that localkube runs in
	LocalkubeContainerName = "localkube"

	// LocalkubeImageName is the image of localkube that is started
	LocalkubeImageName = "redspreadapps/localkube"

	// ContainerDataDir is the path inside the container for etcd data
	ContainerDataDir = "/var/localkube/data"

	// RedspreadName is a Redspread specific identifier for Name
	RedspreadName = "rsprd.com/name"
)

// Controller provides a wrapper around the Docker client for easy control of localkube.
type Controller struct {
	*docker.Client
	log *log.Logger
	out io.Writer
}

// NewController returns a localkube Docker client from a created *docker.Client
func NewController(client *docker.Client, out io.Writer) (*Controller, error) {
	_, err := client.Version()
	if err != nil {
		return nil, fmt.Errorf("Unable to establish connection with Docker daemon: %v", err)
	}

	logger := log.New(out, "", 0)

	return &Controller{
		Client: client,
		log:    logger,
		out:    out,
	}, nil
}

// NewControllerFromEnv creates a new Docker client using environment clues.
func NewControllerFromEnv(out io.Writer) (*Controller, error) {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return nil, fmt.Errorf("could not create Docker client: %v", err)
	}

	return NewController(client, out)
}

// ListLocalkubeCtrs returns a list of localkube containers on this Docker daemon. If a is true only non-running containers will also be displayed
func (d *Controller) ListLocalkubeCtrs(all bool) ([]docker.APIContainers, error) {
	ctrs, err := d.ListContainers(docker.ListContainersOptions{
		All: all,
		Filters: map[string][]string{
			"label": {LocalkubeLabel},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("could not list containers: %v", err)
	}

	return ctrs, nil
}

// OnlyLocalkubeCtr returns the localkube container. If there are multiple possible containers an error is returned
func (d *Controller) OnlyLocalkubeCtr() (ctrId string, running bool, err error) {
	ctrs, err := d.ListLocalkubeCtrs(true)
	if err != nil {
		return "", false, err
	} else if len(ctrs) < 1 {
		return "", false, ErrNoContainer
	} else if len(ctrs) > 1 {
		return "", false, ErrTooManyLocalkubes
	}

	ctr := ctrs[0]
	return ctr.ID, runningStatus(ctr.Status), nil
}

// CreateCtr creates the localkube container. The image is pulled if it doesn't already exist.
func (d *Controller) CreateCtr(name, imageTag string) (ctrId string, running bool, err error) {
	image := fmt.Sprintf("%s:%s", LocalkubeImageName, imageTag)
	ctrOpts := docker.CreateContainerOptions{
		Name: name,
		Config: &docker.Config{
			Hostname: name,
			Image:    image,
			Env: []string{
				fmt.Sprintf("KUBE_ETCD_DATA_DIRECTORY=%s", ContainerDataDir),
			},
			Labels: map[string]string{
				RedspreadName: name,
			},
			StopSignal: "SIGINT",
		},
	}

	d.log.Println("Creating localkube container...")
	ctr, err := d.CreateContainer(ctrOpts)
	if err != nil {
		if err == docker.ErrNoSuchImage {
			// if image does not exist, pull it
			d.log.Printf("Localkube image '%s' not found, pulling now:\n", image)
			if pullErr := d.PullImage(imageTag, false); pullErr != nil {
				return "", false, pullErr
			}
			return d.CreateCtr(name, imageTag)
		}
		return "", false, fmt.Errorf("could not create locakube container: %v", err)
	}
	return ctr.ID, ctr.State.Running, nil
}

// PullImage will pull the localkube image on the connected Docker daemon
func (d *Controller) PullImage(imageTag string, silent bool) error {
	pullOpts := docker.PullImageOptions{
		Repository: LocalkubeImageName,
		Tag:        imageTag,
	}

	in, out := io.Pipe()
	// print pull progress if not silent
	if !silent {
		pullOpts.RawJSONStream = true
		pullOpts.OutputStream = out
		outFd, isTerminal := term.GetFdInfo(d.out)
		go jsonmessage.DisplayJSONMessagesStream(in, d.out, outFd, isTerminal)
	}

	err := d.Client.PullImage(pullOpts, docker.AuthConfiguration{})
	if err != nil {
		return fmt.Errorf("failed to pull localkube image: %v", err)
	}
	return nil
}

// StartCtr starts the specified Container with options specific to localkube. If etcdDataDir is empty, no directory will be mounted for etcd.
func (d *Controller) StartCtr(ctrId, etcdDataDir string) error {
	binds := []string{
		"/sys:/sys:rw",
		"/var/lib/docker:/var/lib/docker",
		"/mnt/sda1/var/lib/docker:/mnt/sda1/var/lib/docker",
		"/var/lib/kubelet:/var/lib/kubelet",
		"/var/run:/var/run:rw",
		"/:/rootfs:ro",
	}

	// if provided mount etcd data dir
	if len(etcdDataDir) != 0 {
		dataBind := fmt.Sprintf("%s:%s", etcdDataDir, ContainerDataDir)
		binds = append(binds, dataBind)
	}

	hostCfg := &docker.HostConfig{
		Binds:         binds,
		RestartPolicy: docker.AlwaysRestart(),
		NetworkMode:   "host",
		PidMode:       "host",
		Privileged:    true,
	}

	d.log.Println("Starting localkube container...")
	err := d.StartContainer(ctrId, hostCfg)
	if err != nil {
		return fmt.Errorf("could not start container `%s`: %v", ctrId, err)
	}
	return nil
}

// StopCtr stops the container with the given id. If delete is true, deletes container after stopping.
func (d *Controller) StopCtr(ctrId string, remove bool) error {
	d.log.Printf("Stopping container '%s'\n", ctrId)
	err := d.StopContainer(ctrId, 5)
	if err != nil && !remove {
		return fmt.Errorf("unable to stop localkube container: %v", err)
	}

	if remove {
		removeCtrOpts := docker.RemoveContainerOptions{
			ID: ctrId,
		}

		d.log.Printf("Removing container '%s'\n", ctrId)
		err = d.RemoveContainer(removeCtrOpts)
		if err != nil {
			return fmt.Errorf("unable to remove localkube container: %v", err)
		}
	}
	return nil
}

// runningStatus returns true if a Docker status string indicates the container is running
func runningStatus(status string) bool {
	return strings.HasPrefix(status, "Up")
}

var (
	// ErrNoContainer is returned when the localkube container hasn't been created yet
	ErrNoContainer = errors.New("localkube container doesn't exist")

	// ErrTooManyLocalkubes is returned when there are more than one localkube containers on the Docker daemon
	ErrTooManyLocalkubes = errors.New("multiple localkube containers have been started")
)
