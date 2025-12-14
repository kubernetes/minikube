// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package useragent

import (
	"context"

	"github.com/hashicorp/aws-sdk-go-base/v2/internal/config"
)

type userAgentKey string

const (
	contextScopedUserAgent userAgentKey = "ContextScopedUserAgent"
)

func Context(ctx context.Context, products config.UserAgentProducts) context.Context {
	return context.WithValue(ctx, contextScopedUserAgent, products)
}

func BuildFromContext(ctx context.Context) string {
	ps, ok := ctx.Value(contextScopedUserAgent).(config.UserAgentProducts)
	if !ok {
		return ""
	}

	return ps.BuildUserAgentString()
}
