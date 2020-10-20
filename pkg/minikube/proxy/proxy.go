/*
Copyright 2019 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package proxy

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/out"
)

// EnvVars are variables we plumb through to the underlying container runtime
var EnvVars = []string{"HTTP_PROXY", "HTTPS_PROXY", "NO_PROXY", "http_proxy", "https_proxy", "no_proxy"}

// isInBlock checks if ip is a CIDR block
func isInBlock(ip string, block string) (bool, error) {
	if ip == "" {
		return false, fmt.Errorf("ip is nil")
	}
	if block == "" {
		return false, fmt.Errorf("CIDR is nil")
	}

	if ip == block {
		return true, nil
	}

	i := net.ParseIP(ip)
	if i == nil {
		return false, fmt.Errorf("parsed IP is nil")
	}

	// check the block if it's CIDR
	if strings.Contains(block, "/") {
		_, b, err := net.ParseCIDR(block)
		if err != nil {
			return false, errors.Wrapf(err, "Error Parsing block %s", b)
		}

		if b.Contains(i) {
			return true, nil
		}
	}

	return false, errors.New("Error ip not in block")
}

// ExcludeIP takes ip or CIDR as string and excludes it from the http(s)_proxy
func ExcludeIP(ip string) error {
	if netIP := net.ParseIP(ip); netIP == nil {
		if _, _, err := net.ParseCIDR(ip); err != nil {
			return fmt.Errorf("ExcludeIP(%v) requires IP or CIDR as a parameter", ip)
		}
	}
	return updateEnv(ip, "NO_PROXY")
}

// IsIPExcluded checks if an IP is excluded from http(s)_proxy
func IsIPExcluded(ip string) bool {
	return checkEnv(ip, "NO_PROXY")
}

// updateEnv appends an ip to the environment variable
func updateEnv(ip string, env string) error {
	if ip == "" {
		return fmt.Errorf("IP %s is blank. ", ip)
	}
	if !isValidEnv(env) {
		return fmt.Errorf("%s is not a valid env var name for proxy settings", env)
	}
	if !checkEnv(ip, env) {
		v := os.Getenv(env)
		if v == "" {
			return os.Setenv(env, ip)
		}
		return os.Setenv(env, fmt.Sprintf("%s,%s", v, ip))
	}
	return nil
}

// checkEnv checks if ip in an environment variable
func checkEnv(ip string, env string) bool {
	v := os.Getenv(env)
	if v == "" {
		return false
	}
	//  Checking for IP explicitly, i.e., 192.168.39.224
	if strings.Contains(v, ip) {
		return true
	}
	// Checks if included in IP ranges, i.e., 192.168.39.13/24
	noProxyBlocks := strings.Split(v, ",")
	for _, b := range noProxyBlocks {
		yes, err := isInBlock(ip, b)
		if err != nil {
			klog.Warningf("fail to check proxy env: %v", err)
		}
		if yes {
			return true
		}
	}

	return false
}

// isValidEnv checks if the env for proxy settings
func isValidEnv(env string) bool {
	for _, e := range EnvVars {
		if e == env {
			return true
		}
	}
	return false
}

// UpdateTransport takes a k8s client *rest.config and returns a config without a proxy.
func UpdateTransport(cfg *rest.Config) *rest.Config {
	wt := cfg.WrapTransport // Config might already have a transport wrapper
	cfg.WrapTransport = func(rt http.RoundTripper) http.RoundTripper {
		if wt != nil {
			rt = wt(rt)
		}
		if ht, ok := rt.(*http.Transport); ok {
			ht.Proxy = nil
			rt = ht
		} else {
			klog.Errorf("Error while casting RoundTripper (of type %T) to *http.Transport : %v", rt, ok)
		}
		return rt
	}
	return cfg
}

// SetDockerEnv sets the proxy environment variables in the docker environment.
func SetDockerEnv() []string {
	for _, k := range EnvVars {
		if v := os.Getenv(k); v != "" {
			// convert https_proxy to HTTPS_PROXY for linux
			// TODO (@medyagh): if user has both http_proxy & HTTPS_PROXY set merge them.
			k = strings.ToUpper(k)
			if k == "HTTP_PROXY" || k == "HTTPS_PROXY" {
				isLocalProxy := func(url string) bool {
					return strings.HasPrefix(url, "localhost") || strings.HasPrefix(url, "127.0")
				}

				normalizedURL := v
				if !strings.Contains(v, "://") {
					normalizedURL = "http://" + v // by default, assumes the url is HTTP scheme
				}
				u, err := url.Parse(normalizedURL)
				if err != nil {
					out.WarningT("Error parsing {{.name}}={{.value}}, {{.err}}", out.V{"name": k, "value": v, "err": err})
					continue
				}

				if isLocalProxy(u.Host) {
					out.WarningT("Local proxy ignored: not passing {{.name}}={{.value}} to docker env.", out.V{"name": k, "value": v})
					continue
				}
			}
			config.DockerEnv = append(config.DockerEnv, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// remove duplicates
	seen := map[string]bool{}
	uniqueEnvs := []string{}
	for e := range config.DockerEnv {
		if !seen[config.DockerEnv[e]] {
			seen[config.DockerEnv[e]] = true
			uniqueEnvs = append(uniqueEnvs, config.DockerEnv[e])
		}
	}
	config.DockerEnv = uniqueEnvs

	return config.DockerEnv
}
