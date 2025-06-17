package environment
import "os"
// GetNoProxyVar gets the no_proxy environment variable, lowercase preferred.
func GetNoProxyVar() (string, string) {
	noProxyVar := "no_proxy"
	noProxyValue := os.Getenv("no_proxy")
	if noProxyValue == "" {
		noProxyVar = "NO_PROXY"
		noProxyValue = os.Getenv("NO_PROXY")
	}
	return noProxyVar, noProxyValue
}