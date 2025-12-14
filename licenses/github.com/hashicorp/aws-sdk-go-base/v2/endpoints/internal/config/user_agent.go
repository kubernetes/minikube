// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package config

import (
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

type UserAgentProduct struct {
	Name    string
	Version string
	Comment string
}

type UserAgentProducts []UserAgentProduct

func (ua UserAgentProducts) BuildUserAgentString() string {
	builder := smithyhttp.NewUserAgentBuilder()
	for _, p := range ua {
		p.buildUserAgentPart(builder)
	}
	return builder.Build()
}

func (p UserAgentProduct) buildUserAgentPart(b *smithyhttp.UserAgentBuilder) {
	if p.Name != "" {
		if p.Version != "" {
			b.AddKeyValue(p.Name, p.Version)
		} else {
			b.AddKey(p.Name)
		}
	}
	if p.Comment != "" {
		b.AddKey("(" + p.Comment + ")")
	}
}
