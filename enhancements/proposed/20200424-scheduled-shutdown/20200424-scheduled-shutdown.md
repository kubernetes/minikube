# Scheduled shutdown and pause

* First proposed: 2020-04-20
* Authors: Thomas Stromberg (@tstromberg)
  
## Reviewer Priorities

Please review this proposal with the following priorities:

* Does this fit with minikube's [principles](https://minikube.sigs.k8s.io/docs/concepts/principles/)?
* Are there other approaches to consider?
* Could the implementation be made simpler?
* Are there usability, reliability, or technical debt concerns?

Please leave the above text in your proposal as instructions to the reader.

## Summary

Add the ability to schedule a future shutdown or pause event.

This is useful for two sets of users:

* command-lines which interact with a minikube on per-invocation basis. These command-line tools may not be aware of when the final invocation is, but would like low-latency between-commands
* IDE's which start minikube on an as-needed basis. Not all IDE's have the ability to trigger exit hooks when closed.

## Goals

* The ability to schedule a pause or shutdown event
* The ability to defer or update the scheduled event
* "minikube start" transparently clears pending scheduled events

## Non-Goals

* Automatically idle detection. This is a related, but complimentary idea.

## Design Details

### Proposed interface:

* `minikube pause --after 5m`
* `minikube stop --after 5m`

* Each scheduled pause would overwrite the previous scheduled event.
* Each scheduled stop would overwrite the previous scheduled event.
* Each call to `minikube start` would clear scheduled events

As a `keep-alive` implementation, tools will repeat the command to reset the clock, and move the event 5 minutes into the future.

### Implementation idea #1: host-based

* If `--schedule` is used, the minikube command will daemonize, storing a pid in a well-known location, such as `$HOME/.minikube/profiles/<name>/scheduled_pause.pid`.
* If the pid already exists, the previous process will be killed, cancelling the scheduled event.

Advantages:

* Able to re-use all of the existing `pause` and `stop` implementation within minikube.
* Built-in handling for multiple architectures
* Does not consume memory reserved for the VM

Disadvantages:

* Runs a background task on the host
* Daemonization may require different handling on Windows

### Implementation idea #2: guest-based

minikube would connect to the control-plane via SSH, and run the equivalent of:

```shell
killall minikube-schedule
sleep 300

for node in $other-nodes; do
  ssh $node halt
done
halt
```

Advantages:

* Consistent execution environment

Disadvantages:

* Requires creation of a helper binary that runs within the VM
* Untested: some drivers may not fully release resources if shutdown from inside the VM
