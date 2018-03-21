# Minikube Release Notes

# Version 0.25.1 - 3/28/2018
* Update to Kubernetes 1.9.4[#2613](https://github.com/kubernetes/minikube/pull/2613e)

# Version 0.25.0 - 1/26/2018
* Add freshpod addon [#2423](https://github.com/kubernetes/minikube/pull/2423)
* List addons in consistent sort order [#2446](https://github.com/kubernetes/minikube/pull/2446)
* [Minikube ISO] Upgrade Docker to 17.09 [#2427](https://github.com/kubernetes/minikube/pull/2427)
* [Minikube ISO] Change cri-o socket location to upstream default [#2262](https://github.com/kubernetes/minikube/pull/2262)
* [Minikube ISO] Update crio to v1.0.3 [#2311](https://github.com/kubernetes/minikube/pull/2311)
* Change Dashboard from Replication Controller to Deployment [#2409](https://github.com/kubernetes/minikube/pull/2409)
* Upgrade kube-addon-manager to v6.5 [#2400](https://github.com/kubernetes/minikube/pull/2400)
* Upgrade heapster to v1.5.0 [#2335](https://github.com/kubernetes/minikube/pull/2335)
* Upgrade ingress controller to v0.9.0 [#2292](https://github.com/kubernetes/minikube/pull/2292)
* Upgrade docker machine to g49dfaa70 [#2299](https://github.com/kubernetes/minikube/pull/2299)
* Added ingress integration tests [#2254](https://github.com/kubernetes/minikube/pull/2254)
* Converted image registries to k8s.gcr.io [#2356](https://github.com/kubernetes/minikube/pull/2356)
* Added cache list command [#2272](https://github.com/kubernetes/minikube/pull/2272)
* Upgrade to Kubernetes 1.9 [#2343](https://github.com/kubernetes/minikube/pull/2343)
* [hyperkit] Support NFS Sharing [#2337](https://github.com/kubernetes/minikube/pull/2337)

# Version 0.24.1 - 11/30/2017
* Add checksum verification for localkube
* Bump minikube iso to v0.23.6

# Version 0.24.0 - 11/29/2017
* Deprecated xhyve and kvm drivers [#2227](https://github.com/kubernetes/minikube/pull/2227)
* Added support for a "rootfs" layer in .minikube/files [#2110](https://github.com/kubernetes/minikube/pull/2110)
* Added a `cache` command to cache non-minikube images [#2203](https://github.com/kubernetes/minikube/pull/2203)
* Updated Dashboard addon to v1.8.0 [#2223](https://github.com/kubernetes/minikube/pull/2223)
* Switched the virtualbox driver to use virtio networking [#2211](https://github.com/kubernetes/minikube/pull/2211)
* Better error message in hyperkit driver [#2215](https://github.com/kubernetes/minikube/pull/2215)
* Update heapster addon to v1.5.0 [#2182](https://github.com/kubernetes/minikube/pull/2182)
* Moved the storage provisioner to run in a pod [#2137](https://github.com/kubernetes/minikube/pull/2137)
* Added support for tcp and udp services to the ingress addon [#2142](https://github.com/kubernetes/minikube/pull/2142)
* Bug fix to use the minikube context instead of the current kubectl context [#2128](https://github.com/kubernetes/minikube/pull/2128)
* Added zsh autocompletion [#2194](https://github.com/kubernetes/minikube/pull/2194)

# Version 0.23.0 - 10/26/2017
* Upgraded to go 1.9 [#2113](https://github.com/kubernetes/minikube/pull/2113)
* Localkube is no longer packaged in minikube bin-data [#2089](https://github.com/kubernetes/minikube/pull/2089)
* Upgraded to Kubernetes 1.8 [#2088](https://github.com/kubernetes/minikube/pull/2088)
* Added more verbose logging to minikube start [#2078](https://github.com/kubernetes/minikube/pull/2078)
* Added CoreDNS as an Addon
* Updated Ingress Addon to v0.9.0-beta.15
* Updated Dashboard to v1.7.0
* Force the none driver to use netgo [#2074](https://github.com/kubernetes/minikube/pull/2074)
* [kvm driver] Driver now returns state.Running for DOM_SHUTDOWN [#2109](https://github.com/kubernetes/minikube/pull/2109)
* [localkube] Added support for CRI-O
* [kubeadm] Added support for CRI-O [#2052](https://github.com/kubernetes/minikube/pull/2052)
* [kubeadm] Added support for feature gates [#2037](https://github.com/kubernetes/minikube/pull/2037)
* [Minikube ISO] Bumped to version v0.23.6 [#2091](https://github.com/kubernetes/minikube/pull/2091)
* [Minikube ISO] Upgraded to Docker 17.05-ce [#1542](https://github.com/kubernetes/minikube/pull/1542)
* [Minikube ISO] Upgraded to CRI-O v1.0.0 [#2069](https://github.com/kubernetes/minikube/pull/2069)

# Version 0.22.3 - 10/3/2017
* Update dnsmasq to 1.14.5 [2022](https://github.com/kubernetes/minikube/pull/2022)
* Windows cache path fix [2000](https://github.com/kubernetes/minikube/pull/2000)
* Windows path fix [1981](https://github.com/kubernetes/minikube/pull/1982)
* Components (apiserver, controller-manager, scheduler, kubelet) can now be configured in the kubeadm bootstrapper with the --extra-config flag [1985](https://github.com/kubernetes/minikube/pull/1985)
* Kubeadm bootstrapper updated to work with Kubernetes v1.8.0 [1985](https://github.com/kubernetes/minikube/pull/1985)
* OpenAPI registration fix cherry-picked for compatibility with kubectl v1.8.0 [2031](https://github.com/kubernetes/minikube/pull/2031)

* [MINIKUBE ISO] Added cri-o runtime [1998](https://github.com/kubernetes/minikube/pull/1998)

# Version 0.22.2 - 9/15/2017
* Fix path issue on windows [1954](https://github.com/kubernetes/minikube/pull/1959)
* Added experimental kubeadm bootstrapper [1903](https://github.com/kubernetes/minikube/pull/1903)
* Fixed Hyper-V KVP daemon [1958](https://github.com/kubernetes/minikube/pull/1958)

# Version 0.22.1 - 9/6/2017
* Fix for chmod error on windows [1933](https://github.com/kubernetes/minikube/pull/1933)

# Version 0.22.0 - 9/6/2017
* Made secure serving the default for all components and disabled insecure serving [#1694](https://github.com/kubernetes/minikube/pull/1694)
* Increased minikube boot speed by caching docker images [#1881](https://github.com/kubernetes/minikube/pull/1881)
* Added .minikube/files directory which gets moved into the VM at /files each VM start[#1917](https://github.com/kubernetes/minikube/pull/1917)
* Update kubernetes to v1.7.5[1912](https://github.com/kubernetes/minikube/pull/1912)
* Update etcd to v3 [#1720](https://github.com/kubernetes/minikube/pull/1720)
* Added experimental hyperkit driver in tree[#1776](https://github.com/kubernetes/minikube/pull/1776)
* Added experimental kvm driver in tree[#1828](https://github.com/kubernetes/minikube/pull/1828)

* [MINIKUBE ISO] Update cni-bin to v0.6.0-rc1 [#1760](https://github.com/kubernetes/minikube/pull/1760)

# Version 0.21.0 - 7/25/2017
* Added check for extra arguments to minikube delete [#1718](https://github.com/kubernetes/minikube/pull/1718)
* Add GCR URL Env Var to Registry-Creds addon [#1436](https://github.com/kubernetes/minikube/pull/1436)
* Bump version of Registry-Creds addon to v1.8 [#1711](https://github.com/kubernetes/minikube/pull/1711)
* Add duration as a configurable type for the configurator [#1715](https://github.com/kubernetes/minikube/pull/1715)
* Added msize and 9p-version flags to mount [#1705](https://github.com/kubernetes/minikube/pull/1705)
* Fixed password shown in plaintext when configuring Registry-Creds addon [#1708](https://github.com/kubernetes/minikube/pull/1708)
* Updated Ingress controller addon to v0.9-beta.11 [#1703](https://github.com/kubernetes/minikube/pull/1703)
* Set kube-proxy sync defaults to reduce localkube CPU load [#1699](https://github.com/kubernetes/minikube/pull/1699)
* Updated default kubernetes version to v1.7.0 [#1693](https://github.com/kubernetes/minikube/pull/1693)
* Updated kube-dns to v1.14.2 [#1693](https://github.com/kubernetes/minikube/pull/1693)
* Updated addon-manager to v6.4-beta.2 [#1693](https://github.com/kubernetes/minikube/pull/1693)
* Fix fetching localkube from internet when the default version is specified [#1688](https://github.com/kubernetes/minikube/pull/1688)
* Removed show-libmachine-logs and use-vendored-driver flags from minikube [#1685](https://github.com/kubernetes/minikube/pull/1685)
* Added logging message before waiting for the VM IP address [#1681](https://github.com/kubernetes/minikube/pull/1681)
* Added a --disable-driver-mounts flag to `minikube start` to disable xhyve and vbox fs mounts [#1646](https://github.com/kubernetes/minikube/pull/1646)
* Added dockerized builds for minikube and localkube with `BUILD_IN_DOCKER=y make` [#1656](https://github.com/kubernetes/minikube/pull/1656)
* Added script to automatically update Arch AUR and brew cask [#1642](https://github.com/kubernetes/minikube/pull/1642)
* Added wait and interval time flags to minikube service command [#1651](https://github.com/kubernetes/minikube/pull/1651)
* Fixed flags to use 9p syntax for uid and gid [#1643](https://github.com/kubernetes/minikube/pull/1643)

* [Minikube ISO] Bump ISO Version to v0.23.0
* [Minikube ISO] Added optional makefile variable `$ISO_DOCKER_EXTRA_ARGS` passed into `make out/minikube.iso` [#1657](https://github.com/kubernetes/minikube/pull/1657)
* [Minikube ISO] Upgraded docker to v1.12.6 [#1658](https://github.com/kubernetes/minikube/pull/1658)
* [Minikube ISO] Added CephFS kernel modules [#1669](https://github.com/kubernetes/minikube/pull/1669)
* [Minikube ISO] Enabled VSOCK kernel modules [#1686](https://github.com/kubernetes/minikube/pull/1686)
* [Minikube ISO] Enable IPSET kernel module [#1697](https://github.com/kubernetes/minikube/pull/1697)
* [Minikube ISO] Add ebtables util and enable kernel module [#1713](https://github.com/kubernetes/minikube/pull/1713)

## Version 0.20.0 - 6/17/2017
* Updated default Kubernetes version to 1.6.4
* Added Local Registry Addon `minikube addons enable registry` [#1583](https://github.com/kubernetes/minikube/pull/1583)
* Fixed kube-DNS addon failures
* Bumped default ISO version to 0.20.0
* Fixed mtime issue on macOS [#1594](https://github.com/kubernetes/minikube/pull/1594)
* Use --dns-domain for k8s API server cert generation [#1589](https://github.com/kubernetes/minikube/pull/1589)
* Added `minikube update-context` command [#1578](https://github.com/kubernetes/minikube/pull/1578)
* Added kubeconfig context and minikube ip to `minikube status` [#1578](https://github.com/kubernetes/minikube/pull/1578)
* Use native golang ssh [#1571](https://github.com/kubernetes/minikube/pull/1571)
* Don't treat stopping stoppped hosts as error [#1606](https://github.com/kubernetes/minikube/pull/1606)
* Bumped ingress addon to 0.9-beta.8
* Removed systemd dependency for None driver [#1592](https://github.com/kubernetes/minikube/pull/1592)

* [Minikube ISO] Enabled IP_VS, MACVLAN, and VXLAN Kernel modules
* [Minikube ISO] Increase number of inodes
* [Minikube ISO] Use buildroot branch 2017-02

## Version 0.19.1 - 5/30/2017
* Fixed issue where using TPRs could cause localkube to crash
* Added mount daemon that can be started using `minikube start --mount --mount-string="/path/to/mount"`.  Cleanup of mount handled by `minikube delete`
* Added minikube "none" driver which does not require a VM but instead launches k8s components on the host.  This allows minikube to be used in cloud environments that don't support nested virtualizations.  This can be launched by running `sudo minikube start --vm-driver=none --use-vendored-driver`
* Update kube-dns to 1.14.2
* Update kubernetes to 1.6.4
* Added `minikube ssh-key` command which retrieves the ssh key information for the minikubeVM
* Fixed vbox interface issue with minikube mount

## Version 0.19.0 - 5/3/2017
* Updated nginx ingress to v0.9-beta.4
* Updated kube-dns to 1.14.1
* Added optional `--profile` flag to all `minikube` commands to support multiple minikube instances
* Increased localkube boot speed by removing dependency on the network being up
* Improved integration tests to be more stable
* Fixed issue where using TPRs could cause localkube to crash

## Version 0.18.0 - 4/6/2017
* Upgraded default kubernetes version to v1.6.0
* Mount command on macOS xhyve
* Pods can now write to files mounted by `minikube mount`
* Added `addon configure` command
* Made DNS domain configurable with `--dns-domain` flag to `minikube start`
* Upgraded Kubernetes Dashboard to 1.6.0
* Removed Boot2Docker ISO support
* Added `addons disable default-storageclass` command to disable default dynamic provisioner
* Added support for private docker registry in registry-creds addon
* Added `--f` flag to `minikube logs` to stream logs
* Added `--docker-opts` flag to `minikube start` to propagate docker options to the daemon
* Updated heapster addon to v1.3.0
* Updated ingress addon to v0.9-beta.3
* Made localkube versions backwards compatible for versions without `--apiserver-name`

* [Minikube ISO] ISO will now be versioned the same as minikube
* [Minikube ISO] Added timezone data
* [Minikube ISO] Added `jq` and `coreutils` packages
* [Minikube ISO] Enabled RDB Kernel module
* [Minikube ISO] Added dockerized build for iso image
* [Minikube ISO] Enabled NFS_v4_2 in kernel
* [Minikube ISO] Added CIFS-utils

## Version 0.17.1 - 3/2/2017
* Removed vendored KVM driver so minikube doesn't have a dependency on libvirt-bin

* [Minikube ISO] Added ethtool
* [Minikube ISO] Added bootlocal.sh script for custom startup options
* [Minikube ISO] Added version info in /etc/VERSION
* [Minikube ISO] Bumped rkt to v1.24.0
* [Minikube ISO] Enabled user namespaces in kernel
* [Minikube ISO] `/tmp/hostpath_pv` and `/tmp/hostpath-provisioner` are now persisted

## Version 0.17.0 - 3/2/2017
* Added external hostpath provisioner to localkube
* Added unit test coverage
* Added API Name as configuration option
* Etcd is now accessible to pods
* Always use native golang SSH
* Added a deprecation warning to boot2docker provisioner
* Added MINIKUBE_HOME environment variable
* Added `minikube mount` command for 9p server

## Version 0.16.0 - 2/2/2017
* Updated minikube ISO to [v1.0.6](https://github.com/kubernetes/minikube/tree/v0.16.0/deploy/iso/minikube-iso/CHANGELOG.md)
* Updated Registry Creds addon to v1.5
* Added check for minimum disk size
* Updated kubernetes to v1.5.2

* [Minikube ISO] Added back in curl, git, and rsync
* [Minikube ISO] Enabled CONFIG_TUN in kernel
* [Minikube ISO] Added NFS packages
* [Minikube ISO] Enabled swapon on start/stop
* [Minikube ISO] Updated CNI to v0.4.0
* [Minikube ISO] Fix permissions for /data directory
* [Minikube ISO] Updated RKT to v1.23.0
* [Minikube ISO] Added in CoreOS toolbox binary
* [Minikube ISO] Fixed vboxFS permission error

## Version 0.15.0 - 1/10/2017
* Update Dashboard to v1.5.1, fixes a CSRF vulnerability in the dashboard
* Updated Kube-DNS addon to v1.9
* Now supports kubenet as a network plugin
* Added --feature-gates flag to enable alpha and experimental features in kube components
* Added --keep-context flag to keep the current kubectl context when starting minikube
* Added environment variable to enable trace profiling in minikube binary
* Updated default ISO to buildroot based minikube.iso v1.0.2
* Localkube now runs as a systemd unit in the minikube VM
* Switched integration tests to use golang subtest framework

## Version 0.14.0 - 12/14/2016
* Update to k8s v1.5.1
* Update Addon-manager to v6.1
* Update Dashboard to v1.5
* Run localkube as systemd unit in minikube-iso
* Add ingress addon
* Add aws-creds addon
* Iso-url is now configurable through `minikube config set`
* Refactor integration tests

## Version 0.13.1 - 12/5/2016
* Fix `service list` command
* Dashboard dowgnraded to v1.4.2, correctly shows PetSets again

## Version 0.13.0 - 12/1/2016
* Added heapster addon, disabled by default
* Added `minikube addon open` command
* Added Linux Virtualbox Integration tests
* Added Linux KVM Integration tests
* Added Minikube ISO Integration test on OS X
* Multiple fixes to Minikube ISO
* Updated docker-machine, pflag libraries
* Added support for net.PortRange to the configurator
* Fix bug for handling multiple kubeconfigs in env var
* Update dashboard version to 1.5.0

## Version 0.12.2 - 10/31/2016
* Fixed dashboard command
* Added support for net.IP to the configurator
* Updated dashboard version to 1.4.2

## Version 0.12.1 - 10/28/2016
* Added docker-env support to the buildroot provisioner
* `minikube service` command now supports multiple ports
* Added `minikube service list` command
* Added `minikube completion bash` command to generate bash completion
* Add progress bars for downloading, switch to go-download
* Run kube-dns as addon instead of vendored in kube2sky
* Remove static UUID for xhyve driver
* Add option to specify network name for KVM

## Version 0.12.0 - 10/21/2016
* Added support for the KUBECONFIG env var during 'minikube start'
* Updated default k8s version to v1.4.3
* Updated addon-manager to v5.1
* Added `config view` subcommand
* Increased memory default to 2048 and cpus default to 2
* Set default `log_dir` to `~/.minikube/logs`
* Added `minikube addons` command to enable or disable cluster addons
* Added format flag to service command
* Added flag Hyper-v Virtual Switch
* Added support for IPv6 addresses in docker env

## Version 0.11.0 - 10/6/2016
* Added a "configurator" allowing users to configure the Kubernetes components with arbitrary values.
* Made Kubernetes v1.4.0 the default version in minikube
* Pre-built binaries are now built with go 1.7.1
* Added opt-in error reporting
* Bug fixes

## Version 0.10.0 - 9/15/2016
* Updated the Kubernetes dashboard to v1.4.0
* Added experimental rkt support
* Enabled DynamicProvisioning of volumes
* Improved the output of the `minikube status` command
* Added `minikube config get` and `minikube config set` commands
* Fixed a bug ensuring that the node IP is routable
* Renamed the created VM from minikubeVM to minikube

## Version 0.9.0 - 9/1/2016
* Added Hyper-V support for Windows
* Added debug-level logging for show-libmachine-logs
* Added ISO checksum validation for cached ISOs
* New .minikube/addons directory where users can put addons to be initialized in minikube
* --https flag on `minikube service` for services that run over ssl/tls
* xhyve driver will now receive the same IP across starts/delete

## Version 0.8.0 - 8/17/2016
* Added a --registry-mirror flag to `minikube start`.
* Updated Kubernetes components to v1.3.5.
* Changed the `dashboard` and `service` commands to wait for the underlying services to be ready.
* Added the `DOCKER_API_VERSION` environment variable to `minikube docker-env`.
* Updated the Kubernetes dashboard to v1.1.1.
* Improved error messages during `minikube start`.
* Added the ability to specify a CIDR for the virtualbox driver.
* Configured the `/data` directory inside the Minikube VM to be persisted across reboots.
* Added the ability for minikube to accept environment variables of the form `MINIKUBE_` in place of certain command line flags.
* Minikube will now cache downloaded localkube versions.

## Version 0.7.1 - 7/27/2016
* Fixed a filepath issue which caused `minikube start` to not work properly on Windows

## Version 0.7.0 - 7/26/2016
* Added experimental support for Windows.
* Changed the etc DNS port to avoid a conflict with deis/router.
* Added a `insecure-registry` flag to `minikube start` to support insecure docker registries.
* Added a `--docker-env` flag to `minikube start` which allows for environment variables to be passed to the Docker daemon.
* Updated Kubernetes components to 1.3.3.
* Enabled all available (including alpha) Kubernetes APIs.
* Added ISO caching.
* Added a `--unset` flag to `minikube docker-env` to unset the environment variables.
* Added a `--no-proxy` flag to `minikube docker-env` to add a machine IP to NO_PROXY environment variable.
* Added additional supported shells for `minikube docker-env` (fish, cmd, powershell, tcsh, bash, zsh).

## Version 0.6.0 - 7/13/2016
* Added a `--disk-size` flag to `minikube start`.
* Fixed a bug regarding auth tokens not being reconfigured properly after VM restart
* Added a new `get-k8s-versions` command, to get the available kubernetes versions so that users know what versions are available when trying to select the kubernetes version to use
* Makefile Updates
* Documentation Updates

## Version 0.5.0 - 7/6/2016
* Updated Kubernetes components to v1.3.0
* Added experimental support for KVM and XHyve based drivers. See the [drivers documentation](DRIVERS.md) for usage.
* Fixed a bug causing cluster state to be deleted after a `minikube stop`.
* Fixed a bug causing the minikube logs to fill up rapidly.
* Added a new `minikube service` command, to open a browser to the URL for a given service.
* Added a `--cpus` flag to `minikube start`.

## Version 0.4.0 - 6/27/2016
* Updated Kubernetes components to v1.3.0-beta.1
* Updated the Kubernetes Dashboard to v1.1.0
* Added a check for updates to minikube.
* Added a driver for VMWare Fusion on OSX.
* Added a flag to customize the memory of the minikube VM.
* Documentation updates
* Fixed a bug in Docker certificate generation. Certificates will now be
  regenerated whenever `minikube start` is run.

## Version 0.3.0 - 6/10/2016
 * Added a `minikube dashboard` command to open the Kubernetes Dashboard.
 * Updated Docker to version 1.11.1.
 * Updated Kubernetes components to v1.3.0-alpha.5-330-g760c563.
 * Generated documentation for all commands. Documentation [is here](https://github.com/kubernetes/minikube/blob/master/docs/minikube.md).


## Version 0.2.0 - 6/3/2016
 * conntrack is now bundled in the ISO.
 * DNS is now working.
 * Minikube now uses the iptables based proxy mode.
 * Internal libmachine logging is now hidden by default.
 * There is a new `minikube ssh` command to ssh into the minikube VM.
 * Dramatically improved integration test coverage
 * Switched to glog instead of fmt.Print*

## Version 0.1.0 - 5/29/2016
 * Initial minikube release.
