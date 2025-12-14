// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logging

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/textproto"
	"regexp"
	"strconv"
	"strings"

	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/hashicorp/aws-sdk-go-base/v2/internal/slices"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/semconv/v1.17.0/httpconv"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"golang.org/x/text/message"
)

const (
	maxRequestBodyLen = 1024

	MaxResponseBodyLen = 4096
)

func DecomposeHTTPRequest(ctx context.Context, req *http.Request) (map[string]any, error) {
	var attributes []attribute.KeyValue

	attributes = append(attributes, httpconv.ClientRequest(req)...)
	// Remove empty `http.flavor`
	attributes = slices.Filter(attributes, func(attr attribute.KeyValue) bool {
		return attr.Key != semconv.HTTPFlavorKey || attr.Value.Emit() != ""
	})

	attributes = append(attributes, decomposeRequestHeaders(req)...)

	bodyLogger := requestBodyLogger(ctx)
	err := bodyLogger.Log(ctx, req, &attributes)
	if err != nil {
		return nil, err
	}

	result := make(map[string]any, len(attributes))
	for _, attribute := range attributes {
		result[string(attribute.Key)] = attribute.Value.AsInterface()
	}

	return result, nil
}

func decomposeRequestHeaders(req *http.Request) []attribute.KeyValue {
	header := req.Header.Clone()

	// Handled directly from the Request
	header.Del("Content-Length")
	header.Del("User-Agent")

	results := make([]attribute.KeyValue, 0, len(header)+1)

	attempt := header.Values("Amz-Sdk-Request")
	if len(attempt) > 0 {
		if resendAttribute, ok := resendCountAttribute(attempt[0]); ok {
			results = append(results, resendAttribute)
		}
	}

	auth := header.Values("Authorization")
	if len(auth) > 0 {
		if authHeader, ok := authorizationHeaderAttribute(auth[0]); ok {
			results = append(results, authHeader)
		}
	}
	header.Del("Authorization")

	securityToken := header.Values("X-Amz-Security-Token")
	if len(securityToken) > 0 {
		results = append(results, RequestHeaderAttributeKey("X-Amz-Security-Token").String("*****"))
	}
	header.Del("X-Amz-Security-Token")

	results = append(results, httpconv.RequestHeader(header)...)

	results = cleanUpHeaderAttributes(results)

	return results
}

type RequestBodyLogger interface {
	Log(ctx context.Context, req *http.Request, attrs *[]attribute.KeyValue) error
}

type ResponseBodyLogger interface {
	Log(ctx context.Context, resp *http.Response, attrs *[]attribute.KeyValue) error
}

func requestBodyLogger(ctx context.Context) RequestBodyLogger {
	if awsmiddleware.GetServiceID(ctx) == "S3" {
		if op := awsmiddleware.GetOperationName(ctx); op == "PutObject" || op == "UploadPart" {
			return &s3ObjectRequestBodyLogger{}
		}
	}

	return &defaultRequestBodyLogger{}
}

var _ RequestBodyLogger = &defaultRequestBodyLogger{}

type defaultRequestBodyLogger struct{}

func (l *defaultRequestBodyLogger) Log(ctx context.Context, req *http.Request, attrs *[]attribute.KeyValue) error {
	reqBytes, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return err
	}

	reader := textproto.NewReader(bufio.NewReader(bytes.NewReader(reqBytes)))

	if _, err = reader.ReadLine(); err != nil {
		return err
	}

	if _, err = reader.ReadMIMEHeader(); err != nil {
		return err
	}

	body, err := ReadTruncatedBody(reader, maxRequestBodyLen)
	if err != nil {
		return err
	}

	*attrs = append(*attrs, attribute.String("http.request.body", body))

	return nil
}

var _ RequestBodyLogger = &s3ObjectRequestBodyLogger{}

type s3ObjectRequestBodyLogger struct{}

func (l *s3ObjectRequestBodyLogger) Log(ctx context.Context, req *http.Request, attrs *[]attribute.KeyValue) error {
	length := outgoingLength(req)
	contentType := req.Header.Get("Content-Type")

	body := s3BodyRedacted(length, contentType)

	*attrs = append(*attrs, attribute.String("http.request.body", body))

	return nil
}

var _ ResponseBodyLogger = &S3ObjectResponseBodyLogger{}

type S3ObjectResponseBodyLogger struct{}

func (l *S3ObjectResponseBodyLogger) Log(ctx context.Context, resp *http.Response, attrs *[]attribute.KeyValue) error {
	length := resp.ContentLength
	contentType := resp.Header.Get("Content-Type")

	body := s3BodyRedacted(length, contentType)

	*attrs = append(*attrs, attribute.String("http.response.body", body))

	return nil
}

func s3BodyRedacted(length int64, contentType string) string {
	body := fmt.Sprintf("[Redacted: %s", formatByteSize(length))

	if contentType != "" {
		body += fmt.Sprintf(", Type: %s", contentType)
	}

	body += "]"

	return body
}

const byteSizeStep = 1024.0

func formatByteSize(size int64) string {
	p := message.NewPrinter(message.MatchLanguage("en"))

	if size <= 1024*1.5 {
		return p.Sprintf("%d bytes", size)
	}

	sizef := float64(size) / byteSizeStep
	var unit string
	for _, unit = range []string{"KB", "MB", "GB"} {
		if sizef < byteSizeStep {
			break
		}
		sizef /= byteSizeStep
	}
	return p.Sprintf("%.1f %s (%d bytes)", sizef, unit, size)
}

// outgoingLength is a copy of the unexported
// (*http.Request).outgoingLength method.
func outgoingLength(req *http.Request) int64 {
	if req.Body == nil || req.Body == http.NoBody {
		return 0
	}
	if req.ContentLength != 0 {
		return req.ContentLength
	}
	return -1
}

func RequestHeaderAttributeKey(k string) attribute.Key {
	return attribute.Key(requestHeaderAttributeName(k))
}

func requestHeaderAttributeName(k string) string {
	return fmt.Sprintf("http.request.header.%s", normalizeHeaderName(k))
}

func normalizeHeaderName(k string) string {
	canonical := http.CanonicalHeaderKey(k)
	lower := strings.ToLower(canonical)
	return strings.ReplaceAll(lower, "-", "_")
}

func authorizationHeaderAttribute(v string) (attribute.KeyValue, bool) {
	parts := regexp.MustCompile(`\s+`).Split(v, 2) //nolint:mnd
	if len(parts) != 2 {                           //nolint:mnd
		return attribute.KeyValue{}, false
	}
	scheme := parts[0]
	if scheme == "" {
		return attribute.KeyValue{}, false
	}
	params := parts[1]
	if params == "" {
		return attribute.KeyValue{}, false
	}

	key := RequestHeaderAttributeKey("Authorization")
	if strings.HasPrefix(scheme, "AWS4-") {
		components := regexp.MustCompile(`,\s+`).Split(params, -1)
		var builder strings.Builder
		builder.Grow(len(params))
		for i, component := range components {
			parts := strings.SplitAfterN(component, "=", 2)
			name := parts[0]
			value := parts[1]
			if name != "SignedHeaders=" && name != "Credential=" {
				// "Signature" or an unknown field
				value = "*****"
			}
			builder.WriteString(name)
			builder.WriteString(value)
			if i < len(components)-1 {
				builder.WriteString(", ")
			}
		}
		return key.String(fmt.Sprintf("%s %s", scheme, MaskAWSSensitiveValues(builder.String()))), true
	} else {
		return key.String(fmt.Sprintf("%s %s", scheme, strings.Repeat("*", len(params)))), true
	}
}

func resendCountAttribute(v string) (kv attribute.KeyValue, ok bool) {
	re := regexp.MustCompile(`attempt=(\d+);`)
	match := re.FindStringSubmatch(v)
	if len(match) != 2 { //nolint:mnd
		return
	}

	attempt, err := strconv.Atoi(match[1])
	if err != nil {
		return
	}

	if attempt > 1 {
		return attribute.Int("http.resend_count", attempt), true
	}

	return
}

func DecomposeResponseHeaders(resp *http.Response) []attribute.KeyValue {
	header := resp.Header.Clone()

	// Handled directly from the Response
	header.Del("Content-Length")

	results := make([]attribute.KeyValue, 0, len(header))

	results = append(results, httpconv.ResponseHeader(header)...)

	results = cleanUpHeaderAttributes(results)

	return results
}

func ResponseHeaderAttributeKey(k string) attribute.Key {
	return attribute.Key(responseHeaderAttributeName(k))
}

func responseHeaderAttributeName(k string) string {
	return fmt.Sprintf("http.response.header.%s", normalizeHeaderName(k))
}

// cleanUpHeaderAttributes converts header attributes with a single element to a string
func cleanUpHeaderAttributes(attrs []attribute.KeyValue) []attribute.KeyValue {
	return slices.ApplyToAll(attrs, func(attr attribute.KeyValue) attribute.KeyValue {
		if l := attr.Value.AsStringSlice(); len(l) == 1 {
			return attr.Key.String(l[0])
		}
		return attr
	})
}

func ReadTruncatedBody(reader *textproto.Reader, len int) (string, error) {
	var builder strings.Builder
	for {
		line, err := reader.ReadLine()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", err
		}
		fmt.Fprintln(&builder, line)
		if builder.Len() >= len {
			fmt.Fprint(&builder, "[truncated...]")
			break
		}
	}

	body := builder.String()
	body = MaskAWSSensitiveValues(body)

	return body, nil
}
