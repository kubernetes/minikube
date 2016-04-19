package swarm

const (
	DiscoveryServiceEndpoint = "https://discovery-stage.hub.docker.com/v1"
)

type Options struct {
	IsSwarm        bool
	Address        string
	Discovery      string
	Master         bool
	Host           string
	Image          string
	Strategy       string
	Heartbeat      int
	Overcommit     float64
	ArbitraryFlags []string
	Env            []string
	IsExperimental bool
}
