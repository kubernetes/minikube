# Minikube Release Notes

## Version 1.0.1 - 2019-04-25

* update-context is confusing with profiles [#4049](https://github.com/kubernetes/minikube/pull/4049)
* BugFix:  ExecRunner.Copy now parses permissions strings as octal [#4139](https://github.com/kubernetes/minikube/pull/4139)
* Add user-friendly error messages for VBOX_THIRD_PARTY & HYPERV_NO_VSWITCH [#4152](https://github.com/kubernetes/minikube/pull/4152)
* Don't enable kubelet at boot, for consistency with other components [#4110](https://github.com/kubernetes/minikube/pull/4110)
* Assert that docker has started rather than explicitly restarting it  [#4116](https://github.com/kubernetes/minikube/pull/4116)
* fix tunnel integration tests for driver None [#4105](https://github.com/kubernetes/minikube/pull/4105)
* Download ISO image before Docker images, as it's required first [#4141](https://github.com/kubernetes/minikube/pull/4141)
* Reroute logs printed directly to stdout [#4115](https://github.com/kubernetes/minikube/pull/4115)
* Update default Kubernetes version to 1.14.1 [#4133](https://github.com/kubernetes/minikube/pull/4133)
* Systemd returns error on inactive, so allow that [#4095](https://github.com/kubernetes/minikube/pull/4095)
* Add known issue: VirtualBox won't boot a 64bits VM when Hyper-V is activated [#4112](https://github.com/kubernetes/minikube/pull/4112)
* Upgrade Docker, from 18.06.2-ce to 18.06.3-ce [#4022](https://github.com/kubernetes/minikube/pull/4022)
* Use Reference, allow caching images with both Tag and Digest [#3899](https://github.com/kubernetes/minikube/pull/3899)
* Added REGISTRY_STORAGE_DELETE_ENABLED environment variable for Registry addon [#4080](https://github.com/kubernetes/minikube/pull/4080)
* Add --download-only option to start command [#3737](https://github.com/kubernetes/minikube/pull/3737)
* Escape ‘%’ in console.OutStyle arguments [#4026](https://github.com/kubernetes/minikube/pull/4026)
* Add port name to service struct used in minikube service [#4011](https://github.com/kubernetes/minikube/pull/4011)
* Update Hyper-V daemons [#4030](https://github.com/kubernetes/minikube/pull/4030)
* Avoid surfacing "error: no objects passed to apply" non-error from addon-manager [#4076](https://github.com/kubernetes/minikube/pull/4076)
* Don't cache images when --vmdriver=none [#4059](https://github.com/kubernetes/minikube/pull/4059)
* Enable CONFIG_NF_CONNTRACK_ZONES  [#3755](https://github.com/kubernetes/minikube/pull/3755)
* Fixed status checking with non-default apiserver-port. [#4058](https://github.com/kubernetes/minikube/pull/4058)
* Escape systemd special chars in docker-env [#3997](https://github.com/kubernetes/minikube/pull/3997)
* Add conformance test script [#4040](https://github.com/kubernetes/minikube/pull/4040)
* ```#compdef``` must be the first line [#4015](https://github.com/kubernetes/minikube/pull/4015)

Huge thank you for this release towards our contributors:
- Abdulla Bin Mustaqeem
- Anders F Björklund
- Andy Daniels
- Arnaud Jardiné
- Artiom Diomin
- Balint Pato
- Benn Linger
- Calin Don
- Cristian Măgherușan-Stanciu @magheru_san
- Dmitry Budaev
- Don McCasland
- Douglas Thrift
- Elijah Oyekunle
- Filip Havlíček
- flyingcircle
- Guang Ya Liu
- Himanshu Pandey
- Igor Akkerman
- Ihor Dvoretskyi
- Jan Janik
- Jat
- jay vyas
- Joel Smith
- Joji Mekkatt
- karmab
- Marcos Diez
- Marco Vito Moscaritolo
- Martynas Pumputis
- Mas
- Miel Donkers
- morvencao
- Oleg Atamanenko
- RA489
- Sharif Elgamal
- Steven Davidovitz
- Thomas Strömberg
- Tom Reznik
- u5surf
- Yaroslav Skopets
- Yoan Blanc
- Zhongcheng Lao

## Version 1.0.0 - 2019-03-27

* Update default Kubernetes version to v1.14.0 [#3967](https://github.com/kubernetes/minikube/pull/3967)
  * NOTE: To avoid interaction issues, we also recommend updating kubectl to a recent release (v1.13+)
* Upgrade addon-manager to v9.0 for compatibility with Kubernetes v1.14 [#3984](https://github.com/kubernetes/minikube/pull/3984)
* Add --image-repository flag so that users can select an alternative repository mirror [#3714](https://github.com/kubernetes/minikube/pull/3714)
* Rename MINIKUBE_IN_COLOR to MINIKUBE_IN_STYLE [#3976](https://github.com/kubernetes/minikube/pull/3976)
* mount: Allow names to be passed in for gid/uid  [#3989](https://github.com/kubernetes/minikube/pull/3989)
* mount: unmount on sigint/sigterm, add --options and --mode, improve UI [#3855](https://github.com/kubernetes/minikube/pull/3855)
* --extra-config now work for kubeadm as well [#3879](https://github.com/kubernetes/minikube/pull/3879)
* start: Set the default value of --cache to true [#3917](https://github.com/kubernetes/minikube/pull/3917)
* Remove the swap partition from minikube.iso [#3927](https://github.com/kubernetes/minikube/pull/3927)
* Add solution catalog to help users who run into known problems [#3931](https://github.com/kubernetes/minikube/pull/3931)
* Automatically propagate proxy environment variables to docker env [#3834](https://github.com/kubernetes/minikube/pull/3834)
* More reliable unmount w/ SIGINT, particularly on kvm2 [#3985](https://github.com/kubernetes/minikube/pull/3985)
* Remove arch suffixes in image names [#3942](https://github.com/kubernetes/minikube/pull/3942)
* Issue #3253, improve kubernetes-version error string [#3596](https://github.com/kubernetes/minikube/pull/3596)
* Update kubeadm bootstrap logic so it does not wait for addon-manager [#3958](https://github.com/kubernetes/minikube/pull/3958)
* Add explicit kvm2 flag for hidden KVM signature [#3947](https://github.com/kubernetes/minikube/pull/3947)
* Remove the rkt container runtime [#3944](https://github.com/kubernetes/minikube/pull/3944)
* Store the toolbox on the disk instead of rootfs [#3951](https://github.com/kubernetes/minikube/pull/3951)
* fix CHANGE_MINIKUBE_NONE_USER regression from recent changes [#3875](https://github.com/kubernetes/minikube/pull/3875)
* Do not wait for k8s-app pods when starting with CNI [#3896](https://github.com/kubernetes/minikube/pull/3896)
* Replace server name in updateKubeConfig if --apiserver-name exists #3878 [#3897](https://github.com/kubernetes/minikube/pull/3897)
* feature-gates via minikube config set [#3861](https://github.com/kubernetes/minikube/pull/3861)
* Upgrade crio to v1.13.1, skip install.tools target as it isn't necessary [#3919](https://github.com/kubernetes/minikube/pull/3919)
* Update Ingress-NGINX to 0.23 Release [#3877](https://github.com/kubernetes/minikube/pull/3877)
* Add addon-manager, dashboard, and storage-provisioner to minikube logs [#3982](https://github.com/kubernetes/minikube/pull/3982)
* logs: Add kube-proxy, dmesg, uptime, uname + newlines between log sources [#3872](https://github.com/kubernetes/minikube/pull/3872)
* Skip "pull" command if using Kubernetes 1.10, which does not support it. [#3832](https://github.com/kubernetes/minikube/pull/3832)
* Allow building minikube for any architecture [#3887](https://github.com/kubernetes/minikube/pull/3887)
* Windows installer using installation path for x64 applications [#3895](https://github.com/kubernetes/minikube/pull/3895)
* caching: Fix containerd, improve console messages, add integration tests [#3767](https://github.com/kubernetes/minikube/pull/3767)
* Fix `minikube addons open heapster` [#3826](https://github.com/kubernetes/minikube/pull/3826)

We couldn't have gotten here without the folks who contributed to this release:

- Anders F Björklund
- Andy Daniels
- Calin Don
- Cristian Măgherușan-Stanciu @magheru_san
- Dmitry Budaev
- Guang Ya Liu
- Igor Akkerman
- Joel Smith
- Marco Vito Moscaritolo
- Marcos Diez
- Martynas Pumputis
- RA489
- Sharif Elgamal
- Steven Davidovitz
- Thomas Strömberg
- Zhongcheng Lao
- flyingcircle
- jay vyas
- morvencao
- u5surf

We all stand on the shoulders of the giants who came before us. A special shout-out to all [813 people who have contributed to minikube](https://github.com/kubernetes/minikube/graphs/contributors), and especially our former maintainers who made minikube into what it is today:

- Matt Rickard
- Dan Lorenc
- Aaron Prindle

## Version 0.35.0 - 2019-03-06

* Update default Kubernetes version to v1.13.4 (latest stable) [#3807](https://github.com/kubernetes/minikube/pull/3807)
* Update docker/machine to fix the AMD bug [#3809](https://github.com/kubernetes/minikube/pull/3809)
* Enable tap and vhost-net in minikube iso [#3758](https://github.com/kubernetes/minikube/pull/3758)
* Enable kernel modules necessary for IPVS [#3783](https://github.com/kubernetes/minikube/pull/3783)
* Add Netfilter `xt_socket` module to complete support for Transparent Proxying (TPROXY) [#3712](https://github.com/kubernetes/minikube/pull/3712)
* Change DefaultMountVersion to 9p2000.L [#3796](https://github.com/kubernetes/minikube/pull/3796)
* fix incorrect style name mount [#3789](https://github.com/kubernetes/minikube/pull/3789)
* When missing a hypervisor, omit the bug report prompt [#3787](https://github.com/kubernetes/minikube/pull/3787)
* Fix minikube logs for other container runtimes [#3780](https://github.com/kubernetes/minikube/pull/3780)
* Improve reliability of kube-proxy configmap updates (retry, block until pods are up) [#3774](https://github.com/kubernetes/minikube/pull/3774)
* update libvirtd [#3711](https://github.com/kubernetes/minikube/pull/3711)
* Add flag for disabling the VirtualBox VTX check [#3734](https://github.com/kubernetes/minikube/pull/3734)
* Add make target for building a rpm file [#3742](https://github.com/kubernetes/minikube/pull/3742)
* Improve building of deb package (versioning and permissions) [#3745](https://github.com/kubernetes/minikube/pull/3745)
* chown command should be against user $HOME, not roots home directory. [#3719](https://github.com/kubernetes/minikube/pull/3719)

Thank you to the following contributors who made this release possible:

- Anders F Björklund
- Artiom Diomin
- Don McCasland
- Elijah Oyekunle
- Filip Havlíček
- Ihor Dvoretskyi
- karmab
- Mas
- Miel Donkers
- Thomas Strömberg
- Tom Reznik
- Yaroslav Skopets
- Yoan Blanc

## Version 0.34.1 - 2019-02-16

* Make non-zero ssh error codes less dramatic [#3703](https://github.com/kubernetes/minikube/pull/3703)
* Only call trySSHPowerOff if we are using hyperv [#3702](https://github.com/kubernetes/minikube/pull/3702)
* Improve reporting when docker host/service is down [#3698](https://github.com/kubernetes/minikube/pull/3698)
* Use the new ISO version, for features and security [#3699](https://github.com/kubernetes/minikube/pull/3699)
* Added and unified driver usage instructions. [#3690](https://github.com/kubernetes/minikube/pull/3690)

Thank you to the folks who contributed to this bugfix release:

- Anders F Björklund
- Joerg Schad
- Thomas Strömberg

## Version 0.34.0 - 2019-02-15

* Initial implementation of 'console' package for stylized & localized console output 😂 [#3638](https://github.com/kubernetes/minikube/pull/3638)
* Podman 1.0.0 [#3584](https://github.com/kubernetes/minikube/pull/3584)
* fix netstat -f error on linux distros [#3592](https://github.com/kubernetes/minikube/pull/3592)
* addons: Fixes multiple files behavior in files rootfs [#3501](https://github.com/kubernetes/minikube/pull/3501)
* Make hyperkit driver more robust: detect crashing, misinstallation, other process names [#3660](https://github.com/kubernetes/minikube/pull/3660)
* Include pod output in 'logs' command & display detected problems during start [#3673](https://github.com/kubernetes/minikube/pull/3673)
* Upgrade Docker, from 18.06.1-ce to 18.06.2-ce [#3666](https://github.com/kubernetes/minikube/pull/3666)
* Upgrade opencontainers/runc to 0a012df [#3669](https://github.com/kubernetes/minikube/pull/3669)
* Clearer output when re-using VM's so that users know what they are waiting on [#3659](https://github.com/kubernetes/minikube/pull/3659)
* Disable kubelet disk eviction by default [#3671](https://github.com/kubernetes/minikube/pull/3671)
* Run poweroff before delete, only call uninstall if driver is None [#3665](https://github.com/kubernetes/minikube/pull/3665)
* Add DeleteCluster to bootstrapper [#3656](https://github.com/kubernetes/minikube/pull/3656)
* Enable CNI for alternative runtimes [#3617](https://github.com/kubernetes/minikube/pull/3617)
* machine: add parallels support [#953](https://github.com/kubernetes/minikube/pull/953)
* When copying assets from .minikube/files on windows, directories get squashed during transfer. ie /etc/ssl/certs/test.pem becomes ~minikube/etcsslcerts/test.pem. This pull request ensures any window style directories are converted into unix style. [#3258](https://github.com/kubernetes/minikube/pull/3258)
* Updated the default kubernetes version [#3625](https://github.com/kubernetes/minikube/pull/3625)
* Update crictl to v1.13.0 [#3616](https://github.com/kubernetes/minikube/pull/3616)
* Upgrade libmachine to version 0.16.1 [#3619](https://github.com/kubernetes/minikube/pull/3619)
* updated to fedora-29 [#3607](https://github.com/kubernetes/minikube/pull/3607)
* fix stale hyperkit.pid making minikube start hang [#3593](https://github.com/kubernetes/minikube/pull/3593)
* CRI: try to use "sudo podman load" instead of "docker load" [#2757](https://github.com/kubernetes/minikube/pull/2757)
* Use mac as identifier for dhcp [#3572](https://github.com/kubernetes/minikube/pull/3572)
* Still generate docker.service unit, even if unused [#3560](https://github.com/kubernetes/minikube/pull/3560)
* Initial commit of logviewer addon [#3391](https://github.com/kubernetes/minikube/pull/3391)
* Add images and improve parsing for kubernetes 1.11  [#3262](https://github.com/kubernetes/minikube/pull/3262)
* Stop containerd from running, if it is not desired [#3549](https://github.com/kubernetes/minikube/pull/3549)
* Re-remove kube-dns addon [#3556](https://github.com/kubernetes/minikube/pull/3556)
* Update docker env during minikube start if VM has already been created [#3387](https://github.com/kubernetes/minikube/pull/3387)
* Remove redundant newline in `minikube status` [#3565](https://github.com/kubernetes/minikube/pull/3565)
* Fix for issue #3044 - mounted timestamps incorrect with windows host [#3285](https://github.com/kubernetes/minikube/pull/3285)

Huge thank you for this release towards our contributors:

- Abhilash Pallerlamudi
- Alberto Alvarez
- Anders Björklund
- Balint Pato
- Bassam Tabbara
- Denis Denisov
- Hidekazu Nakamura
- Himanshu Pandey
- ivans3
- jay vyas
- Jeff Wu
- Kauê Doretto Grecchi
- Leif Ringstad
- Mark Gibbons
- Nicholas Goozeff
- Nicholas Irving
- Rob Richardson
- Roy Lenferink
- Skip Baney
- Thomas Strömberg
- todd densmore
- YAMAMOTO Takashi
- Yugo Horie
- Zhongcheng Lao

## Version 0.33.1 - 2019-01-18

* Install upstream runc into /usr/bin/docker-runc [#3545](https://github.com/kubernetes/minikube/pull/3545)

## Version 0.33.0 - 2019-01-17

* Set default Kubernetes version to v1.13.2 (latest stable) [#3527](https://github.com/kubernetes/minikube/pull/3527)
* Update to opencontainers/runc HEAD as of 2019-01-15 [#3535](https://github.com/kubernetes/minikube/pull/3535)
* Update to crio-bin v1.13.0 [#3515](https://github.com/kubernetes/minikube/pull/3515)
* Write /etc/crictl.yaml when starting [#3194](https://github.com/kubernetes/minikube/pull/3194)
* Improve failure output when kubeadm init fails [#3533](https://github.com/kubernetes/minikube/pull/3533)
* Add new VMware unified driver to supported list [#3534](https://github.com/kubernetes/minikube/pull/3534)
* Fix Windows cache path issues with directory hierarchies and lower-case drive letters [#3252](https://github.com/kubernetes/minikube/pull/3252)
* Avoid out directory, when listing test files [#3229](https://github.com/kubernetes/minikube/pull/3229)
* Do not include the default CNI config by default [#3441](https://github.com/kubernetes/minikube/pull/3441)
* Adding more utils tests [#3494](https://github.com/kubernetes/minikube/pull/3494)
* Add a storage-provisioner-gluster addon [#3521](https://github.com/kubernetes/minikube/pull/3521)
* Improve the default crio-bin configuration [#3190](https://github.com/kubernetes/minikube/pull/3190)
* Allow to specify api server port through CLI fix #2781 [#3108](https://github.com/kubernetes/minikube/pull/3108)
* add brew install instructions for hyperkit [#3140](https://github.com/kubernetes/minikube/pull/3140)
* Added defaultDiskSize setup to hyperkit driver [#3531](https://github.com/kubernetes/minikube/pull/3531)
* Enable ipvlan kernel module [#3510](https://github.com/kubernetes/minikube/pull/3510)
* issue# 3499: minikube status missing newline at end of output [#3502](https://github.com/kubernetes/minikube/pull/3502)
* apiserver health: try up to 5 minutes, add newline [#3528](https://github.com/kubernetes/minikube/pull/3528)
* Pass network-plugin value to kubelet [#3442](https://github.com/kubernetes/minikube/pull/3442)
* Fix missing a line break for minikube status [#3523](https://github.com/kubernetes/minikube/pull/3523)
* Documentation - Updating golang requirement to 1.11 [#3508](https://github.com/kubernetes/minikube/pull/3508)
* Updating e2e tests instructions [#3509](https://github.com/kubernetes/minikube/pull/3509)
* Defer dashboard deployment until "minikube dashboard" is executed [#3485](https://github.com/kubernetes/minikube/pull/3485)
* Change minikube-hostpath storage class addon from Reconcile to EnsureExists [#3497](https://github.com/kubernetes/minikube/pull/3497)
* Tell user given driver has been ignored if exising VM is different [#3374](https://github.com/kubernetes/minikube/pull/3374)

Thank you to all to everyone who contributed to this massive release:

- Amim Knabben
- Anders F Björklund
- Andrew Regner
- bpopovschi
- Fabio Rapposelli
- Jason Cwik
- Jeff Wu
- Kazuki Suda
- Mark Gibbons
- Martynas Pumputis
- Matt Dorn
- Michal Franc
- Narendra Kangralkar
- Niels de Vos
- Sebastien Collin
- Thomas Strömberg
- wujeff

## Version 0.32.0 - 12/21/2018

* Make Kubernetes v1.12.4 the default [#3482](https://github.com/kubernetes/minikube/pull/3482)
* Update kubeadm restart commands to support v1.13.x [#3483](https://github.com/kubernetes/minikube/pull/3483)
* Make "stop" retry on failure. [#3479](https://github.com/kubernetes/minikube/pull/3479)
* VirtualBox time cleanup: sync on boot, don't run timesyncd [#3476](https://github.com/kubernetes/minikube/pull/3476)
* Stream cmd output to tests when -v is enabled, and stream SSH output to logs [#3475](https://github.com/kubernetes/minikube/pull/3475)
* Document None driver docker compatibility [#3367](https://github.com/kubernetes/minikube/pull/3367)
* Enable host DNS resolution in virtualbox driver by default [#3453](https://github.com/kubernetes/minikube/pull/3453)
* Fix CRI socket in Kubernetes >= 1.12.0 kubeadmin config [#3452](https://github.com/kubernetes/minikube/pull/3452)
* Bump dashboard version to v1.10.1 [#3466](https://github.com/kubernetes/minikube/pull/3466)
* Hide KVM signature when using GPU passthrough to support more GPU models [#3459](https://github.com/kubernetes/minikube/pull/3459)
* Allow ServiceCIDR to be configured via 'service-cluster-ip-range' flag. [#3463](https://github.com/kubernetes/minikube/pull/3463)
* Save old cluster config in memory before overwriting [#3450](https://github.com/kubernetes/minikube/pull/3450)
* Change restart policy on gvisor pod [#3445](https://github.com/kubernetes/minikube/pull/3445)

Shout-out to the amazing members of the minikube community who made this release possible:

- Alasdair Tran
- Balint Pato
- Charles-Henri de Boysson
- Chris Eason
- Cory Locklear
- Jeffrey Sica
- JoeWrightss
- RA489
- Thomas Strömberg

## Version 0.31.0 - 12/08/2018

* Enable gvisor addon in minikube [#3399](https://github.com/kubernetes/minikube/pull/3399)
* LoadBalancer emulation with `minikube tunnel` [#3015](https://github.com/kubernetes/minikube/pull/3015)
* Add NET_PRIO cgroup to iso [#3396](https://github.com/kubernetes/minikube/pull/3396)
* Implement a check to see if an ISO URL is valid [#3287](https://github.com/kubernetes/minikube/pull/3287)
* Update Ingress-NGINX to 0.21 Release [#3365](https://github.com/kubernetes/minikube/pull/3365)
* Add schedutils to the guest VM for the ionice command (used by k8s 1.12) [#3419](https://github.com/kubernetes/minikube/pull/3419)
* Remove both the CoreDNS and KubeDNS addons. Let Kubeadm install the correct DNS addon. [#3332](https://github.com/kubernetes/minikube/pull/3332)
* Upgrade Docker, from 17.12.1-ce to 18.06.1-ce [#3223](https://github.com/kubernetes/minikube/pull/3223)
* Include ISO URL and reduce stutter in download error message [#3221](https://github.com/kubernetes/minikube/pull/3221)
* Add apiserver check to "status", and block "start" until it's healthy. [#3401](https://github.com/kubernetes/minikube/pull/3401)
* Containerd improvements
  * Only restart docker service if container runtime is docker [#3426](https://github.com/kubernetes/minikube/pull/3426)
  * Restart containerd after stopping alternate runtimes [#3343](https://github.com/kubernetes/minikube/pull/3343)
* CRI-O improvements
  * Stop docker daemon, when running cri-o [#3211](https://github.com/kubernetes/minikube/pull/3211)
  * Upgrade to crio v1.11.8 [#3313](https://github.com/kubernetes/minikube/pull/3313)
  * Add config parameter for the cri socket path [#3154](https://github.com/kubernetes/minikube/pull/3154)
* Ton of Build and CI improvements
* Ton of documentation updates

Huge thank you for this release towards our contributors:

- Akihiro Suda
- Alexander Ilyin
- Anders Björklund
- Balint Pato
- Bartel Sielski
- Bily Zhang
- dlorenc
- Fernando Diaz
- Ihor Dvoretskyi
- jay vyas
- Joey
- mikeweiwei
- mooncake
- Nguyen Hai Truong
- Peeyush gupta
- peterlobster
- Prakhar Goyal
- priyawadhwa
- SataQiu
- Thomas Strömberg
- xichengliudui
- Yongkun Anfernee Gui

## Version 0.30.0 - 10/04/2018

* **Fix for [CVE-2018-1002103](https://github.com/kubernetes/minikube/issues/3208): Dashboard vulnerable to DNS rebinding attack** [#3210](https://github.com/kubernetes/minikube/pull/3210)
* Initial support for Kubernetes 1.12+ [#3180](https://github.com/kubernetes/minikube/pull/3180)
* Enhance the Ingress Addon [#3099](https://github.com/kubernetes/minikube/pull/3099)
* Upgrade cni and cni-plugins to release version [#3152](https://github.com/kubernetes/minikube/pull/3152)
* ensure that /dev has settled before operating [#3195](https://github.com/kubernetes/minikube/pull/3195)
* Upgrade gluster client in ISO to 4.1.5 [#3162](https://github.com/kubernetes/minikube/pull/3162)
* update nginx ingress controller version to 0.19.0 [#3123](https://github.com/kubernetes/minikube/pull/3123)
* Install crictl from binary instead of from source [#3160](https://github.com/kubernetes/minikube/pull/3160)
* Switch the source of libmachine to machine-drivers. [#3185](https://github.com/kubernetes/minikube/pull/3185)
* Add psmisc package, for pstree command [#3161](https://github.com/kubernetes/minikube/pull/3161)
* Significant improvements to kvm2 networking [#3148](https://github.com/kubernetes/minikube/pull/3148)

Huge thank you for this release towards our contributors:

- Anders F Björklund
- Bob Killen
- David Genest
- Denis Gladkikh
- dlorenc
- Fernando Diaz
- Marcus Heese
- oilbeater
- Raunak Ramakrishnan
- Rui Cao
- samuela
- Sven Anderson
- Thomas Strömberg

## Version 0.29.0 - 09/27/2018

* Issue #3037 change dependency management to dep [#3136](https://github.com/kubernetes/minikube/pull/3136)
* Update dashboard version to v1.10.0 [#3122](https://github.com/kubernetes/minikube/pull/3122)
* fix: --format outputs any string, --https only subsitute http URL scheme [#3114](https://github.com/kubernetes/minikube/pull/3114)
* Change default docker storage driver to overlay2 [#3121](https://github.com/kubernetes/minikube/pull/3121)
* Add env variable for default ES_JAVA_OPTS [#3086](https://github.com/kubernetes/minikube/pull/3086)
* fix(cli): `minikube start --mount --mountsting` without write permission [#2671](https://github.com/kubernetes/minikube/pull/2671)
* Allow certificates to be optionally embedded in .kube/config [#3065](https://github.com/kubernetes/minikube/pull/3065)
* Fix the --cache-images flag. [#3090](https://github.com/kubernetes/minikube/pull/3090)
* support containerd  [#3040](https://github.com/kubernetes/minikube/pull/3040)
* Fix vmwarefusion driver [#3029](https://github.com/kubernetes/minikube/pull/3029)
* Make CoreDNS default addon [#3072](https://github.com/kubernetes/minikube/pull/3072)
* Update CoreDNS deployment [#3073](https://github.com/kubernetes/minikube/pull/3073)
* Replace 9p mount calls to syscall.Rename with os.Rename, which is capable of renaming on top of existing files. [#3047](https://github.com/kubernetes/minikube/pull/3047)
* Revert "Remove untainting logic." [#3050](https://github.com/kubernetes/minikube/pull/3050)
* Upgrade kpod 0.1 to podman 0.4.1 [#3026](https://github.com/kubernetes/minikube/pull/3026)
* Linux install: Set owner to root [#3021](https://github.com/kubernetes/minikube/pull/3021)
* Remove localkube bootstrapper and associated `get-k8s-versions` command [#2911](https://github.com/kubernetes/minikube/pull/2911)
* Update to go 1.10.1 everywhere. [#2777](https://github.com/kubernetes/minikube/pull/2777)
* Allow to override build date with SOURCE_DATE_EPOCH [#3009](https://github.com/kubernetes/minikube/pull/3009)

Huge Thank You for this release to our contributors:

- Aaron Prindle
- AdamDang
- Anders F Björklund
- Arijit Basu
- Asbjørn Apeland
- Balint Pato
- balopat
- Bennett Ellis
- Bernhard M. Wiedemann
- Daemeron
- Damian Kubaczka
- Daniel Santana
- dlorenc
- Jason Stangroome
- Jeffrey Sica
- Joao Carlos
- Kumbirai Tanekha
- Matt Rickard
- Nate Bessette
- NsLib
- peak-load
- Praveen Kumar
- RA489
- Raghavendra Talur
- ruicao
- Sandeep Rajan
- Thomas Strömberg
- Tijs Gommeren
- Viktor Safronov
- wangxy518
- yanxuean

## Version 0.28.2 - 7/20/2018

* Nvidia driver installation fixed [#2996](https://github.com/kubernetes/minikube/pull/2986)

## Version 0.28.1 - 7/16/2018

* vboxsf Host Mounting fixed (Linux Kernel version downgraded to 4.15 from 4.16) [#2986](https://github.com/kubernetes/minikube/pull/2986)
* cri-tools udpated to 1.11.1 [#2986](https://github.com/kubernetes/minikube/pull/2986)
* Feature Gates support added to kubeadm bootstrapper [#2951](https://github.com/kubernetes/minikube/pull/2951)
* Kubernetes 1.11 build support added [#2943](https://github.com/kubernetes/minikube/pull/2943)
* GPU support for kvm2 driver added [#2936](https://github.com/kubernetes/minikube/pull/2936)
* nginx ingress controller updated to 0.16.2 [#2930](https://github.com/kubernetes/minikube/pull/2930)
* heketi and gluster dependencies added to minikube ISO [#2925](https://github.com/kubernetes/minikube/pull/2925)

## Version 0.28.0 - 6/12/2018

* Minikube status command fixes [#2894](https://github.com/kubernetes/minikube/pull/2894)
* Boot changes to support virsh console [#2887](https://github.com/kubernetes/minikube/pull/2887)
* ISO changes to update to Linux 4.16 [#2883](https://github.com/kubernetes/minikube/pull/2883)
* ISO changes to support openvswitch/vxlan [#2876](https://github.com/kubernetes/minikube/pull/2876)
* Docker API version bumped to 1.35 [#2867](https://github.com/kubernetes/minikube/pull/2867)
* Added hyperkit options for enterprise VPN support [#2850](https://github.com/kubernetes/minikube/pull/2850)
* Caching correct images for k8s version [#2849](https://github.com/kubernetes/minikube/pull/2849)
* Cache images feature made synchronous, off by default [#2847](https://github.com/kubernetes/minikube/pull/2847)
* CoreDNS updated to 1.1.3 [#2836](https://github.com/kubernetes/minikube/pull/2836)
* Heapster updated to 1.5.3 [#2821](https://github.com/kubernetes/minikube/pull/2821)
* Fix for clock skew in certificate creation [#2823](https://github.com/kubernetes/minikube/pull/2823)

## Version 0.27.0 - 5/14/2018

* Start the default network for the kvm2 driver [#2806](https://github.com/kubernetes/minikube/pull/2806)
* Fix 1.9.x versions of Kubernetes with the kubeadm bootstrapper [#2791](https://github.com/kubernetes/minikube/pull/2791)
* Switch the ingress addon from an RC to a Deployment [#2788](https://github.com/kubernetes/minikube/pull/2788)
* Update nginx ingress controller to 0.14.0 [#2780](https://github.com/kubernetes/minikube/pull/2780)
* Disable dnsmasq on network for kvm driver [#2745](https://github.com/kubernetes/minikube/pull/2745)

## Version 0.26.1 - 4/17/2018

* Mark hyperkit, kvm2 and none drivers as supported [#2734](https://github.com/kubernetes/minikube/pull/2723) and [#2728](https://github.com/kubernetes/minikube/pull/2728)
* Bug fix for hyper-v driver [#2719](https://github.com/kubernetes/minikube/pull/2719)
* Add back CRI preflight ignore [#2723](https://github.com/kubernetes/minikube/pull/2723)
* Fix preflight checks on clusters <1.9 [#2721](https://github.com/kubernetes/minikube/pull/2721)

## Version 0.26.0 - 4/3/2018

* Update to Kubernetes 1.10 [#2657](https://github.com/kubernetes/minikube/pull/2657)
* Update Nginx Ingress Plugin to 0.12.0 [#2644](https://github.com/kubernetes/minikube/pull/2644)
* [Minikube ISO] Add SSHFS Support to the Minikube ISO [#2600](https://github.com/kubernetes/minikube/pull/2600)
* Upgrade Docker to 17.12 [#2597](https://github.com/kubernetes/minikube/pull/2597)
* Deactivate HSTS in Ingress by default [#2591](https://github.com/kubernetes/minikube/pull/2591)
* Add ValidatingAdmissionWebhook admission controller [#2590](https://github.com/kubernetes/minikube/pull/2590)
* Upgrade docker-machine to fix Hyper-v name conflict [#2586](https://github.com/kubernetes/minikube/pull/2586)
* Upgrade Core DNS Addon to 1.0.6 [#2584](https://github.com/kubernetes/minikube/pull/2584)
* Add metrics server Addon [#2566](https://github.com/kubernetes/minikube/pull/2566)
* Allow nesting in KVM driver [#2555](https://github.com/kubernetes/minikube/pull/2555)
* Add MutatingAdmissionWebhook admission controller [#2547](https://github.com/kubernetes/minikube/pull/2547)
* [Minikube ISO] Add Netfilter module to the ISO for Calico [#2490](https://github.com/kubernetes/minikube/pull/2490)
* Add memory and request limit to EFK Addon [#2465](https://github.com/kubernetes/minikube/pull/2465)

## Version 0.25.0 - 1/26/2018

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

## Version 0.24.1 - 11/30/2017

* Add checksum verification for localkube
* Bump minikube iso to v0.23.6

## Version 0.24.0 - 11/29/2017

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

## Version 0.23.0 - 10/26/2017

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

## Version 0.22.3 - 10/3/2017

* Update dnsmasq to 1.14.5 [2022](https://github.com/kubernetes/minikube/pull/2022)
* Windows cache path fix [2000](https://github.com/kubernetes/minikube/pull/2000)
* Windows path fix [1981](https://github.com/kubernetes/minikube/pull/1982)
* Components (apiserver, controller-manager, scheduler, kubelet) can now be configured in the kubeadm bootstrapper with the --extra-config flag [1985](https://github.com/kubernetes/minikube/pull/1985)
* Kubeadm bootstrapper updated to work with Kubernetes v1.8.0 [1985](https://github.com/kubernetes/minikube/pull/1985)
* OpenAPI registration fix cherry-picked for compatibility with kubectl v1.8.0 [2031](https://github.com/kubernetes/minikube/pull/2031)

* [MINIKUBE ISO] Added cri-o runtime [1998](https://github.com/kubernetes/minikube/pull/1998)

## Version 0.22.2 - 9/15/2017

* Fix path issue on windows [1954](https://github.com/kubernetes/minikube/pull/1959)
* Added experimental kubeadm bootstrapper [1903](https://github.com/kubernetes/minikube/pull/1903)
* Fixed Hyper-V KVP daemon [1958](https://github.com/kubernetes/minikube/pull/1958)

## Version 0.22.1 - 9/6/2017

* Fix for chmod error on windows [1933](https://github.com/kubernetes/minikube/pull/1933)

## Version 0.22.0 - 9/6/2017

* Made secure serving the default for all components and disabled insecure serving [#1694](https://github.com/kubernetes/minikube/pull/1694)
* Increased minikube boot speed by caching docker images [#1881](https://github.com/kubernetes/minikube/pull/1881)
* Added .minikube/files directory which gets moved into the VM at /files each VM start[#1917](https://github.com/kubernetes/minikube/pull/1917)
* Update kubernetes to v1.7.5[1912](https://github.com/kubernetes/minikube/pull/1912)
* Update etcd to v3 [#1720](https://github.com/kubernetes/minikube/pull/1720)
* Added experimental hyperkit driver in tree[#1776](https://github.com/kubernetes/minikube/pull/1776)
* Added experimental kvm driver in tree[#1828](https://github.com/kubernetes/minikube/pull/1828)

* [MINIKUBE ISO] Update cni-bin to v0.6.0-rc1 [#1760](https://github.com/kubernetes/minikube/pull/1760)

## Version 0.21.0 - 7/25/2017

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
* Don't treat stopping stopped hosts as error [#1606](https://github.com/kubernetes/minikube/pull/1606)
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
