# Add disk-usage command

* First proposed: 2021-11-16
* Authors: Marcus Puckett (@mpuckett159)

## Reviewer Priorities

Please review this proposal with the following priorities:

*   Does this fit with minikube's [principles](https://minikube.sigs.k8s.io/docs/concepts/principles/)?
*   Are there other approaches to consider?
*   Could the implementation be made simpler?
*   Are there usability, reliability, or technical debt concerns?

Please leave the above text in your proposal as instructions to the reader.

## Summary

[Source Issue](https://github.com/kubernetes/minikube/issues/12967)

The above issue proposed the implementation of a `disk-usage` command that would allow users to see how much disk space was being utilized by minikube. There are two sets of data here:
 * host disk usage
 * node disk usage

These two data sets are differentiated by what system they are describing. Both are important metrics to be able to track and see.

Host disk usage monitors the amount of disk consumption that minikube is doing at the host level. This would be, by default, the amount of disk space consumed by the `~/.minikube` folder. This can be easily viewed on Linux machines by running `du -sh ~/.minikube`. Implementing this level of disk consumption is relatively trivial in Go as we can just walk the directory and sum all the file sizes.

Node disk usage is the amount of disk consumed within the node VM/container. This data is more complicated to gather because we will need to support many different hypervisors, each with their own unique API. Some may provide this information easily and some may not. We will need to fall back on SSH and string parsing to ensure that this feature has full coverage of all the use cases available. We can start by implementing the SSH commands, and then implement hypervisor specific commands individually until full coverage is complete.

Another challenge is that some (most?) hypervisors do not provide in-guest disk metrics natively, and no matter what we will be relying on ssh command output text parsing to determine in guest disk utilization for some of these drivers. It does not appear that the node describe API will be usable either as it only returns the capacity of the system resources. The more research I've done the more I'm seeing that this will likely be getting done via ssh out string parsing, except for the KVM driver.

## Goals

* Allow users to easily view system disk consumption at both the node and host level.

## Non-Goals

* Other system resource consumption metrics (e.g. RAM or CPU consumption)

## Design Details

As per the linked issue, the new command should be named `disk-usage`.

The command will allow users to view disk consumption in both bytes and "human readable" conversions, e.g. 1000000000B and 1GB.

There should be both compact and verbose views. For the host disk usage this will result in either only displaying the total disk usage of the `~/.minikube` directory, vs displaying a breakdown of the child directories of root `.minikube` directory. It may also be useful here to display any sufficiently large sub-directories that are not necessarily a direct child of the `.minikube` directory.

Verbose Example:
```
❯ minikube disk-usage
    ▪ .minikube - 23.65GB
    ▪ addons - 0B
    ▪ bin - 11.43MB
    ▪ cache - 3.276GB
    ▪ certs - 8.272kB
    ▪ config - 2B
    ▪ files - 3.416kB
    ▪ logs - 175.9kB
    ▪ machines - 20.37GB
    ▪ profiles - 72.45kB
```

Compact Example:
```
❯ minikube disk-usage -c
    ▪ .minikube - 23.65GB
```
NOTE: Not married to the flag example or output styles here just an example.

The default usage should display both host and node consumption details, with optional flags to specify showing one or the other.

## Alternatives Considered

People could always just use `minikube ssh -- df -h /var/lib/docker` (and any other resource monitor related commands) to get node resource information. After doing more research it doesn't look like there are many options besides just creating a wrapper around this command.