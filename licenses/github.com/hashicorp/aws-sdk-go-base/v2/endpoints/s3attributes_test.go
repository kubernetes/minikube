// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package awsbase

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go/middleware"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"go.opentelemetry.io/otel/attribute"
)

func TestS3AttributesAbortMultipartUploadInput(t *testing.T) {
	input := middleware.InitializeInput{
		Parameters: &s3.AbortMultipartUploadInput{
			Bucket:   aws.String("test-bucket"),
			Key:      aws.String("test-key"),
			UploadId: aws.String("abcd"),
		},
	}
	var output middleware.InitializeOutput

	attributes := s3AttributeBuilder(context.TODO(), input, output)

	assertAttributesMatch(t, attributes,
		[]attribute.KeyValue{
			attribute.String("aws.s3.bucket", "test-bucket"),
			attribute.String("aws.s3.key", "test-key"),
			attribute.String("aws.s3.upload_id", "abcd"),
		},
	)
}

func TestS3AttributesCompleteMultipartUploadInput(t *testing.T) {
	input := middleware.InitializeInput{
		Parameters: &s3.CompleteMultipartUploadInput{
			Bucket:   aws.String("test-bucket"),
			Key:      aws.String("test-key"),
			UploadId: aws.String("abcd"),
		},
	}
	var output middleware.InitializeOutput

	attributes := s3AttributeBuilder(context.TODO(), input, output)

	assertAttributesMatch(t, attributes,
		[]attribute.KeyValue{
			attribute.String("aws.s3.bucket", "test-bucket"),
			attribute.String("aws.s3.key", "test-key"),
			attribute.String("aws.s3.upload_id", "abcd"),
		},
	)
}

func TestS3AttributesCreateBucketInput(t *testing.T) {
	input := middleware.InitializeInput{
		Parameters: &s3.CreateBucketInput{
			Bucket: aws.String("test-bucket"),
		},
	}
	var output middleware.InitializeOutput

	attributes := s3AttributeBuilder(context.TODO(), input, output)

	assertAttributesMatch(t, attributes,
		[]attribute.KeyValue{
			attribute.String("aws.s3.bucket", "test-bucket"),
		},
	)
}

func TestS3AttributesCreateMultipartUploadInput(t *testing.T) {
	input := middleware.InitializeInput{
		Parameters: &s3.CreateMultipartUploadInput{
			Bucket: aws.String("test-bucket"),
			Key:    aws.String("test-key"),
		},
	}
	var output middleware.InitializeOutput

	attributes := s3AttributeBuilder(context.TODO(), input, output)

	assertAttributesMatch(t, attributes,
		[]attribute.KeyValue{
			attribute.String("aws.s3.bucket", "test-bucket"),
			attribute.String("aws.s3.key", "test-key"),
		},
	)
}

func TestS3AttributesDeleteBucketInput(t *testing.T) {
	input := middleware.InitializeInput{
		Parameters: &s3.DeleteBucketInput{
			Bucket: aws.String("test-bucket"),
		},
	}
	var output middleware.InitializeOutput

	attributes := s3AttributeBuilder(context.TODO(), input, output)

	assertAttributesMatch(t, attributes,
		[]attribute.KeyValue{
			attribute.String("aws.s3.bucket", "test-bucket"),
		},
	)
}

func TestS3AttributesDeleteObjectInput(t *testing.T) {
	input := middleware.InitializeInput{
		Parameters: &s3.DeleteObjectInput{
			Bucket: aws.String("test-bucket"),
			Key:    aws.String("test-key"),
		},
	}
	var output middleware.InitializeOutput

	attributes := s3AttributeBuilder(context.TODO(), input, output)

	assertAttributesMatch(t, attributes,
		[]attribute.KeyValue{
			attribute.String("aws.s3.bucket", "test-bucket"),
			attribute.String("aws.s3.key", "test-key"),
		},
	)
}

func TestS3AttributesDeleteObjectsInput(t *testing.T) {
	input := middleware.InitializeInput{
		Parameters: &s3.DeleteObjectsInput{
			Bucket: aws.String("test-bucket"),
			Delete: &s3types.Delete{
				Objects: []s3types.ObjectIdentifier{
					{
						Key:       aws.String("test-key"),
						VersionId: nil,
					},
				},
				Quiet: aws.Bool(false),
			},
		},
	}
	var output middleware.InitializeOutput

	attributes := s3AttributeBuilder(context.TODO(), input, output)

	assertAttributesMatch(t, attributes,
		[]attribute.KeyValue{
			attribute.String("aws.s3.bucket", "test-bucket"),
			attribute.String("aws.s3.delete", "Objects=[{Key=test-key}],Quiet=false"),
		},
	)
}

func TestS3AttributesGetObjectInput(t *testing.T) {
	input := middleware.InitializeInput{
		Parameters: &s3.GetObjectInput{
			Bucket: aws.String("test-bucket"),
			Key:    aws.String("test-key"),
		},
	}
	var output middleware.InitializeOutput

	attributes := s3AttributeBuilder(context.TODO(), input, output)

	assertAttributesMatch(t, attributes,
		[]attribute.KeyValue{
			attribute.String("aws.s3.bucket", "test-bucket"),
			attribute.String("aws.s3.key", "test-key"),
		},
	)
}

func TestS3AttributesHeadBucketInput(t *testing.T) {
	input := middleware.InitializeInput{
		Parameters: &s3.HeadBucketInput{
			Bucket: aws.String("test-bucket"),
		},
	}
	var output middleware.InitializeOutput

	attributes := s3AttributeBuilder(context.TODO(), input, output)

	assertAttributesMatch(t, attributes,
		[]attribute.KeyValue{
			attribute.String("aws.s3.bucket", "test-bucket"),
		},
	)
}

func TestS3AttributesHeadObjectInput(t *testing.T) {
	input := middleware.InitializeInput{
		Parameters: &s3.HeadObjectInput{
			Bucket: aws.String("test-bucket"),
			Key:    aws.String("test-key"),
		},
	}
	var output middleware.InitializeOutput

	attributes := s3AttributeBuilder(context.TODO(), input, output)

	assertAttributesMatch(t, attributes,
		[]attribute.KeyValue{
			attribute.String("aws.s3.bucket", "test-bucket"),
			attribute.String("aws.s3.key", "test-key"),
		},
	)
}

func TestS3AttributesListBucketsInput(t *testing.T) {
	input := middleware.InitializeInput{
		Parameters: &s3.ListBucketsInput{},
	}
	var output middleware.InitializeOutput

	attributes := s3AttributeBuilder(context.TODO(), input, output)

	assertAttributesMatch(t, attributes,
		[]attribute.KeyValue{},
	)
}

func TestS3AttributesListObjectsInput(t *testing.T) {
	input := middleware.InitializeInput{
		Parameters: &s3.ListObjectsInput{
			Bucket: aws.String("test-bucket"),
		},
	}
	var output middleware.InitializeOutput

	attributes := s3AttributeBuilder(context.TODO(), input, output)

	assertAttributesMatch(t, attributes,
		[]attribute.KeyValue{
			attribute.String("aws.s3.bucket", "test-bucket"),
		},
	)
}

func TestS3AttributesListObjectsV2Input(t *testing.T) {
	input := middleware.InitializeInput{
		Parameters: &s3.ListObjectsV2Input{
			Bucket: aws.String("test-bucket"),
		},
	}
	var output middleware.InitializeOutput

	attributes := s3AttributeBuilder(context.TODO(), input, output)

	assertAttributesMatch(t, attributes,
		[]attribute.KeyValue{
			attribute.String("aws.s3.bucket", "test-bucket"),
		},
	)
}

func TestS3AttributesPutObjectInput(t *testing.T) {
	input := middleware.InitializeInput{
		Parameters: &s3.PutObjectInput{
			Bucket: aws.String("test-bucket"),
			Key:    aws.String("test-key"),
		},
	}
	var output middleware.InitializeOutput

	attributes := s3AttributeBuilder(context.TODO(), input, output)

	assertAttributesMatch(t, attributes,
		[]attribute.KeyValue{
			attribute.String("aws.s3.bucket", "test-bucket"),
			attribute.String("aws.s3.key", "test-key"),
		},
	)
}

func TestS3AttributesUploadPartInput(t *testing.T) {
	input := middleware.InitializeInput{
		Parameters: &s3.UploadPartInput{
			Bucket:     aws.String("test-bucket"),
			Key:        aws.String("test-key"),
			PartNumber: aws.Int32(1234),
			UploadId:   aws.String("abcd"),
		},
	}
	var output middleware.InitializeOutput

	attributes := s3AttributeBuilder(context.TODO(), input, output)

	assertAttributesMatch(t, attributes,
		[]attribute.KeyValue{
			attribute.String("aws.s3.bucket", "test-bucket"),
			attribute.String("aws.s3.key", "test-key"),
			attribute.Int("aws.s3.part_number", 1234),
			attribute.String("aws.s3.upload_id", "abcd"),
		},
	)
}

func TestS3AttributesSerializeDeleteShorthand(t *testing.T) {
	testcases := map[string]struct {
		input    *s3types.Delete
		expected string
	}{
		"single no version not quiet": {
			input: &s3types.Delete{
				Objects: []s3types.ObjectIdentifier{
					{
						Key:       aws.String("test-key"),
						VersionId: nil,
					},
				},
				Quiet: aws.Bool(false),
			},
			expected: "Objects=[{Key=test-key}],Quiet=false",
		},
		"single version quiet": {
			input: &s3types.Delete{
				Objects: []s3types.ObjectIdentifier{
					{
						Key:       aws.String("test-key"),
						VersionId: aws.String("abc123"),
					},
				},
				Quiet: aws.Bool(true),
			},
			expected: "Objects=[{Key=test-key,VersionId=abc123}],Quiet=true",
		},
		"multiple version quiet": {
			input: &s3types.Delete{
				Objects: []s3types.ObjectIdentifier{
					{
						Key:       aws.String("test-key1"),
						VersionId: aws.String("abc123"),
					},
					{
						Key:       aws.String("test-key2"),
						VersionId: aws.String("xyz789"),
					},
				},
				Quiet: aws.Bool(true),
			},
			expected: "Objects=[{Key=test-key1,VersionId=abc123},{Key=test-key2,VersionId=xyz789}],Quiet=true",
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			out := serializeDeleteShorthand(testcase.input)

			if a, e := out, testcase.expected; a != e {
				t.Fatalf("expected %q, got %q", e, a)
			}
		})
	}
}

func assertAttributesMatch(t *testing.T, x, y []attribute.KeyValue) {
	t.Helper()

	if diff := cmp.Diff(x, y,
		cmpopts.SortSlices(less[attribute.Key]),
		cmp.Comparer(compareAttributeKeyValues),
	); diff != "" {
		t.Fatalf("unexpected credentials: (- got, + expected)\n%s", diff)
	}
}

func less[T ~string](x, y T) bool {
	return x < y
}

func compareAttributeKeyValues(x, y attribute.KeyValue) bool {
	if x.Key != y.Key {
		return false
	}
	if x.Value.Type() != y.Value.Type() {
		return false
	}
	return x.Value.AsInterface() == y.Value.AsInterface()
}
