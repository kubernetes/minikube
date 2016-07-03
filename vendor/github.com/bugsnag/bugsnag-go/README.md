Bugsnag Notifier for Golang
===========================

The Bugsnag Notifier for Golang gives you instant notification of panics, or
unexpected errors, in your golang app. Any unhandled panics will trigger a
notification to be sent to your Bugsnag project.

[Bugsnag](http://bugsnag.com) captures errors in real-time from your web,
mobile and desktop applications, helping you to understand and resolve them
as fast as possible. [Create a free account](http://bugsnag.com) to start
capturing exceptions from your applications.

## How to Install

1. Download the code

    ```shell
    go get github.com/bugsnag/bugsnag-go
    ```

### Using with net/http apps

For a golang app based on [net/http](https://godoc.org/net/http), integrating
Bugsnag takes two steps. You should also use these instructions if you're using
the [gorilla toolkit](http://www.gorillatoolkit.org/), or the
[pat](https://github.com/bmizerany/pat/) muxer.

1. Configure bugsnag at the start of your `main()` function:

    ```go
    import "github.com/bugsnag/bugsnag-go"

    func main() {
        bugsnag.Configure(bugsnag.Configuration{
            APIKey: "YOUR_API_KEY_HERE",
            ReleaseStage: "production",
            // more configuration options
        })

        // rest of your program.
    }
    ```

2. Wrap your server in a [bugsnag.Handler](https://godoc.org/github.com/bugsnag/bugsnag-go/#Handler)

    ```go
    // a. If you're using the builtin http mux, you can just pass
    //    bugsnag.Handler(nil) to http.ListenAndServer
    http.ListenAndServe(":8080", bugsnag.Handler(nil))

    // b. If you're creating a server manually yourself, you can set
    //    its handlers the same way
    srv := http.Server{
        Handler: bugsnag.Handler(nil)
    }

    // c. If you're not using the builtin http mux, wrap your own handler
    // (though make sure that it doesn't already catch panics)
    http.ListenAndServe(":8080", bugsnag.Handler(handler))
    ```

### Using with Revel apps

There are two steps to get panic handling in [revel](https://revel.github.io) apps.

1. Add the `bugsnagrevel.Filter` immediately after the `revel.PanicFilter` in `app/init.go`:

    ```go

    import "github.com/bugsnag/bugsnag-go/revel"

    revel.Filters = []revel.Filter{
        revel.PanicFilter,
        bugsnagrevel.Filter,
        // ...
    }
    ```

2. Set bugsnag.apikey in the top section of `conf/app.conf`.

    ```
    module.static=github.com/revel/revel/modules/static

    bugsnag.apikey=YOUR_API_KEY_HERE

    [dev]
    ```

### Using with Google App Engine

1. Configure bugsnag at the start of your `init()` function:

    ```go
    import "github.com/bugsnag/bugsnag-go"

    func init() {
        bugsnag.Configure(bugsnag.Configuration{
            APIKey: "YOUR_API_KEY_HERE",
        })

        // ...
    }
    ```

2. Wrap *every* http.Handler or http.HandlerFunc with Bugsnag:

    ```go
    // a. If you're using HandlerFuncs
    http.HandleFunc("/", bugsnag.HandlerFunc(
        func (w http.ResponseWriter, r *http.Request) {
            // ...
        }))

    // b. If you're using Handlers
    http.Handle("/", bugsnag.Handler(myHttpHandler))
    ```

3. In order to use Bugsnag, you must provide the current
[`appengine.Context`](https://developers.google.com/appengine/docs/go/reference#Context), or
current `*http.Request` as rawData (This is done automatically for `bugsnag.Handler` and `bugsnag.HandlerFunc`).
The easiest way to do this is to create a new instance of the notifier.

    ```go
    c := appengine.NewContext(r)
    notifier := bugsnag.New(c)

    if err != nil {
        notifier.Notify(err)
    }

    go func () {
        defer notifier.Recover()

        // ...
    }()
    ```


## Notifying Bugsnag manually

Bugsnag will automatically handle any panics that crash your program and notify
you of them. If you've integrated with `revel` or `net/http`, then you'll also
be notified of any panics() that happen while processing a request.

Sometimes however it's useful to manually notify Bugsnag of a problem. To do this,
call [`bugsnag.Notify()`](https://godoc.org/github.com/bugsnag/bugsnag-go/#Notify)

```go
if err != nil {
    bugsnag.Notify(err)
}
```

### Manual panic handling

To avoid a panic in a goroutine from crashing your entire app, you can use
[`bugsnag.Recover()`](https://godoc.org/github.com/bugsnag/bugsnag-go/#Recover)
to stop a panic from unwinding the stack any further. When `Recover()` is hit,
it will send any current panic to Bugsnag and then stop panicking. This is
most useful at the start of a goroutine:

```go
go func() {
    defer bugsnag.Recover()

    // ...
}()
```

Alternatively you can use
[`bugsnag.AutoNotify()`](https://godoc.org/github.com/bugsnag/bugsnag-go/#Recover)
to notify bugsnag of a panic while letting the program continue to panic. This
is useful if you're using a Framework that already has some handling of panics
and you are retrofitting bugsnag support.

```go
defer bugsnag.AutoNotify()
```

## Sending Custom Data

Most functions in the Bugsnag API, including `bugsnag.Notify()`,
`bugsnag.Recover()`, `bugsnag.AutoNotify()`, and `bugsnag.Handler()` let you
attach data to the notifications that they send. To do this you pass in rawData,
which can be any of the supported types listed here. To add support for more
types of rawData see [OnBeforeNotify](#custom-data-with-onbeforenotify).

### Custom MetaData

Custom metaData appears as tabs on Bugsnag.com. You can set it by passing
a [`bugsnag.MetaData`](https://godoc.org/github.com/bugsnag/bugsnag-go/#MetaData)
object as rawData.

```go
bugsnag.Notify(err,
    bugsnag.MetaData{
        "Account": {
            "Name": Account.Name,
            "Paying": Account.Plan.Premium,
        },
    })
```

### Request data

Bugsnag can extract interesting data from
[`*http.Request`](https://godoc.org/net/http/#Request) objects, and
[`*revel.Controller`](https://godoc.org/github.com/revel/revel/#Controller)
objects. These are automatically passed in when handling panics, and you can
pass them yourself.

```go
func (w http.ResponseWriter, r *http.Request) {
    bugsnag.Notify(err, r)
}
```

### User data

User data is searchable, and the `Id` powers the count of users affected. You
can set which user an error affects by passing a
[`bugsnag.User`](https://godoc.org/github.com/bugsnag/bugsnag-go/#User) object as
rawData.

```go
bugsnag.Notify(err,
    bugsnag.User{Id: "1234", Name: "Conrad", Email: "me@cirw.in"})
```

### Error Class

Errors in your Bugsnag dashboard are grouped by their "error class" and by line number.
You can override the error class by passing a
[`bugsnag.ErrorClass`](https://godoc.org/github.com/bugsnag/bugsnag-go/#ErrorClass) object as
rawData.

```go
bugsnag.Notify(err, bugsnag.ErrorClass{"I/O Timeout"})
```

### Context

The context shows up prominently in the list view so that you can get an idea
of where a problem occurred. You can set it by passing a
[`bugsnag.Context`](https://godoc.org/github.com/bugsnag/bugsnag-go/#Context)
object as rawData.

```go
bugsnag.Notify(err, bugsnag.Context{"backgroundJob"})
```

### Severity

Bugsnag supports three severities, `SeverityError`, `SeverityWarning`, and `SeverityInfo`.
You can set the severity of an error by passing one of these objects as rawData.

```go
bugsnag.Notify(err, bugsnag.SeverityInfo)
```

## Configuration

You must call `bugsnag.Configure()` at the start of your program to use Bugsnag, you pass it
a [`bugsnag.Configuration`](https://godoc.org/github.com/bugsnag/bugsnag-go/#Configuration) object
containing any of the following values.

### APIKey

The Bugsnag API key can be found on your [Bugsnag dashboard](https://bugsnag.com) under "Settings".

```go
bugsnag.Configure(bugsnag.Configuration{
    APIKey: "YOUR_API_KEY_HERE",
})
```

### Endpoint

The Bugsnag endpoint defaults to `https://notify.bugsnag.com/`. If you're using Bugsnag enterprise,
you should set this to the endpoint of your local instance.

```go
bugsnag.Configure(bugsnag.Configuration{
    Endpoint: "http://bugsnag.internal:49000/",
})
```

### ReleaseStage

The ReleaseStage tracks where your app is deployed. You should set this to `production`, `staging`,
`development` or similar as appropriate.

```go
bugsnag.Configure(bugsnag.Configuration{
    ReleaseStage: "development",
})
```

### NotifyReleaseStages

The list of ReleaseStages to notify in. By default Bugsnag will notify you in all release stages, but
you can use this to silence development errors.

```go
bugsnag.Configure(bugsnag.Configuration{
    NotifyReleaseStages: []string{"production", "staging"},
})
```

### AppVersion

If you use a versioning scheme for deploys of your app, Bugsnag can use the `AppVersion` to only
re-open errors if they occur in later version of the app.

```go
bugsnag.Configure(bugsnag.Configuration{
    AppVersion: "1.2.3",
})
```

### Hostname

The hostname is used to track where exceptions are coming from in the Bugsnag dashboard. The
default value is obtained from `os.Hostname()` so you won't often need to change this.

```go
bugsnag.Configure(bugsnag.Configuration{
    Hostname: "go1",
})
```

### ProjectPackages

In order to determine where a crash happens Bugsnag needs to know which packages you consider to
be part of your app (as opposed to a library). By default this is set to `[]string{"main*"}`. Strings
are matched to package names using [`filepath.Match`](http://godoc.org/path/filepath#Match).

```go
bugsnag.Configure(bugsnag.Configuration{
    ProjectPackages: []string{"main", "github.com/domain/myapp/*"},
}
```

### ParamsFilters

Sometimes sensitive data is accidentally included in Bugsnag MetaData. You can remove it by
setting `ParamsFilters`. Any key in the `MetaData` that includes any string in the filters
will be redacted. The default is `[]string{"password", "secret"}`, which prevents fields like
`password`, `password_confirmation` and `secret_answer` from being sent.

```go
bugsnag.Configure(bugsnag.Configuration{
    ParamsFilters: []string{"password", "secret"},
}
```

### Logger

The Logger to write to in case of an error inside Bugsnag. This defaults to the global logger.

```go
bugsnag.Configure(bugsnag.Configuration{
    Logger: app.Logger,
}
```

### PanicHandler

The first time Bugsnag is configured, it wraps the running program in a panic
handler using [panicwrap](http://godoc.org/github.com/ConradIrwin/panicwrap). This
forks a sub-process which monitors unhandled panics. To prevent this, set
`PanicHandler` to `func() {}` the first time you call
`bugsnag.Configure`. This will prevent bugsnag from being able to notify you about
unhandled panics.

```go
bugsnag.Configure(bugsnag.Configuration{
    PanicHandler: func() {},
})
```

### Synchronous

Bugsnag usually starts a new goroutine before sending notifications. This means
that notifications can be lost if you do a bugsnag.Notify and then immediately
os.Exit. To avoid this problem, set Bugsnag to Synchronous (or just `panic()`
instead ;).

```go
bugsnag.Configure(bugsnag.Configuration{
    Synchronous: true
})
```

Or just for one error:

```go
bugsnag.Notify(err, bugsnag.Configuration{Synchronous: true})
```

### Transport

The transport configures how Bugsnag makes http requests. By default we use
[`http.DefaultTransport`](http://godoc.org/net/http#RoundTripper) which handles
HTTP proxies automatically using the `$HTTP_PROXY` environment variable.

```go
bugsnag.Configure(bugsnag.Configuration{
    Transport: http.DefaultTransport,
})
```

## Custom data with OnBeforeNotify

While it's nice that you can pass `MetaData` directly into `bugsnag.Notify`,
`bugsnag.AutoNotify`, and `bugsnag.Recover`, this can be a bit cumbersome and
inefficient — you're constructing the meta-data whether or not it will actually
be used.  A better idea is to pass raw data in to these functions, and add an
`OnBeforeNotify` filter that converts them into `MetaData`.

For example, lets say our system processes jobs:

```go
type Job struct{
    Retry     bool
    UserId    string
    UserEmail string
    Name      string
    Params    map[string]string
}
```

You can pass a job directly into Bugsnag.notify:

```go
bugsnag.Notify(err, job)
```

And then add a filter to extract information from that job and attach it to the
Bugsnag event:

```go
bugsnag.OnBeforeNotify(
    func(event *bugsnag.Event, config *bugsnag.Configuration) error {

        // Search all the RawData for any *Job pointers that we're passed in
        // to bugsnag.Notify() and friends.
        for _, datum := range event.RawData {
            if job, ok := datum.(*Job); ok {
                // don't notify bugsnag about errors in retries
                if job.Retry {
                    return fmt.Errorf("not notifying about retried jobs")
                }

                // add the job as a tab on Bugsnag.com
                event.MetaData.AddStruct("Job", job)

                // set the user correctly
                event.User = &User{Id: job.UserId, Email: job.UserEmail}
            }
        }

        // continue notifying as normal
        return nil
    })
```

## Advanced Usage

If you want to have multiple different configurations around in one program,
you can use `bugsnag.New()` to create multiple independent instances of
Bugsnag. You can use these without calling `bugsnag.Configure()`, but bear in
mind that until you call `bugsnag.Configure()` unhandled panics will not be
sent to bugsnag.

```go
notifier := bugsnag.New(bugsnag.Configuration{
    APIKey: "YOUR_OTHER_API_KEY",
})
```

In fact any place that lets you pass in `rawData` also allows you to pass in
configuration.  For example to send http errors to one bugsnag project, you
could do:

```go
bugsnag.Handler(nil, bugsnag.Configuration{APIKey: "YOUR_OTHER_API_KEY"})
```

### GroupingHash

If you need to override Bugsnag's grouping algorithm, you can set the
`GroupingHash` in an `OnBeforeNotify`:

```go
bugsnag.OnBeforeNotify(
    func (event *bugsnag.Event, config *bugsnag.Configuration) error {
        event.GroupingHash = calculateGroupingHash(event)
        return nil
    })
```

### Skipping lines in stacktrace

If you have your own logging wrapper all of your errors will appear to
originate from inside it. You can avoid this problem by constructing
an error with a stacktrace manually, and then passing that to Bugsnag.notify:

```go
import (
    "github.com/bugsnag/bugsnag-go"
    "github.com/bugsnag/bugsnag-go/errors"
)

func LogError(e error) {
    // 1 removes one line of stacktrace, so the caller of LogError
    // will be at the top.
    e = errors.New(e, 1)
    bugsnag.Notify(e)
}
```

