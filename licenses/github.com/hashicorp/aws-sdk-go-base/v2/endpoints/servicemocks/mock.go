// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicemocks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"time"
)

const (
	MockEc2MetadataAccessKey    = `Ec2MetadataAccessKey`
	MockEc2MetadataSecretKey    = `Ec2MetadataSecretKey`
	MockEc2MetadataSessionToken = `Ec2MetadataSessionToken`

	MockEcsCredentialsAccessKey    = `EcsCredentialsAccessKey`
	MockEcsCredentialsSecretKey    = `EcsCredentialsSecretKey`
	MockEcsCredentialsSessionToken = `EcsCredentialsSessionToken`

	MockEnvAccessKey    = `EnvAccessKey`
	MockEnvSecretKey    = `EnvSecretKey`
	MockEnvSessionToken = `EnvSessionToken`

	MockStaticAccessKey = `StaticAccessKey`
	MockStaticSecretKey = `StaticSecretKey`

	MockStsAssumeRoleAccessKey                               = `AssumeRoleAccessKey`
	MockStsAssumeRoleArn                                     = `arn:aws:iam::555555555555:role/AssumeRole`
	MockStsAssumeRoleArn2                                    = `arn:aws:iam::555555555555:role/AssumeRole2`
	MockStsAssumeRoleExternalId                              = `AssumeRoleExternalId`
	MockStsAssumeRoleInvalidResponseBodyInvalidClientTokenId = `<ErrorResponse xmlns="https://sts.amazonaws.com/doc/2011-06-15/">
<Error>
  <Type>Sender</Type>
  <Code>InvalidClientTokenId</Code>
  <Message>The security token included in the request is invalid.</Message>
</Error>
<RequestId>4d0cf5ec-892a-4d3f-84e4-30e9987d9bdd</RequestId>
</ErrorResponse>`
	MockStsAssumeRolePolicy = `{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*",
  }
}`
	MockStsAssumeRolePolicyArn         = `arn:aws:iam::555555555555:policy/AssumeRolePolicy1`
	MockStsAssumeRoleSecretKey         = `AssumeRoleSecretKey`
	MockStsAssumeRoleSessionName       = `AssumeRoleSessionName`
	MockStsAssumeRoleSessionName2      = `AssumeRoleSessionName2`
	MockStsAssumeRoleSessionToken      = `AssumeRoleSessionToken`
	MockStsAssumeRoleSourceIdentity    = `AssumeRoleSourceIdentity`
	MockStsAssumeRoleTagKey            = `AssumeRoleTagKey`
	MockStsAssumeRoleTagValue          = `AssumeRoleTagValue`
	MockStsAssumeRoleTransitiveTagKey  = `AssumeRoleTagKey`
	MockStsAssumeRoleValidResponseBody = `<AssumeRoleResponse xmlns="https://sts.amazonaws.com/doc/2011-06-15/">
<AssumeRoleResult>
  <AssumedRoleUser>
    <Arn>arn:aws:sts::555555555555:assumed-role/role/AssumeRoleSessionName</Arn>
    <AssumedRoleId>ARO123EXAMPLE123:AssumeRoleSessionName</AssumedRoleId>
  </AssumedRoleUser>
  <Credentials>
    <AccessKeyId>AssumeRoleAccessKey</AccessKeyId>
    <SecretAccessKey>AssumeRoleSecretKey</SecretAccessKey>
    <SessionToken>AssumeRoleSessionToken</SessionToken>
    <Expiration>2099-12-31T23:59:59Z</Expiration>
  </Credentials>
</AssumeRoleResult>
<ResponseMetadata>
  <RequestId>01234567-89ab-cdef-0123-456789abcdef</RequestId>
</ResponseMetadata>
</AssumeRoleResponse>`

	MockStsAssumeRoleWithWebIdentityAccessKey         = `AssumeRoleWithWebIdentityAccessKey`
	MockStsAssumeRoleWithWebIdentityArn               = `arn:aws:iam::666666666666:role/WebIdentityToken`
	MockStsAssumeRoleWithWebIdentitySecretKey         = `AssumeRoleWithWebIdentitySecretKey`
	MockStsAssumeRoleWithWebIdentitySessionName       = `AssumeRoleWithWebIdentitySessionName`
	MockStsAssumeRoleWithWebIdentitySessionToken      = `AssumeRoleWithWebIdentitySessionToken`
	MockStsAssumeRoleWithWebIdentityValidResponseBody = `<AssumeRoleWithWebIdentityResponse xmlns="https://sts.amazonaws.com/doc/2011-06-15/">
<AssumeRoleWithWebIdentityResult>
  <SubjectFromWebIdentityToken>amzn1.account.AF6RHO7KZU5XRVQJGXK6HB56KR2A</SubjectFromWebIdentityToken>
  <Audience>client.6666666666666666666.6666@apps.example.com</Audience>
  <AssumedRoleUser>
    <Arn>arn:aws:sts::666666666666:assumed-role/FederatedWebIdentityRole/AssumeRoleWithWebIdentitySessionName</Arn>
    <AssumedRoleId>ARO123EXAMPLE123:AssumeRoleWithWebIdentitySessionName</AssumedRoleId>
  </AssumedRoleUser>
  <Credentials>
    <SessionToken>AssumeRoleWithWebIdentitySessionToken</SessionToken>
    <SecretAccessKey>AssumeRoleWithWebIdentitySecretKey</SecretAccessKey>
    <Expiration>2099-12-31T23:59:59Z</Expiration>
    <AccessKeyId>AssumeRoleWithWebIdentityAccessKey</AccessKeyId>
  </Credentials>
  <Provider>www.amazon.com</Provider>
</AssumeRoleWithWebIdentityResult>
<ResponseMetadata>
  <RequestId>01234567-89ab-cdef-0123-456789abcdef</RequestId>
</ResponseMetadata>
</AssumeRoleWithWebIdentityResponse>`

	MockStsAssumeRoleWithWebIdentityAlternateArn         = `arn:aws:iam::666666666666:role/Alternate`
	MockStsAssumeRoleWithWebIdentityAlternateSessionName = `AssumeRoleWithWebIdentityAlternateSessionName`

	MockStsGetCallerIdentityAccountID                       = `222222222222`
	MockStsGetCallerIdentityInvalidResponseBodyAccessDenied = `<ErrorResponse xmlns="https://sts.amazonaws.com/doc/2011-06-15/">
<Error>
  <Type>Sender</Type>
  <Code>AccessDenied</Code>
  <Message>User: arn:aws:iam::123456789012:user/Bob is not authorized to perform: sts:GetCallerIdentity</Message>
</Error>
<RequestId>01234567-89ab-cdef-0123-456789abcdef</RequestId>
</ErrorResponse>`
	// MockStsGetCallerIdentityValidResponseBodyExpiredToken uses code "ExpiredToken", seemingly the most common
	// code. Errors usually have an invalid body but this may be fixed at some point.
	MockStsGetCallerIdentityValidResponseBodyExpiredToken = `<ErrorResponse xmlns="https://sts.amazonaws.com/doc/2011-06-15/">
<Error>
  <Type>Sender</Type>
  <Code>ExpiredToken</Code>
  <Message>The security token included in the request is expired</Message>
</Error>
<ResponseMetadata>
  <RequestId>01234567-89ab-cdef-0123-456789abcdef</RequestId>
</ResponseMetadata>
</ErrorResponse>`
	// MockStsGetCallerIdentityInvalidResponseBodyExpiredToken uses code "ExpiredToken", seemingly the most common
	// code. Errors usually have an invalid body.
	MockStsGetCallerIdentityInvalidResponseBodyExpiredToken = `<ErrorResponse xmlns="https://sts.amazonaws.com/doc/2011-06-15/">
<Error>
  <Type>Sender</Type>
  <Code>ExpiredToken</Code>
  <Message>The security token included in the request is expired</Message>
</Error>
<RequestId>01234567-89ab-cdef-0123-456789abcdef</RequestId>
</ErrorResponse>`
	// MockStsGetCallerIdentityValidResponseBodyExpiredTokenException uses code "ExpiredTokenException", a more rare code
	// but used at least by Fargate. Errors usually have an invalid body but this may change.
	MockStsGetCallerIdentityValidResponseBodyExpiredTokenException = `<ErrorResponse xmlns="https://sts.amazonaws.com/doc/2011-06-15/">
<Error>
  <Type>Sender</Type>
  <Code>ExpiredTokenException</Code>
  <Message>The security token included in the request is expired</Message>
</Error>
<ResponseMetadata>
  <RequestId>01234567-89ab-cdef-0123-456789abcdef</RequestId>
</ResponseMetadata>
</ErrorResponse>`
	// MockStsGetCallerIdentityInvalidResponseBodyExpiredTokenException uses code "ExpiredTokenException", a more rare code
	// but used at least by Fargate. Errors usually have an invalid body but this may change.
	MockStsGetCallerIdentityInvalidResponseBodyExpiredTokenException = `<ErrorResponse xmlns="https://sts.amazonaws.com/doc/2011-06-15/">
<Error>
  <Type>Sender</Type>
  <Code>ExpiredTokenException</Code>
  <Message>The security token included in the request is expired</Message>
</Error>
<RequestId>01234567-89ab-cdef-0123-456789abcdef</RequestId>
</ErrorResponse>`
	// MockStsGetCallerIdentityValidResponseBodyRequestExpired uses code "RequestExpired", a code only used in EC2.
	// Errors usually have an invalid body but this may change.
	MockStsGetCallerIdentityValidResponseBodyRequestExpired = `<ErrorResponse xmlns="https://sts.amazonaws.com/doc/2011-06-15/">
<Error>
  <Type>Sender</Type>
  <Code>RequestExpired</Code>
  <Message>The security token included in the request is expired</Message>
</Error>
<ResponseMetadata>
  <RequestId>01234567-89ab-cdef-0123-456789abcdef</RequestId>
</ResponseMetadata>
</ErrorResponse>`
	// MockStsGetCallerIdentityInvalidResponseBodyRequestExpired uses code "RequestExpired", a code only used in EC2.
	// Errors usually have an invalid body but this may change.
	MockStsGetCallerIdentityInvalidResponseBodyRequestExpired = `<ErrorResponse xmlns="https://sts.amazonaws.com/doc/2011-06-15/">
<Error>
  <Type>Sender</Type>
  <Code>RequestExpired</Code>
  <Message>The security token included in the request is expired</Message>
</Error>
<RequestId>01234567-89ab-cdef-0123-456789abcdef</RequestId>
</ErrorResponse>`
	MockStsGetCallerIdentityPartition         = `aws`
	MockStsGetCallerIdentityValidResponseBody = `<GetCallerIdentityResponse xmlns="https://sts.amazonaws.com/doc/2011-06-15/">
  <GetCallerIdentityResult>
    <Arn>arn:aws:iam::222222222222:user/Alice</Arn>
    <UserId>AKIAI44QH8DHBEXAMPLE</UserId>
    <Account>222222222222</Account>
  </GetCallerIdentityResult>
  <ResponseMetadata>
    <RequestId>01234567-89ab-cdef-0123-456789abcdef</RequestId>
  </ResponseMetadata>
</GetCallerIdentityResponse>`
	MockStsGetCallerIdentityValidAssumedRoleResponseBody = `<GetCallerIdentityResponse xmlns="https://sts.amazonaws.com/doc/2011-06-15/">
  <GetCallerIdentityResult>
    <Arn>arn:aws:sts::555555555555:assumed-role/role/AssumeRoleSessionName</Arn>
    <UserId>ARO123EXAMPLE123:AssumeRoleSessionName</UserId>
    <Account>555555555555</Account>
  </GetCallerIdentityResult>
  <ResponseMetadata>
    <RequestId>01234567-89ab-cdef-0123-456789abcdef</RequestId>
  </ResponseMetadata>
</GetCallerIdentityResponse>`

	MockWebIdentityToken = `WebIdentityToken`

	MockSsoAccessKeyID     = "SSO_AKID"
	MockSsoSecretAccessKey = "SSO_SECRET_KEY"
	MockSsoSessionToken    = "SSO_SESSION_TOKEN"
)

var (
	MockStsAssumeRoleInvalidEndpointInvalidClientTokenId = &MockEndpoint{
		Request: &MockRequest{
			Body: url.Values{
				"Action":          []string{"AssumeRole"},
				"DurationSeconds": []string{"900"},
				"RoleArn":         []string{MockStsAssumeRoleArn},
				"RoleSessionName": []string{MockStsAssumeRoleSessionName},
				"Version":         []string{"2011-06-15"},
			}.Encode(),
			Method: http.MethodPost,
			Uri:    "/",
		},
		Response: &MockResponse{
			Body:        MockStsAssumeRoleInvalidResponseBodyInvalidClientTokenId,
			ContentType: "text/xml",
			StatusCode:  http.StatusForbidden,
		},
	}
	MockStsAssumeRoleValidEndpoint = &MockEndpoint{
		Request: &MockRequest{
			Body: url.Values{
				"Action":          []string{"AssumeRole"},
				"DurationSeconds": []string{"900"},
				"RoleArn":         []string{MockStsAssumeRoleArn},
				"RoleSessionName": []string{MockStsAssumeRoleSessionName},
				"Version":         []string{"2011-06-15"},
			}.Encode(),
			Method: http.MethodPost,
			Uri:    "/",
		},
		Response: &MockResponse{
			Body:        MockStsAssumeRoleValidResponseBody,
			ContentType: "text/xml",
			StatusCode:  http.StatusOK,
		},
	}

	MockStsAssumeRoleWithWebIdentityValidEndpoint = &MockEndpoint{
		Request: &MockRequest{
			Body: url.Values{
				"Action":           []string{"AssumeRoleWithWebIdentity"},
				"RoleArn":          []string{MockStsAssumeRoleWithWebIdentityArn},
				"RoleSessionName":  []string{MockStsAssumeRoleWithWebIdentitySessionName},
				"Version":          []string{"2011-06-15"},
				"WebIdentityToken": []string{MockWebIdentityToken},
			}.Encode(),
			Method: http.MethodPost,
			Uri:    "/",
		},
		Response: &MockResponse{
			Body:        MockStsAssumeRoleWithWebIdentityValidResponseBody,
			ContentType: "text/xml",
			StatusCode:  http.StatusOK,
		},
	}

	MockStsGetCallerIdentityInvalidEndpointAccessDenied = &MockEndpoint{
		Request: &MockRequest{
			Body: url.Values{
				"Action":  []string{"GetCallerIdentity"},
				"Version": []string{"2011-06-15"},
			}.Encode(),
			Method: http.MethodPost,
			Uri:    "/",
		},
		Response: &MockResponse{
			Body:        MockStsGetCallerIdentityInvalidResponseBodyAccessDenied,
			ContentType: "text/xml",
			StatusCode:  http.StatusForbidden,
		},
	}
	MockStsGetCallerIdentityInvalidBodyExpiredToken = &MockEndpoint{
		Request: &MockRequest{
			Body: url.Values{
				"Action":  []string{"GetCallerIdentity"},
				"Version": []string{"2011-06-15"},
			}.Encode(),
			Method: http.MethodPost,
			Uri:    "/",
		},
		Response: &MockResponse{
			Body:        MockStsGetCallerIdentityInvalidResponseBodyExpiredToken,
			ContentType: "text/xml",
			StatusCode:  http.StatusForbidden,
		},
	}
	MockStsGetCallerIdentityValidBodyExpiredToken = &MockEndpoint{
		Request: &MockRequest{
			Body: url.Values{
				"Action":  []string{"GetCallerIdentity"},
				"Version": []string{"2011-06-15"},
			}.Encode(),
			Method: http.MethodPost,
			Uri:    "/",
		},
		Response: &MockResponse{
			Body:        MockStsGetCallerIdentityValidResponseBodyExpiredToken,
			ContentType: "text/xml",
			StatusCode:  http.StatusForbidden,
		},
	}
	MockStsGetCallerIdentityInvalidBodyExpiredTokenException = &MockEndpoint{
		Request: &MockRequest{
			Body: url.Values{
				"Action":  []string{"GetCallerIdentity"},
				"Version": []string{"2011-06-15"},
			}.Encode(),
			Method: http.MethodPost,
			Uri:    "/",
		},
		Response: &MockResponse{
			Body:        MockStsGetCallerIdentityInvalidResponseBodyExpiredTokenException,
			ContentType: "text/xml",
			StatusCode:  http.StatusForbidden,
		},
	}
	MockStsGetCallerIdentityValidBodyExpiredTokenException = &MockEndpoint{
		Request: &MockRequest{
			Body: url.Values{
				"Action":  []string{"GetCallerIdentity"},
				"Version": []string{"2011-06-15"},
			}.Encode(),
			Method: http.MethodPost,
			Uri:    "/",
		},
		Response: &MockResponse{
			Body:        MockStsGetCallerIdentityValidResponseBodyExpiredTokenException,
			ContentType: "text/xml",
			StatusCode:  http.StatusForbidden,
		},
	}
	MockStsGetCallerIdentityInvalidBodyRequestExpired = &MockEndpoint{
		Request: &MockRequest{
			Body: url.Values{
				"Action":  []string{"GetCallerIdentity"},
				"Version": []string{"2011-06-15"},
			}.Encode(),
			Method: http.MethodPost,
			Uri:    "/",
		},
		Response: &MockResponse{
			Body:        MockStsGetCallerIdentityInvalidResponseBodyRequestExpired,
			ContentType: "text/xml",
			StatusCode:  http.StatusForbidden,
		},
	}
	MockStsGetCallerIdentityValidBodyRequestExpired = &MockEndpoint{
		Request: &MockRequest{
			Body: url.Values{
				"Action":  []string{"GetCallerIdentity"},
				"Version": []string{"2011-06-15"},
			}.Encode(),
			Method: http.MethodPost,
			Uri:    "/",
		},
		Response: &MockResponse{
			Body:        MockStsGetCallerIdentityValidResponseBodyRequestExpired,
			ContentType: "text/xml",
			StatusCode:  http.StatusForbidden,
		},
	}
	MockStsGetCallerIdentityValidEndpoint = &MockEndpoint{
		Request: &MockRequest{
			Body: url.Values{
				"Action":  []string{"GetCallerIdentity"},
				"Version": []string{"2011-06-15"},
			}.Encode(),
			Method: http.MethodPost,
			Uri:    "/",
		},
		Response: &MockResponse{
			Body:        MockStsGetCallerIdentityValidResponseBody,
			ContentType: "text/xml",
			StatusCode:  http.StatusOK,
		},
	}
	MockStsGetCallerIdentityValidAssumedRoleEndpoint = &MockEndpoint{
		Request: &MockRequest{
			Body: url.Values{
				"Action":  []string{"GetCallerIdentity"},
				"Version": []string{"2011-06-15"},
			}.Encode(),
			Method: http.MethodPost,
			Uri:    "/",
		},
		Response: &MockResponse{
			Body:        MockStsGetCallerIdentityValidAssumedRoleResponseBody,
			ContentType: "text/xml",
			StatusCode:  http.StatusOK,
		},
	}
)

// MockAwsApiServer establishes a httptest server to simulate behaviour of a real AWS API server
func MockAwsApiServer(svcName string, endpoints []*MockEndpoint) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := new(bytes.Buffer)
		if _, err := buf.ReadFrom(r.Body); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error reading from HTTP Request Body: %s", err)
			return
		}
		requestBody := buf.String()

		log.Printf("[DEBUG] Received %s API %q request to %q: %s",
			svcName, r.Method, r.RequestURI, requestBody)

		for _, e := range endpoints {
			if r.Method == e.Request.Method && r.RequestURI == e.Request.Uri && requestBody == e.Request.Body {
				log.Printf("[DEBUG] Mocked %s API responding with %d: %s",
					svcName, e.Response.StatusCode, e.Response.Body)

				w.Header().Set("Content-Type", e.Response.ContentType)
				w.Header().Set("X-Amzn-Requestid", "1b206dd1-f9a8-11e5-becf-051c60f11c4a")
				w.Header().Set("Date", time.Now().Format(time.RFC1123))
				w.WriteHeader(e.Response.StatusCode)

				fmt.Fprintln(w, e.Response.Body)
				return
			}
		}

		w.WriteHeader(http.StatusBadRequest)
	}))

	return ts
}

// AwsMetadataApiMock establishes a httptest server to mock out the internal AWS Metadata
// service. IAM Credentials are retrieved by the EC2RoleProvider, which makes
// API calls to this internal URL. By replacing the server with a test server,
// we can simulate an AWS environment
func AwsMetadataApiMock(responses []*MetadataResponse) func() {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Add("Server", "MockEC2")
		log.Printf("[DEBUG] Mock EC2 metadata server received request: %s", r.RequestURI)
		for _, e := range responses {
			if r.RequestURI == e.Uri {
				fmt.Fprintln(w, e.Body)
				return
			}
		}
		w.WriteHeader(http.StatusBadRequest)
	}))

	os.Setenv("AWS_EC2_METADATA_SERVICE_ENDPOINT", ts.URL)
	return ts.Close
}

// EcsCredentialsApiMock establishes a httptest server to mock out the ECS credentials API.
func EcsCredentialsApiMock() func() {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Add("Server", "MockECS")
		log.Printf("[DEBUG] Mock ECS credentials server received request: %s", r.RequestURI)
		if r.RequestURI == "/creds" {
			_ = json.NewEncoder(w).Encode(map[string]string{
				"AccessKeyId":     MockEcsCredentialsAccessKey,
				"Expiration":      time.Now().UTC().Format(time.RFC3339),
				"RoleArn":         "arn:aws:iam::000000000000:role/EcsCredentials",
				"SecretAccessKey": MockEcsCredentialsSecretKey,
				"Token":           MockEcsCredentialsSessionToken,
			})
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	}))

	os.Setenv("AWS_CONTAINER_CREDENTIALS_FULL_URI", ts.URL+"/creds")
	return ts.Close
}

func SsoCredentialsApiMock() (func(), string) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(fmt.Appendf(nil, //nolint:errcheck
			ssoCredentialsResponse,
			time.Now().
				Add(15*time.Minute). //nolint:mnd
				UnixNano()/int64(time.Millisecond)))
	}))

	return ts.Close, ts.URL
}

const ssoCredentialsResponse = `{
	"roleCredentials": {
	  "accessKeyId": "SSO_AKID",
	  "secretAccessKey": "SSO_SECRET_KEY",
	  "sessionToken": "SSO_SESSION_TOKEN",
	  "expiration": %d
	}
}`

// MockStsAssumeRoleValidEndpointWithOptions returns a valid STS AssumeRole response with configurable request options.
func MockStsAssumeRoleValidEndpointWithOptions(options map[string]string) *MockEndpoint {
	urlValues := url.Values{
		"Action":          []string{"AssumeRole"},
		"DurationSeconds": []string{"900"},
		"RoleArn":         []string{MockStsAssumeRoleArn},
		"RoleSessionName": []string{MockStsAssumeRoleSessionName},
		"Version":         []string{"2011-06-15"},
	}

	for k, v := range options {
		urlValues.Set(k, v)
	}

	return &MockEndpoint{
		Request: &MockRequest{
			Body:   urlValues.Encode(),
			Method: http.MethodPost,
			Uri:    "/",
		},
		Response: &MockResponse{
			Body:        MockStsAssumeRoleValidResponseBody,
			ContentType: "text/xml",
			StatusCode:  http.StatusOK,
		},
	}
}

// MockStsAssumeRoleValidEndpointWithOptions returns a valid STS AssumeRole response with configurable request options.
func MockStsAssumeRoleWithWebIdentityValidWithOptions(options map[string]string) *MockEndpoint {
	urlValues := url.Values{
		"Action":           []string{"AssumeRoleWithWebIdentity"},
		"RoleArn":          []string{MockStsAssumeRoleWithWebIdentityArn},
		"RoleSessionName":  []string{MockStsAssumeRoleWithWebIdentitySessionName},
		"Version":          []string{"2011-06-15"},
		"WebIdentityToken": []string{MockWebIdentityToken},
	}

	for k, v := range options {
		urlValues.Set(k, v)
	}

	return &MockEndpoint{
		Request: &MockRequest{
			Body:   urlValues.Encode(),
			Method: http.MethodPost,
			Uri:    "/",
		},
		Response: &MockResponse{
			Body:        MockStsAssumeRoleWithWebIdentityValidResponseBody,
			ContentType: "text/xml",
			StatusCode:  http.StatusOK,
		},
	}
}

// MockEndpoint represents a basic request and response that can be used for creating simple httptest server routes.
type MockEndpoint struct {
	Request  *MockRequest
	Response *MockResponse
}

// MockRequest represents a basic HTTP request
type MockRequest struct {
	Method string
	Uri    string
	Body   string
}

// MockResponse represents a basic HTTP response.
type MockResponse struct {
	StatusCode  int
	Body        string
	ContentType string
}

// MetadataResponse represents a metadata server response URI and body
type MetadataResponse struct {
	Uri  string `json:"uri"`
	Body string `json:"body"`
}

var Ec2metadata_instanceIdEndpoint = &MetadataResponse{
	Uri:  "/latest/meta-data/instance-id",
	Body: "mock-instance-id",
}

var Ec2metadata_securityCredentialsEndpoints = []*MetadataResponse{
	{
		Uri:  "/latest/api/token",
		Body: "Ec2MetadataApiToken",
	},
	{
		Uri:  "/latest/meta-data/iam/security-credentials/",
		Body: "test_role",
	},
	{
		Uri: "/latest/meta-data/iam/security-credentials/test_role",
		Body: `{
	"Code": "Success",
	"LastUpdated": "2015-12-11T17:17:25Z",
	"Type": "AWS-HMAC",
	"AccessKeyId": "Ec2MetadataAccessKey",
	"SecretAccessKey": "Ec2MetadataSecretKey",
	"Token": "Ec2MetadataSessionToken"
}`,
	},
}

var Ec2metadata_iamInfoEndpoint = &MetadataResponse{
	Uri: "/latest/meta-data/iam/info",
	Body: `{
	"Code": "Success",
	"LastUpdated": "2016-03-17T12:27:32Z",
	"InstanceProfileArn": "arn:aws:iam::000000000000:instance-profile/my-instance-profile",
	"InstanceProfileId": "AIPAABCDEFGHIJKLMN123"
}`,
}

const Ec2metadata_iamInfoEndpoint_expectedAccountID = `000000000000`
const Ec2metadata_iamInfoEndpoint_expectedPartition = `aws`

func Ec2metadata_instanceIdentityEndpoint(region string) *MetadataResponse {
	return &MetadataResponse{
		Uri: "/latest/dynamic/instance-identity/document",
		Body: fmt.Sprintf(`{
	"version": "2017-09-30",
	"instanceId": "mock-instance-id",
	"region": %q
}`, region),
	}
}

const IamResponse_GetUser_valid = `<GetUserResponse xmlns="https://iam.amazonaws.com/doc/2010-05-08/">
  <GetUserResult>
    <User>
      <UserId>AIDACKCEVSQ6C2EXAMPLE</UserId>
      <Path>/division_abc/subdivision_xyz/</Path>
      <UserName>Bob</UserName>
      <Arn>arn:aws:iam::111111111111:user/division_abc/subdivision_xyz/Bob</Arn>
      <CreateDate>2013-10-02T17:01:44Z</CreateDate>
      <PasswordLastUsed>2014-10-10T14:37:51Z</PasswordLastUsed>
    </User>
  </GetUserResult>
  <ResponseMetadata>
    <RequestId>7a62c49f-347e-4fc4-9331-6e8eEXAMPLE</RequestId>
  </ResponseMetadata>
</GetUserResponse>`

const IamResponse_GetUser_valid_expectedAccountID = `111111111111`
const IamResponse_GetUser_valid_expectedPartition = `aws`

const IamResponse_GetUser_unauthorized = `<ErrorResponse xmlns="https://iam.amazonaws.com/doc/2010-05-08/">
  <Error>
    <Type>Sender</Type>
    <Code>AccessDenied</Code>
    <Message>User: arn:aws:iam::123456789012:user/Bob is not authorized to perform: iam:GetUser on resource: arn:aws:iam::123456789012:user/Bob</Message>
  </Error>
  <RequestId>7a62c49f-347e-4fc4-9331-6e8eEXAMPLE</RequestId>
</ErrorResponse>`

const IamResponse_GetUser_federatedFailure = `<ErrorResponse xmlns="https://iam.amazonaws.com/doc/2010-05-08/">
  <Error>
    <Type>Sender</Type>
    <Code>ValidationError</Code>
    <Message>Must specify userName when calling with non-User credentials</Message>
  </Error>
  <RequestId>7a62c49f-347e-4fc4-9331-6e8eEXAMPLE</RequestId>
</ErrorResponse>`

const IamResponse_ListRoles_valid = `<ListRolesResponse xmlns="https://iam.amazonaws.com/doc/2010-05-08/">
  <ListRolesResult>
    <IsTruncated>true</IsTruncated>
    <Marker>AWceSSsKsazQ4IEplT9o4hURCzBs00iavlEvEXAMPLE</Marker>
    <Roles>
      <member>
        <Path>/</Path>
        <AssumeRolePolicyDocument>%7B%22Version%22%3A%222008-10-17%22%2C%22Statement%22%3A%5B%7B%22Sid%22%3A%22%22%2C%22Effect%22%3A%22Allow%22%2C%22Principal%22%3A%7B%22Service%22%3A%22ec2.amazonaws.com%22%7D%2C%22Action%22%3A%22sts%3AAssumeRole%22%7D%5D%7D</AssumeRolePolicyDocument>
        <RoleId>AROACKCEVSQ6C2EXAMPLE</RoleId>
        <RoleName>elasticbeanstalk-role</RoleName>
        <Arn>arn:aws:iam::444444444444:role/elasticbeanstalk-role</Arn>
        <CreateDate>2013-10-02T17:01:44Z</CreateDate>
      </member>
    </Roles>
  </ListRolesResult>
  <ResponseMetadata>
    <RequestId>7a62c49f-347e-4fc4-9331-6e8eEXAMPLE</RequestId>
  </ResponseMetadata>
</ListRolesResponse>`

const IamResponse_ListRoles_valid_expectedAccountID = `444444444444`
const IamResponse_ListRoles_valid_expectedPartition = `aws`

const IamResponse_ListRoles_unauthorized = `<ErrorResponse xmlns="https://iam.amazonaws.com/doc/2010-05-08/">
  <Error>
    <Type>Sender</Type>
    <Code>AccessDenied</Code>
    <Message>User: arn:aws:iam::123456789012:user/Bob is not authorized to perform: iam:ListRoles on resource: arn:aws:iam::123456789012:role/</Message>
  </Error>
  <RequestId>7a62c49f-347e-4fc4-9331-6e8eEXAMPLE</RequestId>
</ErrorResponse>`
