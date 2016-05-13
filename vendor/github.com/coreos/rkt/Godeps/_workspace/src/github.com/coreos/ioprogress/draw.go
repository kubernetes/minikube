package ioprogress

import (
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/crypto/ssh/terminal"
)

// DrawFunc is the callback type for drawing progress.
type DrawFunc func(int64, int64) error

// DrawTextFormatFunc is a callback used by DrawFuncs that draw text in
// order to format the text into some more human friendly format.
type DrawTextFormatFunc func(int64, int64) string

var defaultDrawFunc DrawFunc

func init() {
	defaultDrawFunc = DrawTerminal(os.Stdout)
}

// isTerminal returns True when w is going to a tty, and false otherwise.
func isTerminal(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		return terminal.IsTerminal(int(f.Fd()))
	}
	return false
}

// DrawTerminal returns a DrawFunc that draws a progress bar to an io.Writer
// that is assumed to be a terminal (and therefore respects carriage returns).
func DrawTerminal(w io.Writer) DrawFunc {
	return DrawTerminalf(w, func(progress, total int64) string {
		return fmt.Sprintf("%d/%d", progress, total)
	})
}

// DrawTerminalf returns a DrawFunc that draws a progress bar to an io.Writer
// that is formatted with the given formatting function.
func DrawTerminalf(w io.Writer, f DrawTextFormatFunc) DrawFunc {
	var maxLength int

	return func(progress, total int64) error {
		if progress == -1 && total == -1 {
			_, err := fmt.Fprintf(w, "\n")
			return err
		}

		// Make sure we pad it to the max length we've ever drawn so that
		// we don't have trailing characters.
		line := f(progress, total)
		if len(line) < maxLength {
			line = fmt.Sprintf(
				"%s%s",
				line,
				strings.Repeat(" ", maxLength-len(line)))
		}
		maxLength = len(line)

		terminate := "\r"
		if !isTerminal(w) {
			terminate = "\n"
		}
		_, err := fmt.Fprint(w, line+terminate)
		return err
	}
}

var byteUnits = []string{"B", "KB", "MB", "GB", "TB", "PB"}

// DrawTextFormatBytes is a DrawTextFormatFunc that formats the progress
// and total into human-friendly byte formats.
func DrawTextFormatBytes(progress, total int64) string {
	return fmt.Sprintf("%s/%s", ByteUnitStr(progress), ByteUnitStr(total))
}

// DrawTextFormatBar returns a DrawTextFormatFunc that draws a progress
// bar with the given width (in characters). This can be used in conjunction
// with another DrawTextFormatFunc to create a progress bar with bytes, for
// example:
//
//     bar := DrawTextFormatBar(20)
//     func(progress, total int64) string {
//         return fmt.Sprintf(
//           "%s %s",
//           bar(progress, total),
//           DrawTextFormatBytes(progress, total))
//     }
//
func DrawTextFormatBar(width int64) DrawTextFormatFunc {
	return DrawTextFormatBarForW(width, nil)
}

// DrawTextFormatBarForW returns a DrawTextFormatFunc as described in the docs
// for DrawTextFormatBar, however if the io.Writer passed in is not a tty then
// the returned function will always return "".
func DrawTextFormatBarForW(width int64, w io.Writer) DrawTextFormatFunc {
	if w != nil && !isTerminal(w) {
		return func(progress, total int64) string {
			return ""
		}
	}

	width -= 2

	return func(progress, total int64) string {
		current := int64((float64(progress) / float64(total)) * float64(width))
		return fmt.Sprintf(
			"[%s%s]",
			strings.Repeat("=", int(current)),
			strings.Repeat(" ", int(width-current)))
	}
}

// ByteUnitStr pretty prints a number of bytes.
func ByteUnitStr(n int64) string {
	var unit string
	size := float64(n)
	for i := 1; i < len(byteUnits); i++ {
		if size < 1000 {
			unit = byteUnits[i-1]
			break
		}

		size = size / 1000
	}

	return fmt.Sprintf("%.3g %s", size, unit)
}
