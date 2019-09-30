# Addon Host OS Hooks

First proposed: 2019-09-30
Authors: Josh Woodcock

## Reviewer Priorities

Please review this proposal with the following priorities:

*   Does this fit with minikube's [principles](https://minikube.sigs.k8s.io/docs/concepts/principles/)?
*   Are there other approaches to consider?
*   Could the implementation be made simpler?
*   Are there usability, reliability, or technical debt concerns?

Please leave the above text in your proposal as instructions to the reader.

## Summary

In order to provide a richer and more simplified development experience, 
some addons may be better equipped if they can be integrated with capabilities that run on the Host OS
triggered by lifecycle actions that occur as part of enabling and disabling addons

### Feature use cases
- Enabling and starting a background process that monitors ingress resources in a kubernetes cluster and updates host 
  DNS resolver configurations
- Enabling and starting a background process that monitors the ingress resource in kubernetes cluster and installs 
  SSL certificates for local domains on the host
- Start a background process that could notify an IDE of the minikube ips that are running so that the 
IDE could be used for connecting to services which use the minikube ip address and NodePort to connect to the database
- Better integrations for tools that currently are wrapping minikube like: 
  - https://github.com/mrbobbytables/oidckube
    In this case you have to use this tool to start the minikube instance. What if instead of doing that
    The tool cool start a background process when a addon is enabled that would enable the same features this tool
    provides?
  - https://docs.seldon.io/projects/seldon-core/en/latest/examples/go_example.html
    This tool has to request the minikube ip from the user. What if it didn't have to 
    through a addon which could update a seldon-core addon config on the host os?
  - https://github.com/superbrothers/minikube-ingress-dns/blob/master/README.md
    This tool could be started and stoped automatically when the addon is enabled
  - https://www.mailgun.com/blog/creating-development-environments-with-kubernetes-devgun
    This tool could be started and stopped automatically when the addon is enabled
    rather than building a wrapper for minikube
  - https://www.telepresence.io/tutorials/minikube-vpn
    This tool could be enabled as a minikube addon if it can be installed and started on the host OS
  - https://developers.facebook.com/docs/whatsapp/installation/dev-multiconnect-minikube/
    Instead of asking the developer to provide the minikube ip what if what's app could just identify it
    automatically. What if I had a whats app business API addon that could be installed and 
    seemlessly integrated with a host os hook?
- A desktop UI could be developed for minikube which could be enabled and installed as an addon which would show 
  the different instances, mounts, etc currently configured in minikube
- A background service that helps multiple minikube clusters interact with each other can be developed to run on the Host OS
  This tool could be enabled and started as an addon if these hooks are enabled
- Tools which need to mount code from the Host OS to the minikube cluster can be developed and enabled through enabling
  of an addon rather than as a separate step a developer has to do after using minikube. A potential improvement
  to how skaffold is currently deploying things directly through containers

## Goals

* Make it easy for third party software to be run on the Host OS when a addon is enabled
* A simple solution is implemented that can achieves the following capabilities:
  * Plugin owners are responsible for and can easily identify a Host OS
  * A shell command that is appropriate for that particular host is able to run on the Host OS for the 
    following events
    * pre-enable: An addon is about to be enabled
    * post-enable: An addon has been enabled
    * pre-disable: An addon is about to be disabled
    * post-disable: An addon has been disabled
  * Any program hook can obtain user input like requesting sudo permission to enable and/or start a systemd service

## Non-Goals

* Lifecycle events for addons other than the ones listed above
* Conditional logic for whether or not to run the addon other than information related to obtaining the Host OS and/or version number
* Running commands in the background
* Running commands with administrative privileges eg: sudo without requesting the user to authenticate
* Changing how a addon is enabled or disabled by default

## Design Details

A configuration file `config.yaml` optionally exists within the `deploy/addons/addonname` directory.
The configuration file optionally specifies the hooks it wants to run when the addon is enabled or disabled
Here is an approximate example config: 
```yaml
apiVersion: minikube/v1beta
kind: AddonConfig
spec:
  hooks:
    host:
      # (optional) Runs just before the addon is enabled
      preEnable:
        # A list of potential commands to be run based on the OS and OS version for unix operating systems
        # Based on the order, the first regex to match the output of the getHostOSInfoCommand will be run
        # Only 1 command per hook will be run
        unix:
          filteredCommands:
            # (required) A run command is a command to run in order to obtain some information about the Host OS and OS version
            - getHostOSInfoCommand: uname
              # (required) A regex pattern to match based on the output from the run command
              regex: Darwin
              # (required) The command to run if the regex matches the output of the run command
              command: echo "I am about to be enabled"
            - getHostOSInfoCommand: lsb_release -d
              # (required) A list regex pattern to try to match. Any match will count as a successful match
              regexAny:
                - ^.*Ununtu 18.*$
                - ^.*Red Hat.*$
              command: echo "I am about to be enabled on RedHat or Ubuntu 18"
            - getHostOSInfoCommand: uname
              regex: Linux
              command: echo "I am about to be enabled on Linux"
          # A list of potential commands to be run based on the OS and OS version for windows operating systems
        windows:
          filteredCommands:
            - getHostOSInfoCommand: systeminfo | findstr /B /C:"OS Version"
              # Windows 10 or greater
              regex: ^.*OS Version:[ \t]+([1]+[0-9])\..*$
              command: Echo "I am about to be enabled"
      # (optional) Runs just after the addon is enabled
      postEnable:
        unix:
          # Command to run on all unix operating systems. Overrides filteredCommands
          command: echo "I am enabled on Linux"
        windows:
          # Command to run on all windows operating systems. Overrides filteredCommands
          command: Echo "I am enabled"
      # (optional) Runs just before the addon is disabled
      preDisable:
        unix:
          filteredCommands:
            - getHostOSInfoCommand: uname
              regex: Darwin
              command: |
                echo "I am about to be disabled"
                echo "Seriously its about to happen"
            - getHostOSInfoCommand: lsb_release -d
              regexAny:
                - ^.*Ununtu 18.*$
                - ^.*Red Hat.*$
              command: echo "I am about to be disabled on RedHat or Ubuntu 18"
            - getHostOSInfoCommand: uname
              regex: Linux
              command: echo "I am about to be disabled on Linux"
        windows:
          filteredCommands:
            - getHostOSInfoCommand: systeminfo | findstr /B /C:"OS Version"
              # Windows 10 or greater
              regex: ^.*OS Version:[ \t]+([1]+[0-9])\..*$
              command: Echo "I am about to be disabled"
      # (optional) Runs just before the addon is disabled
      postDisable:
        # Command to run on all unix operating systems. Overrides filteredCommands
        unix:
          command: echo "I am disabled on Linux"
        # Command to run on all windows operating systems. Overrides filteredCommands
        windows:
          command: Echo "I am disabled"
```
This configuration is loaded along with the other addon configurations (like whether or not it should be enabled/disabled by default)
When a addon is disabled, or enabled we check whether or not a configuration exists and which hooks are configured. 

If a "preEnable" hook exits with a status code other than 1 then the addon installation fails with an error message 
printed to the console and logged
If a "postEnable" hook exits with a status code other than 0 then the addon remains enabled with an error message
printed to the console and logged 
If a "preDisable" hook exits with a status code other than 0 then the addon is still disabled with an error message
printed to the console and logged 
If a "preDisable" hook exits with a status code other than 0 then the addon is still disabled with an error message
printed to the console and logged

### Testing Plan 
Write integration test for an `example` addon which will install an example set of kubernetes configurations and which has all hooks configured.
The output of each configuration will be tested and should match the expected result based on the Host OS

Include integration test for failed hooks
Include an integration test for requesting user input. This should be good enough to test that a addon can request to run a hook with 
sudo permission

## Alternatives Considered

### Background service that monitors addon status
Write a wrapper program which runs as a background service and continuously monitors which addons are enabled or disabled by running
minikube commands on the cli

#### Cons of doing this: 
1. Any action that should be taken as hook by said background service and which also must have sudo permission 
   must always have sudo permission. This decreases user security since you will basically have to give a program 
   permission to run as an admin in the backgroundwhich means it can literally do anything it wants to do.
2. The preEnable and preDisable hooks could not exist which means that depending on what the application is doing
it could end up in strange and unintended
3. The background program will have to be started/enabled/stopped separately from any minikube addons which it might 
be integrated with. This creates a very poor user experience for the developer who has to remember to do these things
and most likely will never disable a background process once it is configured to run even though it is no longer
being used

#### Pros of doing this
1. No approval or review process will be required from the minikube development team
2. No development time is spent on testing and developing this feature