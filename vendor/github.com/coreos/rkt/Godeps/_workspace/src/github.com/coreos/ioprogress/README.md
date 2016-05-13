# ioprogress

ioprogress is a Go (golang) library with implementations of `io.Reader`
and `io.Writer` that draws progress bars. The primary use case for these
are for CLI applications but alternate progress bar writers can be supplied
for alternate environments.

## Example

![Progress](http://g.recordit.co/GO5HxT16QH.gif)

## Installation

Standard `go get`:

```
$ go get github.com/mitchellh/ioprogress
```

## Usage

Here is an example of outputting a basic progress bar to the CLI as
we're "downloading" from some other `io.Reader` (perhaps from a network
connection):

```go
// Imagine this came from some external source, such as a network connection,
// and that we have the full size of it, such as from a Content-Length HTTP
// header.
var r io.Reader

// Create the progress reader
progressR := &ioprogress.Reader{
	Reader: r,
	Size:   rSize,
}

// Copy all of the reader to some local file f. As it copies, the
// progressR will write progress to the terminal on os.Stdout. This is
// customizable.
io.Copy(f, progressR)
```
