// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logging

import (
	"context"
)

type Logger interface {
	Warn(ctx context.Context, msg string, fields ...map[string]any)
	Info(ctx context.Context, msg string, fields ...map[string]any)
	Debug(ctx context.Context, msg string, fields ...map[string]any)
	Trace(ctx context.Context, msg string, fields ...map[string]any)

	SetField(ctx context.Context, key string, value any) context.Context

	SubLogger(ctx context.Context, name string) (context.Context, Logger)
}
