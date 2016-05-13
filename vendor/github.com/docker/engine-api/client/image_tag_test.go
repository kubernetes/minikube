package client

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"golang.org/x/net/context"

	"github.com/docker/engine-api/types"
)

func TestImageTagError(t *testing.T) {
	client := &Client{
		transport: newMockClient(nil, errorMock(http.StatusInternalServerError, "Server error")),
	}

	err := client.ImageTag(context.Background(), "image_id", "repo:tag", types.ImageTagOptions{})
	if err == nil || err.Error() != "Error response from daemon: Server error" {
		t.Fatalf("expected a Server Error, got %v", err)
	}
}

// Note: this is not testing all the InvalidReference as it's the reponsability
// of distribution/reference package.
func TestImageTagInvalidReference(t *testing.T) {
	client := &Client{
		transport: newMockClient(nil, errorMock(http.StatusInternalServerError, "Server error")),
	}

	err := client.ImageTag(context.Background(), "image_id", "aa/asdf$$^/aa", types.ImageTagOptions{})
	if err == nil || err.Error() != `Error parsing reference: "aa/asdf$$^/aa" is not a valid repository/tag` {
		t.Fatalf("expected ErrReferenceInvalidFormat, got %v", err)
	}
}

func TestImageTag(t *testing.T) {
	expectedURL := "/images/image_id/tag"
	tagCases := []struct {
		force               bool
		reference           string
		expectedQueryParams map[string]string
	}{
		{
			force:     false,
			reference: "repository:tag1",
			expectedQueryParams: map[string]string{
				"force": "",
				"repo":  "repository",
				"tag":   "tag1",
			},
		}, {
			force:     true,
			reference: "another_repository:latest",
			expectedQueryParams: map[string]string{
				"force": "1",
				"repo":  "another_repository",
				"tag":   "latest",
			},
		}, {
			force:     true,
			reference: "another_repository",
			expectedQueryParams: map[string]string{
				"force": "1",
				"repo":  "another_repository",
				"tag":   "",
			},
		}, {
			force:     true,
			reference: "test/another_repository",
			expectedQueryParams: map[string]string{
				"force": "1",
				"repo":  "test/another_repository",
				"tag":   "",
			},
		}, {
			force:     true,
			reference: "test/another_repository:tag1",
			expectedQueryParams: map[string]string{
				"force": "1",
				"repo":  "test/another_repository",
				"tag":   "tag1",
			},
		}, {
			force:     true,
			reference: "test/test/another_repository:tag1",
			expectedQueryParams: map[string]string{
				"force": "1",
				"repo":  "test/test/another_repository",
				"tag":   "tag1",
			},
		}, {
			force:     true,
			reference: "test:5000/test/another_repository:tag1",
			expectedQueryParams: map[string]string{
				"force": "1",
				"repo":  "test:5000/test/another_repository",
				"tag":   "tag1",
			},
		}, {
			force:     true,
			reference: "test:5000/test/another_repository",
			expectedQueryParams: map[string]string{
				"force": "1",
				"repo":  "test:5000/test/another_repository",
				"tag":   "",
			},
		},
	}
	for _, tagCase := range tagCases {
		client := &Client{
			transport: newMockClient(nil, func(req *http.Request) (*http.Response, error) {
				if !strings.HasPrefix(req.URL.Path, expectedURL) {
					return nil, fmt.Errorf("expected URL '%s', got '%s'", expectedURL, req.URL)
				}
				if req.Method != "POST" {
					return nil, fmt.Errorf("expected POST method, got %s", req.Method)
				}
				query := req.URL.Query()
				for key, expected := range tagCase.expectedQueryParams {
					actual := query.Get(key)
					if actual != expected {
						return nil, fmt.Errorf("%s not set in URL query properly. Expected '%s', got %s", key, expected, actual)
					}
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
				}, nil
			}),
		}
		err := client.ImageTag(context.Background(), "image_id", tagCase.reference, types.ImageTagOptions{
			Force: tagCase.force,
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}
