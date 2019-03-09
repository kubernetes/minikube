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
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

/*
Represents a Proxy which can be used to proxy communications.
*/
type Proxy interface {
	// The Proxy's protocol
	Protocol() string
	// The Proxy's host (hostname or IP)
	Host() string
	// The Proxy's port
	Port() uint16
	// username, true: A username was specified
	// username, false: A username was not specified, username should be considered "nil"
	Username() (string, bool)
	// password, true: A password was specified
	// password, false: A password was not specified, password should be considered "nil"
	Password() (string, bool)
	// Human readable location where this Proxy was found
	Src() string
	// A fully qualified URL for this Proxy
	URL() *url.URL
	// A human readable representation of this Proxy. User info (if any) will be obfuscated. Use URL() if you need the URL with user info.
	String() string
	MarshalJSON() ([]byte, error)
	toMap() map[string]interface{}
}

func NewProxy(u *url.URL, src string) (Proxy, error) {
	proxy := new(proxy)
	if err := proxy.init(u, src); err != nil {
		return nil, err
	}
	return proxy, nil
}

const (
	defaultPort = 8443
)

var defaultPorts = map[string]uint16{"": defaultPort, protocolHTTP: defaultPort, protocolHTTPS: defaultPort}

type proxy struct {
	protocol string
	host     string
	port     uint16
	user     *url.Userinfo
	src      string
}

func (p *proxy) init(u *url.URL, src string) error {
	if u == nil {
		return errors.New("nil URL")
	}
	scheme := strings.TrimSpace(strings.ToLower(u.Scheme))
	host, port, err := SplitHostPort(u)
	if err != nil {
		return err
	}
	if host == "" {
		return errors.New("empty host")
	}
	if port == 0 {
		port = defaultPorts[scheme]
	}
	if port == 0 {
		port = defaultPorts[""]
	}
	if port == 0 {
		return errors.New("port undefined")
	}
	p.protocol = scheme
	p.host = host
	p.port = port
	p.user = u.User
	p.src = src
	return nil
}

func (p *proxy) Protocol() string {
	return p.protocol
}

func (p *proxy) Host() string {
	return p.host
}

func (p *proxy) Port() uint16 {
	return p.port
}

func (p *proxy) Username() (string, bool) {
	if p.user == nil {
		return "", false
	}
	return p.user.Username(), true
}

func (p *proxy) Password() (string, bool) {
	if p.user == nil {
		return "", false
	}
	return p.user.Password()
}

func (p *proxy) Src() string {
	return p.src
}

func (p *proxy) URL() *url.URL {
	return &url.URL{
		Scheme: p.Protocol(),
		Host:   fmt.Sprintf("%s:%d", p.Host(), p.Port()),
		User:   p.user,
	}
}

func (p *proxy) String() string {
	var auth string
	if _, exists := p.Username(); exists {
		if _, exists := p.Password(); exists {
			auth = "<username>:<password>@"
		} else {
			auth = "<username>@"
		}
	}
	return fmt.Sprintf("%s|%s://%s%s:%d", p.Src(), p.Protocol(), auth, p.Host(), p.Port())
}

func (p *proxy) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.toMap())
}

func (p *proxy) toMap() map[string]interface{} {
	m := map[string]interface{}{
		"protocol": p.Protocol(),
		"host":     p.Host(),
		"port":     p.Port(),
		"src":      p.Src(),
	}
	if usernameStr, exists := p.Username(); exists {
		m["username"] = usernameStr
	} else {
		m["username"] = nil
	}
	if passwordStr, exists := p.Password(); exists {
		m["password"] = passwordStr
	} else {
		m["password"] = nil
	}
	return m
}
