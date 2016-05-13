// Copyright (c) 2014 The SkyDNS Authors. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.

package server

import (
	"testing"

	"github.com/miekg/dns"
	"github.com/skynetservices/skydns/cache"
)

func TestFit(t *testing.T) {
	m := new(dns.Msg)
	m.SetQuestion("miek.nl", dns.TypeA)

	rr, _ := dns.NewRR("www.miek.nl. IN SRV 10 10 8080 blaat.miek.nl.")
	for i := 0; i < 101; i++ {
		m.Answer = append(m.Answer, rr)
	}
	// Uncompresses length is now 4424. Try trimming this to 1927
	Fit(m, 1927, true)

	if m.Len() > 1927 {
		t.Fatalf("failed to fix message, expected < %d, got %d", 1927, m.Len())
	}
}

func TestCacheTruncated(t *testing.T) {
	s := newTestServer(t, true)
	m := &dns.Msg{}
	m.SetQuestion("skydns.test.", dns.TypeSRV)
	m.Truncated = true
	s.rcache.InsertMessage(cache.Key(m.Question[0], false, false), m)

	// Now asking for this should result in a non-truncated answer.
	resp, _ := dns.Exchange(m, "127.0.0.1:"+StrPort)
	if resp.Truncated {
		t.Fatal("truncated bit should be false")
	}
}

// Store a large message in the cache, then query with a smaller bufsize and check
// we get back a smaller message.
// TODO(miek).
/*
func testCacheStoreLarge(t *testing.T) {
	s := newTestServer(t, true)
	defer s.Stop()

	c := new(dns.Client)
	m := new(dns.Msg)

	for i := 0; i < 2000; i++ {
		is := strconv.Itoa(i)
		m := &msg.Service{
			Host: "2001::" + is, Key: "machine" + is + ".machines.skydns.test.",
		}
		addService(t, s, m.Key, 0, m)
		defer delService(t, s, m.Key)
	}
	m.SetQuestion("machines.skydns.test.", dns.TypeSRV)
	resp, _, err := c.Exchange(m, "127.0.0.1:"+StrPort)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", resp)

	if resp.Rcode != dns.RcodeSuccess {
		t.Fatalf("expecting server failure, got %d", resp.Rcode)
	}
}
*/
