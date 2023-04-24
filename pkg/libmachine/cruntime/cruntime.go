package cruntime

import "k8s.io/minikube/pkg/libmachine/auth"

const (
	DefaultPort = 2376
)

type Options struct {
	CustomPort       int
	ArbitraryFlags   []string
	DNS              []string `json:"Dns"`
	GraphDir         string
	Env              []string
	Ipv6             bool
	InsecureRegistry []string
	Labels           []string
	LogLevel         string
	StorageDriver    string
	SelinuxEnabled   bool
	TLSVerify        bool `json:"TlsVerify"`
	RegistryMirror   []string
	InstallURL       string
	AuthOptions      auth.Options
	CRuntimeOptsDir  string
}

type CRuntimeEngine interface {
	// GenConfigFile generates a container runtime configuration file
	// based on container runtime options;
	// it returns the config file back as a string.
	GenConfigFile(opts Options) (string, error)
}
