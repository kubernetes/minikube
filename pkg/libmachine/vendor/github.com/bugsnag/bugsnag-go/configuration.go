package bugsnag

import (
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

// Configuration sets up and customizes communication with the Bugsnag API.
type Configuration struct {
	// Your Bugsnag API key, e.g. "c9d60ae4c7e70c4b6c4ebd3e8056d2b8". You can
	// find this by clicking Settings on https://bugsnag.com/.
	APIKey string
	// The Endpoint to notify about crashes. This defaults to
	// "https://notify.bugsnag.com/", if you're using Bugsnag Enterprise then
	// set it to your internal Bugsnag endpoint.
	Endpoint string

	// The current release stage. This defaults to "production" and is used to
	// filter errors in the Bugsnag dashboard.
	ReleaseStage string
	// The currently running version of the app. This is used to filter errors
	// in the Bugsnag dasboard. If you set this then Bugsnag will only re-open
	// resolved errors if they happen in different app versions.
	AppVersion string
	// The hostname of the current server. This defaults to the return value of
	// os.Hostname() and is graphed in the Bugsnag dashboard.
	Hostname string

	// The Release stages to notify in. If you set this then bugsnag-go will
	// only send notifications to Bugsnag if the ReleaseStage is listed here.
	NotifyReleaseStages []string

	// packages that are part of your app. Bugsnag uses this to determine how
	// to group errors and how to display them on your dashboard. You should
	// include any packages that are part of your app, and exclude libraries
	// and helpers. You can list wildcards here, and they'll be expanded using
	// filepath.Glob. The default value is []string{"main*"}
	ProjectPackages []string

	// Any meta-data that matches these filters will be marked as [REDACTED]
	// before sending a Notification to Bugsnag. It defaults to
	// []string{"password", "secret"} so that request parameters like password,
	// password_confirmation and auth_secret will not be sent to Bugsnag.
	ParamsFilters []string

	// The PanicHandler is used by Bugsnag to catch unhandled panics in your
	// application. The default panicHandler uses mitchellh's panicwrap library,
	// and you can disable this feature by passing an empty: func() {}
	PanicHandler func()

	// The logger that Bugsnag should log to. Uses the same defaults as go's
	// builtin logging package. bugsnag-go logs whenever it notifies Bugsnag
	// of an error, and when any error occurs inside the library itself.
	Logger interface {
		Printf(format string, v ...interface{}) // limited to the functions used
	}
	// The http Transport to use, defaults to the default http Transport. This
	// can be configured if you are in an environment like Google App Engine
	// that has stringent conditions on making http requests.
	Transport http.RoundTripper
	// Whether bugsnag should notify synchronously. This defaults to false which
	// causes bugsnag-go to spawn a new goroutine for each notification.
	Synchronous bool
	// TODO: remember to update the update() function when modifying this struct
}

func (config *Configuration) update(other *Configuration) *Configuration {
	if other.APIKey != "" {
		config.APIKey = other.APIKey
	}
	if other.Endpoint != "" {
		config.Endpoint = other.Endpoint
	}
	if other.Hostname != "" {
		config.Hostname = other.Hostname
	}
	if other.AppVersion != "" {
		config.AppVersion = other.AppVersion
	}
	if other.ReleaseStage != "" {
		config.ReleaseStage = other.ReleaseStage
	}
	if other.ParamsFilters != nil {
		config.ParamsFilters = other.ParamsFilters
	}
	if other.ProjectPackages != nil {
		config.ProjectPackages = other.ProjectPackages
	}
	if other.Logger != nil {
		config.Logger = other.Logger
	}
	if other.NotifyReleaseStages != nil {
		config.NotifyReleaseStages = other.NotifyReleaseStages
	}
	if other.PanicHandler != nil {
		config.PanicHandler = other.PanicHandler
	}
	if other.Transport != nil {
		config.Transport = other.Transport
	}
	if other.Synchronous {
		config.Synchronous = true
	}

	return config
}

func (config *Configuration) merge(other *Configuration) *Configuration {
	return config.clone().update(other)
}

func (config *Configuration) clone() *Configuration {
	clone := *config
	return &clone
}

func (config *Configuration) isProjectPackage(pkg string) bool {
	for _, p := range config.ProjectPackages {
		if d, f := filepath.Split(p); f == "**" {
			if strings.HasPrefix(pkg, d) {
				return true
			}
		}

		if match, _ := filepath.Match(p, pkg); match {
			return true
		}
	}
	return false
}

func (config *Configuration) stripProjectPackages(file string) string {
	for _, p := range config.ProjectPackages {
		if len(p) > 2 && p[len(p)-2] == '/' && p[len(p)-1] == '*' {
			p = p[:len(p)-1]
		} else if p[len(p)-1] == '*' && p[len(p)-2] == '*' {
			p = p[:len(p)-2]
		} else {
			p = p + "/"
		}
		if strings.HasPrefix(file, p) {
			return strings.TrimPrefix(file, p)
		}
	}

	return file
}

func (config *Configuration) logf(fmt string, args ...interface{}) {
	if config != nil && config.Logger != nil {
		config.Logger.Printf(fmt, args...)
	} else {
		log.Printf(fmt, args...)
	}
}

func (config *Configuration) notifyInReleaseStage() bool {
	if config.NotifyReleaseStages == nil {
		return true
	}
	for _, r := range config.NotifyReleaseStages {
		if r == config.ReleaseStage {
			return true
		}
	}
	return false
}
