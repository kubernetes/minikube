// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logging

import (
	"context"
)

type NullLogger struct {
}

var _ Logger = NullLogger{}

func (l NullLogger) SubLogger(ctx context.Context, name string) (context.Context, Logger) {
	return ctx, l
}

func (l NullLogger) Warn(ctx context.Context, msg string, fields ...map[string]any) {
}

func (l NullLogger) Info(ctx context.Context, msg string, fields ...map[string]any) {
}

func (l NullLogger) Debug(ctx context.Context, msg string, fields ...map[string]any) {
}

func (l NullLogger) Trace(ctx context.Context, msg string, fields ...map[string]any) {
}

func (l NullLogger) SetField(ctx context.Context, key string, value any) context.Context {
	return ctx
}
