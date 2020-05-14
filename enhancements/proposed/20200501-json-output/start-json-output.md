# Add JSON Output to Minikube Start

* First proposed: May 12, 2020
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
This feature will allow tools that rely on minikube, such as IDE extensions, to better parse errors and progress logs from STDOUT and STDERR on `minikube start`.
This allows end users to see clear, and ideally actionable, error messages when minikube breaks.

Minikube currently communicates state to users via:
1. Logs via glog (sent to stderr on `minikube start`)
1. Logs to stdout which represent a step on `minikube start` (e.g. "Preparing Kubernetes...")
1. Logs to stdout which don't represent a step on `minikube start`. This can be an unexpected Warning message.
1. Actionable error messages sent to stdout and stderr if minikube detects an issue

This proposal focuses on converting steps 2-4 to JSON.

## Goals

*   All output to stdout is JSON parsable
*   Progress of each step of `minikube start` includes:
    * A name for the current step (e.g. `Preparing Kubernetes`)
    * The number of the current step
    * The expected total number of steps
*   Progress of artifacts as they are being downloaded
*   Unexpected output, like `Warnings`
*   Actionable error messages, which minikube already sends to stdout/stderr, are now parsable (these are the errors from [err_map.go](https://github.com/kubernetes/minikube/blob/master/pkg/minikube/problem/err_map.go))


## Non-Goals

*   Change the way we handle logs via glog
*   Improving error handling in minikube; this is just about how we output errors to users

## Expected Output

Users can specify JSON output via a `--output json` flag. 

There are four types of logs that will be output to stdout, which will be identified by the `Type` field in JSON:

1. **Type: Log** These are regular steps during minikube start, and will include a name, message, and the current step number
1. **Type: Download** This type will be used while downloading artifacts. It will include the artifact name, download progress percentage, and the step number it is associated with.
1. **Type: Warning** These are unexpected warnings that come up during `minikube start`
1. **Type: Error** These are error messages minikube outputs if it detects an error. Sometimes, these error messages include actionable advice.


The JSON structs for each type will look like this:


**Type: Log**

```json
{
  "Name": "Selecting Driver",
  "Message": "Using the hyperkit driver based on user configuration\n",
  "TotalSteps": 9,
  "CurrentStep": 2,
  "Type": "Log"
}
```

**Type: Download**

```json
{
  "Type": "Download",
  "Artifact": "preload.tar.gz",
  "Progress": "10%",
  "CurrentStep": 4,
  "TotalSteps": 9,
}
```

**Type: Warning**
```json
{
  "Message": "Something seems to be wrong....",
  "Type": "Warning"
}
```

**Type: Error**
In the case of actionable error message:

```json
{
  "Type": "Error",
  "ID": "KVM_UNAVAILABLE",
  "Err": {},
  "Advice": "Your host does not support KVM virtualization. Ensure that qemu-kvm is installed, and run 'virt-host-validate' to debug the problem",
  "URL": "http://mikko.repolainen.fi/documents/virtualization-with-kvm",
  "Issues": [
    2991
  ],
  "ShowIssueLink": false
}
```

In the case of an unactionable error message, set the message as whatever error was returned:

```json
{
  "Message": "error provisioning host: Failed to start host: there was an unexpected & unactionable error",
  "Type": "Error"
}
```


`minikube start` output would look something like this:

```
$ minikube start
{"Name":"Minikube Version","Message":"minikube v1.10.0-beta.2 on Darwin 10.14.6","TotalSteps":9,"CurrentStep":1,"Type":"Log"}
{"Name":"Selecting Driver","Message":"Using the hyperkit driver based on existing profile","TotalSteps":9,"CurrentStep":2,"Type":"Log"}
{"Name":"Starting Control Plane","Message":"Starting node minikube in cluster minikube","TotalSteps":9,"CurrentStep":3,"Type":"Log"}
{"Name":"","Message":"Something seems to be wrong....","TotalSteps":0,"CurrentStep":0,"Type":"Warning"}
{"ID":"NON_C_DRIVE","Err":{},"Advice":"Run minikube from the C: drive.","URL":"","Issues":[1574],"ShowIssueLink":false,"Type":"Error"}
```

In this sample, the first three steps are successful steps (Type: Log).
The fourth step is a warning that something seems to be wrong (Type: Warning).
The fifth step is a sample actionable error message (Type: Error).


## Implementation Details

### Type: Log Steps
Since we need to approximate the total number of steps before minikube starts, we need to know the general steps we expect to execute before starting.

I propose creating a registry of logs, which is prefilled with the following steps:

* "Minikube Version"
* "Selecting Driver"
*	"Starting Control Plane"
*	"Download Necessary Artifacts"
*	"Creating Node"
*	"Preparing Kubernetes"
*	"Verifying Kubernetes"
*	"Enabling Addons"
*	"Done"

When a log is called in the code, it will be associated with one of the above steps.

This will allow us to determine which number step we are currently on, and include that information in the output.

_Note: We may skip steps depending on the user's existing setup. For example, we may not need to "Verify Kubernetes" on a soft start. Though current step may jump from Step 5 to Step 8, the total number of expected steps will not change._


This log would be printed at the correct time later in the code by calling a new function, similar to `out.T`:

```go
func Step(stepName string, style StyleEnum, format string, a ...V)
```

In the code itself, this means that this line, which currently looks like this:

```go
out.T(out.Sparkle, `Using the {{.driver}} driver based on existing profile`, out.V{"driver": ds.String()})
```

would now be:

```go
out.Step(out.CreateDriver, out.Sparkle, `Using the {{.driver}} driver based on existing profile`, out.V{"driver": ds.String()})
```

`out.Step` will be responsible for applying the passed in template, and printing out a JSON encoded version of the step.

### Type: Download Steps

To communicate progress on artifacts as they're being downloaded, we want JSON output that looks something like this during download:

```json
{
  "Type": "Download",
  "Artifact": "preload.tar.gz",
  "Progress": "10%",
  "CurrentStep": 4,
}
```

minikube uses the [go-getter](github.com/hashicorp/go-getter) library to download artifacts (preloaded tarballs, ISOs, etc).
Currently, minikube passes in a [DefaultProgressBar](https://github.com/kubernetes/minikube/blob/master/pkg/minikube/download/download.go#L48) to this library, which is used to communicate download progress to the user.

Instead of passing in `DefaultProgressBar` we should be able to write our own object, something like `JSONOutput`, which will print the current progress of the download in the JSON format specified above instead of showing a progress bar in the terminal.

### Type: Warning & Error Steps

If `--output json` is specified, this should be as simple as including code in `out.WarningT` and `out.ErrT` which will convert to JSON and then print the associated log.


#### Testing Plan
Unit tests will cover:
1. That the JSON output of each type of step is correct and parsable

Integration tests will cover:
1. That in the following cases, if `--output json` is specfied, all logs to stdout are correctly in JSON format:
  * Clean start, with no downloaded artifacts
  * Soft start
  * Restart
  * Force an error
  * Force a warning (can use an old --kubernetes-version)
   

## Alternatives Considered

### Cloud Events

I briefly looked into using [Cloud Events](https://github.com/cloudevents/spec) to send events to clients, specifically looking at the [Go SDK](https://github.com/cloudevents/sdk-go).

Pros:
* Standardized way of sending events
* Supported by CNCF

Cons:
* Having minikube set up an HTTP server adds extra complexity to this proposal. If clients can instead parse JSON info from stdout/stderr then that would be the simplest solution. 
