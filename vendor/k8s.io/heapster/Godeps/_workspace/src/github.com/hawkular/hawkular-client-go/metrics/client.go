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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// TODO Instrumentation? To get statistics?

func (c *HawkularClientError) Error() string {
	return fmt.Sprintf("Hawkular returned status code %d, error message: %s", c.Code, c.msg)
}

// Client creation and instance config

const (
	baseURL            string        = "hawkular/metrics"
	defaultConcurrency int           = 1
	timeout            time.Duration = time.Duration(30 * time.Second)
)

// Tenant Override function to replace the Tenant (defaults to Client default)
func Tenant(tenant string) Modifier {
	return func(r *http.Request) error {
		r.Header.Set("Hawkular-Tenant", tenant)
		return nil
	}
}

// Data Add payload to the request
func Data(data interface{}) Modifier {
	return func(r *http.Request) error {
		jsonb, err := json.Marshal(data)
		if err != nil {
			return err
		}

		b := bytes.NewBuffer(jsonb)
		rc := ioutil.NopCloser(b)
		r.Body = rc

		// fmt.Printf("Sending: %s\n", string(jsonb))

		if b != nil {
			r.ContentLength = int64(b.Len())
		}
		return nil
	}
}

// URL Set the request URL
func (c *Client) Url(method string, e ...Endpoint) Modifier {
	// TODO Create composite URLs? Add().Add().. etc? Easier to modify on the fly..
	return func(r *http.Request) error {
		u := c.createURL(e...)
		r.URL = u
		r.Method = method
		return nil
	}
}

// Filters Multiple Filter types to execute
func Filters(f ...Filter) Modifier {
	return func(r *http.Request) error {
		for _, filter := range f {
			filter(r)
		}
		return nil // Or should filter return err?
	}
}

// Param Add query parameters
func Param(k string, v string) Filter {
	return func(r *http.Request) {
		q := r.URL.Query()
		q.Set(k, v)
		r.URL.RawQuery = q.Encode()
	}
}

// TypeFilter Query parameter filtering with type
func TypeFilter(t MetricType) Filter {
	return Param("type", t.shortForm())
}

// TagsFilter Query parameter filtering with tags
func TagsFilter(t map[string]string) Filter {
	j := tagsEncoder(t)
	return Param("tags", j)
}

// IdFilter Query parameter to add filtering by id name
func IdFilter(regexp string) Filter {
	return Param("id", regexp)
}

// StartTimeFilter Query parameter to filter with start time
func StartTimeFilter(startTime time.Time) Filter {
	return Param("start", strconv.Itoa(int(startTime.Unix())))
}

// EndTimeFilter Query parameter to filter with end time
func EndTimeFilter(endTime time.Time) Filter {
	return Param("end", strconv.Itoa(int(endTime.Unix())))
}

// BucketsFilter Query parameter to define amount of buckets
func BucketsFilter(buckets int) Filter {
	return Param("buckets", strconv.Itoa(buckets))
}

// LimitFilter Query parameter to limit result count
func LimitFilter(limit int) Filter {
	return Param("limit", strconv.Itoa(limit))
}

// OrderFilter Query parameter to define the ordering of datapoints
func OrderFilter(order Order) Filter {
	return Param("order", order.String())
}

// StartFromBeginningFilter Return data from the oldest stored datapoint
func StartFromBeginningFilter() Filter {
	return Param("fromEarliest", "true")
}

// StackedFilter Force downsampling of stacked return values
func StackedFilter() Filter {
	return Param("stacked", "true")
}

// PercentilesFilter Query parameter to define the requested percentiles
func PercentilesFilter(percentiles []float64) Filter {
	s := make([]string, 0, len(percentiles))
	for _, v := range percentiles {
		s = append(s, fmt.Sprintf("%v", v))
	}
	j := strings.Join(s, ",")
	return Param("percentiles", j)
}

// The SEND method..

func (c *Client) createRequest() *http.Request {
	req := &http.Request{
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Host:       c.url.Host,
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Hawkular-Tenant", c.Tenant)

	if len(c.Token) > 0 {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	}

	return req
}

// Send Sends a constructed request to the Hawkular-Metrics server
func (c *Client) Send(o ...Modifier) (*http.Response, error) {
	// Initialize
	r := c.createRequest()

	// Run all the modifiers
	for _, f := range o {
		err := f(r)
		if err != nil {
			return nil, err
		}
	}

	rChan := make(chan *poolResponse)
	preq := &poolRequest{r, rChan}

	c.pool <- preq

	presp := <-rChan
	close(rChan)

	return presp.resp, presp.err
}

// Commands

// Create Creates new metric Definition
func (c *Client) Create(md MetricDefinition, o ...Modifier) (bool, error) {
	// Keep the order, add custom prepend
	o = prepend(o, c.Url("POST", TypeEndpoint(md.Type)), Data(md))

	r, err := c.Send(o...)
	if err != nil {
		return false, err
	}

	defer r.Body.Close()

	if r.StatusCode > 399 {
		err = c.parseErrorResponse(r)
		if err, ok := err.(*HawkularClientError); ok {
			if err.Code != http.StatusConflict {
				return false, err
			} else {
				return false, nil
			}
		}
		return false, err
	}
	return true, nil
}

// Definitions Fetch metric definitions
func (c *Client) Definitions(o ...Modifier) ([]*MetricDefinition, error) {
	o = prepend(o, c.Url("GET", TypeEndpoint(Generic)))

	r, err := c.Send(o...)
	if err != nil {
		return nil, err
	}

	defer r.Body.Close()

	if r.StatusCode == http.StatusOK {
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		md := []*MetricDefinition{}
		if b != nil {
			if err = json.Unmarshal(b, &md); err != nil {
				return nil, err
			}
		}
		return md, err
	} else if r.StatusCode > 399 {
		return nil, c.parseErrorResponse(r)
	}

	return nil, nil
}

// Definition Return a single definition
func (c *Client) Definition(t MetricType, id string, o ...Modifier) (*MetricDefinition, error) {
	o = prepend(o, c.Url("GET", TypeEndpoint(t), SingleMetricEndpoint(id)))

	r, err := c.Send(o...)
	if err != nil {
		return nil, err
	}

	defer r.Body.Close()

	if r.StatusCode == http.StatusOK {
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		md := &MetricDefinition{}
		if b != nil {
			if err = json.Unmarshal(b, md); err != nil {
				return nil, err
			}
		}
		return md, err
	} else if r.StatusCode > 399 {
		return nil, c.parseErrorResponse(r)
	}

	return nil, nil
}

// UpdateTags Update tags of a metric (or create if not existing)
func (c *Client) UpdateTags(t MetricType, id string, tags map[string]string, o ...Modifier) error {
	o = prepend(o, c.Url("PUT", TypeEndpoint(t), SingleMetricEndpoint(id), TagEndpoint()), Data(tags))

	r, err := c.Send(o...)
	if err != nil {
		return err
	}

	defer r.Body.Close()

	if r.StatusCode > 399 {
		return c.parseErrorResponse(r)
	}

	return nil
}

// DeleteTags Delete given tags from the definition
func (c *Client) DeleteTags(t MetricType, id string, tags map[string]string, o ...Modifier) error {
	o = prepend(o, c.Url("DELETE", TypeEndpoint(t), SingleMetricEndpoint(id), TagEndpoint(), TagsEndpoint(tags)))

	r, err := c.Send(o...)
	if err != nil {
		return err
	}

	defer r.Body.Close()

	if r.StatusCode > 399 {
		return c.parseErrorResponse(r)
	}

	return nil
}

// Tags Fetch metric definition's tags
func (c *Client) Tags(t MetricType, id string, o ...Modifier) (map[string]string, error) {
	o = prepend(o, c.Url("GET", TypeEndpoint(t), SingleMetricEndpoint(id), TagEndpoint()))

	r, err := c.Send(o...)
	if err != nil {
		return nil, err
	}

	defer r.Body.Close()

	if r.StatusCode == http.StatusOK {
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		tags := make(map[string]string)
		if b != nil {
			if err = json.Unmarshal(b, &tags); err != nil {
				return nil, err
			}
		}
		return tags, nil
	} else if r.StatusCode > 399 {
		return nil, c.parseErrorResponse(r)
	}

	return nil, nil
}

// Write Write datapoints to the server
func (c *Client) Write(metrics []MetricHeader, o ...Modifier) error {
	if len(metrics) > 0 {
		mHs := make(map[MetricType][]MetricHeader)
		for _, m := range metrics {
			if _, found := mHs[m.Type]; !found {
				mHs[m.Type] = make([]MetricHeader, 0, 1)
			}
			mHs[m.Type] = append(mHs[m.Type], m)
		}

		wg := &sync.WaitGroup{}
		errorsChan := make(chan error, len(mHs))

		for k, v := range mHs {
			wg.Add(1)
			go func(k MetricType, v []MetricHeader) {
				defer wg.Done()

				// Should be sorted and splitted by type & tenant..
				on := o
				on = prepend(on, c.Url("POST", TypeEndpoint(k), DataEndpoint()), Data(v))

				r, err := c.Send(on...)
				if err != nil {
					errorsChan <- err
					return
				}

				defer r.Body.Close()

				if r.StatusCode > 399 {
					errorsChan <- c.parseErrorResponse(r)
				}
			}(k, v)
		}
		wg.Wait()
		select {
		case err, ok := <-errorsChan:
			if ok {
				return err
			}
			// If channel is closed, we're done
		default:
			// Nothing to do
		}

	}
	return nil
}

// ReadMetric Read metric datapoints from the server
func (c *Client) ReadMetric(t MetricType, id string, o ...Modifier) ([]*Datapoint, error) {
	o = prepend(o, c.Url("GET", TypeEndpoint(t), SingleMetricEndpoint(id), DataEndpoint()))

	r, err := c.Send(o...)
	if err != nil {
		return nil, err
	}

	defer r.Body.Close()

	if r.StatusCode == http.StatusOK {
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}

		// Check for GaugeBucketpoint and so on for the rest.. uh
		dp := []*Datapoint{}
		if b != nil {
			if err = json.Unmarshal(b, &dp); err != nil {
				return nil, err
			}
		}
		return dp, nil
	} else if r.StatusCode > 399 {
		return nil, c.parseErrorResponse(r)
	}

	return nil, nil
}

// ReadBuckets Read datapoints from the server with in buckets (aggregates)
func (c *Client) ReadBuckets(t MetricType, o ...Modifier) ([]*Bucketpoint, error) {
	o = prepend(o, c.Url("GET", TypeEndpoint(t), DataEndpoint()))

	r, err := c.Send(o...)
	if err != nil {
		return nil, err
	}

	defer r.Body.Close()

	if r.StatusCode == http.StatusOK {
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}

		// Check for GaugeBucketpoint and so on for the rest.. uh
		bp := []*Bucketpoint{}
		if b != nil {
			if err = json.Unmarshal(b, &bp); err != nil {
				return nil, err
			}
		}
		return bp, nil
	} else if r.StatusCode > 399 {
		return nil, c.parseErrorResponse(r)
	}

	return nil, nil
}

// NewHawkularClient Initialization
func NewHawkularClient(p Parameters) (*Client, error) {
	uri, err := url.Parse(p.Url)
	if err != nil {
		return nil, err
	}

	if uri.Path == "" {
		uri.Path = baseURL
	}

	u := &url.URL{
		Host:   uri.Host,
		Path:   uri.Path,
		Scheme: uri.Scheme,
		Opaque: fmt.Sprintf("//%s/%s", uri.Host, uri.Path),
	}

	c := &http.Client{
		Timeout: timeout,
	}
	if p.TLSConfig != nil {
		transport := &http.Transport{TLSClientConfig: p.TLSConfig}
		c.Transport = transport
	}

	if p.Concurrency < 1 {
		p.Concurrency = 1
	}

	client := &Client{
		url:    u,
		Tenant: p.Tenant,
		Token:  p.Token,
		client: c,
		pool:   make(chan *poolRequest, p.Concurrency),
	}

	for i := 0; i < p.Concurrency; i++ {
		go client.sendRoutine()
	}

	return client, nil
}

// Close Safely close the Hawkular-Metrics client and flush remaining work
func (c *Client) Close() {
	close(c.pool)
}

// HTTP Helper functions

func (c *Client) parseErrorResponse(resp *http.Response) error {
	// Parse error messages here correctly..
	reply, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &HawkularClientError{Code: resp.StatusCode,
			msg: fmt.Sprintf("Reply could not be read: %s", err.Error()),
		}
	}

	details := &HawkularError{}

	err = json.Unmarshal(reply, details)
	if err != nil {
		return &HawkularClientError{Code: resp.StatusCode,
			msg: fmt.Sprintf("Reply could not be parsed: %s", err.Error()),
		}
	}

	return &HawkularClientError{Code: resp.StatusCode,
		msg: details.ErrorMsg,
	}
}

// Endpoint URL functions (...)

func (c *Client) createURL(e ...Endpoint) *url.URL {
	mu := *c.url
	for _, f := range e {
		f(&mu)
	}
	return &mu
}

// TypeEndpoint URL endpoint setting metricType
func TypeEndpoint(t MetricType) Endpoint {
	return func(u *url.URL) {
		addToURL(u, t.String())
	}
}

// SingleMetricEndpoint URL endpoint for requesting single metricID
func SingleMetricEndpoint(id string) Endpoint {
	return func(u *url.URL) {
		addToURL(u, url.QueryEscape(id))
	}
}

// TagEndpoint URL endpoint to check tags information
func TagEndpoint() Endpoint {
	return func(u *url.URL) {
		addToURL(u, "tags")
	}
}

// TagsEndpoint URL endpoint which adds tags query
func TagsEndpoint(tags map[string]string) Endpoint {
	return func(u *url.URL) {
		addToURL(u, tagsEncoder(tags))
	}
}

// DataEndpoint URL endpoint for inserting / requesting datapoints
func DataEndpoint() Endpoint {
	return func(u *url.URL) {
		addToURL(u, "data")
	}
}

func addToURL(u *url.URL, s string) *url.URL {
	u.Opaque = fmt.Sprintf("%s/%s", u.Opaque, s)
	return u
}

func tagsEncoder(t map[string]string) string {
	tags := make([]string, 0, len(t))
	for k, v := range t {
		tags = append(tags, fmt.Sprintf("%s:%s", k, v))
	}
	j := strings.Join(tags, ",")
	return j
}
