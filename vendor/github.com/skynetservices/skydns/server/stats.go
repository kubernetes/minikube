// Copyright (c) 2014 The SkyDNS Authors. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.

package server

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/miekg/dns"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	prometheusPort      = os.Getenv("PROMETHEUS_PORT")
	prometheusPath      = os.Getenv("PROMETHEUS_PATH")
	prometheusNamespace = os.Getenv("PROMETHEUS_NAMESPACE")
	prometheusSubsystem = os.Getenv("PROMETHEUS_SUBSYSTEM")
)

var (
	promDnssecOkCount        prometheus.Counter
	promExternalRequestCount *prometheus.CounterVec
	promRequestCount         *prometheus.CounterVec
	promErrorCount           *prometheus.CounterVec
	promCacheMiss            *prometheus.CounterVec
	promRequestDuration      *prometheus.HistogramVec
	promResponseSize         *prometheus.HistogramVec
)

// Metrics registers the DNS metrics to Prometheus, and starts the internal metrics
// server if the environment variable PROMETHEUS_PORT is set.
func Metrics() {
	if prometheusPath == "" {
		prometheusPath = "/metrics"
	}
	if prometheusSubsystem == "" {
		prometheusSubsystem = "skydns"
	}

	RegisterMetrics(prometheusNamespace, prometheusSubsystem)

	if prometheusPort == "" {
		return
	}

	_, err := strconv.Atoi(prometheusPort)
	if err != nil {
		fatalf("bad port for prometheus: %s", prometheusPort)
	}

	http.Handle(prometheusPath, prometheus.Handler())
	go func() {
		fatalf("%s", http.ListenAndServe(":"+prometheusPort, nil))
	}()
	logf("metrics enabled on :%s%s", prometheusPort, prometheusPath)
}

// RegisterMetrics registers DNS specific Prometheus metrics with the provided namespace
// and subsystem.
func RegisterMetrics(prometheusNamespace, prometheusSubsystem string) {
	promRequestCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: prometheusNamespace,
		Subsystem: prometheusSubsystem,
		Name:      "dns_request_count",
		Help:      "Counter of DNS requests made.",
	}, []string{"type"}) // udp, tcp
	prometheus.MustRegister(promRequestCount)

	promDnssecOkCount = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: prometheusNamespace,
		Subsystem: prometheusSubsystem,
		Name:      "dns_dnssec_ok_count",
		Help:      "Counter of DNSSEC requests.",
	})
	prometheus.MustRegister(promDnssecOkCount) // Maybe more bits here?

	promErrorCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: prometheusNamespace,
		Subsystem: prometheusSubsystem,
		Name:      "dns_error_count",
		Help:      "Counter of DNS requests resulting in an error.",
	}, []string{"error"}) // nxdomain, nodata, truncated, refused, overflow
	prometheus.MustRegister(promErrorCount)

	promCacheMiss = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: prometheusNamespace,
		Subsystem: prometheusSubsystem,
		Name:      "dns_cache_miss_count",
		Help:      "Counter of DNS requests that result in a cache miss.",
	}, []string{"type"}) // response, signature
	prometheus.MustRegister(promCacheMiss)

	promRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: prometheusNamespace,
		Subsystem: prometheusSubsystem,
		Name:      "dns_request_duration",
		Help:      "Histogram of the time (in seconds) each request took to resolve.",
		Buckets:   append([]float64{0.001, 0.003}, prometheus.DefBuckets...),
	}, []string{"type"}) // udp, tcp
	prometheus.MustRegister(promRequestDuration)

	promResponseSize = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: prometheusNamespace,
		Subsystem: prometheusSubsystem,
		Name:      "dns_response_size",
		Help:      "Size of the returns response in bytes.",
		// 4k increments after 4096
		Buckets: []float64{0, 512, 1024, 1500, 2048, 4096,
			8192, 12288, 16384, 20480, 24576, 28672, 32768, 36864,
			40960, 45056, 49152, 53248, 57344, 61440, 65536,
		},
	}, []string{"type"}) // udp, tcp
	prometheus.MustRegister(promResponseSize)

	promExternalRequestCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: prometheusNamespace,
		Subsystem: prometheusSubsystem,
		Name:      "dns_request_external_count",
		Help:      "Counter of external DNS requests.",
	}, []string{"type"}) // recursive, stub, lookup
	prometheus.MustRegister(promExternalRequestCount)

}

// metricSizeAndDuration sets the size and duration metrics.
func metricSizeAndDuration(resp *dns.Msg, start time.Time, tcp bool) {
	net := "udp"
	rlen := float64(0)
	if tcp {
		net = "tcp"
	}
	if resp != nil {
		rlen = float64(resp.Len())
	}
	promRequestDuration.WithLabelValues(net).Observe(float64(time.Since(start)) / float64(time.Second))
	promResponseSize.WithLabelValues(net).Observe(rlen)
}

// Counter is the metric interface used by this package
type Counter interface {
	Inc(i int64)
}

type nopCounter struct{}

func (nopCounter) Inc(_ int64) {}

// These are the old stat variables defined by this package. This
// used by graphite.
var (
	// Pondering deletion in favor of the better and more
	// maintained (by me) prometheus reporting.

	StatsForwardCount     Counter = nopCounter{}
	StatsStubForwardCount Counter = nopCounter{}
	StatsLookupCount      Counter = nopCounter{}
	StatsRequestCount     Counter = nopCounter{}
	StatsDnssecOkCount    Counter = nopCounter{}
	StatsNameErrorCount   Counter = nopCounter{}
	StatsNoDataCount      Counter = nopCounter{}

	StatsDnssecCacheMiss Counter = nopCounter{}
)
