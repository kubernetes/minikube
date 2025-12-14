// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package awsbase

import (
	"context"
	"fmt"

	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"github.com/hashicorp/aws-sdk-go-base/v2/useragent"
)

func apnUserAgentMiddleware(apn APNInfo) middleware.BuildMiddleware {
	return middleware.BuildMiddlewareFunc("tfAPNUserAgent",
		func(ctx context.Context, in middleware.BuildInput, next middleware.BuildHandler) (middleware.BuildOutput, middleware.Metadata, error) {
			request, ok := in.Request.(*smithyhttp.Request)
			if !ok {
				return middleware.BuildOutput{}, middleware.Metadata{}, fmt.Errorf("unknown request type %T", in.Request)
			}

			prependUserAgentHeader(request, apn.BuildUserAgentString())

			return next.HandleBuild(ctx, in)
		},
	)
}

// Because the default User-Agent middleware prepends itself to the contents of the User-Agent header,
// we have to run after it and also prepend our custom User-Agent
func prependUserAgentHeader(request *smithyhttp.Request, value string) {
	current := request.Header.Get("User-Agent")
	if len(current) > 0 {
		current = value + " " + current
	} else {
		current = value
	}
	request.Header["User-Agent"] = append(request.Header["User-Agent"][:0], current)
}

func withUserAgentAppender(ua string) func(*middleware.Stack) error {
	return func(stack *middleware.Stack) error {
		return stack.Build.Add(userAgentMiddleware(ua), middleware.After)
	}
}

func userAgentMiddleware(ua string) middleware.BuildMiddleware {
	return middleware.BuildMiddlewareFunc("tfUserAgentAppender",
		func(ctx context.Context, in middleware.BuildInput, next middleware.BuildHandler) (middleware.BuildOutput, middleware.Metadata, error) {
			request, ok := in.Request.(*smithyhttp.Request)
			if !ok {
				return middleware.BuildOutput{}, middleware.Metadata{}, fmt.Errorf("unknown request type %T", in.Request)
			}

			appendUserAgentHeader(request, ua)

			return next.HandleBuild(ctx, in)
		},
	)
}

func userAgentFromContextMiddleware() middleware.BuildMiddleware {
	return middleware.BuildMiddlewareFunc("tfCtxUserAgentAppender",
		func(ctx context.Context, in middleware.BuildInput, next middleware.BuildHandler) (middleware.BuildOutput, middleware.Metadata, error) {
			request, ok := in.Request.(*smithyhttp.Request)
			if !ok {
				return middleware.BuildOutput{}, middleware.Metadata{}, fmt.Errorf("unknown request type %T", in.Request)
			}

			if v := useragent.BuildFromContext(ctx); v != "" {
				appendUserAgentHeader(request, v)
			}

			return next.HandleBuild(ctx, in)
		},
	)
}

func appendUserAgentHeader(request *smithyhttp.Request, value string) {
	current := request.Header.Get("User-Agent")
	if len(current) > 0 {
		current = current + " " + value
	} else {
		current = value
	}
	request.Header["User-Agent"] = append(request.Header["User-Agent"][:0], current)
}
