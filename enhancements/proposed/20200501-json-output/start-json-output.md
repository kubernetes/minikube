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
In addition, we will output error code from [err_map.go](https://github.com/kubernetes/minikube/blob/master/pkg/minikube/problem/err_map.go) as a parsable JSON message.

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
	}
    // Add this function here (this is just pseudocode)
    if json {
        out.DisplayErrorJSON(os.Stderr, p)
    }
    ...
}
```

If JSON output is speciifed, we can call a function to the `out` package, `out.DisplayErrorJSON()


   






_(2+ paragraphs) A short overview of your implementation idea, containing only as much detail as required to convey your idea._

_If you have multiple ideas, list them concisely._

_Include a testing plan to ensure that your enhancement is not broken by future changes._

## Alternatives Considered

_Alternative ideas that you are leaning against._
