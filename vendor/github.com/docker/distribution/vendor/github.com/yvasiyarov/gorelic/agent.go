package gorelic

import (
	"errors"
	"fmt"
	metrics "github.com/yvasiyarov/go-metrics"
	"github.com/yvasiyarov/newrelic_platform_go"
	"log"
	"net/http"
)

const (
	// DefaultNewRelicPollInterval - how often we will report metrics to NewRelic.
	// Recommended values is 60 seconds
	DefaultNewRelicPollInterval = 60

	// DefaultGcPollIntervalInSeconds - how often we will get garbage collector run statistic
	// Default value is - every 10 seconds
	// During GC stat pooling - mheap will be locked, so be carefull changing this value
	DefaultGcPollIntervalInSeconds = 10

	// DefaultMemoryAllocatorPollIntervalInSeconds - how often we will get memory allocator statistic.
	// Default value is - every 60 seconds
	// During this process stoptheword() is called, so be carefull changing this value
	DefaultMemoryAllocatorPollIntervalInSeconds = 60

	//DefaultAgentGuid is plugin ID in NewRelic.
	//You should not change it unless you want to create your own plugin.
	DefaultAgentGuid = "com.github.yvasiyarov.GoRelic"

	//CurrentAgentVersion is plugin version
	CurrentAgentVersion = "0.0.6"

	//DefaultAgentName in NewRelic GUI. You can change it.
	DefaultAgentName = "Go daemon"
)

//Agent - is NewRelic agent implementation.
//Agent start separate go routine which will report data to NewRelic
type Agent struct {
	NewrelicName                string
	NewrelicLicense             string
	NewrelicPollInterval        int
	Verbose                     bool
	CollectGcStat               bool
	CollectMemoryStat           bool
	CollectHTTPStat             bool
	GCPollInterval              int
	MemoryAllocatorPollInterval int
	AgentGUID                   string
	AgentVersion                string
	plugin                      *newrelic_platform_go.NewrelicPlugin
	HTTPTimer                   metrics.Timer
}

//NewAgent build new Agent objects.
func NewAgent() *Agent {
	agent := &Agent{
		NewrelicName:                DefaultAgentName,
		NewrelicPollInterval:        DefaultNewRelicPollInterval,
		Verbose:                     false,
		CollectGcStat:               true,
		CollectMemoryStat:           true,
		GCPollInterval:              DefaultGcPollIntervalInSeconds,
		MemoryAllocatorPollInterval: DefaultMemoryAllocatorPollIntervalInSeconds,
		AgentGUID:                   DefaultAgentGuid,
		AgentVersion:                CurrentAgentVersion,
	}
	return agent
}

//WrapHTTPHandlerFunc  instrument HTTP handler functions to collect HTTP metrics
func (agent *Agent) WrapHTTPHandlerFunc(h tHTTPHandlerFunc) tHTTPHandlerFunc {
	agent.initTimer()
	return func(w http.ResponseWriter, req *http.Request) {
		proxy := newHTTPHandlerFunc(h)
		proxy.timer = agent.HTTPTimer
		proxy.ServeHTTP(w, req)
	}
}

//WrapHTTPHandler  instrument HTTP handler object to collect HTTP metrics
func (agent *Agent) WrapHTTPHandler(h http.Handler) http.Handler {
	agent.initTimer()

	proxy := newHTTPHandler(h)
	proxy.timer = agent.HTTPTimer
	return proxy
}

//Run initialize Agent instance and start harvest go routine
func (agent *Agent) Run() error {
	if agent.NewrelicLicense == "" {
		return errors.New("please, pass a valid newrelic license key")
	}

	agent.plugin = newrelic_platform_go.NewNewrelicPlugin(agent.AgentVersion, agent.NewrelicLicense, agent.NewrelicPollInterval)
	component := newrelic_platform_go.NewPluginComponent(agent.NewrelicName, agent.AgentGUID)
	agent.plugin.AddComponent(component)

	addRuntimeMericsToComponent(component)

	if agent.CollectGcStat {
		addGCMericsToComponent(component, agent.GCPollInterval)
		agent.debug(fmt.Sprintf("Init GC metrics collection. Poll interval %d seconds.", agent.GCPollInterval))
	}
	if agent.CollectMemoryStat {
		addMemoryMericsToComponent(component, agent.MemoryAllocatorPollInterval)
		agent.debug(fmt.Sprintf("Init memory allocator metrics collection. Poll interval %d seconds.", agent.MemoryAllocatorPollInterval))
	}

	if agent.CollectHTTPStat {
		agent.initTimer()
		addHTTPMericsToComponent(component, agent.HTTPTimer)
		agent.debug(fmt.Sprintf("Init HTTP metrics collection."))
	}

	agent.plugin.Verbose = agent.Verbose
	go agent.plugin.Run()
	return nil
}

//Initialize global metrics.Timer object, used to collect HTTP metrics
func (agent *Agent) initTimer() {
	if agent.HTTPTimer == nil {
		agent.HTTPTimer = metrics.NewTimer()
	}

	agent.CollectHTTPStat = true
}

//Print debug messages
func (agent *Agent) debug(msg string) {
	if agent.Verbose {
		log.Println(msg)
	}
}
