# Release Notes

## Version 1.12.2 - 2020-08-03

Features:
* New Addon: Automated GCP Credentials [#8682](https://github.com/kubernetes/minikube/pull/8682)
* status: Add experimental cluster JSON status with state transition support [#8868](https://github.com/kubernetes/minikube/pull/8868)
* Add support for Error type to JSON output [#8796](https://github.com/kubernetes/minikube/pull/8796)
* Implement Warning type for JSON output [#8793](https://github.com/kubernetes/minikube/pull/8793)
* Add stopping as a possible state in deleting, change errorf to warningf [#8896](https://github.com/kubernetes/minikube/pull/8896)
* Use preloaded tarball for cri-o container runtime [#8588](https://github.com/kubernetes/minikube/pull/8588)

Bug Fixes:
* docker/podman: warn if allocated memory is below limit [#8718](https://github.com/kubernetes/minikube/pull/8718)
* Enabling metrics addon when someone enables dashboard [#8842](https://github.com/kubernetes/minikube/pull/8842)
* make base-image respect --image-repository [#8880](https://github.com/kubernetes/minikube/pull/8880)
* UI: suggest to enable `metric-server` for full feature dashboard addon. [#8863](https://github.com/kubernetes/minikube/pull/8863)
* Fix mount issues with Docker/Podman drivers [#8780](https://github.com/kubernetes/minikube/pull/8780)
* Fix upgrading from minikube 1.9 and older [#8782](https://github.com/kubernetes/minikube/pull/8782)
* Make restarts in Docker/Podman drivers more reliable [#8864](https://github.com/kubernetes/minikube/pull/8864)

Version changes:
* update crio to 1.18.3 and kicbase to ubuntu 20.04 [#8895](https://github.com/kubernetes/minikube/pull/8895)
* Podman downgrade to 1.9.3 for the build command [#8774](https://github.com/kubernetes/minikube/pull/8774)
* Upgrade kicbase to v0.0.11 [#8899](https://github.com/kubernetes/minikube/pull/8899)
* update golang version [#8781](https://github.com/kubernetes/minikube/pull/8781)
* Update external-provisioner for storage provisioner for Kubernetes 1.18 [#8610](https://github.com/kubernetes/minikube/pull/8610)
* Upgrade storage provisioner image  [#8909](https://github.com/kubernetes/minikube/pull/8909)

Thank you to our contributors for this release!
- Ajitesh13
- Alonyb
- Anders F Björklund
- Andrii Volin
- Dean Coakley
- Joel Smith
- Johannes M. Scheuermann
- Jose Donizetti
- Lu Fengqi
- Medya Ghazizadeh
- Pablo Caderno
- Priya Wadhwa
- RA489
- Sedat Gokcen
- Sharif Elgamal
- Shubham
- Thomas Strömberg
- Yang Keao
- dddddai
- niedhui

## Version 1.12.1 - 2020-07-17

Features:
* Add support for Calico CNI (--cni=calico) [#8571](https://github.com/kubernetes/minikube/pull/8571)
* Add support for Cilium CNI (--cni=cilium) [#8573](https://github.com/kubernetes/minikube/pull/8573)


Bug Fixes:
* Fix bugs which prevented upgrades from v1.0+ to v1.12 [#8741](https://github.com/kubernetes/minikube/pull/8741)
* Add KicBaseImage to existing config if missing (fixes v1.9.x upgrade) [#8738](https://github.com/kubernetes/minikube/pull/8738)
* multinode: fix control plane not ready on restart [#8698](https://github.com/kubernetes/minikube/pull/8698)
* none CNI: error if portmap plug-in is required but unavailable [#8684](https://github.com/kubernetes/minikube/pull/8684)

Version Upgrades:
* ingress addon: bump to latest version [#8705](https://github.com/kubernetes/minikube/pull/8705)
* Upgrade go version to 1.14.4 [#8660](https://github.com/kubernetes/minikube/pull/8660)

Huge thank you for this release towards our contributors: 
- Anders F Björklund
- Harsh Modi
- James Lucktaylor
- Medya Ghazizadeh
- Michael Vorburger ⛑️
- Prasad Katti
- Priya Wadhwa
- RA489
- Sharif Elgamal
- Sun-Li Beatteay
- Tam Mach
- Thomas Strömberg
- jinhong.kim



## Version 1.12.0 - 2020-07-09

Features:

* new addon : pod-security-policy [#8454](https://github.com/kubernetes/minikube/pull/8454)
* new --extra-config option to config "scheduler" [#8147](https://github.com/kubernetes/minikube/pull/8147)

ISO Changes:

* Upgrade Docker, from 19.03.11 to 19.03.12 [#8643](https://github.com/kubernetes/minikube/pull/8643)
* Upgrade crio to 1.18.2 [#8645](https://github.com/kubernetes/minikube/pull/8645)

Bug fixes:

* none: Fix 'minikube delete' issues when the apiserver is down  [#8664](https://github.com/kubernetes/minikube/pull/8664)

Huge thank you for this release towards our contributors:

- Anders F Björklund
- Ilya Danilkin
- Jani Poikela
- Li Zhijian
- Matt Broberg
- Medya Ghazizadeh
- Priya Wadhwa
- Sharif Elgamal
- Thomas Strömberg
- colvin
- vinu2003

## Version 1.12.0-beta.1 - 2020-07-01

Features:

* Add --cni flag (replaces --enable-default-cni), fix --network-plugin handling [#8545](https://github.com/kubernetes/minikube/pull/8545)
* make docker driver highly preferred [#8623](https://github.com/kubernetes/minikube/pull/8623)
* Reduce coredns replicas from 2 to 1 [#8552](https://github.com/kubernetes/minikube/pull/8552)
* Allow passing in extra args to etcd via command line [#8551](https://github.com/kubernetes/minikube/pull/8551)

Minor Improvements:

* Kernel with CONFIG_IKHEADERS for BPF tools on Kubernetes [#8582](https://github.com/kubernetes/minikube/pull/8582)
* CNI: Update CRIO netconfig with matching subnet [#8570](https://github.com/kubernetes/minikube/pull/8570)
* docker driver: add solution message when container create is stuck [#8629](https://github.com/kubernetes/minikube/pull/8629)
* docker driver: warn if overlay module is not enabled [#8541](https://github.com/kubernetes/minikube/pull/8541)
* virtualbox: double health check timeout, add better errors [#8547](https://github.com/kubernetes/minikube/pull/8547)
* linux: add solution message for noexec mount volumes [#8597](https://github.com/kubernetes/minikube/pull/8597)
* Gracefully exit if container runtime is misspelled [#8593](https://github.com/kubernetes/minikube/pull/8593)
* add verification for enabling ingress, registry and gvisor addons [#8563](https://github.com/kubernetes/minikube/pull/8563)
* Disable containerd from starting up at boot [#8621](https://github.com/kubernetes/minikube/pull/8621)
* Bump Dashboard to v2.0.1 [#8294](https://github.com/kubernetes/minikube/pull/8294)
* Check for iptables file before determining container is running [#8565](https://github.com/kubernetes/minikube/pull/8565)

Bug Fixes:

* --delete-on-failure flag: Ensure deleting failed hosts in all cases [#8628](https://github.com/kubernetes/minikube/pull/8628)
* docker-env: Do not output usage hint when shell=none. [#8531](https://github.com/kubernetes/minikube/pull/8531)
* docker-env: Avoid container suicide if Docker is not installed locally [#8528](https://github.com/kubernetes/minikube/pull/8528)
* Don't verify nf_conntrack for br_netfilter [#8598](https://github.com/kubernetes/minikube/pull/8598)

Huge thank you for this release towards our contributors:

- Alban Crequy
- Anders F Björklund
- Harkishen-Singh
- Jeff Wu
- Marcin Maciaszczyk
- Medya Ghazizadeh
- Priya Wadhwa
- Sharif Elgamal
- Sunny Beatteay
- Thomas Strömberg

## Version 1.12.0-beta.0 - 2020-06-18

Features:

* Adds support for unsetting of env vars [#8506](https://github.com/kubernetes/minikube/pull/8506)
* Require minikube-automount for /run/minikube/env [#8472](https://github.com/kubernetes/minikube/pull/8472)
* Enable support for offline docker driver [#8417](https://github.com/kubernetes/minikube/pull/8417)
* Added option --all to stop all clusters [#8285](https://github.com/kubernetes/minikube/pull/8285)
* add support for microsoft wsl for docker driver [#8368](https://github.com/kubernetes/minikube/pull/8368)
* add tutorial how to use minikube in github actions as a CI step [#8362](https://github.com/kubernetes/minikube/pull/8362)
* Add KubeVirt addon [#8275](https://github.com/kubernetes/minikube/pull/8275)
* Log stacks for slowjam analysis if STACKLOG_PATH is set [#8329](https://github.com/kubernetes/minikube/pull/8329)

Minor Improvements:

* Add heapster alias to metrics-server addon [#8455](https://github.com/kubernetes/minikube/pull/8455)
* Upgrade crio and crio.conf to v1.18.1 [#8404](https://github.com/kubernetes/minikube/pull/8404)
* bump helm-tiller addon to v2.16.8 [#8471](https://github.com/kubernetes/minikube/pull/8471)
* Upgrade falco-probe driver kernel module to 0.23 [#8450](https://github.com/kubernetes/minikube/pull/8450)
* Upgrade conmon to 2.0.17 [#8406](https://github.com/kubernetes/minikube/pull/8406)
* Upgrade podman to 1.9.3 [#8405](https://github.com/kubernetes/minikube/pull/8405)
* Upgrade Docker, from 19.03.8 to 19.03.11 [#8403](https://github.com/kubernetes/minikube/pull/8403)

Bug Fixes:

* Fix host network interface for VBox [#8475](https://github.com/kubernetes/minikube/pull/8475)

Huge thank you for this release towards our contributors:

- Anders F Björklund
- Ashley Schuett
- Harkishen-Singh
- Kenta Iso
- Marcin Niemira
- Medya Ghazizadeh
- Pablo Caderno
- Prasad Katti
- Priya Wadhwa
- Radoslaw Smigielski
- Sharif Elgamal
- Shubham Gopale
- Stanislav Petrov
- Tacio Costa
- Taqui Raza
- Thomas Strömberg
- TrishaChetani
- awgreene
- gashirar
- jjanik
- sakshamkhanna

## Version 1.11.0 - 2020-05-29

Features:

* add 'defaults' sub-command to `minikube config` [#8143](https://github.com/kubernetes/minikube/pull/8143)
* addons: add OLM addon [#8129](https://github.com/kubernetes/minikube/pull/8129)
* addons:: Add Ambassador Ingress controller addon [#8161](https://github.com/kubernetes/minikube/pull/8161)
* bump oldest k8s version supported to 1.13 [#8154](https://github.com/kubernetes/minikube/pull/8154)
* bump default kubernetes version to 1.18.3 [#8307](https://github.com/kubernetes/minikube/pull/8307)
* Bump helm-tiller 2.16.7 and promote tiller ClusterRoleBinding to v1 [#8174](https://github.com/kubernetes/minikube/pull/8174)

Minor Improvements:

* docker/podman drivers: add fall back image in docker hub [#8320](https://github.com/kubernetes/minikube/pull/8320)
* docker/podman drivers: exit with usage when need login to registry [#8225](https://github.com/kubernetes/minikube/pull/8225)
* multinode: copy apiserver certs only to control plane [#8092](https://github.com/kubernetes/minikube/pull/8092)
* docker-env: restart dockerd inside minikube on failure [#8239](https://github.com/kubernetes/minikube/pull/8239)
* wait for kubernetes components on soft start [#8199](https://github.com/kubernetes/minikube/pull/8199)
* improve minikube status display for one node [#8238](https://github.com/kubernetes/minikube/pull/8238)
* improve solution message for wrong kuberentes-version format [#8118](https://github.com/kubernetes/minikube/pull/8118)

Bug fixes:

* fix HTTP_PROXY env not being passed to docker engine [#8198](https://github.com/kubernetes/minikube/pull/8198)
* honor --image-repository even if --image-mirror-country is set [#8249](https://github.com/kubernetes/minikube/pull/8249)
* parallels driver: fix HostIP implementation [#8259](https://github.com/kubernetes/minikube/pull/8259)
* addon registry: avoid getting stuck on registry port 443 [#8208](https://github.com/kubernetes/minikube/pull/8208)
* respect native-ssh param properly [#8290](https://github.com/kubernetes/minikube/pull/8290)
* fixed parsing kubernetes version for keywords "latest" or "stable" [#8230](https://github.com/kubernetes/minikube/pull/8230)
* multinode: make sure multinode clusters survive restarts [#7973](https://github.com/kubernetes/minikube/pull/7973)
* multinode: delete docker volumes when deleting a  node [#8224](https://github.com/kubernetes/minikube/pull/8224)
* multinode: delete worker volumes for docker driver [#8216](https://github.com/kubernetes/minikube/pull/8216)
* multinode: recreate existing control plane node correctly [#8095](https://github.com/kubernetes/minikube/pull/8095)

Huge thank you for this release towards our contributors:

- Anders F Björklund
- Kenta Iso
- Medya Ghazizadeh
- Mikhail Zholobov
- Natale Vinto
- Nicola Ferraro
- Priya Wadhwa
- RA489
- Sharif Elgamal
- Shubham
- kadern0

## Version 1.10.1 - 2020-05-12

Bug fixes:

* virtualbox: fix IP address retrieval [#8106](https://github.com/kubernetes/minikube/pull/8106)
* hyperv: fix virtual switch bug [#8103](https://github.com/kubernetes/minikube/pull/8103)
* Bump Default Kubernetes version v1.18.2 and update newest [8099](https://github.com/kubernetes/minikube/pull/8099)

Huge thank you for this release towards our contributors: 

- cprogrammer1994
- Medya Ghazizadeh
- Sharif Elgamal
- Thomas Strömberg

## Version 1.10.0 - 2020-05-11

Features:

* Add new env variable `MINIKUBE_FORCE_SYSTEMD` to configure force-systemd [#8010](https://github.com/kubernetes/minikube/pull/8010)
* docker/podman: add alternative repository for base image in github packages [#7943](https://github.com/kubernetes/minikube/pull/7943)


Improvements:

* tunnel: change to clean up by default [#7946](https://github.com/kubernetes/minikube/pull/7946)
* docker/podman warn about non-amd64 archs [#8053](https://github.com/kubernetes/minikube/pull/8053)
* docker: Detect windows container and exit with instructions [#7984](https://github.com/kubernetes/minikube/pull/7984)
* make `minikube help` output consistent [#8036](https://github.com/kubernetes/minikube/pull/8036)
* podman: Use noninteractive sudo when running podman [#7959](https://github.com/kubernetes/minikube/pull/7959)
* podman: Wrap the start command with cgroup manager too [#8001](https://github.com/kubernetes/minikube/pull/8001)
* podman: implement copy for podman-remote [#8060](https://github.com/kubernetes/minikube/pull/8060)
* podman: Don't run the extraction tar container for podman [#8057](https://github.com/kubernetes/minikube/pull/8057)
* podman: disable selinux labels when extracting the tarball (permissions error) [#8017](https://github.com/kubernetes/minikube/pull/8017)
* podman: Get the gateway by inspecting container network [#7962](https://github.com/kubernetes/minikube/pull/7962)
* podman-env: add PointToHost function for podman driver [#8062](https://github.com/kubernetes/minikube/pull/8062)
* virtualbox: Quiet initial ssh timeout warning [#8027](https://github.com/kubernetes/minikube/pull/8027)
* update ingress-nginx addon version [#7997](https://github.com/kubernetes/minikube/pull/7997)
* config: Add base image to the cluster config [#7985](https://github.com/kubernetes/minikube/pull/7985)

Bug Fixes:

* wait to add aliases to /etc/hosts before starting kubelet [#8035](https://github.com/kubernetes/minikube/pull/8035)
* fix missing node name in minikube stop output [#8023](https://github.com/kubernetes/minikube/pull/8023)
* addons: fix initial retry delay, double maximum limit [#7999](https://github.com/kubernetes/minikube/pull/7999)
* restart: validate configs with new hostname, add logging [#8022](https://github.com/kubernetes/minikube/pull/8022)
* assign proper internal IPs for nodes [#8018](https://github.com/kubernetes/minikube/pull/8018)
* use the correct binary for unpacking the preload [#7961](https://github.com/kubernetes/minikube/pull/7961)

Huge thank you for this release towards our contributors:

- Anders F Björklund
- Giacomo Mr. Wolf Furlan
- Kenta Iso
- Manuel Alejandro de Brito Fontes
- Medya Ghazizadeh
- Noah Spahn
- Priya Wadhwa
- Sharif Elgamal
- Thomas Strömberg
- anencore94


## Version 1.10.0-beta.2 - 2020-04-29

Improvements:

* Upgrade default Kubernetes to v1.18.1 [#7714](https://github.com/kubernetes/minikube/pull/7714)
* Automatically apply CNI on multinode clusters [#7930](https://github.com/kubernetes/minikube/pull/7930)
* Add Metal LB addon [#7308](https://github.com/kubernetes/minikube/pull/7308)
* Add `(host|control-plane).minikube.internal` to /etc/hosts [#7247](https://github.com/kubernetes/minikube/pull/7247)
* Add "sudo" to podman calls [#7631](https://github.com/kubernetes/minikube/pull/7631)
* Add list option for "minikube node" command [#7851](https://github.com/kubernetes/minikube/pull/7851)
* Add option to force docker to use systemd as cgroup manager [#7815](https://github.com/kubernetes/minikube/pull/7815)
* Improve auto-select memory for multinode clusters [#7928](https://github.com/kubernetes/minikube/pull/7928)
* bump dashboard image v2.0.0 [#7849](https://github.com/kubernetes/minikube/pull/7849)
* Upgrade docker driver base image to v0.0.10 [#7858](https://github.com/kubernetes/minikube/pull/7858)
* docker-env: fall back to bash if can not detect shell. [#7887](https://github.com/kubernetes/minikube/pull/7887)

Bug fixes:

* docker/podman drivers: wait for service before open url [#7898](https://github.com/kubernetes/minikube/pull/7898)
* addon registry-alias: change hosts update container image [#7864](https://github.com/kubernetes/minikube/pull/7864)
* Fix sysctl fs.protected_regular=1 typo [#7882](https://github.com/kubernetes/minikube/pull/7882)
* change emoji for:  notifying new kubernetes version is available [#7835](https://github.com/kubernetes/minikube/pull/7835)
* contained cni: rename default cni file to have higher priority [#7875](https://github.com/kubernetes/minikube/pull/7875)

Huge thank you for this release towards our contributors:

- Anders F Björklund
- Kenta Iso
- Marcin Niemira
- Medya Ghazizadeh
- Priya Wadhwa
- Radoslaw Smigielski
- Sharif Elgamal
- Thomas Strömberg
- Tobias Klauser
- Travis Mehlinger
- Zhongcheng Lao
- ZouYu
- priyawadhwa

## Version 1.10.0-beta.1 - 2020-04-22

Improvements:
* Skip preload download if --image-repository is set [#7707](https://github.com/kubernetes/minikube/pull/7707)


Bug Fixes:
* ISO: persistently mount /var/lib/containerd [#7843](https://github.com/kubernetes/minikube/pull/7843)
* docker/podman: fix delete -p not cleaning up & add integration test [#7819](https://github.com/kubernetes/minikube/pull/7819)


Huge thank you for this release towards our contributors: 
- Anders F Björklund
- Kenta Iso
- Medya Ghazizadeh
- Prasad Katti
- Priya Wadhwa
- Sharif Elgamal
- Thomas Stromberg
- Tobias Klauser


## Version 1.10.0-beta.0 - 2020-04-20

Improvements:
* faster containerd start by preloading images [#7793](https://github.com/kubernetes/minikube/pull/7793)
* Add fish completion support [#7777](https://github.com/kubernetes/minikube/pull/7777)
* Behavior change: start with no arguments uses existing cluster config [#7449](https://github.com/kubernetes/minikube/pull/7449)
* conformance: add --wait=all, reduce quirks [#7716](https://github.com/kubernetes/minikube/pull/7716)
* Upgrade minimum supported k8s version to v1.12 [#7723](https://github.com/kubernetes/minikube/pull/7723)
* Add default CNI network for running wth podman [#7754](https://github.com/kubernetes/minikube/pull/7754)
* Behavior change: fallback to alternate drivers on failure [#7389](https://github.com/kubernetes/minikube/pull/7389)
* Add registry addon feature for docker on mac/windows [#7603](https://github.com/kubernetes/minikube/pull/7603)
* Check node pressure & new option "node_ready" for --wait flag [#7752](https://github.com/kubernetes/minikube/pull/7752)
* docker driver: Add Service & Tunnel features to windows   [#7739](https://github.com/kubernetes/minikube/pull/7739)
* Add master node/worker node type to `minikube status` [#7586](https://github.com/kubernetes/minikube/pull/7586)
* Add new wait component apps_running [#7460](https://github.com/kubernetes/minikube/pull/7460)
* none: Add support for OpenRC init (Google CloudShell) [#7539](https://github.com/kubernetes/minikube/pull/7539)
* Upgrade falco-probe module to version 0.21.0 [#7436](https://github.com/kubernetes/minikube/pull/7436)

Bug Fixes:
* Fix multinode cluster creation for VM drivers [#7700](https://github.com/kubernetes/minikube/pull/7700)
* tunnel: Fix resolver file permissions, add DNS forwarding test [#7753](https://github.com/kubernetes/minikube/pull/7753)
* unconfine apparmor for kic [#7658](https://github.com/kubernetes/minikube/pull/7658)
* Fix `minikube delete` output nodename missing with docker/podman driver [#7553](https://github.com/kubernetes/minikube/pull/7553)
* Respect driver.FlagDefaults even if --extra-config is set [#7509](https://github.com/kubernetes/minikube/pull/7509)
* remove docker/podman overlay network for docker-runtime [#7425](https://github.com/kubernetes/minikube/pull/7425)


Huge thank you for this release towards our contributors:

- Alonyb
- Anders F Björklund
- Anshul Sirur
- Balint Pato
- Batuhan Apaydın
- Brad Walker
- Frank Schwichtenberg
- Kenta Iso
- Medya Ghazizadeh
- Michael Vorburger ⛑️
- Pablo Caderno
- Prasad Katti
- Priya Wadhwa
- Radoslaw Smigielski
- Ruben Baez
- Sharif Elgamal
- Thomas Strömberg
- Vikky Omkar
- ZouYu
- gorbondiga
- loftkun
- nestoralonso
- remraz
- sayboras
- tomocy

Thank you so much to users who helped with community triage:

- ps-feng
- Prasad Katti

And big thank you to those who participated in our docs fixit week:

- matjung
- jlaswell
- remraz


## Version 1.9.2 - 2020-04-04

Minor improvements:

* UX: Remove noisy debug statement [#7407](https://github.com/kubernetes/minikube/pull/7407)
* Feature: Make --wait more flexible [#7375](https://github.com/kubernetes/minikube/pull/7375)
* Docker: adjust warn if slow for ps and volume [#7410](https://github.com/kubernetes/minikube/pull/7410)
* Localization: Update Japanese translations [#7403](https://github.com/kubernetes/minikube/pull/7403)
* Performance: Parallelize updating cluster and setting up certs [#7394](https://github.com/kubernetes/minikube/pull/7394)
* Addons: allow ingress addon for docker/podman drivers only on linux for now [#7393](https://github.com/kubernetes/minikube/pull/7393)

- Anders F Björklund
- Medya Ghazizadeh
- Prasad Katti
- Priya Wadhwa
- Thomas Strömberg
- tomocy

## Version 1.9.1 - 2020-04-02

Improvements:

* add delete-on-failure flag [#7345](https://github.com/kubernetes/minikube/pull/7345)
* Run dashboard with internal kubectl if not in path [#7299](https://github.com/kubernetes/minikube/pull/7299)
* Implement options for the minikube version command [#7325](https://github.com/kubernetes/minikube/pull/7325)
* service list cmd: display target port and name  [#6879](https://github.com/kubernetes/minikube/pull/6879)
* Add rejection reason to 'unable to find driver' error [#7379](https://github.com/kubernetes/minikube/pull/7379)
* Update Japanese translations [#7359](https://github.com/kubernetes/minikube/pull/7359)

Bug fixes:

* Make eviction and image GC settings consistent across kubeadm API versions [#7364](https://github.com/kubernetes/minikube/pull/7364)
* Move errors and warnings to output to stderr [#7382](https://github.com/kubernetes/minikube/pull/7382)
* Correct assumptions for forwarded hostname & IP handling [#7360](https://github.com/kubernetes/minikube/pull/7360)
* Extend maximum stop retry from 30s to 120s [#7363](https://github.com/kubernetes/minikube/pull/7363)
* Use kubectl version --short if --output=json fails [#7356](https://github.com/kubernetes/minikube/pull/7356)
* Fix embed certs by updating kubeconfig after certs are populated [#7309](https://github.com/kubernetes/minikube/pull/7309)
* none: Use LookPath to verify conntrack install [#7305](https://github.com/kubernetes/minikube/pull/7305)
* Show all global flags in options command [#7292](https://github.com/kubernetes/minikube/pull/7292)
* Fix null deref in start host err [#7278](https://github.com/kubernetes/minikube/pull/7278)
* Increase Docker "slow" timeouts to 15s [#7268](https://github.com/kubernetes/minikube/pull/7268)
* none: check for docker and root uid [#7388](https://github.com/kubernetes/minikube/pull/7388)

Thank you to our contributors for this release!

- Anders F Björklund
- Dan Lorenc
- Eberhard Wolff
- John Laswell
- Marcin Niemira
- Medya Ghazizadeh
- Prasad Katti
- Priya Wadhwa
- Sharif Elgamal
- Thomas Strömberg
- Vincent Link
- anencore94
- priyawadhwa
- re;i
- tomocy

## Version 1.9.0 - 2020-03-26

New features & improvements

* Update DefaultKubernetesVersion to v1.18.0 [#7235](https://github.com/kubernetes/minikube/pull/7235)
* Add --vm flag for users who want to autoselect only VM's [#7068](https://github.com/kubernetes/minikube/pull/7068)
* Add 'stable' and 'latest' as valid kubernetes-version values [#7212](https://github.com/kubernetes/minikube/pull/7212)

* gpu addon: privileged mode no longer required [#7149](https://github.com/kubernetes/minikube/pull/7149)
* Add sch_tbf and extend filter ipset kernel module for bandwidth shaping [#7255](https://github.com/kubernetes/minikube/pull/7255)
* Parse --disk-size and --memory sizes with binary suffixes [#7206](https://github.com/kubernetes/minikube/pull/7206)


Bug Fixes

* Re-initalize failed Kubernetes clusters [#7234](https://github.com/kubernetes/minikube/pull/7234)
* do not override hostname if extraConfig is specified [#7238](https://github.com/kubernetes/minikube/pull/7238)
* Enable HW_RANDOM_VIRTIO to fix sshd startup delays [#7208](https://github.com/kubernetes/minikube/pull/7208)
* hyperv Delete: call StopHost before removing VM [#7160](https://github.com/kubernetes/minikube/pull/7160)

Huge thank you for this release towards our contributors: 

- Anders F Björklund
- Medya Ghazizadeh
- Priya Wadhwa
- Sharif Elgamal
- Thomas Strömberg
- Tom
- Vincent Link
- Yang Keao
- Zhongcheng Lao
- vikkyomkar


## Version 1.9.0-beta.2 - 2020-03-21

New features & improvements

* 🎉 Experimental multi-node support 🎊 [#6787](https://github.com/kubernetes/minikube/pull/6787)
* Add kubectl desc nodes to minikube logs [#7105](https://github.com/kubernetes/minikube/pull/7105)
* bumpup helm-tiller v2.16.1 → v2.16.3 [#7130](https://github.com/kubernetes/minikube/pull/7130)
* Update Nvidia GPU plugin [#7132](https://github.com/kubernetes/minikube/pull/7132)
* bumpup istio & istio-provisoner addon 1.4.0 → 1.5.0 [#7120](https://github.com/kubernetes/minikube/pull/7120)
* New addon: registry-aliases [#6657](https://github.com/kubernetes/minikube/pull/6657)
* Upgrade buildroot minor version [#7101](https://github.com/kubernetes/minikube/pull/7101)
* Skip kubeadm if cluster is running & properly configured [#7124](https://github.com/kubernetes/minikube/pull/7124)
* Make certificates per-profile and consistent until IP or names change [#7125](https://github.com/kubernetes/minikube/pull/7125)

Bugfixes

* Prevent minikube from crashing if namespace or service doesn't exist [#5844](https://github.com/kubernetes/minikube/pull/5844)
* Add warning if both vm-driver and driver are specified [#7109](https://github.com/kubernetes/minikube/pull/7109)
* Improve error when docker-env is used with non-docker runtime [#7112](https://github.com/kubernetes/minikube/pull/7112)
* provisioner: only reload docker if necessary, don't install curl [#7115](https://github.com/kubernetes/minikube/pull/7115)

Thank you to our contributors:

- Anders F Björklund
- Iso Kenta
- Kamesh Sampath
- Kenta Iso
- Prasad Katti
- Priya Wadhwa
- Sharif Elgamal
- Tacio Costa
- Thomas Strömberg
- Zhongcheng Lao
- rajula96reddy
- sayboras

## Version 1.9.0-beta.1 - 2020-03-18

New features

* Use Kubernetes v1.18.0-rc.1 by default [#7076](https://github.com/kubernetes/minikube/pull/7076)
* Upgrade Docker driver to preferred (Linux), default on other platforms [#7090](https://github.com/kubernetes/minikube/pull/7090)
* Upgrade Docker, from 19.03.7 to 19.03.8 [#7040](https://github.com/kubernetes/minikube/pull/7040)
* Upgrade Docker, from 19.03.6 to 19.03.7 [#6939](https://github.com/kubernetes/minikube/pull/6939)
* Upgrade dashboard to v2.0.0-rc6 [#7098](https://github.com/kubernetes/minikube/pull/7098)
* Upgrade crio to 1.17.1 [#7099](https://github.com/kubernetes/minikube/pull/7099)
* Updated French translation [#7055](https://github.com/kubernetes/minikube/pull/7055)

Bugfixes

* If user doesn't specify driver, don't validate against existing cluster [#7096](https://github.com/kubernetes/minikube/pull/7096)
* Strip the version prefix before calling semver [#7054](https://github.com/kubernetes/minikube/pull/7054)
* Move some of the driver validation before driver selection [#7080](https://github.com/kubernetes/minikube/pull/7080)
* Fix bug where global config memory was ignored [#7082](https://github.com/kubernetes/minikube/pull/7082)
* Remove controllerManager from the kubeadm v1beta2 template [#7030](https://github.com/kubernetes/minikube/pull/7030)
* Delete: output underlying status failure [#7043](https://github.com/kubernetes/minikube/pull/7043)
* status: error properly if cluster does not exist [#7041](https://github.com/kubernetes/minikube/pull/7041)

Huge thank you for this release towards our contributors:

- Anders F Björklund
- Medya Ghazizadeh
- Priya Wadhwa
- RA489
- Richard Wall
- Sharif Elgamal
- Thomas Strömberg
- Vikky Omkar
- jumahmohammad

## Version 1.8.2 - 2020-03-13

Shiny new improvements:

* allow setting api-server port for docker/podman drivers [#6991](https://github.com/kubernetes/minikube/pull/6991)
* Update NewestKubernetesVersion to 1.18.0-beta.2 [#6988](https://github.com/kubernetes/minikube/pull/6988)
* Add warning if disk image is missing features [#6978](https://github.com/kubernetes/minikube/pull/6978)

Captivating bug fixes:

* Hyper-V: Round suggested memory alloc by 100MB for VM's [#6987](https://github.com/kubernetes/minikube/pull/6987)
* Merge repositories.json after extracting preloaded tarball so that reference store isn't lost [#6985](https://github.com/kubernetes/minikube/pull/6985)
* Fix dockerd internal port changing on restart [#7021](https://github.com/kubernetes/minikube/pull/7021)
* none: Skip driver preload and image caching [#7015](https://github.com/kubernetes/minikube/pull/7015)
* preload: fix bug for windows file separators [#6968](https://github.com/kubernetes/minikube/pull/6968)
* Block on preload download [#7003](https://github.com/kubernetes/minikube/pull/7003)
* Check if lz4 is available before trying to use it [#6941](https://github.com/kubernetes/minikube/pull/6941)
* Allow backwards compatibility with 1.6 and earlier configs [#6969](https://github.com/kubernetes/minikube/pull/6969)

Huge thank you for this release towards our contributors: 
- Anders F Björklund
- Ian Molee
- Kenta Iso
- Medya Ghazizadeh
- Priya Wadhwa
- Sharif Elgamal
- Thomas Strömberg

## Version 1.8.1 - 2020-03-06

Minor bug fix:

* Block on preload download before extracting, fall back to caching images if it fails [#6928](https://github.com/kubernetes/minikube/pull/6928)
* Cleanup remaining PointToHostDockerDaemon calls [#6925](https://github.com/kubernetes/minikube/pull/6925)

Huge thank you for this release towards our contributors: 
- Priya Wadhwa
- Thomas Stromberg
- Medya Ghazizadeh

## Version 1.8.0 - 2020-03-06

Exciting new improvements:

* Promote docker driver priority from "experimental" to "fallback" [#6791](https://github.com/kubernetes/minikube/pull/6791)
* Preload tarball images for kic drivers (docker,podman) [#6720](https://github.com/kubernetes/minikube/pull/6720)
* Preload tarball images for VMs drivers as well [#6898](https://github.com/kubernetes/minikube/pull/6898)
* Add tunnel for docker driver on darwin  [#6460](https://github.com/kubernetes/minikube/pull/6460)
* Add service feature to docker driver on darwin [#6811](https://github.com/kubernetes/minikube/pull/6811)
* Add cri-o runtime to kic drivers (podman,docker) [#6756](https://github.com/kubernetes/minikube/pull/6756)
* Add mount feature to kic drivers (podman,docker) [#6630](https://github.com/kubernetes/minikube/pull/6630)
* Rename --vm-driver flag to --driver for start command [#6888](https://github.com/kubernetes/minikube/pull/6888)
* Add Korean translation [#6910](https://github.com/kubernetes/minikube/pull/6910)
* Add k8s binaries to preloaded tarball [#6870](https://github.com/kubernetes/minikube/pull/6870)
* Add lz4 and tar to iso [#6897](https://github.com/kubernetes/minikube/pull/6897)
* Add packaging of the falco_probe kernel module [#6560](https://github.com/kubernetes/minikube/pull/6560)
* Automatically scale the default memory allocation [#6900](https://github.com/kubernetes/minikube/pull/6900)
* Change cgroup driver from cgroupfs to systemd [#6651](https://github.com/kubernetes/minikube/pull/6651)
* Unify downloaders, add GitHub and Alibaba ISO fallbacks [#6892](https://github.com/kubernetes/minikube/pull/6892)
* Upgrade cni and cni-plugins to spec 0.4.0 [#6784](https://github.com/kubernetes/minikube/pull/6784)
* Label minikube nodes [#6717](https://github.com/kubernetes/minikube/pull/6717)
* Add more Chinese translations [#6813](https://github.com/kubernetes/minikube/pull/6813)
* Update addon registry 2.6.1 → 2.7.1 [#6707](https://github.com/kubernetes/minikube/pull/6707)
* Use 'k8s.gcr.io' instead of 'gcr.io/google-containers' [#6908](https://github.com/kubernetes/minikube/pull/6908)

Important bug fixes:

* Fix inverted certificate symlink creation logic [#6889](https://github.com/kubernetes/minikube/pull/6889)
* Add systemd patch for handling DHCP router [#6659](https://github.com/kubernetes/minikube/pull/6659)
* Docker: handle already in use container name [#6906](https://github.com/kubernetes/minikube/pull/6906)
* Fix delete --all if using non default profile [#6875](https://github.com/kubernetes/minikube/pull/6875)
* Fix native-ssh flag for the ssh command [#6858](https://github.com/kubernetes/minikube/pull/6858)
* Fix start for existing profile with different vm-driver [#6828](https://github.com/kubernetes/minikube/pull/6828)
* Fix: disabling a disabled addon should not error [#6817](https://github.com/kubernetes/minikube/pull/6817)
* Fix: do not change the profile to a none existing profile [#6774](https://github.com/kubernetes/minikube/pull/6774)
* Generate fish compatible docker-env hint [#6744](https://github.com/kubernetes/minikube/pull/6744)
* Specifying control plane IP in kubeadm config template [#6745](https://github.com/kubernetes/minikube/pull/6745)
* hyperv detection: increase timeout from 2s to 8s [#6701](https://github.com/kubernetes/minikube/pull/6701)
* kic: fix service list for docker on darwin [#6830](https://github.com/kubernetes/minikube/pull/6830)
* kic: fix unprivileged port bind tunnel docker for darwin [#6833](https://github.com/kubernetes/minikube/pull/6833)
* profile list: exit zero even if one profile is not ready [#6882](https://github.com/kubernetes/minikube/pull/6882)
* tunnel on docker driver on mac: fix known_hosts issue [#6810](https://github.com/kubernetes/minikube/pull/6810)
* docker-env: fix semicolons required for fish 2.x users [#6915](https://github.com/kubernetes/minikube/pull/6915)

Thank you to everyone who helped with this extraordinary release. We now invite everyone to give the `--driver=docker` option a try!

- Anders Björklund
- Black-Hole
- Csongor Halmai
- Jose Donizetti
- Keith Schaab
- Kenta Iso
- Kevin Pullin
- Medya Ghazizadeh
- Naveen Kumar Sangi
- Nguyen Hai Truong
- Olivier Lemasle
- Pierre Ugaz
- Prasad Katti
- Priya Wadhwa
- Sharif Elgamal
- Song Shukun
- Tam Mach
- Thomas Strömberg
- anencore94
- sayboras
- vikkyomkar

## Version 1.7.3 - 2020-02-20

* Add podman driver [#6515](https://github.com/kubernetes/minikube/pull/6515)
* Create Hyper-V External Switch [#6264](https://github.com/kubernetes/minikube/pull/6264)
* Don't allow creating profile by profile command [#6672](https://github.com/kubernetes/minikube/pull/6672)
* Create the Node subcommands for multi-node refactor [#6556](https://github.com/kubernetes/minikube/pull/6556)
* Improve docker volume clean up [#6695](https://github.com/kubernetes/minikube/pull/6695)
* Add podman-env for connecting with podman-remote [#6351](https://github.com/kubernetes/minikube/pull/6351)
* Update gvisor addon to latest runsc version [#6573](https://github.com/kubernetes/minikube/pull/6573)
* Fix inverted start resource logic [#6700](https://github.com/kubernetes/minikube/pull/6700)
* Fix bug in --install-addons flag [#6696](https://github.com/kubernetes/minikube/pull/6696)
* Fix bug in docker-env and add tests for docker-env command [#6604](https://github.com/kubernetes/minikube/pull/6604)
* Fix kubeConfigPath  [#6568](https://github.com/kubernetes/minikube/pull/6568)
* Fix `minikube start` in order to be able to start VM even if machine does not exist [#5730](https://github.com/kubernetes/minikube/pull/5730)
* Fail fast if waiting for SSH to be available [#6625](https://github.com/kubernetes/minikube/pull/6625)
* Add RPFilter to ISO kernel - required for modern Calico releases [#6690](https://github.com/kubernetes/minikube/pull/6690)
* Update Kubernetes default version to v1.17.3 [#6602](https://github.com/kubernetes/minikube/pull/6602)
* Update crictl to v1.17.0 [#6667](https://github.com/kubernetes/minikube/pull/6667)
* Add conntrack-tools, needed for kubernetes 1.18 [#6626](https://github.com/kubernetes/minikube/pull/6626)
* Stopped and running machines should count as existing [#6629](https://github.com/kubernetes/minikube/pull/6629)
* Upgrade Docker to 19.03.6 [#6618](https://github.com/kubernetes/minikube/pull/6618)
* Upgrade conmon version for podman [#6622](https://github.com/kubernetes/minikube/pull/6622)
* Upgrade podman to 1.6.5 [#6623](https://github.com/kubernetes/minikube/pull/6623)
* Update helm-tiller addon image v2.14.3 → v2.16.1 [#6575](https://github.com/kubernetes/minikube/pull/6575)

Thank you to our wonderful and amazing contributors who contributed to this bug-fix release:

- Anders F Björklund
- Nguyen Hai Truong
- Martynas Pumputis
- Thomas Strömberg
- Medya Ghazizadeh
- Wietse Muizelaar
- Zhongcheng Lao
- Sharif Elgamal
- Priya Wadhwa
- Rohan Maity
- anencore94
- aallbright
- Tam Mach
- edge0701
- go_vargo

## Version 1.7.2 - 2020-02-07

* Fix to delete context when delete minikube [#6541](https://github.com/kubernetes/minikube/pull/6541)
* Fix usage of quotes in cruntime format strings [#6549](https://github.com/kubernetes/minikube/pull/6549)
* Add ca-certificates directory for distros that do not include it [#6545](https://github.com/kubernetes/minikube/pull/6545)
* kubeadm template: Combine apiserver certSANs with extraArgs [#6547](https://github.com/kubernetes/minikube/pull/6547)
* Add --install-addons=false toggle for users who don't want them [#6536](https://github.com/kubernetes/minikube/pull/6536)
* Fix a variety of bugs in `docker-env` output [#6540](https://github.com/kubernetes/minikube/pull/6540)
* Remove kubeadm pull images [#6514](https://github.com/kubernetes/minikube/pull/6514)

Special thanks go out to our contributors for these fixes:

- Anders F Björklund
- anencore94
- David Taylor
- Priya Wadhwa
- Ruben
- Sharif Elgamal
- Thomas Strömberg

## Version 1.7.1 - 2020-02-05

* Create directory using os.MkDirAll, as mkdir -p does not work on windows [#6508](https://github.com/kubernetes/minikube/pull/6508)
* Revert role change from cluster-admin->system:persistent-volume-provisioner [#6511](https://github.com/kubernetes/minikube/pull/6511)
* gvisor fixes for v1.7.0 [#6512](https://github.com/kubernetes/minikube/pull/6512)
* Remove pod list stability double check [#6509](https://github.com/kubernetes/minikube/pull/6509)
* Use cluster-dns IP setup by kubeadm [#6472](https://github.com/kubernetes/minikube/pull/6472)
* Skip driver autodetection if driver is already set [#6503](https://github.com/kubernetes/minikube/pull/6503)
* Customizing host path for dynamically provisioned PersistentVolumes [#6156](https://github.com/kubernetes/minikube/pull/6156)
* Update kubeadm api version from v1beta1 to v1beta2 [#6150](https://github.com/kubernetes/minikube/pull/6150)
* Use profile name as cluster/node name [#6200](https://github.com/kubernetes/minikube/pull/6200)

Thank you to our wonderful and amazing contributors who contributed to this bug-fix release:

- Nanik T
- Ruben
- Sharif Elgamal
- Thomas Strömberg
- tstromberg
- Vijay Katam
- Zhongcheng Lao

## Version 1.7.0 - 2020-02-04

* Add Azure Container Registry support [#6483](https://github.com/kubernetes/minikube/pull/6483)
* Support --force for overriding the ssh check [#6237](https://github.com/kubernetes/minikube/pull/6237)
* Update translation files with new strings [#6491](https://github.com/kubernetes/minikube/pull/6491)
* fix docker-env for kic drivers [#6487](https://github.com/kubernetes/minikube/pull/6487)
* Fix bugs that prevented previously-enabled addons from starting up [#6471](https://github.com/kubernetes/minikube/pull/6471)
* Fix none driver bugs with "pause"  [#6452](https://github.com/kubernetes/minikube/pull/6452)

Thank you to those brave souls who made the final push toward this release:

- Medya Gh
- Priya Wadhwa
- Sharif Elgamal
- Thomas Strömberg

## Version 1.7.0-beta.2 - 2020-01-31

* Add docker run-time for kic driver [#6436](https://github.com/kubernetes/minikube/pull/6436)
* Configure etcd and kube-proxy metrics to listen on minikube node IP [#6322](https://github.com/kubernetes/minikube/pull/6322)
* add container runtime info to profile list [#6409](https://github.com/kubernetes/minikube/pull/6409)
* status: Explicitly state that the cluster does not exist [#6438](https://github.com/kubernetes/minikube/pull/6438)
* Do not use an arch suffix for the coredns name [#6243](https://github.com/kubernetes/minikube/pull/6243)
* Prevent registry-creds configure from failing when a secret does not exist.  [#6380](https://github.com/kubernetes/minikube/pull/6380)
* improve checking modprob netfilter [#6427](https://github.com/kubernetes/minikube/pull/6427)

Huge thank you for this release towards our contributors: 

- Anders Björklund
- Bjørn Harald Fotland
- Chance Zibolski
- Kim Bao Long
- Medya Ghazizadeh
- Priya Wadhwa
- Sharif Elgamal
- Thomas Strömberg
- akshay

## Version 1.7.0-beta.1 - 2020-01-24

* Add 'pause' command to freeze Kubernetes cluster [#5962](https://github.com/kubernetes/minikube/pull/5962)
* kic driver: add multiple profiles and ssh [#6390](https://github.com/kubernetes/minikube/pull/6390)
* Update DefaultKubernetesVersion to v1.17.2 [#6392](https://github.com/kubernetes/minikube/pull/6392)
* Add varlink program for using with podman-remote [#6349](https://github.com/kubernetes/minikube/pull/6349)
* Update Kubernetes libraries to v1.17.2 [#6374](https://github.com/kubernetes/minikube/pull/6374)
* Remove addon manager [#6334](https://github.com/kubernetes/minikube/pull/6334)
* Remove unnecessary crio restart to improve start latency [#6369](https://github.com/kubernetes/minikube/pull/6369)
* Check for nil ref and img before passing them into go-containerregistry [#6236](https://github.com/kubernetes/minikube/pull/6236)
* Change the compression methods used on the iso [#6341](https://github.com/kubernetes/minikube/pull/6341)
* Update the crio.conf instead of overwriting it [#6219](https://github.com/kubernetes/minikube/pull/6219)
* Update Japanese translation [#6339](https://github.com/kubernetes/minikube/pull/6339)
* Stop minikube dashboard from crashing at start [#6325](https://github.com/kubernetes/minikube/pull/6325)

Thanks you to the following contributors:

- Anders F Björklund
- inductor
- Medya Ghazizadeh
- Naoki Oketani
- Priya Wadhwa
- Sharif Elgamal
- sshukun
- Thomas Strömberg

## Version 1.7.0-beta.0 - 2020-01-15

* Use CGroupDriver function from cruntime for kubelet [#6287](https://github.com/kubernetes/minikube/pull/6287)
* Experimental Docker support (kic) using the Kind image [#6151](https://github.com/kubernetes/minikube/pull/6151)
* disable istio provisioner by default [#6315](https://github.com/kubernetes/minikube/pull/6315)
* Add --dry-run option to start [#6256](https://github.com/kubernetes/minikube/pull/6256)
* Improve "addon list" by viewing as a table  [#6274](https://github.com/kubernetes/minikube/pull/6274)
* Disable IPv6 in the minikube VM until it can be properly supported [#6241](https://github.com/kubernetes/minikube/pull/6241)
* Fixes IPv6 address handling in kubeadm [#6214](https://github.com/kubernetes/minikube/pull/6214)
* Upgrade crio to 1.16.1 [#6210](https://github.com/kubernetes/minikube/pull/6210)
* Upgrade podman to 1.6.4 [#6208](https://github.com/kubernetes/minikube/pull/6208)
* Enable or disable addons per profile [#6124](https://github.com/kubernetes/minikube/pull/6124)
* Upgrade buildroot minor version [#6199](https://github.com/kubernetes/minikube/pull/6199)
* Add systemd patch for booting on AMD Ryzen [#6183](https://github.com/kubernetes/minikube/pull/6183)
* update zh translation [#6176](https://github.com/kubernetes/minikube/pull/6176)
* Add istio addon for minikube [#6154](https://github.com/kubernetes/minikube/pull/6154)

Huge thank you for this release towards our contributors:
- Anders Björklund
- andylibrian
- Dao Cong Tien
- Dominic Yin
- fenglixa
- GennadySpb
- Kenta Iso
- Kim Bao Long
- Medya Ghazizadeh
- Nguyen Hai Truong
- Priya Wadhwa
- Sharif Elgamal
- Thomas Strömberg
- ttonline6
- Zhongcheng Lao
- Zhou Hao

## Version 1.6.2  - 2019-12-19

* Offline: always transfer image if lookup fails, always download drivers [#6111](https://github.com/kubernetes/minikube/pull/6111)
* Update ingress-dns addon [#6091](https://github.com/kubernetes/minikube/pull/6091)
* Fix update-context to use KUBECONFIG when the env is set [#6090](https://github.com/kubernetes/minikube/pull/6090)
* Fixed IPv6 format for SSH [#6110](https://github.com/kubernetes/minikube/pull/6110)
* Add hyperkit version check whether user's hyperkit is newer [#5833](https://github.com/kubernetes/minikube/pull/5833)
* start: Remove create/delete retry loop [#6129](https://github.com/kubernetes/minikube/pull/6129)
* Change error text to encourage better issue reports [#6121](https://github.com/kubernetes/minikube/pull/6121)

Huge thank you for this release towards our contributors:
- Anukul Sangwan
- Aresforchina
- Curtis Carter
- Kenta Iso
- Medya Ghazizadeh
- Sharif Elgamal
- Thomas Strömberg
- Zhou Hao
- priyawadhwa
- tstromberg


## Version 1.6.1  - 2019-12-11

A special bugfix release to fix a Windows regression:

* lock names: Remove uid suffix & hash entire path [#6059](https://github.com/kubernetes/minikube/pull/6059)

## Version 1.6.0 - 2019-12-10

* Update default k8s version to v1.17.0 [#6042](https://github.com/kubernetes/minikube/pull/6042)
* Make Kubernetes version sticky for a cluster instead of auto-upgrading [#5798](https://github.com/kubernetes/minikube/pull/5798)
* cache add: load images to all profiles & skip previously cached images [#5987](https://github.com/kubernetes/minikube/pull/5987)
* Update dashboard to 2.0.0b8 and pre-cache it again [#6039](https://github.com/kubernetes/minikube/pull/6039)
* Pre-cache the latest kube-addon-manager [#5935](https://github.com/kubernetes/minikube/pull/5935)
* Add sch_netem kernel module for network emulation [#6038](https://github.com/kubernetes/minikube/pull/6038)
* Don't use bash as the entrypoint for docker [#5818](https://github.com/kubernetes/minikube/pull/5818)
* Make lock names uid and path specific to avoid conflicts [#5912](https://github.com/kubernetes/minikube/pull/5912)
* Remove deprecated annotation storageclass.beta.kubernetes.io [#5954](https://github.com/kubernetes/minikube/pull/5954)
* show status in profile list [#5988](https://github.com/kubernetes/minikube/pull/5988)
* Use newer gvisor version [#6000](https://github.com/kubernetes/minikube/pull/6000)
* Adds dm-crypt support [#5739](https://github.com/kubernetes/minikube/pull/5739)
* Add performance analysis packages to minikube ISO [#5942](https://github.com/kubernetes/minikube/pull/5942)

Thanks goes out to the merry band of Kubernetes contributors that made this release possible:

- Anders F Björklund
- Anukul Sangwan
- Guilherme Pellizzetti
- Jan Ahrens
- Karuppiah Natarajan
- Laura-Marie Henning
- Medya Ghazizadeh
- Nanik T
- Olivier Lemasle
- Priya Wadhwa
- Sharif Elgamal
- Thomas Strömberg
- Vasyl Purchel
- Wietse Muizelaar

## Version 1.6.0-beta.1 - 2019-11-26

* cri-o v1.16.0 [#5970](https://github.com/kubernetes/minikube/pull/5970)
* Update default k8s version to 1.17.0-rc.1 [#5973](https://github.com/kubernetes/minikube/pull/5973)
* Update crictl to v1.16.1 [#5972](https://github.com/kubernetes/minikube/pull/5972)
* Update docker to v19.03.5 [#5914](https://github.com/kubernetes/minikube/pull/5914)
* Fix profile list for non existenting folder  [#5955](https://github.com/kubernetes/minikube/pull/5955)
* Upgrade podman to 1.6.3 [#5971](https://github.com/kubernetes/minikube/pull/5971)
* Fix validation of container-runtime config [#5964](https://github.com/kubernetes/minikube/pull/5964)
* Add option for virtualbox users to set nat-nic-type  [#5960](https://github.com/kubernetes/minikube/pull/5960)
* Upgrade buildroot minor version to 2019.02.7 [#5967](https://github.com/kubernetes/minikube/pull/5967)
* dashboard: Update to latest images (2.0.0-beta6) [#5934](https://github.com/kubernetes/minikube/pull/5934)

Huge thank you for this release towards our contributors: 
- Adam Crowder
- Anders F Björklund
- David Newman
- Harsimran Singh Maan
- Kenta Iso
- Medya Ghazizadeh
- Reuven Harrison
- Sharif Elgamal
- Thomas Stromberg
- yuxiaobo

## Version 1.6.0-beta.0 - 2019-11-15

* Update DefaultKubernetesVersion to v1.17.0-beta.1 to prepare for betas [#5925](https://github.com/kubernetes/minikube/pull/5925)
* Make it possible to recover from a previously aborted StartCluster (Ctrl-C) [#5916](https://github.com/kubernetes/minikube/pull/5916)
* Add retry to SSH connectivity check [#5848](https://github.com/kubernetes/minikube/pull/5848)
* Make --wait=false non-blocking, --wait=true blocks on system pods [#5894](https://github.com/kubernetes/minikube/pull/5894)
* Only copy new or modified files into VM on restart [#5864](https://github.com/kubernetes/minikube/pull/5864)
* Remove heapster addon [#5243](https://github.com/kubernetes/minikube/pull/5243)
* mention fix for AppArmor related permission errors [#5842](https://github.com/kubernetes/minikube/pull/5842)
* Health check previously configured driver & exit if not installed [#5840](https://github.com/kubernetes/minikube/pull/5840)
* Add memory and wait longer for TestFunctional tests, include node logs [#5852](https://github.com/kubernetes/minikube/pull/5852)
* Improve message when selected driver is incompatible with existing cluster [#5854](https://github.com/kubernetes/minikube/pull/5854)
* Update libmachine to point to latest [#5877](https://github.com/kubernetes/minikube/pull/5877)
* none driver: Warn about --cpus, --memory, and --container-runtime settings [#5845](https://github.com/kubernetes/minikube/pull/5845)
* Refactor config.Config to prepare for multinode [#5889](https://github.com/kubernetes/minikube/pull/5889)

Huge thank you for this release towards our contributors: 
- Anders Björklund
- Aresforchina
- Igor Zibarev
- Josh Woodcock
- Medya Ghazizadeh
- Nanik T
- Priya Wadhwa
- RA489
- Ruslan Gustomiasov
- Sharif Elgamal
- Steffen Gransow
- Thomas Strömberg

## Version 1.5.2 - 2019-10-31 (Happy Halloween!)

* service: fix --url mode [#5790](https://github.com/kubernetes/minikube/pull/5790)
* Refactor command runner interface, allow stdin writes [#5530](https://github.com/kubernetes/minikube/pull/5530)
* macOS install docs: minikube is a normal Homebrew formula now [#5750](https://github.com/kubernetes/minikube/pull/5750)
* Allow CPU count check to be disabled using --force [#5803](https://github.com/kubernetes/minikube/pull/5803)
* Make network validation friendlier, especially to corp networks [#5802](https://github.com/kubernetes/minikube/pull/5802)

Thank you to our contributors for this release:

- Anders F Björklund
- Issy Long
- Medya Ghazizadeh
- Thomas Strömberg

## Version 1.5.1 - 2019-10-29

* Set Docker open-files limit ( 'ulimit -n') to be consistent with other runtimes [#5761](https://github.com/kubernetes/minikube/pull/5761)
* Use fixed uid/gid for the default user account [#5767](https://github.com/kubernetes/minikube/pull/5767)
* Set --wait=false to default but still wait for apiserver [#5757](https://github.com/kubernetes/minikube/pull/5757)
* kubelet: Pass --config to use kubeadm generated configuration [#5697](https://github.com/kubernetes/minikube/pull/5697)
* Refactor to remove opening browser and just return url(s) [#5718](https://github.com/kubernetes/minikube/pull/5718)

Huge thank you for this release towards our contributors:

- Anders F Björklund
- Medya Ghazizadeh
- Nanik T
- Priya Wadhwa
- Sharif Elgamal
- Thomas Strömberg

## Version 1.5.0 - 2019-10-25

* Default to best-available local hypervisor rather than VirtualBox [#5700](https://github.com/kubernetes/minikube/pull/5700)
* Update default Kubernetes version to v1.16.2 [#5731](https://github.com/kubernetes/minikube/pull/5731)
* Add json output for status [#5611](https://github.com/kubernetes/minikube/pull/5611)
* gvisor: Use chroot instead of LD_LIBRARY_PATH [#5735](https://github.com/kubernetes/minikube/pull/5735)
* Hide innocuous viper ConfigFileNotFoundError [#5732](https://github.com/kubernetes/minikube/pull/5732)

Thank you to our contributors!

- Anders F Björklund
- duohedron
- Javis Zhou
- Josh Woodcock
- Kenta Iso
- Marek Schwarz
- Medya Ghazizadeh
- Nanik T
- Rob Bruce
- Sharif Elgamal
- Thomas Strömberg

## Version 1.5.0-beta.0 - 2019-10-21

* Fix node InternalIP not matching host-only address [#5427](https://github.com/kubernetes/minikube/pull/5427)
* Add helm-tiller addon [#5363](https://github.com/kubernetes/minikube/pull/5363)
* Add ingress-dns addon [#5507](https://github.com/kubernetes/minikube/pull/5507)
* Add validation checking for minikube profile [#5624](https://github.com/kubernetes/minikube/pull/5624)
* add ability to override autoupdating drivers [#5640](https://github.com/kubernetes/minikube/pull/5640)
* Add option to  configure  dnsDomain in kubeAdm [#5566](https://github.com/kubernetes/minikube/pull/5566)
* Added flags to purge configuration with minikube delete [#5548](https://github.com/kubernetes/minikube/pull/5548)
* Upgrade Buildroot to 2019.02 and VirtualBox to 5.2 [#5609](https://github.com/kubernetes/minikube/pull/5609)
* Add libmachine debug logs back [#5574](https://github.com/kubernetes/minikube/pull/5574)
* Add JSON output for addons list [#5601](https://github.com/kubernetes/minikube/pull/5601)
* Update default Kubernetes version to 1.16.1 [#5593](https://github.com/kubernetes/minikube/pull/5593)
* Upgrade nginx ingress controller to 0.26.1 [#5514](https://github.com/kubernetes/minikube/pull/5514)
* Initial translations for fr, es, de, ja, and zh-CN [#5466](https://github.com/kubernetes/minikube/pull/5466)
* PL translation [#5491](https://github.com/kubernetes/minikube/pull/5491)
* Warn if incompatible kubectl version is in use [#5596](https://github.com/kubernetes/minikube/pull/5596)
* Fix crash when deleting the cluster but it doesn't exist [#4980](https://github.com/kubernetes/minikube/pull/4980)
* Add json output for profile list [#5554](https://github.com/kubernetes/minikube/pull/5554)
* Allow addon enabling and disabling when minikube is not running [#5565](https://github.com/kubernetes/minikube/pull/5565)
* Added option to delete all profiles [#4780](https://github.com/kubernetes/minikube/pull/4780)
* Replace registry-creds addon ReplicationController with Deployment [#5586](https://github.com/kubernetes/minikube/pull/5586)
* Performance and security enhancement for ingress-dns addon [#5614](https://github.com/kubernetes/minikube/pull/5614)
* Add addons flag to 'minikube start' in order to enable specified addons [#5543](https://github.com/kubernetes/minikube/pull/5543)
* Warn when a user tries to set a profile name that is unlikely to be valid [#4999](https://github.com/kubernetes/minikube/pull/4999)
* Make error message more human readable [#5563](https://github.com/kubernetes/minikube/pull/5563)
* Adjusted Terminal Style Detection [#5508](https://github.com/kubernetes/minikube/pull/5508)
* Fixes image repository flags when using CRI-O and containerd runtime [#5447](https://github.com/kubernetes/minikube/pull/5447)
* fix "minikube update-context" command fail [#5626](https://github.com/kubernetes/minikube/pull/5626)
* Fix pods not being scheduled when ingress deployment is patched [#5519](https://github.com/kubernetes/minikube/pull/5519)
* Fix order of parameters to CurrentContext funcs [#5439](https://github.com/kubernetes/minikube/pull/5439)
* Add solution for VERR_VMX_MSR_ALL_VMX_DISABLED [#5460](https://github.com/kubernetes/minikube/pull/5460)
* fr: fix translations of environment & existent [#5483](https://github.com/kubernetes/minikube/pull/5483)
* optimizing Chinese translation [#5201](https://github.com/kubernetes/minikube/pull/5201)
* Change systemd unit files perm to 644 [#5492](https://github.com/kubernetes/minikube/pull/5492)

Huge thank you for this release towards our contributors:
- Anders F Björklund
- bhanu011
- chentanjun
- Cornelius Weig
- Doug A
- hwdef
- James Peach
- Josh Woodcock
- Kenta Iso
- Marcin Niemira
- Medya Ghazizadeh
- Nanik T
- Pranav Jituri
- Samuel Almeida
- serhatcetinkaya
- Sharif Elgamal
- tanjunchen
- Thomas Strömberg
- u5surf
- yugo horie
- yuxiaobo
- Zhongcheng Lao
- Zoltán Reegn


## Version 1.4.0 - 2019-09-17

Notable user-facing changes:

* Update default Kubernetes version to v1.16.0 [#5395](https://github.com/kubernetes/minikube/pull/5395)
* Upgrade dashboard to 2.0.0b4 [#5403](https://github.com/kubernetes/minikube/pull/5403)
* Upgrade addon-manager to v9.0.2, improve startup and reconcile latency [#5405](https://github.com/kubernetes/minikube/pull/5405)
* Add --interactive flag to prevent stdin prompts [#5397](https://github.com/kubernetes/minikube/pull/5397)
* Automatically install docker-machine-driver-hyperkit if missing or incompatible [#5354](https://github.com/kubernetes/minikube/pull/5354)
* Driver defaults to the one previously used by the cluster [#5372](https://github.com/kubernetes/minikube/pull/5372)
* Include port names in the 'minikube service' cmd's output [#5290](https://github.com/kubernetes/minikube/pull/5290)
* Include ISO files as part of a GitHub release [#5388](https://github.com/kubernetes/minikube/pull/5388)

Thank you to our contributors for making the final push to our biggest release yet:

- Jan Janik
- Jose Donizetti
- Josh Woodcock
- Medya Ghazizadeh
- Thomas Strömberg
- chentanjun

## Version 1.4.0-beta.2 - 2019-09-13

Notable user-facing changes:

* Update default Kubernetes release to v1.16.0-rc.2 [#5320](https://github.com/kubernetes/minikube/pull/5320)
* Retire Kubernetes v1.10 support [#5342](https://github.com/kubernetes/minikube/pull/5342)
* Remove "Ignoring --vm-driver" warning [#5016](https://github.com/kubernetes/minikube/pull/5016)
* Upgrade crio to 1.15.2 [#5338](https://github.com/kubernetes/minikube/pull/5338)

Thank you to our contributors:

- Anders F Björklund
- John Pfuntner
- RA489
- Thomas Strömberg

## Version 1.4.0-beta.1 - 2019-09-11

Notable user-facing changes:

* Automatically download the Linux kvm2 driver [#5085](https://github.com/kubernetes/minikube/pull/5085)
* Hyper-V now uses "Default Switch" out of the box / upgrade to latest machine-drivers/machine [#5311](https://github.com/kubernetes/minikube/pull/5311)
* docker: Skip HTTP_PROXY=localhost [#5289](https://github.com/kubernetes/minikube/pull/5289)
* Add error if a non-default profile name is used with the none driver [#5321](https://github.com/kubernetes/minikube/pull/5321)
* dashboard: When run as root, show URL instead of opening browser [#5292](https://github.com/kubernetes/minikube/pull/5292)
* Add 'native-ssh' flag to 'minikube start' and 'minikube ssh' [#4510](https://github.com/kubernetes/minikube/pull/4510)
* Upgrade Docker, from 18.09.8 to 18.09.9 [#5303](https://github.com/kubernetes/minikube/pull/5303)
* Upgrade crio to 1.15.1 [#5304](https://github.com/kubernetes/minikube/pull/5304)

Thank you to our recent contributors:

- Anders F Björklund
- Deepika Pandhi
- Marcin Niemira
- Matt Morrissette
- Sharif Elgamal
- Thomas Strömberg
- Zachariusz Karwacki
- josedonizetti

## Version 1.4.0-beta.0 - 2019-09-04

* Upgrade default Kubernetes version to v1.16.0-beta1 [#5250](https://github.com/kubernetes/minikube/pull/5250)
* Move root filesystem from rootfs to tmpfs [#5133](https://github.com/kubernetes/minikube/pull/5133)
* Support adding untrusted root CA certificates (corp certs) [#5015](https://github.com/kubernetes/minikube/pull/5015)
* none: Add a minimum CPUs check [#5086](https://github.com/kubernetes/minikube/pull/5086)
* Exit if --kubernetes-version is older than the oldest supported version [#4759](https://github.com/kubernetes/minikube/pull/4759)
* `make` now works on Windows [#5253](https://github.com/kubernetes/minikube/pull/5253)
* logs: include exited containers, controller manager, double line count [#5249](https://github.com/kubernetes/minikube/pull/5249)
* Announce environmental overrides up front [#5212](https://github.com/kubernetes/minikube/pull/5212)
* Upgrade addons to use apps/v1 instead of extensions/v1beta1  [#5028](https://github.com/kubernetes/minikube/pull/5028)
* Re-Added time synchronization between host/VM  [#4991](https://github.com/kubernetes/minikube/pull/4991)
* Exit if uid=0, add --force flag to override [#5179](https://github.com/kubernetes/minikube/pull/5179)
* Move program data files onto persistent storage [#5032](https://github.com/kubernetes/minikube/pull/5032)
* Add wait-timeout flag to start command and refactor util/kubernetes [#5121](https://github.com/kubernetes/minikube/pull/5121)
* Update URL should be concatenated without a / [#5109](https://github.com/kubernetes/minikube/pull/5109)
* delete: Clean up machine directory if DeleteHost fails to [#5106](https://github.com/kubernetes/minikube/pull/5106)
* config: add insecure-registry [#4844](https://github.com/kubernetes/minikube/pull/4844)
* config: add container-runtime [#4834](https://github.com/kubernetes/minikube/pull/4834)
* Improve handling KUBECONFIG environment variable with invalid entries [#5056](https://github.com/kubernetes/minikube/pull/5056)
* Upgrade containerd to 1.2.8. [#5194](https://github.com/kubernetes/minikube/pull/5194)
* Update gvisor runsc version [#4494](https://github.com/kubernetes/minikube/pull/4494)
* Upgrade nginx to security patch v0.25.1 [#5197](https://github.com/kubernetes/minikube/pull/5197)


Thank you to our contributors:

- AllenZMC
- Alok Kumar
- Anders F Björklund
- bpopovschi
- Carlos Sanchez
- chentanjun
- Deepika Pandhi
- Diego Mendes
- ethan
- Guangming Wang
- Ian Lewis
- Ivan Ogasawara
- Jituri, Pranav
- josedonizetti
- Marcin Niemira
- Max K
- Medya Ghazizadeh
- Michaël Bitard
- Miguel Moll
- Olivier Lemasle
- Pankaj Patil
- Phillip Ahereza
- Pranav Jituri
- Praveen Sastry
- Priya Wadhwa
- RA489
- Rishabh Budhiraja
- serhatcetinkaya
- Sharif Elgamal
- Thomas Strömberg
- Vydruth
- William Zhang
- xieyanker
- Zhongcheng Lao
- Zoltán Reegn

## Version 1.3.1 - 2019-08-13

* Update code references to point to new documentation site [#5052](https://github.com/kubernetes/minikube/pull/5052)
* Localization support for help text [#4814](https://github.com/kubernetes/minikube/pull/4814)
* Fix progress bar on Windows + git bash [#5025](https://github.com/kubernetes/minikube/pull/5025)
* Restore --disable-driver-mounts flag [#5026](https://github.com/kubernetes/minikube/pull/5026)
* Fixed the template for dashboard output [#5004](https://github.com/kubernetes/minikube/pull/5004)
* Use a temp dest to atomically download the iso [#5000](https://github.com/kubernetes/minikube/pull/5000)

Thank you to our merry band of contributors for assembling this last minute bug fix release.

- Jituri, Pranav
- Medya Ghazizadeh
- Pranav Jituri
- Ramiro Berrelleza
- Sharif Elgamal
- Thomas Strömberg
- josedonizetti

## Version 1.3.0 - 2019-08-05

* Added a new command: profile list [#4811](https://github.com/kubernetes/minikube/pull/4811)
* Update latest kubernetes version to v1.15.2 [#4986](https://github.com/kubernetes/minikube/pull/4986)
* Update latest kubernetes version to v1.15.1 [#4915](https://github.com/kubernetes/minikube/pull/4915)
* logs: Add container status & cruntime logs [#4960](https://github.com/kubernetes/minikube/pull/4960)
* Automatically set flags for MINIKUBE_ prefixed env vars [#4607](https://github.com/kubernetes/minikube/pull/4607)
* hyperv: Run "sudo poweroff" before stopping VM [#4758](https://github.com/kubernetes/minikube/pull/4758)
* Decrease ReasonableStartTime from 10 minutes to 5 minutes [#4961](https://github.com/kubernetes/minikube/pull/4961)
* Remove ingress-nginx default backend [#4786](https://github.com/kubernetes/minikube/pull/4786)
* Upgrade nginx ingress to 0.25.0 [#4785](https://github.com/kubernetes/minikube/pull/4785)
* Bump k8s.io/kubernetes to 1.15.0 [#4719](https://github.com/kubernetes/minikube/pull/4719)
* Upgrade Docker, from 18.09.7 to 18.09.8 [#4818](https://github.com/kubernetes/minikube/pull/4818)
* Upgrade Docker, from 18.09.6 to 18.09.7 [#4657](https://github.com/kubernetes/minikube/pull/4657)
* Upgrade crio to 1.15.0 [#4703](https://github.com/kubernetes/minikube/pull/4703)
* Update crictl to v1.15.0 [#4761](https://github.com/kubernetes/minikube/pull/4761)
* Upgrade Podman to 1.4 [#4610](https://github.com/kubernetes/minikube/pull/4610)
* Upgrade libmachine to master [#4817](https://github.com/kubernetes/minikube/pull/4817)
* Add linux packaging for the kvm2 driver binary [#4556](https://github.com/kubernetes/minikube/pull/4556)
* Unset profile when it is deleted [#4922](https://github.com/kubernetes/minikube/pull/4922)
* more reliable stop for none driver [#4871](https://github.com/kubernetes/minikube/pull/4871)
* Fix regression caused by registry-proxy [#4805](https://github.com/kubernetes/minikube/pull/4805)
* Warn if hyperkit version is old [#4691](https://github.com/kubernetes/minikube/pull/4691)
* Add warn if kvm driver version is old [#4676](https://github.com/kubernetes/minikube/pull/4676)
* Add T versions of the console convenience functions [#4796](https://github.com/kubernetes/minikube/pull/4796)
* Remove deprecated drivers: kvm-old and xhyve [#4781](https://github.com/kubernetes/minikube/pull/4781)
* Don't disable other container engines when --vm_driver=none [#4545](https://github.com/kubernetes/minikube/pull/4545)
* Proxy: handle lower case proxy env vars [#4602](https://github.com/kubernetes/minikube/pull/4602)
* virtualbox: Make DNS settings configurable [#4619](https://github.com/kubernetes/minikube/pull/4619)
* Add support to custom qemu uri on kvm2 driver [#4401](https://github.com/kubernetes/minikube/pull/4401)
* Update Ingress-NGINX to 0.24.1 Release [#4583](https://github.com/kubernetes/minikube/pull/4583)

A big thanks goes out to our crew of merry contributors:

- Aida Ghazizadeh
- Anders F Björklund
- Ben Ebsworth
- Benjamin Howell
- cclauss
- Christophe VILA
- Deepjyoti Mondal
- fang duan
- Francis
- Gustavo Belfort
- Himanshu Pandey
- Jituri, Pranav
- josedonizetti
- Jose Donizetti
- Kazuki Suda
- Kyle Bai
- Marcos Diez
- Medya Ghazizadeh
- Nabarun Pal
- Om Kumar
- Pranav Jituri
- RA489
- serhat çetinkaya
- Sharif Elgamal
- Stuart P. Bentley
- Thomas Strömberg
- Zoltán Reegn

## Version 1.2.0 - 2019-06-24

* Update Kubernetes default version to v1.15.0 [#4534](https://github.com/kubernetes/minikube/pull/4534)
* Allow --kubernetes-version to be specified without the leading v [#4568](https://github.com/kubernetes/minikube/pull/4568)
* Enable running containers with Podman [#4421](https://github.com/kubernetes/minikube/pull/4421)
* Provide warning message for unnecessary sudo [#4455](https://github.com/kubernetes/minikube/pull/4455)
* Universally redirect stdlog messages to glog [#4562](https://github.com/kubernetes/minikube/pull/4562)
* Add ability to localize all strings output to console [#4464](https://github.com/kubernetes/minikube/pull/4464)
* Upgrade CNI config version to 0.3.0 [#4410](https://github.com/kubernetes/minikube/pull/4410)
* Register registry-proxy.yaml.tmpl with registry addons [#4529](https://github.com/kubernetes/minikube/pull/4529)
* Stop updating /etc/rkt/net.d config files [#4407](https://github.com/kubernetes/minikube/pull/4407)
* Fix "mount failed: File exists" issue when unmount fails [#4393](https://github.com/kubernetes/minikube/pull/4393)
* Don't try to load cached images for none driver [#4522](https://github.com/kubernetes/minikube/pull/4522)
* Add support for Kubernetes v1.15.0-beta.1 [#4469](https://github.com/kubernetes/minikube/pull/4469)
* Switch kubectl current-context on profile change [#4504](https://github.com/kubernetes/minikube/pull/4504)
* Add kvm network name validation [#4380](https://github.com/kubernetes/minikube/pull/4380)
* Detect status before enable/disable addon [#4424](https://github.com/kubernetes/minikube/pull/4424)
* Automatically add extra options for none driver on ubuntu [#4465](https://github.com/kubernetes/minikube/pull/4465)

Thank you to the following wonderful people for their contribution to this release:

- Anders F Björklund
- Deepjyoti Mondal
- Francis
- Jose Donizetti
- Medya Ghazizadeh
- Om Kumar
- Sharif Elgamal
- Thomas Strömberg
- Y.Horie
- fenglixa
- josedonizetti

## Version 1.1.1 - 2019-06-07

* Upgrade to kubernetes 1.14.3 [#4444](https://github.com/kubernetes/minikube/pull/4444)
* fix ShowDriverDeprecationNotification config setting [#4431](https://github.com/kubernetes/minikube/pull/4431)
* Cache: don't use ssh runner for the none driver [#4439](https://github.com/kubernetes/minikube/pull/4439)
* Fixing file path for windows [#4434](https://github.com/kubernetes/minikube/pull/4434)
* Improve type check for driver none [#4419](https://github.com/kubernetes/minikube/pull/4419)
* Dashboard: add --disable-settings-authorizer to avoid settings 403 forbidden [#4405](https://github.com/kubernetes/minikube/pull/4405)
* dashboard: detect nonexistent profile instead of causing a panic [#4396](https://github.com/kubernetes/minikube/pull/4396)
* Fixed addon-manager failing with non-default --apiserver-port [#4386](https://github.com/kubernetes/minikube/pull/4386)
* Fix kvm gpu log [#4381](https://github.com/kubernetes/minikube/pull/4381)
* Windows installer: Use PowerShell to update PATH value to avoid 1024 char truncation [#4362](https://github.com/kubernetes/minikube/pull/4362)
* Increase apiserver wait time from 1 minute to 3 minutes [#4372](https://github.com/kubernetes/minikube/pull/4372)
* Sync guest system clock if desynchronized from host [#4283](https://github.com/kubernetes/minikube/pull/4283)
* docker-env: Remove DOCKER_API_VERSION [#4364](https://github.com/kubernetes/minikube/pull/4364)
* Disable hyperv dynamic memory for hyperv driver [#2797](https://github.com/kubernetes/minikube/pull/2797)
* Fix kvm remove when domain is not defined [#4355](https://github.com/kubernetes/minikube/pull/4355)
* Enable registry-proxy [#4341](https://github.com/kubernetes/minikube/pull/4341)
* Make buildah --no-pivot default, using env var [#4321](https://github.com/kubernetes/minikube/pull/4321)
* Pass minikube stdin to the kubectl command [#4354](https://github.com/kubernetes/minikube/pull/4354)
* kernel: Add config for tc u32 filter and mirred action [#4340](https://github.com/kubernetes/minikube/pull/4340)
* Enable GatewayPorts in sshd_config, for proxying in services into minikube [#4338](https://github.com/kubernetes/minikube/pull/4338)
* Fix kvm remove when domain is not running [#4344](https://github.com/kubernetes/minikube/pull/4344)
* kvm2: Add support for --kvm-network to ensureNetwork [#4323](https://github.com/kubernetes/minikube/pull/4323)
* Get current profile if no arguments given [#4335](https://github.com/kubernetes/minikube/pull/4335)
* Skip kvm network deletion if private network doesn't exist [#4331](https://github.com/kubernetes/minikube/pull/4331)

Huge thank you for this release towards our contributors:

- Abdulla Bin Mustaqeem
- Anders Björklund
- Andy Daniels
- Archana Shinde
- Arnaud Jardiné
- Artiom Diomin
- Balint Pato
- Benn Linger
- Calin Don
- Chris Eason
- Cristian Măgherușan-Stanciu @magheru_san
- Deepika Pandhi
- Dmitry Budaev
- Don McCasland
- Douglas Thrift
- Elijah Oyekunle
- Filip Havlíček
- Guang Ya Liu
- Himanshu Pandey
- Igor Akkerman
- Ihor Dvoretskyi
- Jan Janik
- Jat
- Joel Smith
- Joji Mekkatt
- Marco Vito Moscaritolo
- Marcos Diez
- Martynas Pumputis
- Mas
- Maximilian Hess
- Medya Gh
- Miel Donkers
- Mike Lewis
- Oleg Atamanenko
- Om Kumar
- Pradip-Khakurel
- Pranav Jituri
- RA489
- Shahid Iqbal
- Sharif Elgamal
- Steven Davidovitz
- Thomas Bechtold
- Thomas Strömberg
- Tiago Ilieve
- Tobias Bradtke
- Toliver Jue
- Tom Reznik
- Yaroslav Skopets
- Yoan Blanc
- Zhongcheng Lao
- Zoran Regvart
- fenglixa
- flyingcircle
- jay vyas
- josedonizetti
- karmab
- kerami
- morvencao
- salamani
- u5surf
- wj24021040

## Version 1.1.0 - 2019-05-21

* Allow macOS to resolve service FQDNs during 'minikube tunnel' [#3464](https://github.com/kubernetes/minikube/pull/3464)
* Expose ‘—pod-network-cidr’ argument in minikube [#3892](https://github.com/kubernetes/minikube/pull/3892)
* Upgrade default Kubernetes release to v1.14.2 [#4279](https://github.com/kubernetes/minikube/pull/4279)
* Update to Podman 1.3 & CRIO v1.14.1 [#4299](https://github.com/kubernetes/minikube/pull/4299)
* Upgrade Docker, from 18.06.3-ce to 18.09.5 [#4204](https://github.com/kubernetes/minikube/pull/4204)
* Upgrade Docker, from 18.09.5 to 18.09.6 [#4296](https://github.com/kubernetes/minikube/pull/4296)
* Add Go modules support [#4241](https://github.com/kubernetes/minikube/pull/4241)
* Add more solutions messages [#4257](https://github.com/kubernetes/minikube/pull/4257)
* Add new kubectl command [#4193](https://github.com/kubernetes/minikube/pull/4193)
* Add solution text for common kvm2 and VirtualBox problems [#4198](https://github.com/kubernetes/minikube/pull/4198)
* Adding support for s390x [#4091](https://github.com/kubernetes/minikube/pull/4091)
* Allow minikube to function with misconfigured NO_PROXY value [#4229](https://github.com/kubernetes/minikube/pull/4229)
* Disable SystemVerification preflight on Kubernetes releases <1.13 [#4306](https://github.com/kubernetes/minikube/pull/4306)
* Don't attempt to pull docker images on relaunch [#4129](https://github.com/kubernetes/minikube/pull/4129)
* Fix location of Kubernetes binaries in cache directory [#4244](https://github.com/kubernetes/minikube/pull/4244)
* Fix registry addon ReplicationController template [#4220](https://github.com/kubernetes/minikube/pull/4220)
* Make default output of 'minikube start' consume fewer lines in the terminal [#4197](https://github.com/kubernetes/minikube/pull/4197)
* Make handling of stale mount pid files more robust [#4191](https://github.com/kubernetes/minikube/pull/4191)
* Make sure to start Docker, before getting version [#4307](https://github.com/kubernetes/minikube/pull/4307)
* Restart kube-proxy using kubeadm & add bootstrapper.WaitCluster [#4276](https://github.com/kubernetes/minikube/pull/4276)
* Return host IP when using vmware as vm driver. [#4255](https://github.com/kubernetes/minikube/pull/4255)
* Select an accessible image repository for some users [#3937](https://github.com/kubernetes/minikube/pull/3937)
* Set apiserver oom_adj to -10 to avoid OOMing before other pods [#4282](https://github.com/kubernetes/minikube/pull/4282)
* Standardize ASCII prefix for info, warning, and error messages [#4162](https://github.com/kubernetes/minikube/pull/4162)
* Unset the current-context after minikube stop [#4177](https://github.com/kubernetes/minikube/pull/4177)
* Validate kvm network exists [#4308](https://github.com/kubernetes/minikube/pull/4308)
* storageclass no longer beta #4148 [#4153](https://github.com/kubernetes/minikube/pull/4153)

Thank you to the contributors whose work made v1.1 into something we could all be proud of:

- Anders F Björklund
- Chris Eason
- Deepika Pandhi
- Himanshu Pandey
- Jan Janik
- Marcos Diez
- Maximilian Hess
- Medya Gh
- Sharif Elgamal
- Thomas Strömberg
- Tiago Ilieve
- Tobias Bradtke
- Zhongcheng Lao
- Zoran Regvart
- josedonizetti
- kerami
- salamani

## Version 1.0.1 - 2019-04-29

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
* Don't cache images when --vm-driver=none [#4059](https://github.com/kubernetes/minikube/pull/4059)
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
* Tell user given driver has been ignored if existing VM is different [#3374](https://github.com/kubernetes/minikube/pull/3374)

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
* fix: --format outputs any string, --https only substitute http URL scheme [#3114](https://github.com/kubernetes/minikube/pull/3114)
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
* cri-tools updated to 1.11.1 [#2986](https://github.com/kubernetes/minikube/pull/2986)
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
* Generated documentation for all commands. Documentation [is here](https://minikube.sigs.k8s.io/docs/).

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
