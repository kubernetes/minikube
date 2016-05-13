// Copyright (c) 2014 The SkyDNS Authors. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.

package cache

import (
	"testing"
	"time"

	"github.com/miekg/dns"
)

const testTTL = 2

type testcase struct {
	m           *dns.Msg
	dnssec, tcp bool
}

func newMsg(zone string, typ uint16) *dns.Msg {
	m := &dns.Msg{}
	m.SetQuestion(zone, typ)
	return m
}

func TestInsertMessage(t *testing.T) {
	c := New(10, testTTL)

	testcases := []testcase{
		{newMsg("miek.nl.", dns.TypeMX), false, false},
		{newMsg("miek2.nl.", dns.TypeNS), false, false},
		{newMsg("miek3.nl.", dns.TypeMX), true, false},
	}

	for _, tc := range testcases {
		c.InsertMessage(Key(tc.m.Question[0], tc.dnssec, tc.tcp), tc.m)

		m1 := c.Hit(tc.m.Question[0], tc.dnssec, tc.tcp, tc.m.Id)
		if m1.Question[0].Qtype != tc.m.Question[0].Qtype {
			t.Fatalf("bad Qtype, expected %d, got %d:", tc.m.Question[0].Qtype, m1.Question[0].Qtype)
		}
		if m1.Question[0].Name != tc.m.Question[0].Name {
			t.Fatalf("bad Qtype, expected %s, got %s:", tc.m.Question[0].Name, m1.Question[0].Name)
		}

		m1 = c.Hit(tc.m.Question[0], !tc.dnssec, tc.tcp, tc.m.Id)
		if m1 != nil {
			t.Fatalf("bad cache hit, expected <nil>, got %s:", m1)
		}
		m1 = c.Hit(tc.m.Question[0], !tc.dnssec, !tc.tcp, tc.m.Id)
		if m1 != nil {
			t.Fatalf("bad cache hit, expected <nil>, got %s:", m1)
		}
		m1 = c.Hit(tc.m.Question[0], tc.dnssec, !tc.tcp, tc.m.Id)
		if m1 != nil {
			t.Fatalf("bad cache hit, expected <nil>, got %s:", m1)
		}
	}
}

func TestExpireMessage(t *testing.T) {
	c := New(10, testTTL-1)

	tc := testcase{newMsg("miek.nl.", dns.TypeMX), false, false}
	c.InsertMessage(Key(tc.m.Question[0], tc.dnssec, tc.tcp), tc.m)

	m1 := c.Hit(tc.m.Question[0], tc.dnssec, tc.tcp, tc.m.Id)
	if m1.Question[0].Qtype != tc.m.Question[0].Qtype {
		t.Fatalf("bad Qtype, expected %d, got %d:", tc.m.Question[0].Qtype, m1.Question[0].Qtype)
	}
	if m1.Question[0].Name != tc.m.Question[0].Name {
		t.Fatalf("bad Qtype, expected %s, got %s:", tc.m.Question[0].Name, m1.Question[0].Name)
	}

	time.Sleep(testTTL)

	m1 = c.Hit(tc.m.Question[0], tc.dnssec, tc.tcp, tc.m.Id)
	if m1.Question[0].Qtype != tc.m.Question[0].Qtype {
		t.Fatalf("bad Qtype, expected %d, got %d:", tc.m.Question[0].Qtype, m1.Question[0].Qtype)
	}
	if m1.Question[0].Name != tc.m.Question[0].Name {
		t.Fatalf("bad Qtype, expected %s, got %s:", tc.m.Question[0].Name, m1.Question[0].Name)
	}
}
