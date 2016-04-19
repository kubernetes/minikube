package mcndockerclient

import (
	"fmt"

	"github.com/docker/machine/libmachine/auth"
)

type URLer interface {
	// URL returns the Docker host URL
	URL() (string, error)
}

type AuthOptionser interface {
	// AuthOptions returns the authOptions
	AuthOptions() *auth.Options
}

type DockerHost interface {
	URLer
	AuthOptionser
}

type RemoteDocker struct {
	HostURL    string
	AuthOption *auth.Options
}

// URL returns the Docker host URL
func (rd *RemoteDocker) URL() (string, error) {
	if rd.HostURL == "" {
		return "", fmt.Errorf("Docker Host URL not set")
	}

	return rd.HostURL, nil
}

// AuthOptions returns the authOptions
func (rd *RemoteDocker) AuthOptions() *auth.Options {
	return rd.AuthOption
}
