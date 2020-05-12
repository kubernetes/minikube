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
1. Logs to stdout which don't represent a step on `minikube start`. This can be an unexpected warning message or an `option`.

This proposal focuses **only** on converting outputs representing clear steps in `minikube start` to JSON (option 2), and making sure error code sent to stderr is parsable.

## Goals

### Stderr
*   Error code from [err_map.go](https://github.com/kubernetes/minikube/blob/master/pkg/minikube/problem/err_map.go) is parsable from stderr

### Stdout
*   Progress of each step of `minikube start` is communicated in JSON and includes:
  1. A name for the current step (e.g. `Preparing Kubernetes`)
  1. The number of the current step
  1. The expected total number of steps
*   Progress of artifacts as they are being downloaded is communicated in JSON



## Non-Goals

*   Change the way we handle logs via glog
*   Change the way we handle non-steps to stdout (Warnings, etc.)
*   Improving error handling in minikube; this is just about how we output errors to users

## Expected Output

Users can specify JSON output via a `--output json` flag. 

### Stderr

If minikube fails, and an actionable error message exists, the following JSON will be printed to stderr:

```json
{
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

### Stdout
If `--output json` is specified, each step of `minikube start` will be output as JSON:

```json
{
  "Name": "Selecting Driver",
  "Message": "✨  Using the hyperkit driver based on user configuration\n",
  "TotalSteps": 9,
  "CurrentStep": 2,
  "Type": "Log"
}
```

and each update on a downloaded artifact will be output as:

```json
{
  "Type": "Download",
  "Artifact": "preload.tar.gz",
  "Progress": "10%"
}
```

`minikube start` output would look something like this:

```
$ minikube start
{"Name":"Minikube Version","Message":"minikube v1.10.0-beta.2 on Darwin 10.14.6","TotalSteps":9,"CurrentStep":1, "Type":"Log"}
{"Name":"Selecting Driver","Message":"Using the hyperkit driver,"TotalSteps":9,"CurrentStep":2,"Type":"Log"}
{"Name":"Starting Control Plane","Message":"Starting node minikube in cluster minikube","TotalSteps":9,"CurrentStep":3,"Type":"Log"}
{"Name":"Download Necessary Artifacts","Message":"Downloading Kubernetes v1.18.1 preload","TotalSteps":9,"CurrentStep":4,"Type":"Log"}
  {"Type":"Download", "Artifact":"preload.tar.gz", "Progress": "10%"}
  {"Type":"Download", "Artifact":"preload.tar.gz", "Progress": "61%"}
  {"Type":"Download", "Artifact":"preload.tar.gz", "Progress": "73%"}
  {"Type":"Download", "Artifact":"preload.tar.gz", "Progress": "87%"}
  {"Type":"Download", "Artifact":"preload.tar.gz", "Progress": "100%"}
{"Name":"Creating Node","Message":"Creating hyperkit VM","TotalSteps":9,"CurrentStep":5,"Type":"Log"}
{"Name":"Preparing Kubernetes","Message":"Preparing Kubernetes v1.18.1 on Docker 19.03.8","TotalSteps":9,"CurrentStep":6,"Type":"Log"}
{"Name":"Verifying Kubernetes","Message":"🔎  Verifying Kubernetes components","TotalSteps":9,"CurrentStep":7,"Type":"Log"}
{"Name":"Enabling Addons","Message":"🌟  Enabled addons: default-storageclass, storage-provisioner","TotalSteps":9,"CurrentStep":7,"Type":"Log"}
{"Name":"Done","Message":"🏄  Done! kubectl is now configured to use \"minikube\"","TotalSteps":9,"CurrentStep":9,"Type":"Log"}
```

This way, clients can parse the output as it is logged and know the following:
1. What the type of step is (Log vs Download)
1. What step we are currently on
1. The total number of steps
1. The specific message related to that step


## Implementation Details

### Stderr
glog logs can be sent to stderr as usual.
In addition, we will output error code from [err_map.go](https://github.com/kubernetes/minikube/blob/master/pkg/minikube/problem/err_map.go) as a parsable JSON message to stderr.

This will be done by adding the following function to the `out` package:

```go
func DisplayErrorJSON(out io.Writer, p *problem.Problem)
```

which will print the JSON encoding of `p` to `out` (in this case, stderr).


### Stdout - Log Steps
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

_Note: We may skip steps depending on the user's existing setup. For example, we may not need to "Verify Kubernetes" on a soft start. The current step/total step numbers will only be approximations_


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

`out.Step` will be responsible for applying the passed in template, and printing out a JSON encoded version of the step if `--output json` is specified:

```json
{
  "Name": "Selecting Driver",
  "Message": "✨  Using the hyperkit driver based on user configuration\n",
  "TotalSteps": 9,
  "CurrentStep": 2,
  "Type": "Log"
}
```

### Stdout - Download Steps

To communicate progress on artifacts as they're being downloaded, we want JSON output that looks something like this during download:

```json
{
  "Type": "Download",
  "Artifact": "preload.tar.gz",
  "Progress": "10%"
}
```

minikube uses the [go-getter](github.com/hashicorp/go-getter) library to download artifacts (preloaded tarballs, ISOs, etc).
Currently, minikube passes in a [DefaultProgressBar](https://github.com/kubernetes/minikube/blob/master/pkg/minikube/download/download.go#L48) to this library, which is used to communicate download progress to the user.

Instead of passing in `DefaultProgressBar` we should be able to write our own object, something like `JSONOutput`, which will print the current progress of the download in the JSON format specified above instead of showing a progress bar in the terminal.


#### Testing Plan
Both unit tests and integration tests will be required to test these features feature.

Unit tests will cover:
1. That the JSON output of output steps, both type `Log` and type `Download`, is correct and parsable
1. That errors are sent to stderr correctly and are parsable

Integration tests will cover:
1. That in the following cases, if `--output json` is specfied, all logs are correctly in JSON format:
  * Clean start, with no downloaded artifacts
  * Soft start
  * Restart
   

## Alternatives Considered

### Cloud Events

I briefly looked into using [Cloud Events](https://github.com/cloudevents/spec) to send events to clients, specifically looking at the [Go SDK](https://github.com/cloudevents/sdk-go).

Pros:
* Standardized way of sending events
* Supported by CNCF

Cons:
* Having minikube set up an HTTP server adds extra complexity to this proposal. If clients can instead parse JSON info from stdout/stderr then that would be the simplest solution. 
