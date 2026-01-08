package persist

import (
	"github.com/docker/machine/libmachine/host"
)

type Store interface {
	// Exists returns whether a machine exists or not
	Exists(name string) (bool, error)

	// List returns a list of all hosts in the store
	List() ([]string, error)

	// Load loads a host by name
	Load(name string) (*host.Host, error)

	// Remove removes a machine from the store
	Remove(name string) error

	// Save persists a machine in the store
	Save(host *host.Host) error
}

func LoadHosts(s Store, hostNames []string) ([]*host.Host, map[string]error) {
	loadedHosts := []*host.Host{}
	errors := map[string]error{}

	for _, hostName := range hostNames {
		h, err := s.Load(hostName)
		if err != nil {
			errors[hostName] = err
		} else {
			loadedHosts = append(loadedHosts, h)
		}
	}

	return loadedHosts, errors
}

func LoadAllHosts(s Store) ([]*host.Host, map[string]error, error) {
	hostNames, err := s.List()
	if err != nil {
		return nil, nil, err
	}
	loadedHosts, hostInError := LoadHosts(s, hostNames)
	return loadedHosts, hostInError, nil
}
