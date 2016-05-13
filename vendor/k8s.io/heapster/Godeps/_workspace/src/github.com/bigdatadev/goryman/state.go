package goryman

import ()

// State is a wrapper for Riemann states
type State struct {
	Ttl         float32
	Time        int64
	Tags        []string
	Host        string // Defaults to os.Hostname()
	State       string
	Service     string
	Once        bool
	Metric      interface{} // Could be Int, Float32, Float64
	Description string
	Attributes  map[string]string
}
