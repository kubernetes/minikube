// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package awsbase

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/textproto"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	smithylogging "github.com/aws/smithy-go/logging"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"github.com/hashicorp/aws-sdk-go-base/v2/logging"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/semconv/v1.17.0/httpconv"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
)

type debugLogger struct {
	ctx context.Context
}

func (l debugLogger) Logf(classification smithylogging.Classification, format string, v ...any) {
	s := fmt.Sprintf(format, v...)
	if l.ctx != nil {
		logger := logging.RetrieveLogger(l.ctx)
		switch classification {
		case smithylogging.Debug:
			logger.Debug(l.ctx, s)
		case smithylogging.Warn:
			logger.Warn(l.ctx, s)
		}
	} else {
		s = strings.ReplaceAll(s, "\r", "") // Works around https://github.com/jen20/teamcity-go-test/pull/2
		log.Printf("[%s] missing_context: %s "+string(logging.AwsSdkKey)+"="+awsSdkGoV2Val, classification, s)
	}
}

func (l debugLogger) WithContext(ctx context.Context) smithylogging.Logger {
	return &debugLogger{
		ctx: ctx,
	}
}

const awsSdkGoV2Val = "aws-sdk-go-v2"

func awsSDKv2Attr() attribute.KeyValue {
	return logging.AwsSdkKey.String(awsSdkGoV2Val)
}

type logAttributeExtractor struct{}

func (l *logAttributeExtractor) ID() string {
	return "TF_AWS_LogAttributeExtractor"
}

func (l *logAttributeExtractor) HandleInitialize(ctx context.Context, in middleware.InitializeInput, next middleware.InitializeHandler) (
	out middleware.InitializeOutput, metadata middleware.Metadata, err error) {
	logger := logging.RetrieveLogger(ctx)

	region := awsmiddleware.GetRegion(ctx)
	serviceID := awsmiddleware.GetServiceID(ctx)

	attributes := []attribute.KeyValue{
		otelaws.SystemAttr(),
		otelaws.ServiceAttr(serviceID),
		otelaws.RegionAttr(region),
		otelaws.OperationAttr(awsmiddleware.GetOperationName(ctx)),
		awsSDKv2Attr(),
	}

	setters := map[string]otelaws.AttributeBuilder{
		dynamodb.ServiceID: otelaws.DynamoDBAttributeBuilder,
		s3.ServiceID:       s3AttributeBuilder,
		sqs.ServiceID:      otelaws.SQSAttributeBuilder,
	}
	if setter, ok := setters[serviceID]; ok {
		attributes = append(attributes, setter(ctx, in, out)...)
	}

	for _, attribute := range attributes {
		ctx = logger.SetField(ctx, string(attribute.Key), attribute.Value.AsInterface())
	}

	return next.HandleInitialize(ctx, in)
}

// Replaces the built-in logging middleware from https://github.com/aws/smithy-go/blob/main/transport/http/middleware_http_logging.go
// We want access to the request and response structs, and cannot get it from the built-in.
// The typical route of adding logging to the http.RoundTripper doesn't work for the AWS SDK for Go v2 without forcing us to manually implement
// configuration that the SDK handles for us.
type requestResponseLogger struct{}

// ID is the middleware identifier.
func (r *requestResponseLogger) ID() string {
	return "TF_AWS_RequestResponseLogger"
}

func (r *requestResponseLogger) HandleDeserialize(ctx context.Context, in middleware.DeserializeInput, next middleware.DeserializeHandler,
) (
	out middleware.DeserializeOutput, metadata middleware.Metadata, err error,
) {
	logger := logging.RetrieveLogger(ctx)

	region := awsmiddleware.GetRegion(ctx)

	if signingRegion := awsmiddleware.GetSigningRegion(ctx); signingRegion != region { //nolint:staticcheck // Not retrievable elsewhere
		ctx = logger.SetField(ctx, string(logging.SigningRegionKey), signingRegion)
	}
	if awsmiddleware.GetEndpointSource(ctx) == aws.EndpointSourceCustom {
		ctx = logger.SetField(ctx, string(logging.CustomEndpointKey), true)
	}

	smithyRequest, ok := in.Request.(*smithyhttp.Request)
	if !ok {
		return out, metadata, fmt.Errorf("unknown request type %T", in.Request)
	}

	rc := smithyRequest.Build(ctx)

	requestFields, err := logging.DecomposeHTTPRequest(ctx, rc)
	if err != nil {
		return out, metadata, fmt.Errorf("decomposing request: %w", err)
	}
	logger.Debug(ctx, "HTTP Request Sent", requestFields)

	smithyRequest, err = smithyRequest.SetStream(rc.Body)
	if err != nil {
		return out, metadata, err
	}
	in.Request = smithyRequest

	start := time.Now()

	out, metadata, err = next.HandleDeserialize(ctx, in)

	elapsed := time.Since(start)

	if err == nil {
		smithyResponse, ok := out.RawResponse.(*smithyhttp.Response)
		if !ok {
			return out, metadata, fmt.Errorf("unknown response type: %T", out.RawResponse)
		}

		responseFields, err := decomposeHTTPResponse(ctx, smithyResponse.Response, elapsed)
		if err != nil {
			return out, metadata, fmt.Errorf("decomposing response: %w", err)
		}
		logger.Debug(ctx, "HTTP Response Received", responseFields)
	}

	return out, metadata, err
}

func decomposeHTTPResponse(ctx context.Context, resp *http.Response, elapsed time.Duration) (map[string]any, error) {
	var attributes []attribute.KeyValue

	attributes = append(attributes, attribute.Int64("http.duration", elapsed.Milliseconds()))

	attributes = append(attributes, httpconv.ClientResponse(resp)...)

	attributes = append(attributes, logging.DecomposeResponseHeaders(resp)...)

	bodyLogger := responseBodyLogger(ctx)
	err := bodyLogger.Log(ctx, resp, &attributes)
	if err != nil {
		return nil, err
	}

	result := make(map[string]any, len(attributes))
	for _, attribute := range attributes {
		result[string(attribute.Key)] = attribute.Value.AsInterface()
	}

	return result, nil
}

func responseBodyLogger(ctx context.Context) logging.ResponseBodyLogger {
	if awsmiddleware.GetServiceID(ctx) == "S3" {
		if op := awsmiddleware.GetOperationName(ctx); op == "GetObject" {
			return &logging.S3ObjectResponseBodyLogger{}
		}
	}

	return &defaultResponseBodyLogger{}
}

var _ logging.ResponseBodyLogger = &defaultResponseBodyLogger{}

type defaultResponseBodyLogger struct{}

func (l *defaultResponseBodyLogger) Log(ctx context.Context, resp *http.Response, attrs *[]attribute.KeyValue) error {
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Restore the body reader
	resp.Body = io.NopCloser(bytes.NewBuffer(content))

	reader := textproto.NewReader(bufio.NewReader(bytes.NewReader(content)))

	body, err := logging.ReadTruncatedBody(reader, logging.MaxResponseBodyLen)
	if err != nil {
		return err
	}

	*attrs = append(*attrs, attribute.String("http.response.body", body))

	return nil
}

// May be contributed to go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws
// See: https://github.com/open-telemetry/opentelemetry-go-contrib/issues/4321
func s3AttributeBuilder(ctx context.Context, in middleware.InitializeInput, out middleware.InitializeOutput) []attribute.KeyValue {
	s3Attributes := []attribute.KeyValue{}

	switch v := in.Parameters.(type) {
	case *s3.AbortMultipartUploadInput:
		s3Attributes = append(s3Attributes, semconv.AWSS3Bucket(aws.ToString(v.Bucket)))

		s3Attributes = append(s3Attributes, semconv.AWSS3Key(aws.ToString(v.Key)))

		s3Attributes = append(s3Attributes, semconv.AWSS3UploadID(aws.ToString(v.UploadId)))

	case *s3.CompleteMultipartUploadInput:
		s3Attributes = append(s3Attributes, semconv.AWSS3Bucket(aws.ToString(v.Bucket)))

		s3Attributes = append(s3Attributes, semconv.AWSS3Key(aws.ToString(v.Key)))

		s3Attributes = append(s3Attributes, semconv.AWSS3UploadID(aws.ToString(v.UploadId)))

	case *s3.CreateBucketInput:
		s3Attributes = append(s3Attributes, semconv.AWSS3Bucket(aws.ToString(v.Bucket)))

	case *s3.CreateMultipartUploadInput:
		s3Attributes = append(s3Attributes, semconv.AWSS3Bucket(aws.ToString(v.Bucket)))

		s3Attributes = append(s3Attributes, semconv.AWSS3Key(aws.ToString(v.Key)))

	case *s3.DeleteBucketInput:
		s3Attributes = append(s3Attributes, semconv.AWSS3Bucket(aws.ToString(v.Bucket)))

	case *s3.DeleteObjectInput:
		s3Attributes = append(s3Attributes, semconv.AWSS3Bucket(aws.ToString(v.Bucket)))

		s3Attributes = append(s3Attributes, semconv.AWSS3Key(aws.ToString(v.Key)))

	case *s3.DeleteObjectsInput:
		s3Attributes = append(s3Attributes, semconv.AWSS3Bucket(aws.ToString(v.Bucket)))

		s3Attributes = append(s3Attributes, semconv.AWSS3Delete(serializeDeleteShorthand(v.Delete)))

	case *s3.GetObjectInput:
		s3Attributes = append(s3Attributes, semconv.AWSS3Bucket(aws.ToString(v.Bucket)))

		s3Attributes = append(s3Attributes, semconv.AWSS3Key(aws.ToString(v.Key)))

	case *s3.HeadBucketInput:
		s3Attributes = append(s3Attributes, semconv.AWSS3Bucket(aws.ToString(v.Bucket)))

	case *s3.HeadObjectInput:
		s3Attributes = append(s3Attributes, semconv.AWSS3Bucket(aws.ToString(v.Bucket)))

		s3Attributes = append(s3Attributes, semconv.AWSS3Key(aws.ToString(v.Key)))

	case *s3.ListBucketsInput:
		// ListBucketsInput defines no attributes

	case *s3.ListObjectsInput:
		s3Attributes = append(s3Attributes, semconv.AWSS3Bucket(aws.ToString(v.Bucket)))

	case *s3.ListObjectsV2Input:
		s3Attributes = append(s3Attributes, semconv.AWSS3Bucket(aws.ToString(v.Bucket)))

	case *s3.PutObjectInput:
		s3Attributes = append(s3Attributes, semconv.AWSS3Bucket(aws.ToString(v.Bucket)))

		s3Attributes = append(s3Attributes, semconv.AWSS3Key(aws.ToString(v.Key)))

	case *s3.UploadPartInput:
		s3Attributes = append(s3Attributes, semconv.AWSS3Bucket(aws.ToString(v.Bucket)))

		s3Attributes = append(s3Attributes, semconv.AWSS3Key(aws.ToString(v.Key)))

		s3Attributes = append(s3Attributes, semconv.AWSS3PartNumber(int(aws.ToInt32(v.PartNumber))))

		s3Attributes = append(s3Attributes, semconv.AWSS3UploadID(aws.ToString(v.UploadId)))
	}

	return s3Attributes
}

func serializeDeleteShorthand(d *s3types.Delete) string {
	var builder strings.Builder

	fmt.Fprint(&builder, "Objects=[")
	count := len(d.Objects)
	for i, object := range d.Objects {
		fmt.Fprint(&builder, "{")

		fmt.Fprintf(&builder, "Key=%s", aws.ToString(object.Key))

		if object.VersionId != nil {
			fmt.Fprintf(&builder, ",VersionId=%s", aws.ToString(object.VersionId))
		}

		fmt.Fprint(&builder, "}")
		if i+1 != count {
			fmt.Fprint(&builder, ",")
		}
	}
	fmt.Fprint(&builder, "],")

	fmt.Fprintf(&builder, "Quiet=%t", aws.ToBool(d.Quiet))

	return builder.String()
}
