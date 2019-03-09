// +build windows

// Copyright (C) 2018, Rapid7 LLC, Boston, MA, USA.
// All rights reserved. This material contains unpublished, copyrighted
// work including confidential and proprietary information of Rapid7.
package winhttp

import (
	"golang.org/x/sys/windows"
	"unicode/utf16"
	"unsafe"
)

//noinspection SpellCheckingInspection,GoNameStartsWithPackageName,GoSnakeCaseUsage
const (
	// WINHTTP_AUTOPROXY_AUTO_DETECT - Attempt to automatically discover the URL of the PAC file using both DHCP and DNS queries to the local network.
	WINHTTP_AUTOPROXY_AUTO_DETECT = 0x00000001
	// WINHTTP_AUTOPROXY_CONFIG_URL - Download the PAC file from the URL specified by lpszAutoConfigUrl in the WINHTTP_AUTOPROXY_OPTIONS structure.
	WINHTTP_AUTOPROXY_CONFIG_URL = 0x00000002
	// WINHTTP_AUTO_DETECT_TYPE_DHCP - Use DHCP to locate the proxy auto-configuration file.
	WINHTTP_AUTO_DETECT_TYPE_DHCP = 0x00000001
	// WINHTTP_AUTO_DETECT_TYPE_DNS_A - Use DNS to attempt to locate the proxy auto-configuration file at a well-known location on the domain of the local computer.
	WINHTTP_AUTO_DETECT_TYPE_DNS_A = 0x00000002
	// WINHTTP_ACCESS_TYPE_NO_PROXY - Resolves all host names directly without a proxy.
	WINHTTP_ACCESS_TYPE_NO_PROXY = 0x00000001
)

//noinspection SpellCheckingInspection
type (
	Lpwstr    *uint16
	Dword     uint32
	HInternet uintptr
	Allocated interface {
		Free() error
	}
	/*
		typedef struct __unnamed_struct_4 {
		  DWORD   dwFlags;
		  DWORD   dwAutoDetectFlags;
		  LPCWSTR lpszAutoConfigUrl;
		  LPVOID  lpvReserved;
		  DWORD   dwReserved;
		  BOOL    fAutoLogonIfChallenged;
		} WINHTTP_AUTOPROXY_OPTIONS;
	*/
	AutoProxyOptions struct {
		DwFlags                Dword
		DwAutoDetectFlags      Dword
		LpszAutoConfigUrl      Lpwstr
		lpvReserved            uintptr
		dwReserved             uint32
		FAutoLogonIfChallenged bool
	}
	/*
		typedef struct WINHTTP_PROXY_INFO {
		  DWORD  dwAccessType;
		  LPWSTR lpszProxy;
		  LPWSTR lpszProxyBypass;
		}  *LPWINHTTP_PROXY_INFO;
	*/
	ProxyInfo struct {
		DwAccessType    Dword
		LpszProxy       Lpwstr
		LpszProxyBypass Lpwstr
	}
	/*
		typedef struct WINHTTP_CURRENT_USER_IE_PROXY_CONFIG {
		  BOOL   fAutoDetect;
		  LPWSTR lpszAutoConfigUrl;
		  LPWSTR lpszProxy;
		  LPWSTR lpszProxyBypass;
		};
	*/
	CurrentUserIEProxyConfig struct {
		FAutoDetect       bool
		LpszAutoConfigUrl Lpwstr
		LpszProxy         Lpwstr
		LpszProxyBypass   Lpwstr
	}
)

/*
MSDN:
```
The WinHttpOpen function initializes, for an application, the use of WinHTTP functions and returns a WinHTTP-session handle.
	WINHTTPAPI HINTERNET WinHttpOpen(
	  LPCWSTR pszAgentW,
	  DWORD   dwAccessType,
	  LPCWSTR pszProxyW,
	  LPCWSTR pszProxyBypassW,
	  DWORD   dwFlags
	);
pszAgentW
	A pointer to a string variable that contains the name of the application or entity calling the WinHTTP functions. This name is used as the user agent in the HTTP protocol.
dwAccessType
	Type of access required.
pszProxyW
	A pointer to a string variable that contains the name of the proxy server to use when proxy access is specified by setting dwAccessType to WINHTTP_ACCESS_TYPE_NAMED_PROXY. The WinHTTP functions recognize only CERN type proxies for HTTP. If dwAccessType is not set to WINHTTP_ACCESS_TYPE_NAMED_PROXY, this parameter must be set to WINHTTP_NO_PROXY_NAME.
pszProxyBypassW
	A pointer to a string variable that contains an optional semicolon delimited list of host names or IP addresses, or both, that should not be routed through the proxy when dwAccessType is set to WINHTTP_ACCESS_TYPE_NAMED_PROXY. The list can contain wildcard characters. Do not use an empty string, because the WinHttpOpen function uses it as the proxy bypass list. If this parameter specifies the "<local>" macro in the list as the only entry, this function bypasses any host name that does not contain a period. If dwAccessType is not set to WINHTTP_ACCESS_TYPE_NAMED_PROXY, this parameter must be set to WINHTTP_NO_PROXY_BYPASS.
dwFlags
	Unsigned long integer value that contains the flags that indicate various options affecting the behavior of this function.
Returns a valid session handle if successful, or NULL otherwise. To retrieve extended error information, call GetLastError.
```
*/
//noinspection SpellCheckingInspection
func Open(pszAgentW Lpwstr, dwAccessType Dword, pszProxyW Lpwstr, pszProxyBypassW Lpwstr, dwFlags Dword) (HInternet, error) {
	if err := openP.Find(); err != nil {
		return 0, err
	}
	r, _, err := openP.Call(
		uintptr(unsafe.Pointer(pszAgentW)),
		uintptr(dwAccessType),
		uintptr(unsafe.Pointer(pszProxyW)),
		uintptr(unsafe.Pointer(pszProxyBypassW)),
		uintptr(dwFlags),
	)
	if rNil(r) {
		return 0, err
	}
	return HInternet(r), nil
}

/*
MSDN:
```
The WinHttpCloseHandle function closes a single HINTERNET handle.
	BOOLAPI WinHttpCloseHandle(
	  IN HINTERNET hInternet
	);
hInternet
	Valid HINTERNET handle to be closed.
Returns TRUE if the handle is successfully closed, or FALSE otherwise.
```
*/
func CloseHandle(hInternet HInternet) error {
	if err := closeHandleP.Find(); err != nil {
		return err
	}
	r, _, err := closeHandleP.Call(uintptr(hInternet))
	if rTrue(r) {
		return nil
	}
	return err
}

/*
The WinHttpSetTimeouts function sets time-outs involved with HTTP transactions.
	BOOLAPI WinHttpSetTimeouts(
	  IN HINTERNET hInternet,
	  IN int       nResolveTimeout,
	  IN int       nConnectTimeout,
	  IN int       nSendTimeout,
	  IN int       nReceiveTimeout
	);
hInternet
	The HINTERNET handle returned by WinHttpOpen or WinHttpOpenRequest.
nResolveTimeout
	A value of type integer that specifies the time-out value, in milliseconds, to use for name resolution. If resolution takes longer than this time-out value, the action is canceled. The initial value is zero, meaning no time-out (infinite).
nConnectTimeout
	A value of type integer that specifies the time-out value, in milliseconds, to use for server connection requests. If a connection request takes longer than this time-out value, the request is canceled. The initial value is 60,000 (60 seconds).
	TCP/IP can time out while setting up the socket during the three leg SYN/ACK exchange, regardless of the value of this parameter.
nSendTimeout
	A value of type integer that specifies the time-out value, in milliseconds, to use for sending requests. If sending a request takes longer than this time-out value, the send is canceled. The initial value is 30,000 (30 seconds).
nReceiveTimeout
	A value of type integer that specifies the time-out value, in milliseconds, to receive a response to a request. If a response takes longer than this time-out value, the request is canceled. The initial value is 30,000 (30 seconds).
Returns TRUE if successful, or FALSE otherwise.
*/
func SetTimeouts(hInternet HInternet, nResolveTimeout int, nConnectTimeout int, nSendTimeout int, nReceiveTimeout int) error {
	if err := setTimeoutsP.Find(); err != nil {
		return err
	}
	r, _, err := setTimeoutsP.Call(
		uintptr(hInternet),
		uintptr(nResolveTimeout),
		uintptr(nConnectTimeout),
		uintptr(nSendTimeout),
		uintptr(nReceiveTimeout))
	if rTrue(r) {
		return nil
	}
	return err
}

/*
MSDN:
```
The WinHttpGetProxyForUrl function retrieves the proxy data for the specified URL.
	BOOLAPI WinHttpGetProxyForUrl(
	  IN HINTERNET                 hSession,
	  IN LPCWSTR                   lpcwszUrl,
	  IN WINHTTP_AUTOPROXY_OPTIONS *pAutoProxyOptions,
	  OUT WINHTTP_PROXY_INFO       *pProxyInfo
	);
hSession
	The WinHTTP session handle returned by the WinHttpOpen function.
lpcwszUrl
	A pointer to a null-terminated Unicode string that contains the URL of the HTTP request that the application is preparing to send.
pAutoProxyOptions
	A pointer to a WINHTTP_AUTOPROXY_OPTIONS structure that specifies the auto-proxy options to use.
pProxyInfo
	A pointer to a WINHTTP_PROXY_INFO structure that receives the proxy setting. This structure is then applied to the request handle using the WINHTTP_OPTION_PROXY option. Free the lpszProxy and lpszProxyBypass strings contained in this structure (if they are non-NULL) using the GlobalFree function.
If the function succeeds, the function returns TRUE.
If the function fails, it returns FALSE. For extended error data, call GetLastError.
```
*/
//noinspection SpellCheckingInspection
func GetProxyForUrl(hInternet HInternet, lpcwszUrl Lpwstr, pAutoProxyOptions *AutoProxyOptions) (*ProxyInfo, error) {
	if err := getProxyForUrlP.Find(); err != nil {
		return nil, err
	}
	p := new(ProxyInfo)
	r, _, err := getProxyForUrlP.Call(
		uintptr(hInternet),
		uintptr(unsafe.Pointer(lpcwszUrl)),
		uintptr(unsafe.Pointer(pAutoProxyOptions)),
		uintptr(unsafe.Pointer(p)))
	if rTrue(r) {
		return p, nil
	}
	return nil, err
}

/*
The WinHttpGetIEProxyConfigForCurrentUser function retrieves the Internet Explorer proxy configuration for the current user.
	BOOLAPI WinHttpGetIEProxyConfigForCurrentUser(
	  IN OUT WINHTTP_CURRENT_USER_IE_PROXY_CONFIG *pProxyConfig
	);
pProxyConfig
	A pointer, on input, to a WINHTTP_CURRENT_USER_IE_PROXY_CONFIG structure. On output, the structure contains the Internet Explorer proxy settings for the current active network connection (for example, LAN, dial-up, or VPN connection).
Returns TRUE if successful, or FALSE otherwise.
*/
//noinspection SpellCheckingInspection
func GetIEProxyConfigForCurrentUser() (*CurrentUserIEProxyConfig, error) {
	if err := getIEProxyConfigForCurrentUserP.Find(); err != nil {
		return nil, err
	}
	p := new(CurrentUserIEProxyConfig)
	r, _, err := getIEProxyConfigForCurrentUserP.Call(uintptr(unsafe.Pointer(p)))
	if rTrue(r) {
		return p, nil
	}
	return nil, err
}

/*
The WinHttpGetDefaultProxyConfiguration function retrieves the default WinHTTP proxy configuration from the registry.
	WINHTTPAPI BOOL WinHttpGetDefaultProxyConfiguration(
	  IN OUT WINHTTP_PROXY_INFO *pProxyInfo
	);
pProxyInfo
	A pointer to a variable of type WINHTTP_PROXY_INFO that receives the default proxy configuration.
Returns TRUE if successful or FALSE otherwise.
*/
func GetDefaultProxyConfiguration() (*ProxyInfo, error) {
	pInfo := new(ProxyInfo)
	if err := getDefaultProxyConfigurationP.Find(); err != nil {
		return nil, err
	}
	r, _, err := getDefaultProxyConfigurationP.Call(uintptr(unsafe.Pointer(pInfo)))
	if rTrue(r) {
		return pInfo, nil
	}
	return nil, err
}

//noinspection SpellCheckingInspection
func LpwstrToString(d Lpwstr) string {
	if d == nil {
		return ""
	}
	s := make([]uint16, 0, 256)
	p := uintptr(unsafe.Pointer(d))
	pMax := p + lpwstrMaxBytes
	for ; p < pMax; p += 2 {
		c := *(*uint16)(unsafe.Pointer(p))
		// NUL char is EOF
		if c == 0 {
			return string(utf16.Decode(s))
		}
		s = append(s, c)
	}
	return ""
}

//noinspection SpellCheckingInspection
func StringToLpwstr(s string) *uint16 {
	if s == "" {
		return nil
	}
	// If s contains \x00, we'll just silently return nil, better than a panic
	r, err := windows.UTF16PtrFromString(s)
	if err != nil {
		return nil
	}
	return r
}

//noinspection SpellCheckingInspection
func (p *ProxyInfo) Free() error {
	if p == nil {
		return nil
	}
	rerr := globalFree(p.LpszProxy)
	if err := globalFree(p.LpszProxyBypass); rerr == nil && err != nil {
		rerr = err
	}
	return rerr
}

//noinspection SpellCheckingInspection
func (p *CurrentUserIEProxyConfig) Free() error {
	if p == nil {
		return nil
	}
	rerr := globalFree(p.LpszAutoConfigUrl)
	if err := globalFree(p.LpszProxy); rerr == nil && err != nil {
		rerr = err
	}
	if err := globalFree(p.LpszProxyBypass); rerr == nil && err != nil {
		rerr = err
	}
	return rerr
}

/************* BEGIN PRIVATE IMPL *************/

// Set a sane ceiling in case we don't find \x00\x00
// Maximum number of bytes for any returned lpwstr
const lpwstrMaxBytes = 1024 * 512

//noinspection SpellCheckingInspection
var (
	kd                              = windows.NewLazySystemDLL("kernel32.dll")
	globalFreeP                     = kd.NewProc("GlobalFree")
	whd                             = windows.NewLazySystemDLL("winhttp.dll")
	openP                           = whd.NewProc("WinHttpOpen")
	closeHandleP                    = whd.NewProc("WinHttpCloseHandle")
	setTimeoutsP                    = whd.NewProc("WinHttpSetTimeouts")
	getProxyForUrlP                 = whd.NewProc("WinHttpGetProxyForUrl")
	getIEProxyConfigForCurrentUserP = whd.NewProc("WinHttpGetIEProxyConfigForCurrentUser")
	getDefaultProxyConfigurationP   = whd.NewProc("WinHttpGetDefaultProxyConfiguration")
)

func globalFree(hMem *uint16) error {
	if hMem == nil {
		return nil
	}
	if err := globalFreeP.Find(); err != nil {
		return err
	}
	r, _, err := globalFreeP.Call(uintptr(unsafe.Pointer(hMem)))
	if rNil(r) {
		return nil
	}
	return err
}

func rNil(r uintptr) bool {
	return r == 0
}

func rTrue(r uintptr) bool {
	return r == 1
}
