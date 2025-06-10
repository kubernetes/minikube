package environment
import "os"
// GetNoProxyVar 获取 no_proxy 环境变量，优先使用小写。
func GetNoProxyVar() (string, string) {
	noProxyVar := "no_proxy"
	noProxyValue := os.Getenv("no_proxy")
	if noProxyValue == "" {
		noProxyVar = "NO_PROXY"
		noProxyValue = os.Getenv("NO_PROXY")
	}
	return noProxyVar, noProxyValue
}