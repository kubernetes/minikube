// +build appengine

package bugsnag

import (
	"appengine"
	"appengine/urlfetch"
	"appengine/user"
	"fmt"
	"log"
	"net/http"
)

func defaultPanicHandler() {}

func init() {
	OnBeforeNotify(appengineMiddleware)
}

func appengineMiddleware(event *Event, config *Configuration) (err error) {
	var c appengine.Context

	for _, datum := range event.RawData {
		if r, ok := datum.(*http.Request); ok {
			c = appengine.NewContext(r)
			break
		} else if context, ok := datum.(appengine.Context); ok {
			c = context
			break
		}
	}

	if c == nil {
		return fmt.Errorf("No appengine context given")
	}

	// You can only use the builtin http library if you pay for appengine,
	// so we use the appengine urlfetch service instead.
	config.Transport = &urlfetch.Transport{
		Context: c,
	}

	// Anything written to stderr/stdout is discarded, so lets log to the request.
	config.Logger = log.New(appengineWriter{c}, config.Logger.Prefix(), config.Logger.Flags())

	// Set the releaseStage appropriately
	if config.ReleaseStage == "" {
		if appengine.IsDevAppServer() {
			config.ReleaseStage = "development"
		} else {
			config.ReleaseStage = "production"
		}
	}

	if event.User == nil {
		u := user.Current(c)
		if u != nil {
			event.User = &User{
				Id:    u.ID,
				Email: u.Email,
			}
		}
	}

	return nil
}

// Convert an appengine.Context into an io.Writer so we can create a log.Logger.
type appengineWriter struct {
	appengine.Context
}

func (c appengineWriter) Write(b []byte) (int, error) {
	c.Warningf(string(b))
	return len(b), nil
}
