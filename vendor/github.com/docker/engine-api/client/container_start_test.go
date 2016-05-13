package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"golang.org/x/net/context"
)

func TestContainerStartError(t *testing.T) {
	client := &Client{
		transport: newMockClient(nil, errorMock(http.StatusInternalServerError, "Server error")),
	}
	err := client.ContainerStart(context.Background(), "nothing")
	if err == nil || err.Error() != "Error response from daemon: Server error" {
		t.Fatalf("expected a Server Error, got %v", err)
	}
}

func TestContainerStart(t *testing.T) {
	client := &Client{
		transport: newMockClient(nil, func(req *http.Request) (*http.Response, error) {
			// we're not expecting any payload, but if one is supplied, check it is valid.
			if req.Header.Get("Content-Type") == "application/json" {
				var startConfig interface{}
				if err := json.NewDecoder(req.Body).Decode(&startConfig); err != nil {
					return nil, fmt.Errorf("Unable to parse json: %s", err)
				}
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
			}, nil
		}),
	}

	err := client.ContainerStart(context.Background(), "container_id")
	if err != nil {
		t.Fatal(err)
	}
}
