package commands

import (
	"errors"
	"fmt"

	"time"

	"k8s.io/minikube/pkg/libmachine/libmachine"
	"k8s.io/minikube/pkg/libmachine/libmachine/persist"
	"k8s.io/minikube/pkg/libmachine/libmachine/state"
)

const (
	activeDefaultTimeout = 10
)

var (
	errNoActiveHost  = errors.New("No active host found")
	errActiveTimeout = errors.New("Error getting active host: timeout")
)

func cmdActive(c CommandLine, api libmachine.API) error {
	if len(c.Args()) > 0 {
		return ErrTooManyArguments
	}

	hosts, hostsInError, err := persist.LoadAllHosts(api)
	if err != nil {
		return fmt.Errorf("Error getting active host: %s", err)
	}

	timeout := time.Duration(c.Int("timeout")) * time.Second
	items := getHostListItems(hosts, hostsInError, timeout)

	active, err := activeHost(items)

	if err != nil {
		return err
	}

	fmt.Println(active.Name)
	return nil
}

func activeHost(items []HostListItem) (HostListItem, error) {
	timeout := false
	for _, item := range items {
		if item.ActiveHost || item.ActiveSwarm {
			return item, nil
		}
		if item.State == state.Timeout {
			timeout = true
		}
	}
	if timeout {
		return HostListItem{}, errActiveTimeout
	}
	return HostListItem{}, errNoActiveHost
}
