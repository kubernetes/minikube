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

type providerLinux struct {
	provider
}

/*
Create a new Provider which is used to retrieve Proxy configurations.
Params:
	configFile: Optional. Path to a configuration file which specifies proxies.
*/
func NewProvider(configFile string) Provider {
	c := new(providerLinux)
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
func (p *providerLinux) GetProxy(protocol string, targetUrlStr string) Proxy {
	return p.provider.get(protocol, ParseTargetURL(targetUrlStr, protocol))
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
func (p *providerLinux) GetHTTPProxy(targetUrl string) Proxy {
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
func (p *providerLinux) GetHTTPSProxy(targetUrl string) Proxy {
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
func (p *providerLinux) GetFTPProxy(targetUrl string) Proxy {
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
func (p *providerLinux) GetSOCKSProxy(targetUrl string) Proxy {
	return p.GetProxy(protocolSOCKS, targetUrl)
}
