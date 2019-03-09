// Copyright 2018, Rapid7, Inc.
// License: BSD-3-clause
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
// * Redistributions of source code must retain the above copyright notice,
// this list of conditions and the following disclaimer.
// * Redistributions in binary form must reproduce the above copyright
// notice, this list of conditions and the following disclaimer in the
// documentation and/or other materials provided with the distribution.
// * Neither the name of the copyright holder nor the names of its contributors
// may be used to endorse or promote products derived from this software
// without specific prior written permission.
package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Provider interface {
	/*
		Returns the Proxy configuration for the given traffic protocol and targetUrl.
		If none is found, or an error occurs, nil is returned.
		Params:
			protocol: The protocol of traffic the proxy is to be used for. (i.e. http, https, ftp, socks)
			targetUrl: The URL the proxy is to be used for. (i.e. https://test.endpoint.rapid7.com)
		Returns:
			Proxy: A proxy was found.
			nil: A proxy was not found, or an error occurred.
	*/
	GetProxy(protocol string, targetUrl string) Proxy

	/*
		Returns the Proxy configuration for HTTP traffic and the given targetUrl.
		If none is found, or an error occurs, nil is returned.
		Params:
			targetUrl: The URL the proxy is to be used for. (i.e. http://test.endpoint.rapid7.com)
		Returns:
			Proxy: A proxy was found.
			nil: A proxy was not found, or an error occurred.
	*/
	GetHTTPProxy(targetUrl string) Proxy

	/*
		Returns the Proxy configuration for HTTPS traffic and the given targetUrl.
		If none is found, or an error occurs, nil is returned.
		Params:
			targetUrl: The URL the proxy is to be used for. (i.e. https://test.endpoint.rapid7.com)
		Returns:
			Proxy: A proxy was found.
			nil: A proxy was not found, or an error occurred.
	*/
	GetHTTPSProxy(targetUrl string) Proxy

	/*
		Returns the Proxy configuration for FTP traffic and the given targetUrl.
		If none is found, or an error occurs, nil is returned.
		Params:
			targetUrl: The URL the proxy is to be used for. (i.e. ftp://test.endpoint.rapid7.com)
		Returns:
			Proxy: A proxy was found.
			nil: A proxy was not found, or an error occurred.
	*/
	GetFTPProxy(targetUrl string) Proxy

	/*
		Returns the Proxy configuration for generic TCP/UDP traffic and the given targetUrl.
		If none is found, or an error occurs, nil is returned.
		Params:
			targetUrl: The URL the proxy is to be used for. (i.e. ftp://test.endpoint.rapid7.com)
		Returns:
			Proxy: A proxy was found.
			nil: A proxy was not found, or an error occurred.
	*/
	GetSOCKSProxy(targetUrl string) Proxy
	/*
		Set the timeouts used by this provider making a call which requires external resources (i.e. WPAD/PAC).
		Should any of these timeouts be exceeded, that particular call will be cancelled.
		To this end, this timeout does not represent the complete timeout for any call to this provider,
		but rather are applied to individual implementations uniquely.
		Additionally, this timeout is not guaranteed to be respected by the implementation, and may vary.
		Params:
			resolve: Time in milliseconds to use for name resolution. Provider default is 5000.
			connect: Time in milliseconds to use for server connection requests. Provider default is 5000.
					 TCP/IP can time out while setting up the socket during the
					 three leg SYN/ACK exchange, regardless of the value of this parameter.
			send: Time in milliseconds to use for sending requests. Provider default is 20000.
			receive: Time in milliseconds to receive a response to a request. Provider default is 20000.
	*/
	SetTimeouts(resolve int, connect int, send int, receive int)
}

const (
	protocolHTTP          = "http"
	protocolHTTPS         = "https"
	protocolFTP           = "ftp"
	protocolSOCKS         = "socks"
	proxyKeyFormat        = "%s_PROXY"
	noProxyKeyUpper       = "NO_PROXY"
	noProxyKeyLower       = "no_proxy"
	prefixSOCKS           = protocolSOCKS
	prefixAll             = "all"
	targetUrlWildcard     = "*"
	domainDelimiter       = "."
	bypassLocal           = "<local>"
	srcConfigurationFile  = "ConfigurationFile"
	srcEnvironmentFmt     = "Environment[%s]"
	defaultResolveTimeout = 5000
	defaultConnectTimeout = 5000
	defaultSendTimeout    = 20000
	defaultReceiveTimeout = 20000
)

type getEnvAdapter func(string) string

type commandAdapter func(context.Context, string, ...string) *exec.Cmd

type provider struct {
	configFile     string
	getEnv         getEnvAdapter
	proc           commandAdapter
	resolveTimeout int
	connectTimeout int
	sendTimeout    int
	receiveTimeout int
}

func (p *provider) init(configFile string) {
	p.configFile = configFile
	p.getEnv = os.Getenv
	p.proc = exec.CommandContext
	p.resolveTimeout = defaultResolveTimeout
	p.connectTimeout = defaultConnectTimeout
	p.sendTimeout = defaultSendTimeout
	p.receiveTimeout = defaultReceiveTimeout
}

/*
Set the timeouts used by this provider making a call which requires external resources (i.e. WPAD/PAC).
Should any of these timeouts be exceeded, that particular call will be cancelled.
To this end, this timeout does not represent the complete timeout for any call to this provider,
but rather are applied to individual implementations uniquely.
Additionally, this timeout is not guaranteed to be respected by the implementation, and may vary.
Params:
	resolve: Time in milliseconds to use for name resolution. Provider default is 5000.
	connect: Time in milliseconds to use for server connection requests. Provider default is 5000.
			 TCP/IP can time out while setting up the socket during the
			 three leg SYN/ACK exchange, regardless of the value of this parameter.
	send: Time in milliseconds to use for sending requests. Provider default is 20000.
	receive: Time in milliseconds to receive a response to a request. Provider default is 20000.
*/
func (p *provider) SetTimeouts(resolve int, connect int, send int, receive int) {
	p.resolveTimeout = resolve
	p.connectTimeout = connect
	p.sendTimeout = send
	p.receiveTimeout = receive
}

/*
Returns the Proxy configuration for the given traffic protocol and targetUrl.
If none is found, or an error occurs, nil is returned.
Params:
	protocol: The protocol of traffic the proxy is to be used for. (i.e. http, https, ftp, socks)
	targetUrl: The URL the proxy is to be used for. (i.e. https://test.endpoint.rapid7.com)
Returns:
	Proxy: A proxy was found.
	nil: A proxy was not found, or an error occurred.
*/
func (p *provider) get(protocol string, targetUrl *url.URL) Proxy {
	proxy := p.readConfigFileProxy(protocol)
	if proxy != nil {
		return proxy
	}
	return p.readSystemEnvProxy(protocol, targetUrl)
}

/*
Unmarshal the proxy.config file, and return the first proxy matched for the given protocol.
If no proxy is found, or an error occurs reading the proxy.config file, nil is returned.
Params:
	protocol: The protocol of traffic the proxy is to be used for. (i.e. http, https, ftp, socks)
Returns:
	Proxy: A proxy is found in proxy.config for the given protocol.
	nil: No proxy is found or an error occurs reading the proxy.config file.
*/
func (p *provider) readConfigFileProxy(protocol string) Proxy {
	proxyJson, err := p.unmarshalProxyConfigFile()
	if err != nil {
		log.Printf("[proxy.Provider.readConfigFileProxy]: %s\n", err)
		return nil
	}
	uStr, exists := proxyJson[protocol]
	if !exists {
		return nil
	}
	uUrl, uErr := ParseURL(uStr, "")
	var uProxy Proxy
	if uErr == nil {
		uProxy, uErr = NewProxy(uUrl, srcConfigurationFile)
	}
	if uErr != nil {
		log.Printf("[proxy.Provider.readConfigFileProxy]: invalid config file proxy, skipping \"%s\": \"%s\"\n", protocol, uStr)
		return nil
	}
	return uProxy
}

/*
Unmarshal the proxy.config file into a simple map[string]string structure.
Returns:
	map[string]string, nil: Unmarshal of proxy.config is successful.
	nil, error: Unmarshal of proxy.config is not successful.
*/
func (p *provider) unmarshalProxyConfigFile() (map[string]string, error) {
	m := map[string]string{}
	if p.configFile == "" {
		return m, nil
	}
	f := filepath.Join(p.configFile)
	stat, err := os.Stat(f)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		return nil, errors.New(fmt.Sprintf("proxy configuration file not present: %s", f))
	} else if stat.IsDir() {
		return nil, errors.New(fmt.Sprintf("proxy configuration file is a directory: %s", f))
	} else if stat.Size() <= 0 {
		return nil, errors.New(fmt.Sprintf("proxy configuration file empty: %s", f))
	} else if stat.Size() > 1048576 {
		return nil, errors.New(fmt.Sprintf("proxy configuration file too large: %s", f))
	}
	out, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to read proxy configuration file: %s: %s", f, err))
	}

	if err = json.Unmarshal(out, &m); err != nil {
		return nil, errors.New(fmt.Sprintf("failed to unmarshal proxy configuration file: %s: %s", f, err))
	}
	// Sanitize the protocols so we can be case insensitive
	for protocol, v := range m {
		delete(m, protocol)
		m[strings.ToLower(protocol)] = v
	}
	return m, nil
}

/*
Find the proxy configured by environment variables for the given traffic protocol and targetUrl.
If no proxy is found, or an error occurs, nil is returned.
Params:
	protocol: The protocol of traffic the proxy is to be used for. (i.e. http, https, ftp, socks)
	targetUrl: The URL the proxy is to be used for. (i.e. https://test.endpoint.rapid7.com)
Returns:
	proxy: A proxy is found through environment variables for the given traffic protocol.
	nil: No proxy is found or an error occurs reading the environment variables.
*/
func (p *provider) readSystemEnvProxy(prefix string, targetUrl *url.URL) Proxy {
	// SOCKS configuration is set as ALL_PROXY and all_proxy on Linux. Replace here for all OSs to keep consistent
	if strings.HasPrefix(prefix, prefixSOCKS) {
		prefix = prefixAll
	}
	keys := []string{
		strings.ToUpper(fmt.Sprintf(proxyKeyFormat, prefix)),
		strings.ToLower(fmt.Sprintf(proxyKeyFormat, prefix))}
	noProxyValues := map[string]string{
		noProxyKeyUpper: p.getEnv(noProxyKeyUpper),
		noProxyKeyLower: p.getEnv(noProxyKeyLower)}
K:
	for _, key := range keys {
		proxy, err := p.parseEnvProxy(key)
		if err != nil {
			if !isNotFound(err) {
				log.Printf("[proxy.Provider.readSystemEnvProxy]: failed to parse \"%s\" value: %s\n", key, err)
			}
			continue
		}
		bypass := false
		for noProxyKey, proxyBypass := range noProxyValues {
			if proxyBypass == "" {
				continue
			}
			bypass = p.isProxyBypass(targetUrl, proxyBypass, ",")
			log.Printf("[proxy.Provider.readSystemEnvProxy]: \"%s\"=\"%s\", targetUrl=%s, bypass=%t", noProxyKey, proxyBypass, targetUrl, bypass)
			if bypass {
				continue K
			}
		}
		return proxy
	}
	return nil
}

/*
Return true if the given targetUrl should bypass a proxy for the given proxyBypass value and sep.
For example:
	("test.endpoint.rapid7.com", "rapid7.com", ",") -> true
	("test.endpoint.rapid7.com", ".rapid7.com", ",") -> true
	("test.endpoint.rapid7.com", "*.rapid7.com", ",") -> true
	("test.endpoint.rapid7.com", "*", ",") -> true
	("test.endpoint.rapid7.com", "test.endpoint.rapid7.com", ",") -> true
	("test.endpoint.rapid7.com", "someHost,anotherHost", ",") -> false
	("test.endpoint.rapid7.com", "", ",") -> false
Params:
	targetUrl: The URL the proxy is to be used for. (i.e. https://test.endpoint.rapid7.com)
	proxyBypass: The proxy bypass value.
	sep: The separator to use with the proxy bypass value.
Returns:
	true: The proxy should be bypassed for the given targetUrl
	false: Otherwise
*/
func (p *provider) isProxyBypass(targetUrl *url.URL, proxyBypass string, sep string) bool {
	targetHost, _, _ := SplitHostPort(targetUrl)
	for _, s := range strings.Split(proxyBypass, sep) {
		s = strings.TrimSpace(s)
		if s == "" {
			// No value
			continue
		} else if s == bypassLocal {
			// Windows uses <local> for local domains
			if IsLoopbackHost(targetHost) {
				return true
			}
		}
		// Exact match
		if m, err := filepath.Match(s, targetHost); err != nil {
			return false
		} else if m {
			return true
		}
		// Prefix "* for wildcard matches (rapid7.com -> *.rapid7.com)
		if strings.Index(s, targetUrlWildcard) != 0 {
			// (rapid7.com -> .rapid7.com)
			if strings.Index(s, domainDelimiter) != 0 {
				s = domainDelimiter + s
			}
			s = targetUrlWildcard + s
		}
		if m, err := filepath.Match(s, targetHost); err != nil {
			return false
		} else if m {
			return true
		}
	}
	return false
}

/*
Read the given environment variable by key, returning the proxy if it is valid.
Returns nil if no proxy is configured, or an error occurs.
Params:
	key: The environment variable key
Returns:
	proxy: A proxy was found for the given environment variable key and is valid.
	false: Otherwise
*/
func (p *provider) parseEnvProxy(key string) (Proxy, error) {
	proxyUrl, err := p.parseEnvURL(key)
	if err != nil {
		return nil, err
	}
	proxy, err := NewProxy(proxyUrl, fmt.Sprintf(srcEnvironmentFmt, key))
	if err != nil {
		return nil, err
	}
	return proxy, nil
}

/*
Parse the optionally valid URL string from the given environment variable key's value.
Params:
	key: The name of the environment variable
Returns:
	url.URL: If the environment variable was populated, the parsed value. Otherwise nil.
	error: If the environment variable was populated, but we failed to parse it.
*/
func (p *provider) parseEnvURL(key string) (*url.URL, error) {
	value := strings.TrimSpace(p.getEnv(key))
	if value != "" {
		return ParseURL(value, "")
	}
	return nil, new(notFoundError)
}

type notFoundError struct{}

func (e notFoundError) Error() string {
	return "No proxy found"
}

type timeoutError struct{}

func (e timeoutError) Error() string {
	return "Timed out"
}

/*
Returns:
	true: The error represents a Proxy not being found
	false: Otherwise
s*/
func isNotFound(e error) bool {
	switch e.(type) {
	case *notFoundError:
		return true
	case notFoundError:
		return true
	default:
		return false
	}
}

/*
Returns:
	true: The error represents a Time out
	false: Otherwise
s*/
func isTimedOut(e error) bool {
	switch e.(type) {
	case *timeoutError:
		return true
	case timeoutError:
		return true
	default:
		return false
	}
}
