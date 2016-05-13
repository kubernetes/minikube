package mellow

import (
	"fmt"
	"net/http"
	"os"

	"github.com/bugsnag/bugsnag-go"
)

func init() {
	bugsnag.OnBeforeNotify(func(event *bugsnag.Event, config *bugsnag.Configuration) error {
		event.MetaData.AddStruct("original", event.Error.StackFrames())
		return nil
	})
	bugsnag.Configure(bugsnag.Configuration{
		APIKey: "066f5ad3590596f9aa8d601ea89af845",
	})

	http.HandleFunc("/", bugsnag.HandlerFunc(handler))
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "welcome")
	notifier := bugsnag.New(r)
	notifier.Notify(fmt.Errorf("oh hia"), bugsnag.MetaData{"env": {"values": os.Environ()}})
	fmt.Fprint(w, "welcome\n")

	panic("zoomg")
}
