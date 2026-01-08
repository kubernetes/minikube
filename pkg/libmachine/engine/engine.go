package engine

const (
	DefaultPort = 2376
)

type Options struct {
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
}
