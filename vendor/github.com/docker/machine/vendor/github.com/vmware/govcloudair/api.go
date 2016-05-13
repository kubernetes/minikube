/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

// Package govcloudair provides a simple binding for vCloud Air REST APIs.
package govcloudair

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/cenkalti/backoff"
	types "github.com/vmware/govcloudair/types/v56"
)

// Client provides a client to vCloud Air, values can be populated automatically using the Authenticate method.
type Client struct {
	VAToken       string      // vCloud Air authorization token
	VAEndpoint    url.URL     // vCloud Air API endpoint
	Region        string      // Region where the compute resource lives.
	VCDToken      string      // Access Token (authorization header)
	VCDAuthHeader string      // Authorization header
	VCDVDCHREF    url.URL     // HREF of the backend VDC you're using
	Http          http.Client // HttpClient is the client to use. Default will be used if not provided.
}

// VCHS API

type services struct {
	Service []struct {
		Region      string `xml:"region,attr"`
		ServiceID   string `xml:"serviceId,attr"`
		ServiceType string `xml:"serviceType,attr"`
		Type        string `xml:"type,attr"`
		HREF        string `xml:"href,attr"`
	} `xml:"Service"`
}

type session struct {
	Link []*types.Link `xml:"Link"`
}

type computeResources struct {
	VdcRef []struct {
		Status string        `xml:"status,attr"`
		Name   string        `xml:"name,attr"`
		Type   string        `xml:"type,attr"`
		HREF   string        `xml:"href,attr"`
		Link   []*types.Link `xml:"Link"`
	} `xml:"VdcRef"`
}

type vCloudSession struct {
	VdcLink []struct {
		AuthorizationToken  string `xml:"authorizationToken,attr"`
		AuthorizationHeader string `xml:"authorizationHeader,attr"`
		Name                string `xml:"name,attr"`
		HREF                string `xml:"href,attr"`
	} `xml:"VdcLink"`
}

//

func (c *Client) vaauthorize(user, pass string) (u url.URL, err error) {

	if user == "" {
		user = os.Getenv("VCLOUDAIR_USERNAME")
	}

	if pass == "" {
		pass = os.Getenv("VCLOUDAIR_PASSWORD")
	}

	s := c.VAEndpoint
	s.Path += "/vchs/sessions"

	// No point in checking for errors here
	req := c.NewRequest(map[string]string{}, "POST", s, nil)

	// Set Basic Authentication Header
	req.SetBasicAuth(user, pass)

	// Add the Accept header for vCA
	req.Header.Add("Accept", "application/xml;version=5.6")

	resp, err := checkResp(c.Http.Do(req))
	if err != nil {
		return url.URL{}, err
	}
	defer resp.Body.Close()

	// Store the authentication header
	c.VAToken = resp.Header.Get("X-Vchs-Authorization")

	session := new(session)

	if err = decodeBody(resp, session); err != nil {
		return url.URL{}, fmt.Errorf("error decoding session response: %s", err)
	}

	// Loop in the session struct to find right service and compute resource.
	for _, s := range session.Link {
		if s.Type == "application/xml;class=vnd.vmware.vchs.servicelist" && s.Rel == "down" {
			u, err := url.ParseRequestURI(s.HREF)
			return *u, err
		}
	}
	return url.URL{}, fmt.Errorf("couldn't find a Service List in current session")
}

func (c *Client) vaacquireservice(s url.URL, cid string) (u url.URL, err error) {

	if cid == "" {
		cid = os.Getenv("VCLOUDAIR_COMPUTEID")
	}

	req := c.NewRequest(map[string]string{}, "GET", s, nil)

	// Add the Accept header for vCA
	req.Header.Add("Accept", "application/xml;version=5.6")

	// Set Authorization Header for vCA
	req.Header.Add("x-vchs-authorization", c.VAToken)

	resp, err := checkResp(c.Http.Do(req))
	if err != nil {
		return url.URL{}, fmt.Errorf("error processing compute action: %s", err)
	}

	services := new(services)

	if err = decodeBody(resp, services); err != nil {
		return url.URL{}, fmt.Errorf("error decoding services response: %s", err)
	}

	// Loop in the Services struct to find right service and compute resource.
	for _, s := range services.Service {
		if s.ServiceID == cid {
			c.Region = s.Region
			u, err := url.ParseRequestURI(s.HREF)
			return *u, err
		}
	}
	return url.URL{}, fmt.Errorf("couldn't find a Compute Resource in current service list")
}

func (c *Client) vaacquirecompute(s url.URL, vid string) (u url.URL, err error) {

	if vid == "" {
		vid = os.Getenv("VCLOUDAIR_VDCID")
	}

	req := c.NewRequest(map[string]string{}, "GET", s, nil)

	// Add the Accept header for vCA
	req.Header.Add("Accept", "application/xml;version=5.6")

	// Set Authorization Header
	req.Header.Add("x-vchs-authorization", c.VAToken)

	resp, err := checkResp(c.Http.Do(req))
	if err != nil {
		return url.URL{}, fmt.Errorf("error processing compute action: %s", err)
	}

	computeresources := new(computeResources)

	if err = decodeBody(resp, computeresources); err != nil {
		return url.URL{}, fmt.Errorf("error decoding computeresources response: %s", err)
	}

	// Iterate through the ComputeResources struct searching for the right
	// backend server
	for _, s := range computeresources.VdcRef {
		if s.Name == vid {
			for _, t := range s.Link {
				if t.Name == vid {
					u, err := url.ParseRequestURI(t.HREF)
					return *u, err
				}
			}
		}
	}
	return url.URL{}, fmt.Errorf("couldn't find a VDC Resource in current Compute list")
}

func (c *Client) vagetbackendauth(s url.URL, cid string) error {

	if cid == "" {
		cid = os.Getenv("VCLOUDAIR_COMPUTEID")
	}

	req := c.NewRequest(map[string]string{}, "POST", s, nil)

	// Add the Accept header for vCA
	req.Header.Add("Accept", "application/xml;version=5.6")

	// Set Authorization Header
	req.Header.Add("x-vchs-authorization", c.VAToken)

	// Adding exponential backoff to retry
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = time.Duration(30 * time.Second)

	ticker := backoff.NewTicker(b)

	var err error
	var resp *http.Response

	for t := range ticker.C {
		resp, err = checkResp(c.Http.Do(req))
		if err != nil {
			fmt.Println(err, "retrying...", t)
			continue
		}
		ticker.Stop()
		break

	}

	if err != nil {
		return fmt.Errorf("error processing backend url action: %s", err)
	}

	defer resp.Body.Close()

	vcloudsession := new(vCloudSession)

	if err = decodeBody(resp, vcloudsession); err != nil {
		return fmt.Errorf("error decoding vcloudsession response: %s", err)
	}

	// Get the backend session information
	for _, s := range vcloudsession.VdcLink {
		if s.Name == cid {
			// Fetch the authorization token
			c.VCDToken = s.AuthorizationToken

			// Fetch the authorization header
			c.VCDAuthHeader = s.AuthorizationHeader

			u, err := url.ParseRequestURI(s.HREF)
			if err != nil {
				return fmt.Errorf("error decoding href: %s", err)
			}
			c.VCDVDCHREF = *u
			return nil
		}
	}
	return fmt.Errorf("error finding the right backend resource")
}

// NewClient returns a new empty client to authenticate against the vCloud Air
// service, the vCloud Air endpoint can be overridden by setting the
// VCLOUDAIR_ENDPOINT environment variable.
func NewClient() (*Client, error) {

	var u *url.URL
	var err error

	if os.Getenv("VCLOUDAIR_ENDPOINT") != "" {
		u, err = url.ParseRequestURI(os.Getenv("VCLOUDAIR_ENDPOINT"))
		if err != nil {
			return &Client{}, fmt.Errorf("cannot parse endpoint coming from VCLOUDAIR_ENDPOINT")
		}
	} else {
		// Implicitly trust this URL parse.
		u, _ = url.ParseRequestURI("https://vchs.vmware.com/api")
	}

	Client := Client{
		VAEndpoint: *u,
		// Patching things up as we're hitting several TLS timeouts.
		Http: http.Client{Transport: &http.Transport{TLSHandshakeTimeout: 120 * time.Second}},
	}
	return &Client, nil
}

// Authenticate is an helper function that performs a complete login in vCloud
// Air and in the backend vCloud Director instance.
func (c *Client) Authenticate(username, password, computeid, vdcid string) (Vdc, error) {
	// Authorize
	vaservicehref, err := c.vaauthorize(username, password)
	if err != nil {
		return Vdc{}, fmt.Errorf("error Authorizing: %s", err)
	}

	// Get Service
	vacomputehref, err := c.vaacquireservice(vaservicehref, computeid)
	if err != nil {
		return Vdc{}, fmt.Errorf("error Acquiring Service: %s", err)
	}

	// Get Compute
	vavdchref, err := c.vaacquirecompute(vacomputehref, vdcid)
	if err != nil {
		return Vdc{}, fmt.Errorf("error Acquiring Compute: %s", err)
	}

	// Get Backend Authorization
	if err = c.vagetbackendauth(vavdchref, computeid); err != nil {
		return Vdc{}, fmt.Errorf("error Acquiring Backend Authorization: %s", err)
	}

	v, err := c.retrieveVDC()
	if err != nil {
		return Vdc{}, fmt.Errorf("error Acquiring VDC: %s", err)
	}

	return v, nil

}

// NewRequest creates a new HTTP request and applies necessary auth headers if
// set.
func (c *Client) NewRequest(params map[string]string, method string, u url.URL, body io.Reader) *http.Request {

	p := url.Values{}

	// Build up our request parameters
	for k, v := range params {
		p.Add(k, v)
	}

	// Add the params to our URL
	u.RawQuery = p.Encode()

	// Build the request, no point in checking for errors here as we're just
	// passing a string version of an url.URL struct and http.NewRequest returns
	// error only if can't process an url.ParseRequestURI().
	req, _ := http.NewRequest(method, u.String(), body)

	if c.VCDAuthHeader != "" && c.VCDToken != "" {
		// Add the authorization header
		req.Header.Add(c.VCDAuthHeader, c.VCDToken)
		// Add the Accept header for VCD
		req.Header.Add("Accept", "application/*+xml;version=5.6")
	}

	return req

}

// Disconnect performs a disconnection from the vCloud Air API endpoint.
func (c *Client) Disconnect() error {
	if c.VCDToken == "" && c.VCDAuthHeader == "" && c.VAToken == "" {
		return fmt.Errorf("cannot disconnect, client is not authenticated")
	}

	s := c.VAEndpoint
	s.Path += "/vchs/session"

	req := c.NewRequest(map[string]string{}, "DELETE", s, nil)

	// Add the Accept header for vCA
	req.Header.Add("Accept", "application/xml;version=5.6")

	// Set Authorization Header
	req.Header.Add("x-vchs-authorization", c.VAToken)

	if _, err := checkResp(c.Http.Do(req)); err != nil {
		return fmt.Errorf("error processing session delete for vchs: %s", err)
	}

	return nil
}

// parseErr takes an error XML resp and returns a single string for use in error messages.
func parseErr(resp *http.Response) error {

	errBody := new(types.Error)

	// if there was an error decoding the body, just return that
	if err := decodeBody(resp, errBody); err != nil {
		return fmt.Errorf("error parsing error body for non-200 request: %s", err)
	}

	return fmt.Errorf("API Error: %d: %s", errBody.MajorErrorCode, errBody.Message)
}

// decodeBody is used to XML decode a response body
func decodeBody(resp *http.Response, out interface{}) error {

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Unmarshal the XML.
	if err = xml.Unmarshal(body, &out); err != nil {
		return err
	}

	return nil
}

// checkResp wraps http.Client.Do() and verifies the request, if status code
// is 2XX it passes back the response, if it's a known invalid status code it
// parses the resultant XML error and returns a descriptive error, if the
// status code is not handled it returns a generic error with the status code.
func checkResp(resp *http.Response, err error) (*http.Response, error) {
	if err != nil {
		return resp, err
	}

	switch i := resp.StatusCode; {
	// Valid request, return the response.
	case i == 200 || i == 201 || i == 202 || i == 204:
		return resp, nil
	// Invalid request, parse the XML error returned and return it.
	case i == 400 || i == 401 || i == 403 || i == 404 || i == 405 || i == 406 || i == 409 || i == 415 || i == 500 || i == 503 || i == 504:
		return nil, parseErr(resp)
	// Unhandled response.
	default:
		return nil, fmt.Errorf("unhandled API response, please report this issue, status code: %s", resp.Status)
	}
}
