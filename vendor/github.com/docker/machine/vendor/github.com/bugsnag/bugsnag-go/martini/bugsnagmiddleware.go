/*
Package bugsnagmartini provides a martini middleware that sends
panics to Bugsnag. You should use this middleware in combination
with martini.Recover() if you want to send error messages to your
clients:

	func main() {
		m := martini.New()
		// used to stop panics bubbling and return a 500 error.
		m.Use(martini.Recovery())

		// used to send panics to Bugsnag.
		m.Use(bugsnagmartini.AutoNotify(bugsnag.Configuration{
			APIKey: "YOUR_API_KEY_HERE",
		})

		// ...
	}

This middleware also makes bugsnag available to martini handlers via
the context.

	func myHandler(w http.ResponseWriter, r *http.Request, bugsnag *bugsnag.Notifier) {
		// ...
		bugsnag.Notify(err)
		// ...
	}

*/
package bugsnagmartini

import (
	"net/http"

	"github.com/bugsnag/bugsnag-go"
	"github.com/go-martini/martini"
)

// AutoNotify sends any panics to bugsnag, and then re-raises them.
// You should use this after another middleware that
// returns an error page to the client, for example martini.Recover().
// The arguments can be any RawData to pass to Bugsnag, most usually
// you'll pass a bugsnag.Configuration object.
func AutoNotify(rawData ...interface{}) martini.Handler {

	// set the release stage based on the martini environment.
	rawData = append([]interface{}{bugsnag.Configuration{ReleaseStage: martini.Env}},
		rawData...)

	return func(r *http.Request, c martini.Context) {

		// create a notifier that has the current request bound to it
		notifier := bugsnag.New(append(rawData, r)...)
		defer notifier.AutoNotify(r)
		c.Map(notifier)
		c.Next()
	}
}
