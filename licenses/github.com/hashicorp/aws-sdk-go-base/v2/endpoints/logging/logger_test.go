// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logging

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-log/tflogtest"
)

func testLoggerWarn(t *testing.T, rootName string, factory func(ctx context.Context, name string, output io.Writer) (context.Context, Logger)) {
	t.Helper()

	loggerName := "test"
	expectedModule := rootName + "." + loggerName

	var buf bytes.Buffer
	ctx := context.Background()
	ctx, logger := factory(ctx, loggerName, &buf)

	logger.Warn(ctx, "message", map[string]any{
		"one": int(1),
		"two": "two",
	})

	lines, err := tflogtest.MultilineJSONDecode(&buf)
	if err != nil {
		t.Fatalf("decoding log lines: %s", err)
	}

	expected := []map[string]any{
		{
			"@level":   "warn",
			"@module":  expectedModule,
			"@message": "message",
			"one":      float64(1),
			"two":      "two",
		},
	}

	if diff := cmp.Diff(expected, lines); diff != "" {
		t.Errorf("unexpected logger output difference: %s", diff)
	}
}

func testLoggerSetField(t *testing.T, rootName string, factory func(ctx context.Context, name string, output io.Writer) (context.Context, Logger)) {
	t.Helper()

	loggerName := "test"
	expectedModule := rootName + "." + loggerName

	var buf bytes.Buffer
	originalCtx := context.Background()
	originalCtx, logger := factory(originalCtx, loggerName, &buf)

	newCtx := logger.SetField(originalCtx, "key", "value")

	logger.Warn(newCtx, "new logger")
	logger.Warn(newCtx, "new logger", map[string]any{
		"key": "other value",
	})
	logger.Warn(originalCtx, "original logger")

	lines, err := tflogtest.MultilineJSONDecode(&buf)
	if err != nil {
		t.Fatalf("ctxWithField: decoding log lines: %s", err)
	}

	expected := []map[string]any{
		{
			"@level":   "warn",
			"@module":  expectedModule,
			"@message": "new logger",
			"key":      "value",
		},
		{
			"@level":   "warn",
			"@module":  expectedModule,
			"@message": "new logger",
			"key":      "other value",
		},
		{
			"@level":   "warn",
			"@module":  expectedModule,
			"@message": "original logger",
		},
	}

	if diff := cmp.Diff(expected, lines); diff != "" {
		t.Errorf("unexpected logger output difference: %s", diff)
	}
}
