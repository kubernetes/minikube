package ioprogress

import (
	"io"
	"time"
)

// Reader is an implementation of io.Reader that draws the progress of
// reading some data.
type Reader struct {
	// Reader is the underlying reader to read from
	Reader io.Reader

	// Size is the total size of the data coming out of the reader.
	Size int64

	// DrawFunc is the callback to invoke to draw the progress bar. By
	// default, this will be DrawTerminal(os.Stdout).
	//
	// DrawInterval is the minimum time to wait between reads to update the
	// progress bar.
	DrawFunc     DrawFunc
	DrawInterval time.Duration

	progress int64
	lastDraw time.Time
}

// Read reads from the underlying reader and invokes the DrawFunc if
// appropriate. The DrawFunc is executed when there is data that is
// read (progress is made) and at least DrawInterval time has passed.
func (r *Reader) Read(p []byte) (int, error) {
	// If we haven't drawn before, initialize the progress bar
	if r.lastDraw.IsZero() {
		r.initProgress()
	}

	// Read from the underlying source
	n, err := r.Reader.Read(p)

	// Always increment the progress even if there was an error
	r.progress += int64(n)

	// If we don't have any errors, then draw the progress. If we are
	// at the end of the data, then finish the progress.
	if err == nil {
		// Only draw if we read data or we've never read data before (to
		// initialize the progress bar).
		if n > 0 {
			r.drawProgress()
		}
	}
	if err == io.EOF {
		r.finishProgress()
	}

	return n, err
}

func (r *Reader) drawProgress() {
	// If we've drawn before, then make sure that the draw interval
	// has passed before we draw again.
	interval := r.DrawInterval
	if interval == 0 {
		interval = time.Second
	}
	if !r.lastDraw.IsZero() {
		nextDraw := r.lastDraw.Add(interval)
		if time.Now().Before(nextDraw) {
			return
		}
	}

	// Draw
	f := r.drawFunc()
	f(r.progress, r.Size)

	// Record this draw so that we don't draw again really quickly
	r.lastDraw = time.Now()
}

func (r *Reader) finishProgress() {
	f := r.drawFunc()
	f(r.progress, r.Size)

	// Print a newline
	f(-1, -1)

	// Reset lastDraw so we don't finish again
	var zeroDraw time.Time
	r.lastDraw = zeroDraw
}

func (r *Reader) initProgress() {
	var zeroDraw time.Time
	r.lastDraw = zeroDraw
	r.drawProgress()
	r.lastDraw = zeroDraw
}

func (r *Reader) drawFunc() DrawFunc {
	if r.DrawFunc == nil {
		return defaultDrawFunc
	}

	return r.DrawFunc
}
