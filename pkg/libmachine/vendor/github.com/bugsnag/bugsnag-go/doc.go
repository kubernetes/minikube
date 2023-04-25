/*
Package bugsnag captures errors in real-time and reports them to Bugsnag (http://bugsnag.com).

Using bugsnag-go is a three-step process.

1. As early as possible in your program configure the notifier with your APIKey. This sets up
handling of panics that would otherwise crash your app.

	func init() {
		bugsnag.Configure(bugsnag.Configuration{
			APIKey: "YOUR_API_KEY_HERE",
		})
	}

2. Add bugsnag to places that already catch panics. For example you should add it to the HTTP server
when you call ListenAndServer:

	http.ListenAndServe(":8080", bugsnag.Handler(nil))

If that's not possible, for example because you're using Google App Engine, you can also wrap each
HTTP handler manually:

	http.HandleFunc("/" bugsnag.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		...
	})

3. To notify Bugsnag of an error that is not a panic, pass it to bugsnag.Notify. This will also
log the error message using the configured Logger.

	if err != nil {
		bugsnag.Notify(err)
	}

For detailed integration instructions see https://bugsnag.com/docs/notifiers/go.

Configuration

The only required configuration is the Bugsnag API key which can be obtained by clicking "Settings"
on the top of https://bugsnag.com/ after signing up. We also recommend you set the ReleaseStage
and AppVersion if these make sense for your deployment workflow.

RawData

If you need to attach extra data to Bugsnag notifications you can do that using
the rawData mechanism.  Most of the functions that send errors to Bugsnag allow
you to pass in any number of interface{} values as rawData. The rawData can
consist of the Severity, Context, User or MetaData types listed below, and
there is also builtin support for *http.Requests.

	bugsnag.Notify(err, bugsnag.SeverityError)

If you want to add custom tabs to your bugsnag dashboard you can pass any value in as rawData,
and then process it into the event's metadata using a bugsnag.OnBeforeNotify() hook.

	bugsnag.Notify(err, account)

	bugsnag.OnBeforeNotify(func (e *bugsnag.Event, c *bugsnag.Configuration) {
		for datum := range e.RawData {
			if account, ok := datum.(Account); ok {
				e.MetaData.Add("account", "name", account.Name)
				e.MetaData.Add("account", "url", account.URL)
			}
		}
	})

If necessary you can pass Configuration in as rawData, or modify the Configuration object passed
into OnBeforeNotify hooks. Configuration passed in this way only affects the current notification.
*/
package bugsnag
