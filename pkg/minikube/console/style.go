/*
Copyright 2019 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package console

import (
	"fmt"

	"golang.org/x/text/message"
)

// style describes how to stylize a message.
type style struct {
	// Prefix is a string to place in the beginning of a message
	Prefix string
	// OmitNewline omits a newline at the end of a message.
	OmitNewline bool
}

// styles is a map of style name to style struct
// For consistency, ensure that emojis added render with the same width across platforms.
var styles = map[string]style{
	"happy":      {Prefix: "😄"},
	"success":    {Prefix: "✅"},
	"failure":    {Prefix: "❌"},
	"conflict":   {Prefix: "💥"},
	"fatal":      {Prefix: "💣"},
	"notice":     {Prefix: "📌"},
	"ready":      {Prefix: "🏄"},
	"restarting": {Prefix: "🔄"},
	"stopping":   {Prefix: "✋"},
	"stopped":    {Prefix: "🛑"},
	"warning":    {Prefix: "⚠️"},
	"waiting":    {Prefix: "⌛"},
	"usage":      {Prefix: "💡"},
	"launch":     {Prefix: "🚀"},
	"thumbs-up":  {Prefix: "👍"},
	"option":     {Prefix: "   ▪ "},
	"crushed":    {Prefix: "💔"},

	// Specialized purpose styles
	"iso-download":      {Prefix: "💿"},
	"file-download":     {Prefix: "💾"},
	"caching":           {Prefix: "🤹"},
	"starting-vm":       {Prefix: "🔥"},
	"starting-none":     {Prefix: "🤹"},
	"deleting-vm":       {Prefix: "🔥"},
	"copying":           {Prefix: "✨"},
	"connectivity":      {Prefix: "📶"},
	"mounting":          {Prefix: "📁"},
	"celebrate":         {Prefix: "🎉"},
	"container-runtime": {Prefix: "🎁"},
	"Docker":            {Prefix: "🐳"},
	"CRIO":              {Prefix: "🎁"}, // This should be a snow-flake, but the emoji has a strange width on macOS
	"containerd":        {Prefix: "📦"},
	"permissions":       {Prefix: "🔑"},
	"enabling":          {Prefix: "🔌"},
	"pulling":           {Prefix: "🚜"},
	"verifying":         {Prefix: "🤔"},
	"verifying-noline":  {Prefix: "🤔", OmitNewline: true},
	"kubectl":           {Prefix: "💗"},
	"meh":               {Prefix: "🙄"},
	"embarassed":        {Prefix: "🤦"},
	"tip":               {Prefix: "💡"},
}

// Add a prefix to a string
func applyPrefix(prefix, format string) string {
	if prefix == "" {
		return format
	}
	// TODO(tstromberg): Ensure compatibility with RTL languages.
	return prefix + "  " + format
}

// Apply styling to a format string
func applyStyle(style string, useColor bool, format string, a ...interface{}) (string, error) {
	p := message.NewPrinter(preferredLanguage)
	out := p.Sprintf(format, a...)

	s, ok := styles[style]
	if !s.OmitNewline {
		out += "\n"
	}

	// Similar to CSS styles, if no style matches, output an unformatted string.
	if !ok {
		return p.Sprintf(format, a...), fmt.Errorf("unknown style: %q", style)
	}

	prefix := s.Prefix
	if !useColor && prefix != "" {
		prefix = "-"
	}
	out = applyPrefix(prefix, out)
	return out, nil
}
