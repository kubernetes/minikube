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
	"github.com/rapid7/go-get-proxied/winhttp"
	"log"
	"net/url"
	"reflect"
	"strings"
)

/*
Create a new Provider which is used to retrieve Proxy configurations.
Params:
	configFile: Optional. Path to a configuration file which specifies proxies.
*/
func NewProvider(configFile string) Provider {
	c := new(providerWindows)
	c.init(configFile)
	return c
}

/*
Returns the Proxy configuration for the given proxy protocol and targetUrl.
If none is found, or an error occurs, nil is returned.
This function searches the following locations in the following order:
	* Configuration file: proxy.config
	* Environment: HTTPS_PROXY, https_proxy, ...
	* IE Proxy Config: AutoDetect
	* IE Proxy Config: AutoConfig URL
	* IE Proxy Config: Manual
	* WinHTTP Default
Params:
	protocol: The protocol of traffic the proxy is to be used for. (i.e. http, https, ftp, socks)
	targetUrl: The URL the proxy is to be used for. (i.e. https://test.endpoint.rapid7.com)
Returns:
	Proxy: A proxy was found
	nil: A proxy was not found, or an error occurred
*/
func (p *providerWindows) GetProxy(protocol string, targetUrlStr string) Proxy {
	targetUrl := ParseTargetURL(targetUrlStr, protocol)
	proxy := p.provider.get(protocol, targetUrl)
	if proxy != nil {
		return proxy
	}
	return p.readWinHttpProxy(protocol, targetUrl)
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
func (p *providerWindows) GetHTTPProxy(targetUrl string) Proxy {
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
func (p *providerWindows) GetHTTPSProxy(targetUrl string) Proxy {
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
func (p *providerWindows) GetFTPProxy(targetUrl string) Proxy {
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
func (p *providerWindows) GetSOCKSProxy(targetUrl string) Proxy {
	return p.GetProxy(protocolSOCKS, targetUrl)
}

const (
	userAgent        = "ir_agent"
	srcAutoDetect    = "WinHTTP:AutoDetect"
	srcAutoConfigUrl = "WinHTTP:AutoConfigUrl"
	srcNamedProxy    = "WinHTTP:NamedProxy"
	srcWinHttp       = "WinHTTP:WinHttpDefault"
)

type providerWindows struct {
	provider
}

//noinspection SpellCheckingInspection
func (p *providerWindows) readWinHttpProxy(protocol string, targetUrl *url.URL) Proxy {
	// Internet Options
	ieProxyConfig, err := p.getIeProxyConfigCurrentUser()
	if err != nil {
		log.Printf("[proxy.Provider.readWinHttpProxy] Failed to read IE proxy config: %s\n", err)
	} else {
		defer p.freeWinHttpResource(ieProxyConfig)
		if ieProxyConfig.FAutoDetect {
			proxy, err := p.getProxyAutoDetect(protocol, targetUrl)
			if err == nil {
				return proxy
			} else if !isNotFound(err) {
				log.Printf("[proxy.Provider.readWinHttpProxy] No proxy discovered via AutoDetect: %s\n", err)
			}
		}
		if autoConfigUrl := winhttp.LpwstrToString(ieProxyConfig.LpszAutoConfigUrl); autoConfigUrl != "" {
			proxy, err := p.getProxyAutoConfigUrl(protocol, targetUrl, autoConfigUrl)
			if err == nil {
				return proxy
			} else if !isNotFound(err) {
				log.Printf("[proxy.Provider.readWinHttpProxy] No proxy discovered via AutoConfigUrl, %s: %s\n", autoConfigUrl, err)
			}
		}
		proxy, err := p.parseProxyInfo(srcNamedProxy, protocol, targetUrl, ieProxyConfig.LpszProxy, ieProxyConfig.LpszProxyBypass)
		if err == nil {
			return proxy
		} else if !isNotFound(err) {
			log.Printf("[proxy.Provider.readWinHttpProxy] Failed to parse named proxy: %s\n", err)
		}
	}
	// netsh winhttp
	proxy, err := p.getProxyWinHttpDefault(protocol, targetUrl)
	if err == nil {
		return proxy
	} else if !isNotFound(err) {
		log.Printf("[proxy.Provider.readWinHttpProxy] Failed to parse WinHttp default proxy info: %s\n", err)
	}
	return nil
}

/*
CurrentUserIEProxyConfig represents the "Internet Options" window you are probably familiar with
when working with explorer.
Returns:
	CurrentUserIEProxyConfig, nil: No errors occurred
	nil, error: An error occurred
*/
func (p *providerWindows) getIeProxyConfigCurrentUser() (*winhttp.CurrentUserIEProxyConfig, error) {
	ieProxyConfig, err := winhttp.GetIEProxyConfigForCurrentUser()
	if err != nil {
		return nil, err
	}
	return ieProxyConfig, nil
}

/*
Returns the Proxy found through automatic detection, if any.
Params:
	protocol: The protocol of traffic the proxy is to be used for. (i.e. http, https, ftp, socks)
	targetUrl: The URL the proxy is to be used for. (i.e. https://test.endpoint.rapid7.com)
Returns:
	Proxy, nil: A proxy was found
	nil, notFoundError: No proxy was found
	nil, error: An error occurred
*/
func (p *providerWindows) getProxyAutoDetect(protocol string, targetUrl *url.URL) (Proxy, error) {
	return p.getProxyForUrl(srcAutoDetect, protocol, targetUrl,
		&winhttp.AutoProxyOptions{
			DwFlags:                winhttp.WINHTTP_AUTOPROXY_AUTO_DETECT,
			DwAutoDetectFlags:      winhttp.WINHTTP_AUTO_DETECT_TYPE_DHCP | winhttp.WINHTTP_AUTO_DETECT_TYPE_DNS_A,
			FAutoLogonIfChallenged: true,
		})
}

/*
Returns the Proxy found with the PAC file retrieved from autoConfigUrl, if any.
Params:
	protocol: The protocol of traffic the proxy is to be used for. (i.e. http, https, ftp, socks)
	targetUrl: The URL the proxy is to be used for. (i.e. https://test.endpoint.rapid7.com)
	autoConfigUrl: The URL to be used to retrieve the PAC file.
Returns:
	Proxy, nil: A proxy was found
	nil, notFoundError: No proxy was found
	nil, error: An error occurred
*/
func (p *providerWindows) getProxyAutoConfigUrl(protocol string, targetUrl *url.URL, autoConfigUrl string) (Proxy, error) {
	return p.getProxyForUrl(srcAutoConfigUrl, protocol, targetUrl,
		&winhttp.AutoProxyOptions{
			DwFlags:                winhttp.WINHTTP_AUTOPROXY_CONFIG_URL,
			LpszAutoConfigUrl:      winhttp.StringToLpwstr(autoConfigUrl),
			FAutoLogonIfChallenged: true,
		})
}

/*
Returns the Proxy found through the WinHTTP default settings.
This is typically configured via: `netsh winhttp set proxy`
Params:
	protocol: The protocol of traffic the proxy is to be used for. (i.e. http, https, ftp, socks)
	targetUrl: The URL the proxy is to be used for. (i.e. https://test.endpoint.rapid7.com)
Returns:
	Proxy, nil: A proxy was found
	nil, notFoundError: No proxy was found
	nil, error: An error occurred
*/
func (p *providerWindows) getProxyWinHttpDefault(protocol string, targetUrl *url.URL) (Proxy, error) {
	pInfo, err := winhttp.GetDefaultProxyConfiguration()
	if err != nil {
		return nil, err
	}
	defer p.freeWinHttpResource(pInfo)
	return p.parseProxyInfo(srcWinHttp, protocol, targetUrl, pInfo.LpszProxy, pInfo.LpszProxyBypass)
}

/*
Returns the Proxy found through either automatic detection or a automatic configuration URL.
Params:
	src: If a proxy is constructed, the human readable source to associated it with.
	protocol: The protocol of traffic the proxy is to be used for. (i.e. http, https, ftp, socks)
	targetUrl: The URL the proxy is to be used for. (i.e. https://test.endpoint.rapid7.com)
	autoProxyOptions: Use this to inform WinHTTP what route to take when doing the lookup (automatic detection, or automatic configuration URL)
Returns:
	Proxy, nil: A proxy was found
	nil, notFoundError: No proxy was found
	nil, error: An error occurred
*/
func (p *providerWindows) getProxyForUrl(src string, protocol string, targetUrl *url.URL, autoProxyOptions *winhttp.AutoProxyOptions) (Proxy, error) {
	pInfo, err := p.getProxyInfoForUrl(targetUrl, autoProxyOptions)
	if err != nil {
		return nil, err
	}
	defer p.freeWinHttpResource(pInfo)
	return p.parseProxyInfo(src, protocol, targetUrl, pInfo.LpszProxy, pInfo.LpszProxyBypass)
}

/*
Returns the ProxyInfo found through either automatic detection or a automatic configuration URL.
Params:
	targetUrl: The URL the proxy is to be used for. (i.e. https://test.endpoint.rapid7.com)
	autoProxyOptions: Use this to inform WinHTTP what route to take when doing the lookup (automatic detection, or automatic configuration URL)
Returns:
	ProxyInfo, nil: A proxy was found
	nil, error: An error occurred
*/
func (p *providerWindows) getProxyInfoForUrl(targetUrl *url.URL, autoProxyOptions *winhttp.AutoProxyOptions) (*winhttp.ProxyInfo, error) {
	h, err := winhttp.Open(
		winhttp.StringToLpwstr(userAgent),
		winhttp.WINHTTP_ACCESS_TYPE_NO_PROXY,
		winhttp.StringToLpwstr(""),
		winhttp.StringToLpwstr(""),
		0)
	if err != nil {
		return nil, err
	}
	defer p.closeHandle(h)
	err = winhttp.SetTimeouts(h, p.resolveTimeout, p.connectTimeout, p.sendTimeout, p.receiveTimeout)
	if err != nil {
		return nil, err
	}
	proxyInfo, err := winhttp.GetProxyForUrl(h, winhttp.StringToLpwstr(targetUrl.String()), autoProxyOptions)
	if err != nil {
		return nil, err
	}
	return proxyInfo, nil
}

/*
Parse the lpszProxy and lpszProxyBypass into a Proxy configuration (if any).
Params:
	src: If a proxy is constructed, the human readable source to associated it with.
	protocol: The protocol of traffic the proxy is to be used for. (i.e. http, https, ftp, socks)
	targetUrl: The URL the proxy is to be used for. (i.e. https://test.endpoint.rapid7.com)
	lpszProxy: The Lpwstr which represents the proxy value (if any). This value can be optionally separated by protocol.
	lpszProxyBypass: The Lpwstr which represents the proxy bypass value (if any).
Returns:
	Proxy, nil: A proxy was found
	nil, notFoundError: No proxy was found or was bypassed
	nil, error: An error occurred
*/
//noinspection SpellCheckingInspection
func (p *providerWindows) parseProxyInfo(src string, protocol string, targetUrl *url.URL, lpszProxy winhttp.Lpwstr, lpszProxyBypass winhttp.Lpwstr) (Proxy, error) {
	proxyUrlStr := p.parseLpszProxy(protocol, winhttp.LpwstrToString(lpszProxy))
	if proxyUrlStr == "" {
		return nil, new(notFoundError)
	}
	proxyUrl, err := ParseURL(proxyUrlStr, "")
	if err != nil {
		return nil, err
	}
	proxyBypass := winhttp.LpwstrToString(lpszProxyBypass)
	if proxyBypass != "" {
		bypass := p.isLpszProxyBypass(targetUrl, proxyBypass)
		log.Printf("[proxy.Provider.parseProxyInfo]: lpszProxyBypass=\"%s\", targetUrl=%s, bypass=%t", proxyBypass, targetUrl, bypass)
		if bypass {
			return nil, new(notFoundError)
		}
	}
	return NewProxy(proxyUrl, src)
}

/*
Parse the lpszProxy into a single proxy URL, represented as a string.
For example:
	("https", "1.2.3.4") -> "1.2.3.4"
	("https", "https=1.2.3.4;http=4.5.6.7") -> "1.2.3.4"
	("https", "") -> ""
	("https", "http=4.5.6.7") -> ""
Params:
	protocol: The protocol of traffic the proxy is to be used for. (i.e. http, https, ftp, socks)
	lpszProxy: The Lpwstr which represents the proxy value (if any). This value can be optionally separated by protocol.
Returns:
	string: The proxy URL (if any) from the lpszProxy value.
*/
//noinspection SpellCheckingInspection
func (p *providerWindows) parseLpszProxy(protocol string, lpszProxy string) string {
	m := ""
	for _, s := range strings.Split(lpszProxy, ";") {
		parts := strings.SplitN(s, "=", 2)
		// No protocol?
		if len(parts) < 2 {
			// Assign a match, but keep looking in case we have a protocol specific match
			m = s
		} else if strings.TrimSpace(parts[0]) == protocol {
			m = parts[1]
			break
		}
	}
	return m
}

/*
Return true if the given targetUrl should bypass a proxy for the given lpszProxyBypass value.
For example:
	("test.endpoint.rapid7.com", "rapid7.com") -> true
	("test.endpoint.rapid7.com", "someHost;anotherHost") -> false
	("test.endpoint.rapid7.com", "") -> false
Params:
	targetUrl: The URL the proxy is to be used for. (i.e. https://test.endpoint.rapid7.com)
	lpszProxyBypass: The Lpwstr which represents the proxy bypass value (if any).
Returns:
	true: The proxy should be bypassed for the given targetUrl
	false: Otherwise
*/
//noinspection SpellCheckingInspection
func (p *providerWindows) isLpszProxyBypass(targetUrl *url.URL, lpszProxyBypass string) bool {
	return p.isProxyBypass(targetUrl, lpszProxyBypass, ";")
}

/*
Close the given handle. This should always be called when use of a handle is no longer required.
Params:
	h: The handle
*/
func (p *providerWindows) closeHandle(h winhttp.HInternet) {
	if err := winhttp.CloseHandle(h); err != nil {
		log.Printf("[proxy.Provider.closeHandle] Failed to close handle \"%d\": %s\n", h, err)
	}
}

/*
Free an allocated object. This should always be called for structs returned from winhttp calls.
Params:
	r: The resource
*/
func (p *providerWindows) freeWinHttpResource(r winhttp.Allocated) {
	if r == nil {
		return
	}
	if err := r.Free(); err != nil {
		log.Printf("[proxy.Provider.readWinHttp] Failed to free struct \"%s\": %s\n", reflect.TypeOf(r), err)
	}
}
