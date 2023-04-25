package egoscale

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"
)

// Get populates the given resource or fails
func (client *Client) Get(g Gettable) error {
	ctx, cancel := context.WithTimeout(context.Background(), client.Timeout)
	defer cancel()

	return client.GetWithContext(ctx, g)
}

// GetWithContext populates the given resource or fails
func (client *Client) GetWithContext(ctx context.Context, g Gettable) error {
	return g.Get(ctx, client)
}

// Delete removes the given resource of fails
func (client *Client) Delete(g Deletable) error {
	ctx, cancel := context.WithTimeout(context.Background(), client.Timeout)
	defer cancel()

	return client.DeleteWithContext(ctx, g)
}

// DeleteWithContext removes the given resource of fails
func (client *Client) DeleteWithContext(ctx context.Context, g Deletable) error {
	return g.Delete(ctx, client)
}

// List lists the given resource (and paginate till the end)
func (client *Client) List(g Listable) ([]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), client.Timeout)
	defer cancel()

	return client.ListWithContext(ctx, g)
}

// ListWithContext lists the given resources (and paginate till the end)
func (client *Client) ListWithContext(ctx context.Context, g Listable) ([]interface{}, error) {
	s := make([]interface{}, 0)

	req, err := g.ListRequest()
	if err != nil {
		return s, err
	}

	client.PaginateWithContext(ctx, req, func(item interface{}, e error) bool {
		if item != nil {
			s = append(s, item)
			return true
		}
		err = e
		return false
	})

	return s, err
}

// AsyncListWithContext lists the given resources (and paginate till the end)
//
//
//	// NB: goroutine may leak if not read until the end. Create a proper context!
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	outChan, errChan := client.AsyncListWithContext(ctx, new(egoscale.VirtualMachine))
//
//	for {
//		select {
//		case i, ok := <- outChan:
//			if ok {
//				vm := i.(egoscale.VirtualMachine)
//				// ...
//			} else {
//				outChan = nil
//			}
//		case err, ok := <- errChan:
//			if ok {
//				// do something
//			}
//			// Once an error has been received, you can expect the channels to be closed.
//			errChan = nil
//		}
//		if errChan == nil && outChan == nil {
//			break
//		}
//	}
//
func (client *Client) AsyncListWithContext(ctx context.Context, g Listable) (<-chan interface{}, <-chan error) {
	outChan := make(chan interface{}, client.PageSize)
	errChan := make(chan error)

	go func() {
		defer close(outChan)
		defer close(errChan)

		req, err := g.ListRequest()
		if err != nil {
			errChan <- err
			return
		}

		client.PaginateWithContext(ctx, req, func(item interface{}, e error) bool {
			if item != nil {
				outChan <- item
				return true
			}
			errChan <- e
			return false
		})
	}()

	return outChan, errChan
}

// Paginate runs the ListCommand and paginates
func (client *Client) Paginate(req ListCommand, callback IterateItemFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), client.Timeout)
	defer cancel()

	client.PaginateWithContext(ctx, req, callback)
}

// PaginateWithContext runs the ListCommand as long as the ctx is valid
func (client *Client) PaginateWithContext(ctx context.Context, req ListCommand, callback IterateItemFunc) {
	pageSize := client.PageSize

	page := 1

	for {
		req.SetPage(page)
		req.SetPageSize(pageSize)
		resp, err := client.RequestWithContext(ctx, req)
		if err != nil {
			callback(nil, err)
			break
		}

		size := 0
		didErr := false
		req.each(resp, func(element interface{}, err error) bool {
			// If the context was cancelled, kill it in flight
			if e := ctx.Err(); e != nil {
				element = nil
				err = e
			}

			if callback(element, err) {
				size++
				return true
			}

			didErr = true
			return false
		})

		if size < pageSize || didErr {
			break
		}

		page++
	}
}

// APIName returns the CloudStack name of the given command
func (client *Client) APIName(request Command) string {
	return request.name()
}

// Response returns the response structure of the given command
func (client *Client) Response(request Command) interface{} {
	switch request.(type) {
	case syncCommand:
		return (request.(syncCommand)).response()
	case AsyncCommand:
		return (request.(AsyncCommand)).asyncResponse()
	default:
		panic(fmt.Errorf("The command %s is not a proper Sync or Async command", request.name()))
	}
}

// NewClientWithTimeout creates a CloudStack API client
//
// Timeout is set to both the HTTP client and the client itself.
func NewClientWithTimeout(endpoint, apiKey, apiSecret string, timeout time.Duration) *Client {
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
		},
	}

	cs := &Client{
		HTTPClient:    client,
		Endpoint:      endpoint,
		APIKey:        apiKey,
		apiSecret:     apiSecret,
		PageSize:      50,
		Timeout:       timeout,
		RetryStrategy: FibonacciRetryStrategy,
	}

	return cs
}

// NewClient creates a CloudStack API client with default timeout (60)
func NewClient(endpoint, apiKey, apiSecret string) *Client {
	timeout := time.Duration(60 * time.Second)
	return NewClientWithTimeout(endpoint, apiKey, apiSecret, timeout)
}

// FibonacciRetryStrategy waits for an increasing amount of time following the Fibonacci sequence
func FibonacciRetryStrategy(iteration int64) time.Duration {
	var a, b, i, tmp int64
	a = 0
	b = 1
	for i = 0; i < iteration; i++ {
		tmp = a + b
		a = b
		b = tmp
	}
	return time.Duration(a) * time.Second
}
