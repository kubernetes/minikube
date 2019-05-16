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
	"os"
	"strings"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
)

// EnvVars are variables we plumb through to the underlying container runtime
var EnvVars = []string{"HTTP_PROXY", "HTTPS_PROXY", "NO_PROXY"}

// isInBlock checks if ip is a CIDR block
func isInBlock(ip string, block string) (bool, error) {
	if ip == "" {
		return false, fmt.Errorf("ip is nil")
	}
	if block == "" {
		return false, fmt.Errorf("CIDR is nil")
	}

	i := net.ParseIP(ip)
	if i == nil {
		return false, fmt.Errorf("parsed IP is nil")
	}
	_, b, err := net.ParseCIDR(block)
	if err != nil {
		return false, errors.Wrapf(err, "Error Parsing block %s", b)
	}

	if b.Contains(i) {
		return true, nil
	}
	return false, errors.Wrapf(err, "Error ip not in block")
}

// ExcludeIP will exclude the ip from the http(s)_proxy
func ExcludeIP(ip string) error {
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
		if yes, _ := isInBlock(ip, b); yes {
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
			glog.Errorf("Error while casting RoundTripper (of type %T) to *http.Transport : %v", rt, ok)
		}
		return rt
	}
	return cfg
}
