/*
   Copyright 2015-2016 Red Hat, Inc. and/or its affiliates
   and other contributors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package metrics

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// HawkularClientError Extracted error information from Hawkular-Metrics server
type HawkularClientError struct {
	msg  string
	Code int
}

// Parameters Initialization parameters to the client
type Parameters struct {
	Tenant      string // Technically optional, but requires setting Tenant() option everytime
	Url         string
	TLSConfig   *tls.Config
	Token       string
	Concurrency int
}

// Client HawkularClient's data structure
type Client struct {
	Tenant string
	url    *url.URL
	client *http.Client
	Token  string
	pool   chan (*poolRequest)
}

type poolRequest struct {
	req   *http.Request
	rChan chan (*poolResponse)
}

type poolResponse struct {
	err  error
	resp *http.Response
}

// HawkularClient HawkularClient base type to define available functions..
type HawkularClient interface {
	Send(*http.Request) (*http.Response, error)
}

// Modifier Modifiers base type
type Modifier func(*http.Request) error

// Filter Filter type for querying
type Filter func(r *http.Request)

// Endpoint Endpoint type to define request URL
type Endpoint func(u *url.URL)

// MetricType restrictions
type MetricType int

const (
	Gauge = iota
	Availability
	Counter
	Generic
)

var longForm = []string{
	"gauges",
	"availability",
	"counters",
	"metrics",
}

var shortForm = []string{
	"gauge",
	"availability",
	"counter",
	"metrics",
}

func (mt MetricType) validate() error {
	if int(mt) > len(longForm) && int(mt) > len(shortForm) {
		return fmt.Errorf("Given MetricType value %d is not valid", mt)
	}
	return nil
}

// String Get string representation of type
func (mt MetricType) String() string {
	if err := mt.validate(); err != nil {
		return "unknown"
	}
	return longForm[mt]
}

func (mt MetricType) shortForm() string {
	if err := mt.validate(); err != nil {
		return "unknown"
	}
	return shortForm[mt]
}

// UnmarshalJSON Custom unmarshaller for MetricType
func (mt *MetricType) UnmarshalJSON(b []byte) error {
	var f interface{}
	err := json.Unmarshal(b, &f)
	if err != nil {
		return err
	}

	if str, ok := f.(string); ok {
		for i, v := range shortForm {
			if str == v {
				*mt = MetricType(i)
				break
			}
		}
	}

	return nil
}

// MarshalJSON Custom marshaller for MetricType
func (mt MetricType) MarshalJSON() ([]byte, error) {
	return json.Marshal(mt.String())
}

// Hawkular-Metrics external structs
// Do I need external.. hmph.

type MetricHeader struct {
	Tenant string      `json:"-"`
	Type   MetricType  `json:"-"`
	Id     string      `json:"id"`
	Data   []Datapoint `json:"data"`
}

// Datapoint Value should be convertible to float64 for numeric values, Timestamp is milliseconds since epoch
type Datapoint struct {
	Timestamp int64             `json:"timestamp"`
	Value     interface{}       `json:"value"`
	Tags      map[string]string `json:"tags,omitempty"`
}

// HawkularError Return payload from Hawkular-Metrics if processing failed
type HawkularError struct {
	ErrorMsg string `json:"errorMsg"`
}

type MetricDefinition struct {
	Tenant        string            `json:"-"`
	Type          MetricType        `json:"type,omitempty"`
	Id            string            `json:"id"`
	Tags          map[string]string `json:"tags,omitempty"`
	RetentionTime int               `json:"dataRetention,omitempty"`
}

// TODO Fix the Start & End to return a time.Time

// Bucketpoint Return structure for bucketed data
type Bucketpoint struct {
	Start       int64        `json:"start"`
	End         int64        `json:"end"`
	Min         float64      `json:"min"`
	Max         float64      `json:"max"`
	Avg         float64      `json:"avg"`
	Median      float64      `json:"median"`
	Empty       bool         `json:"empty"`
	Samples     int64        `json:"samples"`
	Percentiles []Percentile `json:"percentiles"`
}

// Percentile Hawkular-Metrics calculated percentiles representation
type Percentile struct {
	Quantile float64 `json:"quantile"`
	Value    float64 `json:"value"`
}

// Order Basetype for selecting the sorting of datapoints
type Order int

const (
	// ASC Ascending
	ASC = iota
	// DESC Descending
	DESC
)

// String Get string representation of type
func (o Order) String() string {
	switch o {
	case ASC:
		return "ASC"
	case DESC:
		return "DESC"
	}
	return ""
}
