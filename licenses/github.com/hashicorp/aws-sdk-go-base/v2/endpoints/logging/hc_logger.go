// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logging

import (
	"context"

	"github.com/hashicorp/go-hclog"
)

type HcLogger struct{}

var _ Logger = HcLogger{}

func NewHcLogger(ctx context.Context, logger hclog.Logger) (context.Context, HcLogger) {
	ctx = hclog.WithContext(ctx, logger)

	return ctx, HcLogger{}
}

func (l HcLogger) SubLogger(ctx context.Context, name string) (context.Context, Logger) {
	logger := hclog.FromContext(ctx)
	logger = logger.Named(name)
	ctx = hclog.WithContext(ctx, logger)

	return ctx, HcLogger{}
}

func (l HcLogger) Warn(ctx context.Context, msg string, fields ...map[string]any) {
	logger := hclog.FromContext(ctx)
	logger.Warn(msg, flattenFields(fields...)...)
}

func (l HcLogger) Info(ctx context.Context, msg string, fields ...map[string]any) {
	logger := hclog.FromContext(ctx)
	logger.Info(msg, flattenFields(fields...)...)
}

func (l HcLogger) Debug(ctx context.Context, msg string, fields ...map[string]any) {
	logger := hclog.FromContext(ctx)
	logger.Debug(msg, flattenFields(fields...)...)
}

func (l HcLogger) Trace(ctx context.Context, msg string, fields ...map[string]any) {
	logger := hclog.FromContext(ctx)
	logger.Trace(msg, flattenFields(fields...)...)
}

// TODO: how to handle duplicates
func flattenFields(fields ...map[string]any) []any {
	var totalLen int
	for _, m := range fields {
		totalLen = len(m)
	}
	f := make([]any, 0, totalLen*2) //nolint:mnd

	for _, m := range fields {
		for k, v := range m {
			f = append(f, k, v)
		}
	}
	return f
}

func (l HcLogger) SetField(ctx context.Context, key string, value any) context.Context {
	logger := hclog.FromContext(ctx)
	logger = logger.With(key, value)
	ctx = hclog.WithContext(ctx, logger)
	return ctx
}
