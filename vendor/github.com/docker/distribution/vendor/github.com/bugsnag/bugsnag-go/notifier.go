package bugsnag

import (
	"fmt"

	"github.com/bugsnag/bugsnag-go/errors"
)

// Notifier sends errors to Bugsnag.
type Notifier struct {
	Config  *Configuration
	RawData []interface{}
}

// New creates a new notifier.
// You can pass an instance of bugsnag.Configuration in rawData to change the configuration.
// Other values of rawData will be passed to Notify.
func New(rawData ...interface{}) *Notifier {
	config := Config.clone()
	for i, datum := range rawData {
		if c, ok := datum.(Configuration); ok {
			config.update(&c)
			rawData[i] = nil
		}
	}

	return &Notifier{
		Config:  config,
		RawData: rawData,
	}
}

// Notify sends an error to Bugsnag. Any rawData you pass here will be sent to
// Bugsnag after being converted to JSON. e.g. bugsnag.SeverityError, bugsnag.Context,
// or bugsnag.MetaData.
func (notifier *Notifier) Notify(err error, rawData ...interface{}) (e error) {
	event, config := newEvent(errors.New(err, 1), rawData, notifier)

	// Never block, start throwing away errors if we have too many.
	e = middleware.Run(event, config, func() error {
		config.log("notifying bugsnag: %s", event.Message)
		if config.notifyInReleaseStage() {
			if config.Synchronous {
				return (&payload{event, config}).deliver()
			}
			go (&payload{event, config}).deliver()
			return nil
		}
		return fmt.Errorf("not notifying in %s", config.ReleaseStage)
	})

	if e != nil {
		config.log("bugsnag.Notify: %v", e)
	}
	return e
}

// AutoNotify notifies Bugsnag of any panics, then repanics.
// It sends along any rawData that gets passed in.
// Usage: defer AutoNotify()
func (notifier *Notifier) AutoNotify(rawData ...interface{}) {
	if err := recover(); err != nil {
		rawData = notifier.addDefaultSeverity(rawData, SeverityError)
		notifier.Notify(errors.New(err, 2), rawData...)
		panic(err)
	}
}

// Recover logs any panics, then recovers.
// It sends along any rawData that gets passed in.
// Usage: defer Recover()
func (notifier *Notifier) Recover(rawData ...interface{}) {
	if err := recover(); err != nil {
		rawData = notifier.addDefaultSeverity(rawData, SeverityWarning)
		notifier.Notify(errors.New(err, 2), rawData...)
	}
}

func (notifier *Notifier) dontPanic() {
	if err := recover(); err != nil {
		notifier.Config.log("bugsnag/notifier.Notify: panic! %s", err)
	}
}

// Add a severity to raw data only if the default is not set.
func (notifier *Notifier) addDefaultSeverity(rawData []interface{}, s severity) []interface{} {

	for _, datum := range append(notifier.RawData, rawData...) {
		if _, ok := datum.(severity); ok {
			return rawData
		}
	}

	return append(rawData, s)
}
