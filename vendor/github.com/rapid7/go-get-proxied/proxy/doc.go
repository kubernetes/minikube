// Package proxy allows you to retrieve a system configured proxy for a given protocol and target URL.
//
// The priority of retrieval is the following:
//
// 	Windows:
//		Configuration File
//		Environment Variable: HTTPS_PROXY, HTTP_PROXY, FTP_PROXY, or ALL_PROXY. `NO_PROXY` is respected.
//		Internet Options: Automatically detect settings (WPAD)
//		Internet Options: Use automatic configuration script (PAC)
//		Internet Options: Manual proxy server
//		WINHTTP: (netsh winhttp)
//
//	Linux:
//		Configuration File
//		Environment Variable: HTTPS_PROXY, HTTP_PROXY, FTP_PROXY, or ALL_PROXY. `NO_PROXY` is respected.
//
//	MacOS:
//		Configuration File
//		Environment Variable: HTTPS_PROXY, HTTP_PROXY, FTP_PROXY, or ALL_PROXY. `NO_PROXY` is respected.
//		Network Settings: scutil
//
// Example Usage
//
// The following is a complete example using assert in a standard test function:
//
//		package main
//
//		import (
//			"github.com/rapid7/go-get-proxied/proxy"
//		)
//
//		func main() {
//	    	p := proxy.NewProvider("").Get("https", "https://rapid7.com")
//			if p != nil {
//				fmt.Printf("Found proxy: %s\n", p)
//			}
// 	   }
//
package proxy
