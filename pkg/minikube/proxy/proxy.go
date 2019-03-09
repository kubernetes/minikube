// Package proxy contains helpers for interacting with HTTP proxies
package proxy

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/golang/glog"
	ggp "github.com/rapid7/go-get-proxied/proxy"
	"k8s.io/minikube/pkg/minikube/config"
)

const bypass = "bypass"
const http = "http"
const https = "https"

// envVars is a map of environment proxy to protocol
var envVars = map[string][]string{
	"http_proxy":  []string{http},
	"https_proxy": []string{https},
	"HTTP_PROXY":  []string{http},
	"HTTPS_PROXY": []string{https},
	"ALL_PROXY":   []string{http, https},
	"NO_PROXY":    []string{bypass},
}

// envVarPriority describes our environment priority order: first wins
var envVarPriority = []string{"HTTP_PROXY", "HTTPS_PROXY", "http_proxy", "https_proxy", "ALL_PROXY", "NO_PROXY"}

// Detected returns whether or not a proxy configuration was detected
func Detected() bool {
	e := Environment()
	if e[http] != "" || e[https] != "" {
		return true
	}
	return false
}

// proxyString converts a ggp.Proxy object into a well-formed string
func proxyString(p ggp.Proxy) string {
	url := p.URL()
	if url.Scheme == "" {
		url.Scheme = "http"
	}
	return url.String()
}

// Environment returns values discovered from the environment
func Environment() map[string]string {
	env := map[string]string{}

	for _, ev := range envVarPriority {
		v := os.Getenv(ev)
		if v == "" {
			continue
		}
		for _, proto := range envVars[ev] {
			if env[proto] == "" {
				env[proto] = v
			}
		}
	}

	glog.Infof("Disabling log output because of ggp ... don't hate me.")
	log.SetOutput(ioutil.Discard)

	// If env is unset, include values detected from the operating system
	if env[https] == "" {
		p := ggp.NewProvider("").GetHTTPSProxy("https://k8s.gcr.io/")
		if p != nil {
			env[https] = proxyString(p)
		}
	}
	if env[http] == "" {
		p := ggp.NewProvider("").GetHTTPProxy("http://storage.googleapis.com/")
		if p != nil {
			env[http] = proxyString(p)
		}
	}
	return env
}

// Recommended returns recommended values for use within a VM
func Recommended(k8s config.KubernetesConfig) map[string]string {
	env := map[string]string{}

	for proto, value := range Environment() {
		if proto != bypass {
			env[fmt.Sprintf("%s_PROXY", strings.ToUpper(proto))] = value
		}
	}
	env["NO_PROXY"] = NoProxy(k8s)
	return env
}

// NoProxy returns a recommended value for NO_PROXY
func NoProxy(k8s config.KubernetesConfig) string {
	current := Environment()[bypass]
	var vals []string
	if current != "" {
		vals = strings.Split(current, ",")
	}
	seen := map[string]bool{}
	for _, v := range vals {
		seen[v] = true
	}

	for _, optVal := range []string{k8s.NodeIP, k8s.ServiceCIDR, "127.0.0.1"} {
		if optVal != "" && !seen[optVal] {
			vals = append(vals, optVal)
		}
	}

	sort.Strings(vals)
	return strings.Join(vals, ",")
}

// EnvContent returns content appropriate for /etc/environment
func EnvContent(k8s config.KubernetesConfig) []byte {
	var b bytes.Buffer
	for k, v := range Recommended(k8s) {
		b.WriteString(fmt.Sprintf("%s=%s\n", k, v))
	}
	return b.Bytes()
}
