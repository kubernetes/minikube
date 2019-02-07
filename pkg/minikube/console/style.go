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
var styles = map[string]style{
	// General purpose
	"happy":      style{Prefix: "😄"},
	"success":    style{Prefix: "✅ "},
	"failure":    style{Prefix: "❌"},
	"conflict":   style{Prefix: "💥"},
	"fatal":      style{Prefix: "💣"},
	"notice":     style{Prefix: "📌"},
	"ready":      style{Prefix: "🏄"},
	"restarting": style{Prefix: "🔁"},
	"stopping":   style{Prefix: "🚦"},
	"stopped":    style{Prefix: "🛑"},
	"warning":    style{Prefix: "⚠️"},
	"waiting":    style{Prefix: "⌛"},
	"usage":      style{Prefix: "💡"},
	"launch":     style{Prefix: "🚀"},

	// Specialized purpose
	"iso-download":      style{Prefix: "💿"},
	"file-download":     style{Prefix: "💾"},
	"caching":           style{Prefix: "🤹"},
	"starting-vm":       style{Prefix: "🔥"},
	"copying":           style{Prefix: "✨"},
	"connectivity":      style{Prefix: "📡"},
	"mounting":          style{Prefix: "📁"},
	"celebrate":         style{Prefix: "🎉"},
	"container-runtime": style{Prefix: "🎁"},
	"enabling":          style{Prefix: "🔌"},
	"pulling":           style{Prefix: "🚜"},
	"verifying":         style{Prefix: "🤔"},
	"kubectl":           style{Prefix: "❤️"},
	"meh":               style{Prefix: "🙄"},
	"embarassed":        style{Prefix: "🤦"},
}

// Add a prefix to a string
func applyPrefix(prefix, format string) string {
	if prefix == "" {
		return format
	}
	// TODO(tstromberg): Ensure compatibility with RTL languages.
	return prefix + " " + format
}

// Apply styling to a format string
func applyStyle(style string, useColor bool, format string, a ...interface{}) (string, error) {
	p := message.NewPrinter(preferredLanguage)
	s, ok := styles[style]
	if !s.OmitNewline {
		format = format + "\n"
	}

	// Similar to CSS styles, if no style matches, output an unformatted string.
	if !ok {
		return p.Sprintf(format, a...), fmt.Errorf("unknown style: %q", style)
	}

	prefix := s.Prefix
	if useColor && prefix != "" {
		prefix = "-"
	}
	format = applyPrefix(prefix, format)
	return p.Sprintf(format, a...), nil
}
