# Add JSON Output to Minikube Start

* First proposed: May 1, 2020
* Authors: Priya Wadhwa (@priyawadhwa)

## Reviewer Priorities

Please review this proposal with the following priorities:

*   Does this fit with minikube's [principles](https://minikube.sigs.k8s.io/docs/concepts/principles/)?
*   Are there other approaches to consider?
*   Could the implementation be made simpler?
*   Are there usability, reliability, or technical debt concerns?

Please leave the above text in your proposal as instructions to the reader.

## Summary

This proposal discusses adding JSON output to `minikube start`. 
This feature will allow tools that rely on minikube, such as IDE extensions, to better parse errors and progress logs from `minikube start`.
This allows end users to see clear, and ideally actionable, error messages when minikube breaks.



## Goals

### Stderr
*   Error code from [err_map.go](https://github.com/kubernetes/minikube/blob/master/pkg/minikube/problem/err_map.go) is parsable from stderr
*   In case of a panic, default to sending all logs to stderr so that the user can see all logs

### Stdout
*   Progress is communicated with the total numbers of steps and the current step the user is on
*   An encoded step name, like `pull_images`, would be useful here


## Non-Goals

*   Improving error handling in minikube; this is just about how we output errors to users

## Design Details
Users can specify JSON output on minikube start via a flag:

```
minikube start --output json
```

### Stderr
Logs can be sent to stderr as usual.
In addition, we will output error code from [err_map.go](https://github.com/kubernetes/minikube/blob/master/pkg/minikube/problem/err_map.go) as a parsable JSON message to stderr.

This will be done by adding the following function to the `out` package:

```go
func DisplayErrorJSON(out io.Writer, p *problem.Problem)
```

which will print the JSON encoding of `p` to `out`.

If JSON output is specified, and minikube finds an applicable error message in `err_map.go`, then we can call `DisplayErrorJSON` when handling the error in the [WithError](https://github.com/kubernetes/minikube/blob/master/pkg/minikube/exit/exit.go#L57) function:

```go
// WithError outputs an error and exits.
func WithError(msg string, err error) {
	...
	p := problem.FromError(err, runtime.GOOS)
	if p != nil {
        WithProblem(msg, err, p)
        // Add this function here (this is just pseudocode)
        if json {
            out.DisplayErrorJSON(os.Stderr, p)
        }
	}
    ...
}
```

#### Testing Plan
This feature will be covered by unit tests exclusively.
If a problem is found, and json is specified, then we just want to make sure that the JSON output is parsable.

### Stdout
As mentioned above, the requirements for stdout are:
1. Steps are numbered, and the total number of steps is known
1. Steps have a helpful name

First, we'll need some way of distinguishing between logs.
Each log has three components:
1. The message itself 
1. The emoji associated with the log (StyleEnum)
1. Whether the log is of type Log, Warning, or Error

```go
type Log struct {
	LogType // Will be either Log, Warning or Error
	style   StyleEnum
	message string
}
```


We also need to know the total number of steps before minikube starts.
For that to be possible, we need to know all of the logs we will print before starting.

I propose creating a registry of logs, which is prefilled with all necessary logs before we start minikube.
The registry would look like this:

```go
// Registry holds all user-facing logs
type Registry struct {
	// maps the name of the log to a Log type
	Logs  map[string]Log
	Index int
}
```

at the beginning of `minikube start`, we would initialize a `Registry` type in the `out` package, which would exist as a global variable.
All logs would be added to the registry at this time via `Register`:

```go
// Register registers a log
func Register(name string, style StyleEnum, message string, logType LogType) {
	registry.Logs[name] = log{
		style:   style,
		message: message,
		LogType: logType,
	}
}
```

and before starting minikube we would run `initializeRegistry()`, which will register all logs we expect to output:

```go
func initializeRegistry()
    Register("select_driver", out.Sparkle, `Using the {{.driver}} driver based on existing profile`, Log)
    // Add all other logs here as well

```

This log would be printed at the correct time later in the code by calling:

```go
func Print(name string, a ...V)
```

which would find the correct log in the registry and apply the template to it.

In the code itself, this means that this line, which currently looks like this:

```go
out.T(out.Sparkle, `Using the {{.driver}} driver based on existing profile`, out.V{"driver": ds.String()})
```

would now be:

```go
out.T("select_driver", out.V{"driver": ds.String()})
```

We can use the `Index` field in `Registry` to track which number log we are currently at. 
Since logs have been pre-registered, we know what the total number of expected logs is.

Similarly to stderr, if the JSON flag is specified, we will print the JSON encoding of the `Log` struct to stdout instead of the expected log in `out.T`, `out.Warning` and `out.Err`.


#### Testing Plan
Both unit tests and integration tests will be required to test this feature.

Unit tests will cover:
1. That the JSON output is correct and parsable

Integration tests will cover:
1. That all output to stdout is in JSON, ensuring that all user-facing logs have been registered
   

## Alternatives Considered

I haven't been able to think of an alternate way to do this just yet.
