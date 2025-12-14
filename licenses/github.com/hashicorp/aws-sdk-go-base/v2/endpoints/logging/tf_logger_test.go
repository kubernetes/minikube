// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logging

import (
	"context"
	"io"
	"testing"

	"github.com/hashicorp/terraform-plugin-log/tflogtest"
)

const tflogRootName = "provider"

func TestTfLoggerWarn(t *testing.T) {
	testLoggerWarn(t, tflogRootName, tfLoggerFactory)
}

func TestTfLoggerSetField(t *testing.T) {
	testLoggerSetField(t, tflogRootName, tfLoggerFactory)
}

func tfLoggerFactory(ctx context.Context, name string, output io.Writer) (context.Context, Logger) {
	ctx = tflogtest.RootLogger(ctx, output)

	ctx, rootLogger := NewTfLogger(ctx)
	ctx, logger := rootLogger.SubLogger(ctx, name)

	return ctx, logger
}
