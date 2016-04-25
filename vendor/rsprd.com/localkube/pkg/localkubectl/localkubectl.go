// Package localkubectl allows the lifecycle of the localkube container to be controlled
package localkubectl

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/mitchellh/go-homedir"
)

var (
	// DefaultHostDataDir is the directory on the host which is mounted for localkube data if no directory is specified
	DefaultHostDataDir = "~/.localkube/data"

	// LocalkubeDefaultTag is the tag to use for the localkube image
	LocalkubeDefaultTag = "latest"

	// LocalkubeClusterName is the name the cluster configuration is stored under
	LocalkubeClusterName = "localkube"

	// LocalkubeContext is the name of the context used by localkube
	LocalkubeContext = "localkube"
)

// Command returns a Command with subcommands for starting and stopping a localkube cluster
func Command(out io.Writer) *cli.Command {
	l := log.New(out, "", 0)
	return &cli.Command{
		Name:        "cluster",
		Description: "Manages localkube Kubernetes development environment",
		Subcommands: []cli.Command{
			{
				Name:        "start",
				Usage:       "spread cluster start [-t <tag>] [ClusterDataDirectory]",
				Description: "Starts the localkube cluster",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name: "t",
						Value: LocalkubeDefaultTag,
						Usage: "specifies localkube image tag to use, default is latest",
					},
				},
				Action: func(c *cli.Context) {
					// create new Docker client
					ctlr, err := NewControllerFromEnv(out)
					if err != nil {
						l.Fatal(err)
					}

					// create (if needed) and start localkube container
					err = startCluster(ctlr, l, c)
					if err != nil {
						l.Fatalf("could not start localkube: %v", err)
					}

					// guess which IP the API server will be on
					host, err := identifyHost(ctlr.Endpoint())
					if err != nil {
						l.Fatal(err)
					}

					// use default port
					host = fmt.Sprintf("%s:8080", host)

					// setup localkube kubectl context
					currentContext, err := GetCurrentContext()
					if err != nil {
						l.Fatal(err)
					}

					// set as current if no CurrentContext is set
					setCurrent := (len(currentContext) == 0)

					err = SetupContext(LocalkubeClusterName, LocalkubeContext, host, setCurrent)
					if err != nil {
						l.Fatal(err)
					}

					// display help text messages if context change
					if setCurrent {
						l.Printf("Created `%s` context and set it as current.\n", LocalkubeContext)
					} else if currentContext != LocalkubeContext {
						l.Println(SwitchContextInstructions(LocalkubeContext))
					}
				},
			},
			{
				Name:        "stop",
				Usage:       "spread cluster stop [-r]",
				Description: "Stops the localkube cluster",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name: "r",
						Usage: "removes container",
					},
				},
				Action: func(c *cli.Context) {
					ctlr, err := NewControllerFromEnv(out)
					if err != nil {
						l.Fatal(err)
					}

					remove := c.Bool("r")

					ctrs, err := ctlr.ListLocalkubeCtrs(true)
					if err != nil {
						l.Fatal(err)
					}

					for _, ctr := range ctrs {
						if remove || runningStatus(ctr.Status) {
							ctlr.StopCtr(ctr.ID, remove)
						}
					}
				},
			},
		},
	}
}

// startCluster configures and starts a cluster using command line parameters
func startCluster(ctlr *Controller, log *log.Logger, c *cli.Context) error {
	var err error

	// set data directory
	dataDir := c.Args().First()
	if len(dataDir) == 0 {
		dataDir, err = homedir.Expand(DefaultHostDataDir)
		if err != nil {
			return fmt.Errorf("Unable to expand home directory: %v", err)
		}
	}

	// set tag
	tag := c.String("t")

	// check if localkube container exists
	ctrId, running, err := ctlr.OnlyLocalkubeCtr()
	if err != nil {
		if err == ErrNoContainer {
			// if container doesn't exist, create
			ctrId, running, err = ctlr.CreateCtr(LocalkubeContainerName, tag)
			if err != nil {
				return err
			}
		} else {
			// stop for all other errors
			return err
		}
	}

	// start container if not running
	if !running {
		err = ctlr.StartCtr(ctrId, dataDir)
		if err != nil {
			return err
		}
	} else {
		log.Println("Localkube is already running")
	}
	return nil
}

func identifyHost(endpoint string) (string, error) {
	beginPort := strings.LastIndex(endpoint, ":")
	switch {
	// if using TCP use provided host
	case strings.HasPrefix(endpoint, "tcp://"):
		return endpoint[6:beginPort], nil
	// assuming localhost if Unix
	// TODO: Make this customizable
	case strings.HasPrefix(endpoint, "unix://"):
		return "127.0.0.1", nil
	}
	return "", fmt.Errorf("Could not determine localkube API server from endpoint `%s`", endpoint)
}
