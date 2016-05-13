package newrelic_platform_go

import (
	"log"
	"os"
)

type Agent struct {
	Host    string `json:"host"`
	Version string `json:"version"`
	Pid     int    `json:"pid"`
}

func NewAgent(Version string) *Agent {
	agent := &Agent{
		Version: Version,
	}
	return agent
}

func (agent *Agent) CollectEnvironmentInfo() {
	var err error
	agent.Pid = os.Getpid()
	if agent.Host, err = os.Hostname(); err != nil {
		log.Fatalf("Can not get hostname: %#v \n", err)
	}
}
