# Changelog

# 0.7.0 (2016-4-13)

General
- `DRIVER` environment variable now supported to supply value for `create --driver` flag
- Update to Go 1.6.1
- SSH client has been refactored
- RC versions of Machine will now create and upgrade to boot2docker RCs instead
  of stable versions if available

Drivers
- `azure`
    - Driver has been completely re-written to use resource templates and a significantly easier-to-use authentication model
- `digitalocean`
    - New `--digitalocean-ssh-key-fingerprint` for using existing SSH keys instead of creating new ones
- `virtualbox`
    - Fix issue with `bootlocal.sh`
    - New `--virtualbox-nictype` flag to set driver for NAT network
    - More robust host-only interface collision detection
    - Add support for running VirtualBox on a Windows 32 bit host
    - Change default DNS passthrough handling
- `amazonec2`
    - Specifying multiple security groups to use is now supported
- `exoscale`
    - Add support for user-data
- `hyperv`
    - Machines can now be created by a non-administrator
- `rackspace`
    - New `--rackspace-active-timeout` parameter
- `vmwarefusion`
    - Bind mount shared folder directory by default
- `google`
    - New `--google-use-internal-ip-only` parameter

Provisioners
- General
    - Support for specifying Docker engine port in some cases
- CentOS
    - Now defaults to using upstream `get.docker.com` script instead of custom RPMs.
- boot2docker
    - More robust eth* interface detection
- Swarm
    - Add `--swarm-experimental` parameter to enable experimental Swarm features


# 0.6.0 (2016-02-04)

+ Fix SSH wait before provisioning issue

# 0.6.0-rc4 (2016-02-03)

General

+ `env`
    + Fix shell auto detection

Drivers

+ `exoscale`
    + Fix configuration of exoscale endpoint

# 0.6.0-rc3 (2016-02-01)

- Exit with code 3 if error is during pre-create check

# 0.6.0-rc2 (2016-01-28)

- Fix issue creating Swarms
- Fix `ls` header issue
- Add code to wait for Docker daemon before returning from `start` / `restart`
- Start porting integration tests to Go from BATS
- Add Appveyor for Windows tests
- Update CoreOS provisioner to use `docker daemon`
- Various documentation and error message fixes
- Add ability to create GCE machine using existing VM

# 0.6.0-rc1 (2016-01-18)

General

- Update to Go 1.5.3
- Short form of command invocations is now supported
    - `docker-machine start`, `docker-machine stop` and others will now use
      `default` as the machine name argument if one is not specified
- Fix issue with panics in drivers
- Machine now returns exit code 3 if the pre-create check fails.
    - This is potentially useful for scripting `docker-machine`.
- `docker-machine provision` command added to allow re-running of provisioning
  on instances.
    - This allows users to re-run provisioning if it fails during `create`
      instead of needing to completely start over.

Provisioning

- Most provisioners now use `docker daemon` instead of `docker -d`
- Swarm masters now run with replication enabled
- If `/var/lib` is a BTRFS partition, `btrfs` will now be used as the storage
  driver for the instance

Drivers

- Amazon EC2
    - Default VPC will be used automatically if none is specified
    - Credentials are now be read from the conventional `~/.aws/credentials`
      file automatically
    - Fix a few issues such as nil pointer dereferences
- VMware Fusion
    - Try to get IP from multiple DHCP lease files
- OpenStack
    - Only derive tenant ID if tenant name is supplied

# 0.5.6 (2016-01-11)

General

- `create`
  - Set swarm master to advertise on port 3376
  - Fix swarm restart policy
  - Stop asking for ssh key passwords interactively
- `env`
  - Improve documentation
  - Fix bash on windows
  - Automatic shell detection on Windows
- `help`
  - Don't show the full path to `docker-machine.exe` on windows
- `ls`
  - Allow custom format
  - Improve documentation
- `restart`
  - Improve documentation
- `rm`
  - Improve documentation
  - Better user experience when removing multiple hosts
- `version`
  - Don't show the full path to `docker-machine.exe` on windows
- `start`, `stop`, `restart`, `kill`
  - Better logs and homogeneous behaviour across all drivers

Build

- Introduce CI tests for external binary compatibility
- Add amazon EC2 integration test

Misc

- Improve BugSnags reports: better shell detection, better windows version detection
- Update DockerClient dependency
- Improve bash-completion script
- Improve documentation for bash-completion

Drivers

- Amazon EC2
  - Improve documentation
  - Support optional tags
  - Option to create EbsOptimized instances
- Google
  - Fix remove when instance is stopped
- Openstack
  - Flags to import and reuse existing nova keypairs
- VirtualBox
  - Fix multiple bugs related to host-only adapters
  - Retry commands when `VBoxManage` is not ready
  - Reject VirtualBox versions older that 4.3
  - Fail with a clear message when Hyper-v installation prevents VirtualBox from working
  - Print a warning for Boot2Docker v1.9.1, which is known to have an issue with AUFS
- Vmware Fusion
  - Support soft links in VM paths

Libmachine

- Fix code sample that uses libmachine
- libmachine can be used in external applications


# 0.5.5 (2015-12-28)

General

- `env`
  - Better error message if swarm is down
  - Add quotes to command if there are spaces in the path
  - Fix Powershell env hints
  - Default to cmd shell on windows
  - Detect fish shell
- `scp`
  - Ignore empty ssh key
- `stop`, `start`, `kill`
  - Add feedback to the user
- `rm`
  - Now works when `config.json` is not found
- `ssh`
  - Disable ControlPath
  - Log which SSH client is used
- `ls`
  - Listing is now faster by reducing calls to the driver
  - Shows if the active machine is a swarm cluster

Build

- Automate 90% of the release process
- Upgrade to Go 1.5.2
- Don't build 32bits binaries for Linux and OSX
- Prevent makefile from defaulting to using containers

Misc

- Update docker-machine version
- Updated the bash completion with new options added
- Bugsnag: Retrieve windows version on non-english OS

Drivers

- Amazon EC2
  - Convert API calls to official SDK
  - Make DeviceName configurable
- Digital Ocean
  - Custom SSH port support
- Generic
  - Don't support `kill` since `stop` is not supported
- Google
  - Coreos provisionning
- Hyper-V
  - Lot's of code simplifications
  - Pre-Check that the user is an Administrator
  - Pre-Check that the virtual switch exists
  - Add Environment variables for each flag
  - Fix how Powershell is detected
  - VSwitch name should be saved to config.json
  - Add a flag to set the CPU count
  - Close handle after copying boot2docker.iso into vm folder - will otherwise keep hyper-v from starting vm
  - Update Boot2Docker cache in PreCreateCheck phase
- OpenStack
 - Filter floating IPs by tenant ID
- Virtualbox
  - Reject duplicate hostonlyifs Name/IP with clear message
  - Detect when hostonlyif can't be created. Point to known working version of VirtualBox
  - Don't create the VM if no hardware virtualization is available and add a flag to force create
  - Add `VBox.log` to bugsnag crashreport
  - Update Boot2Docker cache in PreCreateCheck phase
  - Detect Incompatibility with Hyper-v
- VSphere
 - Rewrite driver to work with govmomi instead of wrapping govc
- All
  - Change host restart to use the driver implementation
  - Fix truncated logs
  - Increase heartbeat interval and timeout

Provisioners

- Download latest Boot2Docker if it is out-of-date
- Add swarm config to coreos
- All provisioners now honor `engine-install-url`

# 0.5.4 (2015-12-28)

This is a patch release to fix a regression with STDOUT/STDERR behavior (#2587).

# 0.5.3 (2015-12-14)

**Please note**: With this release Machine will be reverting back to distribution in a single binary, which is more efficient on bandwidth and hard disk space. All the core driver plugins are now included in the main binary. You will want to delete the old driver binaries that you might have in your path.

e.g.:

```console
$ rm /usr/local/bin/docker-machine-driver-{amazonec2,azure,digitalocean,exoscale,generic,google,hyperv,none,openstack,rackspace,softlayer,virtualbox,vmwarefusion,vmwarevcloudair,vmwarevsphere}
```

Non-core driver plugins should still work as intended (in externally distributed binaries of the form `docker-machine-driver-name`.  Please report any issues you encounter them with externally loaded plugins.

General

- Optionally report crashes to Bugsnag to help us improve docker-machine
- Fix multiple nil dereferences in `docker-machine ls` command
- Improve the build and CI
- `docker-machine env` now supports emacs
- Run Swarm containers in provisioning step using Docker API instead of SSH/shell
- Show docker daemon version in `docker-machine ls`
- `docker-machine ls` can filter by engine label
- `docker-machine ls` filters are case insensitive
- `--timeout` flag for `docker-machine ls`
- Logs use `logrus` library
- Swarm container network is now `host`
- Added advertise flag to Swarm manager template
- Fix `help` flag for `docker-machine ssh`
- Add confirmation `-y` flag to `docker-machine rm`
- Fix `docker-machine config` for fish
- Embed all core drivers in `docker-machine` binary to reduce the bundle from 120M to 15M

Drivers

- Generic
	- Support password protected ssh keys though ssh-agent
	- Support DNS names
- Virtualbox
	- Show a warning if virtualbox is too old
	- Recognize yet another Hardware Virtualization issue pattern
	- Fix Hardware Virtualization on Linux/AMD
	- Add the `--virtualbox-host-dns-resolver` flag
	- Allow virtualbox DNSProxy override
- Google
	- Open firewall port for Swarm when needed
- VMware Fusion
	- Explicitly set umask before invoking vmrun in vmwarefusion
	- Activate the plugin only on OSX
	- Add id/gid option to mount when using vmhgfs
	- Fix for vSphere driver boot2docker ISO issues
- Digital Ocean
	- Support for creating Droplets with Cloud-init User Data
- Openstack
	- Sanitize keynames by replacing dots with underscores
- All
	- Most base images are now set to `Ubuntu 15.10`
	- Fix compatibility with drivers developed with docker-machine 0.5.0
	- Better error report for broken/incompatible drivers
	- Don't break `config.json` configuration when the disk is full

Provisioners

- Increase timeout for installing boot2docker
- Support `Ubuntu 15.10`

Misc

- Improve the documentation
- Update known drivers list

# 0.5.2 (2015-11-30)

General

-   Bash autocompletion and helpers fixed
-   Remove `RawDriver` from `config.json` - Driver parameters can now be edited
    directly again in this file.
-   Change fish `env` variable setting to be global
-   Add `docker-machine version` command
-   Move back to normal `codegangsta/cli` upstream
-   `--tls-san` flag for extra SANs

Drivers

-   Fix `GetURL` IPv6 compatibility
-   Add documentation page for available 3rd party drivers
-   VirtualBox
    -   Support for shared folders and virtualization detection on Linux hosts
    -   Improved detection of invalid host-only interface settings
-   Google
    -   Update default images
-   VMware Fusion
    -   Add option to disable shared folder
-   Generic
    -   New environment variables for flags

Provisioners

-   Support for Ubuntu >=15.04.  This means Ubuntu machines can be created which
    work with `overlay` driver of lib network.
-   Fix issue with current netstat / daemon availability checking

# 0.5.1 (2015-11-16)

-   Fixed boot2docker VM import regression
-   Fix regression breaking `docker-machine env -u` to unset environment variables
-   Enhanced virtualization capability detection and `VBoxManage` path detection
-   Properly lock VirtualBox access when running several commands concurrently
-   Allow plugins to write to STDOUT without `--debug` enabled
-   Fix Rackspace driver regression
-   Support colons in `docker-machine scp` filepaths
-   Pass environment variables for provisioned Engines to Swarm as well
-   Various enhancements around boot2docker ISO upgrade (progress bar, increased timeout)

# 0.5.0 (2015-11-1)

-   General
    -   Add pluggable driver model
    -   Clean up code to be more modular and reusable in `libmachine`
    -   Add `--github-api-token` for situations where users are getting rate limited
        by GitHub attempting to get the current `boot2docker.iso` version
    -   Various enhancements around the Makefile and build toolchain (still an active WIP)
    -   Disable SSH multiplex explicitly in commands run with the "External" client
    -   Show "-" for "inactive" machines instead of nothing
    -   Make daemon status detection more robust
-   Provisioners
    -   New CoreOS, SUSE, and Arch Linux provisioners
    -   Fixes around package installation / upgrade code on Debian and Ubuntu
-   CLI
    -   Support for regular expression pattern matching and matching by names in `ls --filter`
    -   `--no-proxy` flag for `env` (sets `NO_PROXY` in addition to other environment variables)
-   Drivers
    -   `openstack`
        -   `--openstack-ip-version` parameter
        -   `--openstack-active-timeout` parameter
    -   `google`
        -   fix destructive behavior of `start` / `stop`
    -   `hyperv`
        -   fix issues with PowerShell
    -   `vmwarefusion`
        -   some issues with shared folders fixed
        -   `--vmwarefusion-configdrive-url` option for configuration via `cloud-init`
    -   `amazonec2`
        -   `--amazonec2-use-private-address` option to use private networking
    -   `virtualbox`
        -   Enhancements around robustness of the created host-only network
        -   Fix IPv6 network mask prefix parsing
        -   `--virtualbox-no-share` option to disable the automatic home directory mount
        -   `--virtualbox-hostonly-nictype` and `--virtualbox-hostonly-nicpromisc` for controlling settings around the created hostonly NIC

# 0.4.1 (2015-08)

-   Fixes `upgrade` functionality on Debian based systems
-   Fixes `upgrade` functionality on Ubuntu based systems

# 0.4.0 (2015-08-11)

## Updates

-   HTTP Proxy support for Docker Engine
-   RedHat distros now use Docker Yum repositories
-   Ability to set environment variables in the Docker Engine
-   Internal libmachine updates for stability

## Drivers

-   Google:
    -   Preemptible instances
    -   Static IP support

## Fixes

-   Swarm Discovery Flag is verified
-   Timeout added to `ls` command to prevent hangups
-   SSH command failure now reports information about error
-   Configuration migration updates

# 0.3.0 (2015-06-18)

## Features

-   Engine option configuration (ability to configure all engine options)
-   Swarm option configuration (ability to configure all swarm options)
-   New Provisioning system to allow for greater flexibility and stability for installing and configuring Docker
-   New Provisioners
    -   Rancher OS
    -   RedHat Enterprise Linux 7.0+ (experimental)
    -   Fedora 21+ (experimental)
    -   Debian 8+ (experimental)
-   PowerShell support (configure Windows Docker CLI)
-   Command Prompt (cmd.exe) support (configure Windows Docker CLI)
-   Filter command help by driver
-   Ability to import Boot2Docker instances
-   Boot2Docker CLI migration guide (experimental)
-   Format option for `inspect` command
-   New logging output format to improve readability and display across platforms
-   Updated "active" machine concept - now is implicit according to `DOCKER_HOST` environment variable.  Note: this removes the implicit "active" machine and can no longer be specified with the `active` command.  You change the "active" host by using the `env` command instead.
-   Specify Swarm version (`--swarm-image` flag)

## Drivers

-   New: Exoscale Driver
-   New: Generic Driver (provision any host with supported base OS and SSH)
-   Amazon EC2
    -   SSH user is configurable
    -   Support for Spot instances
    -   Add option to use private address only
    -   Base AMI updated to 20150417
-   Google
    -   Support custom disk types
    -   Updated base image to v20150316
-   Openstack
    -   Support for Keystone v3 domains
-   Rackspace
    -   Misc fixes including environment variable for Flavor Id and stability
-   Softlayer
    -   Enable local disk as provisioning option
    -   Fixes for SSH access errors
    -   Fixed bug where public IP would always be returned when requesting private
    -   Add support for specifying public and private VLAN IDs
-   VirtualBox
    -   Use Intel network interface driver (adds great stability)
    -   Stability fixes for NAT access
    -   Use DNS pass through
    -   Default CPU to single core for improved performance
    -   Enable shared folder support for Windows hosts
-   VMware Fusion
    -   Boot2Docker ISO updated
    -   Shared folder support

## Fixes

-   Provisioning improvements to ensure Docker is available
-   SSH improvements for provisioning stability
-   Fixed SSH key generation bug on Windows
-   Help formatting for improved readability

## Breaking Changes

-   "Short-Form" name reference no longer supported Instead of "docker-machine " implying the active host you must now use docker-machine
-   VMware shared folders require Boot2Docker 1.7

## Special Thanks

We would like to thank all contributors.  Machine would not be where it is
without you.  We would also like to give special thanks to the following
contributors for outstanding contributions to the project:

-   @frapposelli for VMware updates and fixes
-   @hairyhenderson for several improvements to Softlayer driver, inspect formatting and lots of fixes
-   @ibuildthecloud for rancher os provisioning
-   @sthulb for portable SSH library                                              
-   @vincentbernat for exoscale                                                   
-   @zchee for Amazon updates and great doc updates

# 0.2.0 (2015-04-16)

Core Stability and Driver Updates

## Core

-   Support for system proxy environment
-   New command to regenerate TLS certificates
    -   Note: this will restart the Docker engine to apply
-   Updates to driver operations (create, start, stop, etc) for better reliability
-   New internal `libmachine` package for internal api (not ready for public usage)
-   Updated Driver Interface
    -   [Driver Spec](https://github.com/docker/machine/blob/master/docs/DRIVER_SPEC.md)
    -   Removed host provisioning from Drivers to enable a more consistent install
    -   Removed SSH commands from each Driver for more consistent operations
-   Swarm: machine now uses Swarm default binpacking strategy

## Driver Updates

-   All drivers updated to new Driver interface
-   Amazon EC2
    -   Better checking for subnets on creation
    -   Support for using Private IPs in VPC
    -   Fixed bug with duplicate security group authorization with Swarm
    -   Support for IAM instance profile
    -   Fixed bug where IP was not properly detected upon stop
-   DigitalOcean
    -   IPv6 support
    -   Backup option
    -   Private Networking
-   Openstack / Rackspace
    -   Gophercloud updated to latest version
    -   New insecure flag to disable TLS (use with caution)
-   Google
    -   Google source image updated
    -   Ability to specify auth token via file
-   VMware Fusion
    -   Paravirtualized driver for disk (pvscsi)
    -   Enhanced paravirtualized NIC (vmxnet3)
    -   Power option updates
    -   SSH keys persistent across reboots
    -   Stop now gracefully stops VM
    -   vCPUs now match host CPUs
-   SoftLayer
    -   Fixed provision bug where `curl` was not present
-   VirtualBox
    -   Correct power operations with Saved VM state
    -   Fixed bug where image option was ignored

## CLI

-   Auto-regeneration of TLS certificates when TLS error is detected
    -   Note: this will restart the Docker engine to apply
-   Minor UI updates including improved sorting and updated command docs
-   Bug with `config` and `env` with spaces fixed
    -   Note: you now must use `eval $(docker-machine env machine)` to load environment settings
-   Updates to better support `fish` shell
-   Use `--tlsverify` for both `config` and `env` commands
-   Commands now use eval for better interoperability with shell

## Testing

-   New integration test framework (bats)

# 0.1.0 (2015-02-26)

Initial beta release.

-   Provision Docker Engines using multiple drivers
-   Provide light management for the machines
    -   Create, Start, Stop, Restart, Kill, Remove, SSH
-   Configure the Docker Engine for secure communication (TLS)
-   Easily switch target machine for fast configuration of Docker Engine client
-   Provision Swarm clusters (experimental)

## Included drivers

-   Amazon EC2
-   Digital Ocean
-   Google
-   Microsoft Azure
-   Microsoft Hyper-V
-   Openstack
-   Rackspace
-   VirtualBox
-   VMware Fusion
-   VMware vCloud Air
-   VMware vSphere
