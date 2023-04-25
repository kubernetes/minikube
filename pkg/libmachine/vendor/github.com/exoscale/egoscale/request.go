package egoscale

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Error formats a CloudStack error into a standard error
func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("API error %s %d (%d): %s", e.ErrorCode, e.ErrorCode, e.CsErrorCode, e.ErrorText)
}

// Success computes the values based on the RawMessage, either string or bool
func (e *booleanResponse) IsSuccess() (bool, error) {
	if e.Success == nil {
		return false, fmt.Errorf("Not a valid booleanResponse")
	}

	str := ""
	if err := json.Unmarshal(e.Success, &str); err != nil {
		boolean := false
		if e := json.Unmarshal(e.Success, &boolean); e != nil {
			return false, e
		}
		return boolean, nil
	}
	return str == "true", nil
}

// Error formats a CloudStack job response into a standard error
func (e *booleanResponse) Error() error {
	success, err := e.IsSuccess()

	if err != nil {
		return err
	}

	if success {
		return nil
	}

	fmt.Printf("%#v", e)
	return fmt.Errorf("API error: %s", e.DisplayText)
}

func (exo *Client) parseResponse(resp *http.Response) (json.RawMessage, error) {
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	a, err := rawValues(b)

	if a == nil {
		b, err = rawValue(b)
		if err != nil {
			return nil, err
		}
	}

	if resp.StatusCode >= 400 {
		errorResponse := new(ErrorResponse)
		if e := json.Unmarshal(b, errorResponse); e == nil && errorResponse.ErrorCode > 0 {
			return nil, errorResponse
		}
		return nil, fmt.Errorf("%d %s", resp.StatusCode, b)
	}

	return b, nil
}

// asyncRequest perform an asynchronous job with a context
func (exo *Client) asyncRequest(ctx context.Context, request AsyncCommand) (interface{}, error) {
	var err error

	res := request.asyncResponse()
	exo.AsyncRequestWithContext(ctx, request, func(j *AsyncJobResult, er error) bool {
		if er != nil {
			err = er
			return false
		}
		if j.JobStatus == Success {
			if r := j.Response(res); err != nil {
				err = r
			}
			return false
		}
		return true
	})
	return res, err
}

// syncRequest performs a sync request with a context
func (exo *Client) syncRequest(ctx context.Context, request syncCommand) (interface{}, error) {
	body, err := exo.request(ctx, request)
	if err != nil {
		return nil, err
	}

	response := request.response()
	err = json.Unmarshal(body, response)

	// booleanResponse will alway be valid...
	if err == nil {
		if br, ok := response.(*booleanResponse); ok {
			success, e := br.IsSuccess()
			if e != nil {
				return nil, e
			}
			if !success {
				err = fmt.Errorf("Not a valid booleanResponse")
			}
		}
	}

	if err != nil {
		errResponse := new(ErrorResponse)
		if e := json.Unmarshal(body, errResponse); e == nil && errResponse.ErrorCode > 0 {
			return errResponse, nil
		}
		return nil, err
	}

	return response, nil
}

// BooleanRequest performs the given boolean command
func (exo *Client) BooleanRequest(req Command) error {
	resp, err := exo.Request(req)
	if err != nil {
		return err
	}

	if b, ok := resp.(*booleanResponse); ok {
		return b.Error()
	}

	panic(fmt.Errorf("The command %s is not a proper boolean response. %#v", req.name(), resp))
}

// BooleanRequestWithContext performs the given boolean command
func (exo *Client) BooleanRequestWithContext(ctx context.Context, req Command) error {
	resp, err := exo.RequestWithContext(ctx, req)
	if err != nil {
		return err
	}

	if b, ok := resp.(*booleanResponse); ok {
		return b.Error()
	}

	panic(fmt.Errorf("The command %s is not a proper boolean response. %#v", req.name(), resp))
}

// Request performs the given command
func (exo *Client) Request(request Command) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), exo.Timeout)
	defer cancel()

	switch request.(type) {
	case syncCommand:
		return exo.syncRequest(ctx, request.(syncCommand))
	case AsyncCommand:
		return exo.asyncRequest(ctx, request.(AsyncCommand))
	default:
		panic(fmt.Errorf("The command %s is not a proper Sync or Async command", request.name()))
	}
}

// RequestWithContext preforms a request with a context
func (exo *Client) RequestWithContext(ctx context.Context, request Command) (interface{}, error) {
	switch request.(type) {
	case syncCommand:
		return exo.syncRequest(ctx, request.(syncCommand))
	case AsyncCommand:
		return exo.asyncRequest(ctx, request.(AsyncCommand))
	default:
		panic(fmt.Errorf("The command %s is not a proper Sync or Async command", request.name()))
	}
}

// AsyncRequest performs the given command
func (exo *Client) AsyncRequest(request AsyncCommand, callback WaitAsyncJobResultFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), exo.Timeout)
	defer cancel()
	exo.AsyncRequestWithContext(ctx, request, callback)
}

// AsyncRequestWithContext preforms a request with a context
func (exo *Client) AsyncRequestWithContext(ctx context.Context, request AsyncCommand, callback WaitAsyncJobResultFunc) {
	body, err := exo.request(ctx, request)
	if err != nil {
		callback(nil, err)
		return
	}

	jobResult := new(AsyncJobResult)
	if err := json.Unmarshal(body, jobResult); err != nil {
		r := new(ErrorResponse)
		if e := json.Unmarshal(body, r); e != nil && r.ErrorCode > 0 {
			if !callback(nil, r) {
				return
			}
		}
		if !callback(nil, err) {
			return
		}
	}

	// Successful response
	if jobResult.JobID == "" || jobResult.JobStatus != Pending {
		if !callback(jobResult, nil) {
			return
		}
	}

	for iteration := 0; ; iteration++ {
		time.Sleep(exo.RetryStrategy(int64(iteration)))

		req := &QueryAsyncJobResult{JobID: jobResult.JobID}
		resp, err := exo.syncRequest(ctx, req)
		if err != nil {
			if !callback(nil, err) {
				return
			}
		}

		result, ok := resp.(*QueryAsyncJobResultResponse)
		if !ok {
			if !callback(nil, fmt.Errorf("AsyncJobResult expected, got %t", resp)) {
				return
			}
		}

		res := (*AsyncJobResult)(result)

		if res.JobStatus == Failure {
			if !callback(nil, res.Error()) {
				return
			}
		} else {
			if !callback(res, nil) {
				return
			}
		}
	}
}

// Payload builds the HTTP request from the given command
func (exo *Client) Payload(request Command) (string, error) {
	params := url.Values{}
	err := prepareValues("", &params, request)
	if err != nil {
		return "", err
	}
	if hookReq, ok := request.(onBeforeHook); ok {
		hookReq.onBeforeSend(&params)
	}
	params.Set("apikey", exo.APIKey)
	params.Set("command", request.name())
	params.Set("response", "json")

	// This code is borrowed from net/url/url.go
	// The way it's encoded by net/url doesn't match
	// how CloudStack works.
	var buf bytes.Buffer
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	for _, k := range keys {
		prefix := csEncode(k) + "="
		for _, v := range params[k] {
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(prefix)
			buf.WriteString(csEncode(v))
		}
	}

	return buf.String(), nil
}

// Sign signs the HTTP request and return it
func (exo *Client) Sign(query string) string {
	mac := hmac.New(sha1.New, []byte(exo.apiSecret))
	mac.Write([]byte(strings.ToLower(query)))
	signature := csEncode(base64.StdEncoding.EncodeToString(mac.Sum(nil)))

	return fmt.Sprintf("%s&signature=%s", csQuotePlus(query), signature)
}

// request makes a Request while being close to the metal
func (exo *Client) request(ctx context.Context, req Command) (json.RawMessage, error) {
	payload, err := exo.Payload(req)
	if err != nil {
		return nil, err
	}
	query := exo.Sign(payload)

	method := "GET"
	url := fmt.Sprintf("%s?%s", exo.Endpoint, query)

	var body io.Reader
	// respect Internet Explorer limit of 2048
	if len(url) > 1<<11 {
		url = exo.Endpoint
		method = "POST"
		body = strings.NewReader(query)
	}

	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	request = request.WithContext(ctx)
	request.Header.Add("User-Agent", fmt.Sprintf("exoscale/egoscale (%v)", Version))

	if method == "POST" {
		request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		request.Header.Add("Content-Length", strconv.Itoa(len(query)))
	}

	resp, err := exo.HTTPClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	text, err := exo.parseResponse(resp)
	if err != nil {
		return nil, err
	}

	return text, nil
}
