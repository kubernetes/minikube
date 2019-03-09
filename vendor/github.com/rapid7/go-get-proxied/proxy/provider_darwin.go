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
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type providerDarwin struct {
	provider
}

/*
Create a new Provider which is used to retrieve Proxy configurations.
Params:
	configFile: Optional. Path to a configuration file which specifies proxies.
*/
func NewProvider(configFile string) Provider {
	c := new(providerDarwin)
	c.init(configFile)
	return c
}

/*
Returns the Proxy configuration for the given proxy protocol and targetUrl.
If none is found, or an error occurs, nil is returned.
This function searches the following locations in the following order:
	* Configuration file: proxy.config
	* Environment: HTTPS_PROXY, https_proxy, ...
Params:
	protocol: The protocol of traffic the proxy is to be used for. (i.e. http, https, ftp, socks)
	targetUrl: The URL the proxy is to be used for. (i.e. https://test.endpoint.rapid7.com)
Returns:
	Proxy: A proxy was found
	nil: A proxy was not found, or an error occurred
*/
func (p *providerDarwin) GetProxy(protocol string, targetUrlStr string) Proxy {
	targetUrl := ParseTargetURL(targetUrlStr, protocol)
	if proxy := p.provider.get(protocol, targetUrl); proxy != nil {
		return proxy
	}
	return p.readDarwinNetworkSettingProxy(protocol, targetUrl)
}

/*
Returns the Proxy configuration for HTTP traffic and the given targetUrl.
If none is found, or an error occurs, nil is returned.
Params:
	targetUrl: The URL the proxy is to be used for. (i.e. http://test.endpoint.rapid7.com)
Returns:
	Proxy: A proxy was found.
	nil: A proxy was not found, or an error occurred.
*/
func (p *providerDarwin) GetHTTPProxy(targetUrl string) Proxy {
	return p.GetProxy(protocolHTTP, targetUrl)
}

/*
Returns the Proxy configuration for HTTPS traffic and the given targetUrl.
If none is found, or an error occurs, nil is returned.
Params:
	targetUrl: The URL the proxy is to be used for. (i.e. https://test.endpoint.rapid7.com)
Returns:
	Proxy: A proxy was found.
	nil: A proxy was not found, or an error occurred.
*/
func (p *providerDarwin) GetHTTPSProxy(targetUrl string) Proxy {
	return p.GetProxy(protocolHTTPS, targetUrl)
}

/*
Returns the Proxy configuration for FTP traffic and the given targetUrl.
If none is found, or an error occurs, nil is returned.
Params:
	targetUrl: The URL the proxy is to be used for. (i.e. ftp://test.endpoint.rapid7.com)
Returns:
	Proxy: A proxy was found.
	nil: A proxy was not found, or an error occurred.
*/
func (p *providerDarwin) GetFTPProxy(targetUrl string) Proxy {
	return p.GetProxy(protocolFTP, targetUrl)
}

/*
Returns the Proxy configuration for generic TCP/UDP traffic and the given targetUrl.
If none is found, or an error occurs, nil is returned.
Params:
	targetUrl: The URL the proxy is to be used for. (i.e. ftp://test.endpoint.rapid7.com)
Returns:
	Proxy: A proxy was found.
	nil: A proxy was not found, or an error occurred.
*/
func (p *providerDarwin) GetSOCKSProxy(targetUrl string) Proxy {
	return p.GetProxy(protocolSOCKS, targetUrl)
}

const (
	scUtilBinary          = "scutil"
	scUtilBinaryArgument  = "--proxy"
	scUtilProxyEnabled    = "Enable:1"
	scUtilProxyDisabled   = "Enable:0"
	scUtilPortPrefix      = "Port:"
	scUtilProxyPrefix     = "Proxy:"
	scUtilExceptionsList  = "ExceptionsList"
	exceptionsListPattern = "ExceptionsList.*:.*{(.|\n)*.}"
	srcScUtil             = "State:/Network/Global/Proxies"
)

/*
Returns the Network Setting Proxy found.
If none is found, or an error occurs, nil is returned.
Params:
	protocol: The proxy's protocol (i.e. https)
	targetUrl: The URL the proxy is to be used for. (i.e. https://test.endpoint.myorganization.com)
Returns:
	Proxy: A proxy was found
	nil: A proxy was not found, or an error occurred
*/
func (p *providerDarwin) readDarwinNetworkSettingProxy(protocol string, targetUrl *url.URL) Proxy {
	proxy, err := p.parseScutildata(protocol, targetUrl, scUtilBinary, scUtilBinaryArgument)
	if err != nil {
		if isNotFound(err) {
			log.Printf("[proxy.Provider.readDarwinNetworkSettingProxy]: %s proxy is not enabled.\n", protocol)
		} else if isTimedOut(err) {
			log.Printf("[proxy.Provider.readDarwinNetworkSettingProxy]: Operation timed out. \n")
		} else {
			log.Printf("[proxy.Provider.readDarwinNetworkSettingProxy]: Failed to parse Scutil data, %s\n", err)
		}
	}
	return proxy
}

/*
Returns the Proxy found by parsing the Scutil output.
If none is found, or an error occurs, nil is returned.
Params:
	protocol: The proxy's protocol (i.e. https)
	targetUrl: The URL the proxy is to be used for. (i.e. https://test.endpoint.myorganization.com)
	name: The name of the program (scutil)
	arg: The list of the arguments (--proxy)
Returns:
	Proxy: A proxy was found, nil if no proxy found or an error occurred
	error: the error that has occurred, nil if there is no error
*/
func (p *providerDarwin) parseScutildata(protocol string, targetUrl *url.URL, name string, arg ...string) (Proxy, error) {
	lookupProtocol := strings.ToUpper(protocol) // to cover search for http, HTTP, https, HTTPS

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1) // Die after one second
	defer cancel()

	cmd := p.proc(ctx, name, arg...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, new(timeoutError)
	}
	scutilData := out.String()
	scanner := bufio.NewScanner(strings.NewReader(scutilData))
	/* init values */
	var enable bool
	var port string
	var host string
	var bypassProxyEnable bool
	regexEnable, err := regexp.Compile(lookupProtocol + scUtilProxyEnabled)
	if err != nil {
		return nil, err
	}
	regexDisable, err := regexp.Compile(lookupProtocol + scUtilProxyDisabled)
	if err != nil {
		return nil, err
	}
	regexPort, err := regexp.Compile(lookupProtocol + scUtilPortPrefix)
	if err != nil {
		return nil, err
	}
	regexProxy, err := regexp.Compile(lookupProtocol + scUtilProxyPrefix)
	if err != nil {
		return nil, err
	}
	regexBypassProxy, err := regexp.Compile(scUtilExceptionsList)
	if err != nil {
		return nil, err
	}

	for scanner.Scan() {
		str := strings.Replace(scanner.Text(), " ", "", -1) // removing spaces
		if !bypassProxyEnable {                             // don't search if already found
			bypassProxyListFound := regexBypassProxy.FindStringIndex(str)
			if bypassProxyListFound != nil {
				bypassProxyEnable = true
			}
		}
		if !enable { // don't search if already found
			// if proxy is disabled, stop the search
			protocolDisableFound := regexDisable.FindStringIndex(str)
			if protocolDisableFound != nil {
				break
			}
			protocolEnableFound := regexEnable.FindStringIndex(str)
			if protocolEnableFound != nil {
				enable = true
			}
		}
		if port == "" { // don't search if already found
			portFoundLoc := regexPort.FindStringIndex(str)
			if portFoundLoc != nil {
				port = str[portFoundLoc[1]:]
			}
		}
		if host == "" { // don't search if already found
			proxyFoundLoc := regexProxy.FindStringIndex(str)
			if proxyFoundLoc != nil {
				host = str[proxyFoundLoc[1]:]
			}
		}
	}
	if !enable {
		return nil, new(notFoundError)
	}

	proxyUrlStr := host + ":" + port
	proxyUrl, err := ParseURL(proxyUrlStr, "")
	if err != nil {
		return nil, err
	}
	src := srcScUtil
	proxy, err := NewProxy(proxyUrl, src)
	if err != nil {
		return nil, err
	}
	// if no bypass info exists, return the proxy obtained
	if bypassProxyEnable == false {
		return proxy, nil
	}
	proxyBypass, err := p.readScutilBypassProxy(scutilData)
	if err != nil {
		return nil, err
	}
	if proxyBypass != "" {
		bypass := p.isProxyBypass(targetUrl, proxyBypass, ",")
		log.Printf("[proxy.Provider.parseProxyInfo]: ProxyBypass=\"%s\", targetUrl=%s, bypass=%t", proxyBypass, targetUrl, bypass)
		if bypass {
			return nil, nil
		}
	}
	return proxy, nil
}

/*
Returns the Bypass Proxy Settings found by parsing the Scutil output.
If none is found, or an error occurs, empty string ("") is returned.
Params:
	scutilData: The scutil data content
Returns:
	Bypass Proxy list: List of bypass proxies that are found, "" when none found or an error occurred
	error: the error that has occurred, nil if there is no error
*/
func (p *providerDarwin) readScutilBypassProxy(scutilData string) (string, error) {
	regexBypassProxy, err := regexp.Compile(exceptionsListPattern)
	if err != nil {
		return "", err
	}
	exceptionsListFound := regexBypassProxy.FindStringIndex(scutilData)
	if exceptionsListFound == nil {
		return "", nil
	}
	exceptionsList := scutilData[exceptionsListFound[0]:exceptionsListFound[1]]
	scanner := bufio.NewScanner(strings.NewReader(exceptionsList))
	firstLine := -1
	var bypassProxies string
	for scanner.Scan() {
		if firstLine == -1 { // skip the first line
			firstLine = 0
			continue
		}
		str := strings.Replace(scanner.Text(), " ", "", -1) // removing spaces
		s := fmt.Sprintf("%d:", firstLine)
		regexProxy, err := regexp.Compile(s)
		if err != nil {
			return "", err
		}
		firstLine += 1
		proxyFoundLoc := regexProxy.FindStringIndex(str)
		if proxyFoundLoc != nil {
			bypassProxyUrlStr := str[proxyFoundLoc[1]:]
			bypassProxies = bypassProxies + bypassProxyUrlStr + ","
		}
	}
	return bypassProxies, nil
}
