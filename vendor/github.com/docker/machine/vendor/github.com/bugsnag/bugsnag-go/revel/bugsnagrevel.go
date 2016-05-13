// Package bugsnagrevel adds Bugsnag to revel.
// It lets you pass *revel.Controller into bugsnag.Notify(),
// and provides a Filter to catch errors.
package bugsnagrevel

import (
	"strings"
	"sync"

	"github.com/bugsnag/bugsnag-go"
	"github.com/revel/revel"
)

var once sync.Once

// Filter should be added to the filter chain just after the PanicFilter.
// It sends errors to Bugsnag automatically. Configuration is read out of
// conf/app.conf, you should set bugsnag.apikey, and can also set
// bugsnag.endpoint, bugsnag.releasestage, bugsnag.appversion,
// bugsnag.projectroot, bugsnag.projectpackages if needed.
func Filter(c *revel.Controller, fc []revel.Filter) {
	defer bugsnag.AutoNotify(c)
	fc[0](c, fc[1:])
}

// Add support to bugsnag for reading data out of *revel.Controllers
func middleware(event *bugsnag.Event, config *bugsnag.Configuration) error {
	for _, datum := range event.RawData {
		if controller, ok := datum.(*revel.Controller); ok {
			// make the request visible to the builtin HttpMIddleware
			event.RawData = append(event.RawData, controller.Request.Request)
			event.Context = controller.Action
			event.MetaData.AddStruct("Session", controller.Session)
		}
	}

	return nil
}

func init() {
	revel.OnAppStart(func() {
		bugsnag.OnBeforeNotify(middleware)

		var projectPackages []string
		if packages, ok := revel.Config.String("bugsnag.projectpackages"); ok {
			projectPackages = strings.Split(packages, ",")
		} else {
			projectPackages = []string{revel.ImportPath + "/app/*", revel.ImportPath + "/app"}
		}

		bugsnag.Configure(bugsnag.Configuration{
			APIKey:          revel.Config.StringDefault("bugsnag.apikey", ""),
			Endpoint:        revel.Config.StringDefault("bugsnag.endpoint", ""),
			AppVersion:      revel.Config.StringDefault("bugsnag.appversion", ""),
			ReleaseStage:    revel.Config.StringDefault("bugsnag.releasestage", revel.RunMode),
			ProjectPackages: projectPackages,
			Logger:          revel.ERROR,
		})
	})
}
