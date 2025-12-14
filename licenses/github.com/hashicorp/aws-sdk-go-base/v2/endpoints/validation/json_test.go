// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validation

import (
	"testing"
)

func TestJSONNoDuplicateKeys(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		wantErr bool
	}{
		{
			name:    "invalid",
			s:       "{{{",
			wantErr: true,
		},
		{
			name: "valid",
			s: `{
  "a": "foo",
  "b": {
    "c": "bar",
    "d": [
      {
        "e": "baz"
      },
      {
        "f": "qux",
        "g": "foo"
      }
    ]
  }
}`,
			wantErr: false,
		},
		{
			name: "root",
			s: `{
  "a": "foo",
  "a": "bar"
}`,
			wantErr: true,
		},
		{
			name: "nested object",
			s: `{
  "a": "foo",
  "b": {
    "c": "bar",
    "c": "baz"
  }
}`,
			wantErr: true,
		},
		{
			name: "nested array",
			s: `{
  "a": "foo",
  "b": {
    "c": "bar",
    "d": [
      {
        "e": "foo",
        "e": "bar"
      },
      {
        "f": "baz",
        "g": "qux"
      }
    ]
  }
}`,
			wantErr: true,
		},
		{
			name: "multiple",
			s: `{
  "a": "foo",
  "a": "bar",
  "b": {
    "c": "baz",
    "c": "qux",
    "d": [
      {
        "e": "foo"
      },
      {
        "f": "bar",
        "f": "baz",
        "g": "qux"
      }
    ]
  }
}`,
			wantErr: true,
		},
		{
			name: "aws iam condition keys",
			s: `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "iam:PassRole",
      "Resource": "*",
      "Condition": {
        "StringEquals": {
          "iam:PassedToService": "cloudwatch.amazonaws.com"
        },
        "StringEquals": {
          "iam:PassedToService": "ec2.amazonaws.com"
        }
      }
    }
  ]
}`,

			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := JSONNoDuplicateKeys(tt.s); (err != nil) != tt.wantErr {
				t.Errorf("JSONNoDuplicateKeys() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
