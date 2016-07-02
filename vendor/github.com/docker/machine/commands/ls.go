package commands

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"text/tabwriter"
	"text/template"
	"time"

	"io"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcndockerclient"
	"github.com/docker/machine/libmachine/persist"
	"github.com/docker/machine/libmachine/state"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/skarademir/naturalsort"
)

const (
	lsDefaultTimeout = 10
	tableFormatKey   = "table"
	lsDefaultFormat  = "table {{ .Name }}\t{{ .Active }}\t{{ .DriverName}}\t{{ .State }}\t{{ .URL }}\t{{ .Swarm }}\t{{ .DockerVersion }}\t{{ .Error}}"
)

var (
	headers = map[string]string{
		"Name":          "NAME",
		"Active":        "ACTIVE",
		"ActiveHost":    "ACTIVE_HOST",
		"ActiveSwarm":   "ACTIVE_SWARM",
		"DriverName":    "DRIVER",
		"State":         "STATE",
		"URL":           "URL",
		"SwarmOptions":  "SWARM_OPTIONS",
		"Swarm":         "SWARM",
		"EngineOptions": "ENGINE_OPTIONS",
		"Error":         "ERRORS",
		"DockerVersion": "DOCKER",
		"ResponseTime":  "RESPONSE",
	}
)

type HostListItem struct {
	Name          string
	Active        string
	ActiveHost    bool
	ActiveSwarm   bool
	DriverName    string
	State         state.State
	URL           string
	SwarmOptions  *swarm.Options
	Swarm         string
	EngineOptions *engine.Options
	Error         string
	DockerVersion string
	ResponseTime  time.Duration
}

// FilterOptions -
type FilterOptions struct {
	SwarmName  []string
	DriverName []string
	State      []string
	Name       []string
	Labels     []string
}

func cmdLs(c CommandLine, api libmachine.API) error {
	filters, err := parseFilters(c.StringSlice("filter"))
	if err != nil {
		return err
	}

	hostList, hostInError, err := persist.LoadAllHosts(api)
	if err != nil {
		return err
	}

	hostList = filterHosts(hostList, filters)

	// Just print out the names if we're being quiet
	if c.Bool("quiet") {
		for _, host := range hostList {
			fmt.Println(host.Name)
		}
		return nil
	}

	template, table, err := parseFormat(c.String("format"))
	if err != nil {
		return err
	}

	var w io.Writer
	if table {
		tabWriter := tabwriter.NewWriter(os.Stdout, 5, 1, 3, ' ', 0)
		defer tabWriter.Flush()

		w = tabWriter

		if err := template.Execute(w, headers); err != nil {
			return err
		}
	} else {
		w = os.Stdout
	}

	timeout := time.Duration(c.Int("timeout")) * time.Second
	items := getHostListItems(hostList, hostInError, timeout)

	swarmMasters := make(map[string]string)
	swarmInfo := make(map[string]string)

	for _, host := range hostList {
		if host.HostOptions != nil {
			swarmOptions := host.HostOptions.SwarmOptions
			if swarmOptions.Master {
				swarmMasters[swarmOptions.Discovery] = host.Name
			}

			if swarmOptions.Discovery != "" {
				swarmInfo[host.Name] = swarmOptions.Discovery
			}
		}
	}

	for _, item := range items {
		swarmColumn := ""
		if item.SwarmOptions != nil && item.SwarmOptions.Discovery != "" {
			swarmColumn = swarmMasters[item.SwarmOptions.Discovery]
			if item.SwarmOptions.Master {
				swarmColumn = fmt.Sprintf("%s (master)", swarmColumn)
			}
		}
		item.Swarm = swarmColumn

		if err := template.Execute(w, item); err != nil {
			return err
		}
	}

	return nil
}

func parseFormat(format string) (*template.Template, bool, error) {
	table := false
	finalFormat := format

	if finalFormat == "" {
		finalFormat = lsDefaultFormat
	}

	if strings.HasPrefix(finalFormat, tableFormatKey) {
		table = true
		finalFormat = finalFormat[len(tableFormatKey):]
	}

	finalFormat = strings.Trim(finalFormat, " ")
	r := strings.NewReplacer(`\t`, "\t", `\n`, "\n")
	finalFormat = r.Replace(finalFormat)

	template, err := template.New("").Parse(finalFormat + "\n")
	if err != nil {
		return nil, false, err
	}

	return template, table, nil
}

func parseFilters(filters []string) (FilterOptions, error) {
	options := FilterOptions{}
	for _, f := range filters {
		kv := strings.SplitN(f, "=", 2)
		if len(kv) != 2 {
			return options, errors.New("Unsupported filter syntax.")
		}
		key, value := strings.ToLower(kv[0]), kv[1]

		switch key {
		case "swarm":
			options.SwarmName = append(options.SwarmName, value)
		case "driver":
			options.DriverName = append(options.DriverName, value)
		case "state":
			options.State = append(options.State, value)
		case "name":
			options.Name = append(options.Name, value)
		case "label":
			options.Labels = append(options.Labels, value)
		default:
			return options, fmt.Errorf("Unsupported filter key '%s'", key)
		}
	}
	return options, nil
}

func filterHosts(hosts []*host.Host, filters FilterOptions) []*host.Host {
	if len(filters.SwarmName) == 0 &&
		len(filters.DriverName) == 0 &&
		len(filters.State) == 0 &&
		len(filters.Name) == 0 &&
		len(filters.Labels) == 0 {
		return hosts
	}

	filteredHosts := []*host.Host{}
	swarmMasters := getSwarmMasters(hosts)

	for _, h := range hosts {
		if filterHost(h, filters, swarmMasters) {
			filteredHosts = append(filteredHosts, h)
		}
	}
	return filteredHosts
}

func getSwarmMasters(hosts []*host.Host) map[string]string {
	swarmMasters := make(map[string]string)
	for _, h := range hosts {
		if h.HostOptions != nil {
			swarmOptions := h.HostOptions.SwarmOptions
			if swarmOptions != nil && swarmOptions.Master {
				swarmMasters[swarmOptions.Discovery] = h.Name
			}
		}
	}
	return swarmMasters
}

func filterHost(host *host.Host, filters FilterOptions, swarmMasters map[string]string) bool {
	swarmMatches := matchesSwarmName(host, filters.SwarmName, swarmMasters)
	driverMatches := matchesDriverName(host, filters.DriverName)
	stateMatches := matchesState(host, filters.State)
	nameMatches := matchesName(host, filters.Name)
	labelMatches := matchesLabel(host, filters.Labels)

	return swarmMatches && driverMatches && stateMatches && nameMatches && labelMatches
}

func matchesSwarmName(host *host.Host, swarmNames []string, swarmMasters map[string]string) bool {
	if len(swarmNames) == 0 {
		return true
	}
	for _, n := range swarmNames {
		if host.HostOptions != nil && host.HostOptions.SwarmOptions != nil {
			if strings.EqualFold(n, swarmMasters[host.HostOptions.SwarmOptions.Discovery]) {
				return true
			}
		}
	}
	return false
}

func matchesDriverName(host *host.Host, driverNames []string) bool {
	if len(driverNames) == 0 {
		return true
	}
	for _, n := range driverNames {
		if strings.EqualFold(host.DriverName, n) {
			return true
		}
	}
	return false
}

func matchesState(host *host.Host, states []string) bool {
	if len(states) == 0 {
		return true
	}
	for _, n := range states {
		s, err := host.Driver.GetState()
		if err != nil {
			log.Warn(err)
		}
		if strings.EqualFold(n, s.String()) {
			return true
		}
	}
	return false
}

func matchesName(host *host.Host, names []string) bool {
	if len(names) == 0 {
		return true
	}
	for _, n := range names {
		r, err := regexp.Compile(n)
		if err != nil {
			log.Error(err)
			os.Exit(1) // TODO: Can we get rid of this call, and exit 'properly' ?
		}
		if r.MatchString(host.Driver.GetMachineName()) {
			return true
		}
	}
	return false
}

func matchesLabel(host *host.Host, labels []string) bool {
	if len(labels) == 0 {
		return true
	}

	var englabels = make(map[string]string, len(host.HostOptions.EngineOptions.Labels))

	if host.HostOptions != nil && host.HostOptions.EngineOptions.Labels != nil {
		for _, s := range host.HostOptions.EngineOptions.Labels {
			kv := strings.SplitN(s, "=", 2)
			englabels[kv[0]] = kv[1]
		}
	}

	for _, l := range labels {
		kv := strings.SplitN(l, "=", 2)
		if val, exists := englabels[kv[0]]; exists && strings.EqualFold(val, kv[1]) {
			return true
		}
	}
	return false
}

// PERFORMANCE: The code of this function is complicated because we try
// to call the underlying drivers as less as possible to get the information
// we need.
func attemptGetHostState(h *host.Host, stateQueryChan chan<- HostListItem) {
	requestBeginning := time.Now()
	url := ""
	currentState := state.None
	dockerVersion := "Unknown"
	hostError := ""

	url, err := h.URL()

	// PERFORMANCE: if we have the url, it's ok to assume the host is running
	// This reduces the number of calls to the drivers
	if err == nil {
		if url != "" {
			currentState = state.Running
		} else {
			currentState, err = h.Driver.GetState()
		}
	} else {
		currentState, _ = h.Driver.GetState()
	}

	if err == nil && url != "" {
		// PERFORMANCE: Reuse the url instead of asking the host again.
		// This reduces the number of calls to the drivers
		dockerHost := &mcndockerclient.RemoteDocker{
			HostURL:    url,
			AuthOption: h.AuthOptions(),
		}
		dockerVersion, err = mcndockerclient.DockerVersion(dockerHost)

		if err != nil {
			dockerVersion = "Unknown"
		} else {
			dockerVersion = fmt.Sprintf("v%s", dockerVersion)
		}
	}

	if err != nil {
		hostError = err.Error()
	}
	if hostError == drivers.ErrHostIsNotRunning.Error() {
		hostError = ""
	}

	var swarmOptions *swarm.Options
	var engineOptions *engine.Options
	if h.HostOptions != nil {
		swarmOptions = h.HostOptions.SwarmOptions
		engineOptions = h.HostOptions.EngineOptions
	}

	isMaster := false
	swarmHost := ""
	if swarmOptions != nil {
		isMaster = swarmOptions.Master
		swarmHost = swarmOptions.Host
	}

	activeHost := isActive(currentState, url)
	activeSwarm := isSwarmActive(currentState, url, isMaster, swarmHost)
	active := "-"
	if activeHost {
		active = "*"
	}
	if activeSwarm {
		active = "* (swarm)"
	}

	stateQueryChan <- HostListItem{
		Name:          h.Name,
		Active:        active,
		ActiveHost:    activeHost,
		ActiveSwarm:   activeSwarm,
		DriverName:    h.Driver.DriverName(),
		State:         currentState,
		URL:           url,
		SwarmOptions:  swarmOptions,
		EngineOptions: engineOptions,
		DockerVersion: dockerVersion,
		Error:         hostError,
		ResponseTime:  time.Now().Round(time.Millisecond).Sub(requestBeginning.Round(time.Millisecond)),
	}
}

func getHostState(h *host.Host, hostListItemsChan chan<- HostListItem, timeout time.Duration) {
	// This channel is used to communicate the properties we are querying
	// about the host in the case of a successful read.
	stateQueryChan := make(chan HostListItem)

	go attemptGetHostState(h, stateQueryChan)

	select {
	// If we get back useful information, great.  Forward it straight to
	// the original parent channel.
	case hli := <-stateQueryChan:
		hostListItemsChan <- hli

	// Otherwise, give up after a predetermined duration.
	case <-time.After(timeout):
		hostListItemsChan <- HostListItem{
			Name:         h.Name,
			DriverName:   h.Driver.DriverName(),
			State:        state.Timeout,
			ResponseTime: timeout,
		}
	}
}

func getHostListItems(hostList []*host.Host, hostsInError map[string]error, timeout time.Duration) []HostListItem {
	log.Debugf("timeout set to %s", timeout)

	hostListItems := []HostListItem{}
	hostListItemsChan := make(chan HostListItem)

	for _, h := range hostList {
		go getHostState(h, hostListItemsChan, timeout)
	}

	for range hostList {
		hostListItems = append(hostListItems, <-hostListItemsChan)
	}

	close(hostListItemsChan)

	for name, err := range hostsInError {
		hostListItems = append(hostListItems, newHostListItemInError(name, err))
	}

	sortHostListItemsByName(hostListItems)
	return hostListItems
}

func newHostListItemInError(name string, err error) HostListItem {
	return HostListItem{
		Name:       name,
		DriverName: "not found",
		State:      state.Error,
		Error:      strings.Replace(err.Error(), "\n", " ", -1),
	}
}

func sortHostListItemsByName(items []HostListItem) {
	m := make(map[string]HostListItem, len(items))
	s := make([]string, len(items))
	for i, v := range items {
		name := strings.ToLower(v.Name)
		m[name] = v
		s[i] = name
	}
	sort.Sort(naturalsort.NaturalSort(s))
	for i, v := range s {
		items[i] = m[v]
	}
}

func isActive(currentState state.State, hostURL string) bool {
	return currentState == state.Running && hostURL == os.Getenv("DOCKER_HOST")
}

func isSwarmActive(currentState state.State, hostURL string, isMaster bool, swarmHost string) bool {
	return isMaster && currentState == state.Running && toSwarmURL(hostURL, swarmHost) == os.Getenv("DOCKER_HOST")
}

func urlPort(urlWithPort string) string {
	parts := strings.Split(urlWithPort, ":")
	return parts[len(parts)-1]
}

func toSwarmURL(hostURL string, swarmHost string) string {
	hostPort := urlPort(hostURL)
	swarmPort := urlPort(swarmHost)
	return strings.Replace(hostURL, ":"+hostPort, ":"+swarmPort, 1)
}
