# Release Notes

## Version 1.34.0 - 2024-09-04

Breaking Changes:
* Bump minimum podman version to 4.9.0 [#19457](https://github.com/kubernetes/minikube/pull/19457)
* Disallow using Docker Desktop 4.34.0 
Features:
* Bump default Kubernetes version to v1.31.0 [#19435](https://github.com/kubernetes/minikube/pull/19435)
* Add new driver for macOS: vfkit [#19423](https://github.com/kubernetes/minikube/pull/19423)
* Add Parallels driver support for darwin/arm64 [#19373](https://github.com/kubernetes/minikube/pull/19373)
* Add new volcano addon [#18602](https://github.com/kubernetes/minikube/pull/18602)
* Addons ingress-dns: Added support for all architectures [#19198](https://github.com/kubernetes/minikube/pull/19198)
* Support privileged ports on WSL [#19370](https://github.com/kubernetes/minikube/pull/19370)
* VM drivers with docker container-runtime now use docker-buildx for image building [#19339](https://github.com/kubernetes/minikube/pull/19339)
* Support running x86 QEMU on arm64 [#19228](https://github.com/kubernetes/minikube/pull/19228)
* Add `-o json` option for `addon images` command [#19364](https://github.com/kubernetes/minikube/pull/19364)

Improvements:
* add -d shorthand for --driver [#19356](https://github.com/kubernetes/minikube/pull/19356)
* add -c shorthand for --container-runtime [#19217](https://github.com/kubernetes/minikube/pull/19217)
* kvm2: Don't delete the "default" libvirt network [#18920](https://github.com/kubernetes/minikube/pull/18920)
* Update MINIKUBE_HOME usage [#18648](https://github.com/kubernetes/minikube/pull/18648)
* CNI: Updated permissions to support network policies on kindnet [#19360](https://github.com/kubernetes/minikube/pull/19360)
* GPU: Set `NVIDIA_DRIVER_CAPABILITIES` to `all` when GPU is enabled [#19345](https://github.com/kubernetes/minikube/pull/19345)
* Improved error message when trying to use `mount` on system missing 9P [#18995](https://github.com/kubernetes/minikube/pull/18995)
* Improved error message when enabling KVM addons on non-KVM cluster [#19195](https://github.com/kubernetes/minikube/pull/19195)
* Added warning when loading image with wrong arch [#19229](https://github.com/kubernetes/minikube/pull/19229)
* `profile list --output json` handle empty config folder  [#16900](https://github.com/kubernetes/minikube/pull/16900)
* Check connectivity outside minikube when connectivity issuse [#18859](https://github.com/kubernetes/minikube/pull/18859)

Bugs:
* Fix not creating API server tunnel for QEMU w/ builtin network [#19191](https://github.com/kubernetes/minikube/pull/19191)
* Fix waiting for user input on firewall unblock when `--interactive=false` [#19531](https://github.com/kubernetes/minikube/pull/19531)
* Fix network retry check when subnet already in use for podman [#17779](https://github.com/kubernetes/minikube/pull/17779)
* Fix empty tarball when generating image save [#19312](https://github.com/kubernetes/minikube/pull/19312)
* Fix missing permission for kong-serviceaccount [#19002](https://github.com/kubernetes/minikube/pull/19002)

Version Upgrades:
* Addon cloud-spanner: Update cloud-spanner-emulator/emulator image from 1.5.17 to 1.5.23 [#19341](https://github.com/kubernetes/minikube/pull/19341) [#19501](https://github.com/kubernetes/minikube/pull/19501)
* Addon headlamp: Update headlamp-k8s/headlamp image from v0.23.2 to v0.25.0 [#18992](https://github.com/kubernetes/minikube/pull/18992) [#19152](https://github.com/kubernetes/minikube/pull/19152) [#19349](https://github.com/kubernetes/minikube/pull/19349)
* Addon kong: Update kong image from 3.6.1 to 3.7.1 [#19046](https://github.com/kubernetes/minikube/pull/19046) [#19124](https://github.com/kubernetes/minikube/pull/19124)
* Addon kubevirt: Update bitnami/kubectl image from 1.30.0 to 1.31.0 [#18929](https://github.com/kubernetes/minikube/pull/18929) [#19087](https://github.com/kubernetes/minikube/pull/19087) [#19313](https://github.com/kubernetes/minikube/pull/19313) [#19479](https://github.com/kubernetes/minikube/pull/19479)
* Addon ingress: Update ingress-nginx/controller image from v1.10.1 to v1.11.2 [#19302](https://github.com/kubernetes/minikube/pull/19302) [#19461](https://github.com/kubernetes/minikube/pull/19461)
* Addon inspektor-gadget: Update inspektor-gadget image from v0.27.0 to v0.32.0 [#18872](https://github.com/kubernetes/minikube/pull/18872) [#18931](https://github.com/kubernetes/minikube/pull/18931) [#19011](https://github.com/kubernetes/minikube/pull/19011) [#19166](https://github.com/kubernetes/minikube/pull/19166) [#19411](https://github.com/kubernetes/minikube/pull/19411) [#19554](https://github.com/kubernetes/minikube/pull/19554)
* Addon istio-provisioner: Update istio/operator image from 1.21.2 to 1.23.0 [#18932](https://github.com/kubernetes/minikube/pull/18932) [#19052](https://github.com/kubernetes/minikube/pull/19052) [#19167](https://github.com/kubernetes/minikube/pull/19167) [#19283](https://github.com/kubernetes/minikube/pull/19283) [#19450](https://github.com/kubernetes/minikube/pull/19450)
* Addon nvidia-device-plugin: Update nvidia/k8s-device-plugin image from v0.15.0 to v0.16.2 [#19162](https://github.com/kubernetes/minikube/pull/19162) [#19266](https://github.com/kubernetes/minikube/pull/19266) [#19336](https://github.com/kubernetes/minikube/pull/19336) [#19409](https://github.com/kubernetes/minikube/pull/19409)
* Addon metrics-server: Update metrics-server/metrics-server image from v0.7.1 to v0.7.2 [#19529](https://github.com/kubernetes/minikube/pull/19529)
* Addon YAKD: bump marcnuri/yakd image from 0.0.4 to 0.0.5 [#19145](https://github.com/kubernetes/minikube/pull/19145)
* CNI: Update calico from v3.27.3 to v3.28.1 [#18870](https://github.com/kubernetes/minikube/pull/18870) [#19377](https://github.com/kubernetes/minikube/pull/19377)
* CNI: Update cilium from v1.15.3 to v1.16.1 [#18925](https://github.com/kubernetes/minikube/pull/18925) [#19084](https://github.com/kubernetes/minikube/pull/19084) [#19247](https://github.com/kubernetes/minikube/pull/19247) [#19337](https://github.com/kubernetes/minikube/pull/19337) [#19476](https://github.com/kubernetes/minikube/pull/19476)
* CNI: Update kindnetd from v20240202-8f1494ea to v20240813-c6f155d6 [#18933](https://github.com/kubernetes/minikube/pull/18933) [#19252](https://github.com/kubernetes/minikube/pull/19252) [#19265](https://github.com/kubernetes/minikube/pull/19265) [#19307](https://github.com/kubernetes/minikube/pull/19307) [#19378](https://github.com/kubernetes/minikube/pull/19378) [#19446](https://github.com/kubernetes/minikube/pull/19446)
* CNI: Update flannel from v0.25.1 to v0.25.6 [#18966](https://github.com/kubernetes/minikube/pull/18966) [#19008](https://github.com/kubernetes/minikube/pull/19008) [#19085](https://github.com/kubernetes/minikube/pull/19085) [#19297](https://github.com/kubernetes/minikube/pull/19297) [#19522](https://github.com/kubernetes/minikube/pull/19522)
* Kicbase: Update nerdctld from 0.6.0 to 0.6.1 [#19282](https://github.com/kubernetes/minikube/pull/19282)
* Kicbase: Bump ubuntu:jammy from 20240427 to 20240808 [#19068](https://github.com/kubernetes/minikube/pull/19068) [#19184](https://github.com/kubernetes/minikube/pull/19184) [#19478](https://github.com/kubernetes/minikube/pull/19478)
* Kicbase/ISO: Update buildkit from v0.13.1 to v0.15.2 [#19024](https://github.com/kubernetes/minikube/pull/19024) [#19116](https://github.com/kubernetes/minikube/pull/19116) [#19264](https://github.com/kubernetes/minikube/pull/19264) [#19355](https://github.com/kubernetes/minikube/pull/19355) [#19452](https://github.com/kubernetes/minikube/pull/19452)
* Kicbase/ISO: Update cni-plugins from v1.4.1 to v1.5.1 [#19044](https://github.com/kubernetes/minikube/pull/19044) [#19128](https://github.com/kubernetes/minikube/pull/19128)
* Kicbase/ISO: Update containerd from v1.7.15 to v1.7.21 [#18934](https://github.com/kubernetes/minikube/pull/18934) [#19106](https://github.com/kubernetes/minikube/pull/19106) [#19186](https://github.com/kubernetes/minikube/pull/19186) [#19298](https://github.com/kubernetes/minikube/pull/19298) [#19521](https://github.com/kubernetes/minikube/pull/19521)
* Kicbase/ISO: Update cri-dockerd from v0.3.12 to v0.3.15 [#19199](https://github.com/kubernetes/minikube/pull/19199) [#19249](https://github.com/kubernetes/minikube/pull/19249)
* Kicbase/ISO: Update crun from 1.14.4 to 1.16.1 [#19112](https://github.com/kubernetes/minikube/pull/19112) [#19389](https://github.com/kubernetes/minikube/pull/19389) [#19443](https://github.com/kubernetes/minikube/pull/19443)
* Kicbase/ISO: Update docker from 26.0.2 to 27.2.0 [#18993](https://github.com/kubernetes/minikube/pull/18993) [#19038](https://github.com/kubernetes/minikube/pull/19038) [#19142](https://github.com/kubernetes/minikube/pull/19142) [#19153](https://github.com/kubernetes/minikube/pull/19153) [#19175](https://github.com/kubernetes/minikube/pull/19175) [#19319](https://github.com/kubernetes/minikube/pull/19319) [#19326](https://github.com/kubernetes/minikube/pull/19326) [#19429](https://github.com/kubernetes/minikube/pull/19429) [#19530](https://github.com/kubernetes/minikube/pull/19530)
* Kicbase/ISO: Update nerdctl from 1.7.5 to 1.7.6 [#18869](https://github.com/kubernetes/minikube/pull/18869)
* Kicbase/ISO: Update runc from v1.1.12 to v1.1.13 [#19104](https://github.com/kubernetes/minikube/pull/19104)

For a more detailed changelog, including changes occurring in pre-release versions, see [CHANGELOG.md](https://github.com/kubernetes/minikube/blob/master/CHANGELOG.md).

Thank you to our contributors for this release!

- Anders F Björklund
- Anjali Chaturvedi
- Artem Basalaev
- Benjamin P. Jung
- Daniel Iwaniec
- Dylan Piergies
- Gabriel Pelouze
- Hritesh Mondal
- Jack Brown
- Jeff MAURY
- Marc Nuri
- Matteo Mortari
- Medya Ghazizadeh
- Nir Soffer
- Philippe Miossec
- Predrag Rogic
- Radoslaw Smigielski
- Raghavendra Talur
- Sandipan Panda
- Steven Powell
- Sylvester Carolan
- Tom McLaughlin
- Tony-Sol
- aiyijing
- chubei
- daniel-iwaniec
- hritesh04
- joaquimrocha
- ljtian
- mitchell amihod
- shixiuguo
- sunyuxuan
- thomasjm
- tianlijun
- tianlj
- 錦南路之花
- 锦南路之花

Thank you to our PR reviewers for this release!

- spowelljr (67 comments)
- medyagh (53 comments)
- nirs (14 comments)
- cfergeau (4 comments)
- liangyuanpeng (2 comments)
- ComradeProgrammer (1 comments)
- afbjorklund (1 comments)
- aojea (1 comments)
- bobsira (1 comments)

Thank you to our triage members for this release!

- kundan2707 (55 comments)
- medyagh (29 comments)
- afbjorklund (28 comments)
- T-Lakshmi (20 comments)
- Ritikaa96 (16 comments)

Check out our [contributions leaderboard](https://minikube.sigs.k8s.io/docs/contrib/leaderboard/v1.34.0/) for this release!

## Version 1.33.1 - 2024-05-13

Bugs:
* Fix `DNSSEC validation failed` errors [#18830](https://github.com/kubernetes/minikube/pull/18830)
* Fix `too many open files` errors [#18832](https://github.com/kubernetes/minikube/pull/18832)
* CNI cilium: Fix cilium pods failing to start-up [#18846](https://github.com/kubernetes/minikube/pull/18846)
* Addon ingress: Fix enable failing on arm64 machines using VM driver [#18779](https://github.com/kubernetes/minikube/pull/18779)
* Addon kubeflow: Fix some components missing arm64 images [#18765](https://github.com/kubernetes/minikube/pull/18765)

Version Upgrades:
* Addon cloud-spanner: Update cloud-spanner-emulator/emulator image from 1.5.15 to 1.5.17 [#18773](https://github.com/kubernetes/minikube/pull/18773) [#18811](https://github.com/kubernetes/minikube/pull/18811)
* Addon headlamp: Update headlamp-k8s/headlamp image from v0.23.1 to v0.23.2 [#18793](https://github.com/kubernetes/minikube/pull/18793)
* Addon ingress: Update ingress-nginx/controller image from v1.10.0 to v1.10.1 [#18756](https://github.com/kubernetes/minikube/pull/18756)
* Addon istio-provisioner: Update istio/operator image from 1.21.1 to 1.21.2 [#18757](https://github.com/kubernetes/minikube/pull/18757)
* Addon kubevirt: Update bitnami/kubectl image from 1.29.3 to 1.30.0 [#18711](https://github.com/kubernetes/minikube/pull/18711) [#18771](https://github.com/kubernetes/minikube/pull/18771)
* Addon nvidia-device-plugin: Update nvidia/k8s-device-plugin image from v0.14.5 to v0.15.0 [#18703](https://github.com/kubernetes/minikube/pull/18703)
* CNI cilium: Update from v1.15.1 to v1.15.3 [#18846](https://github.com/kubernetes/minikube/pull/18846)
* High Availability: Update kube-vip from 0.7.1 to v0.8.0 [#18774](https://github.com/kubernetes/minikube/pull/18774)
* Kicbase/ISO: Update docker from 26.0.1 to 26.0.2 [#18706](https://github.com/kubernetes/minikube/pull/18706)
* Kicbase: Bump ubuntu:jammy from 20240227 to 20240427 [#18702](https://github.com/kubernetes/minikube/pull/18702) [#18769](https://github.com/kubernetes/minikube/pull/18769) [#18804](https://github.com/kubernetes/minikube/pull/18804)

For a more detailed changelog, see [CHANGELOG.md](https://github.com/kubernetes/minikube/blob/master/CHANGELOG.md).

Thank you to our contributors for this release!

- Bodhi Hu
- Jérémie Tarot
- Nir Soffer
- Predrag Rogic
- Steven Powell
- cuiyourong
- joaquimrocha

Thank you to our PR reviewers for this release!

- medyagh (9 comments)
- nirs (3 comments)
- llegolas (1 comments)
- spowelljr (1 comments)

Thank you to our triage members for this release!

- medyagh (6 comments)
- afbjorklund (5 comments)
- xcarolan (4 comments)
- nevotheless (3 comments)
- dasumner (2 comments)

Check out our [contributions leaderboard](https://minikube.sigs.k8s.io/docs/contrib/leaderboard/v1.33.1/) for this release!

## Version 1.33.0 - 2024-04-19

Features:
* Add support for Kubernetes v1.30 [#18669](https://github.com/kubernetes/minikube/pull/18669)
* Support exposing clusterIP services via `minikube service` [#17877](https://github.com/kubernetes/minikube/pull/17877)

Minor Improvements:
* Add active kubecontext to `minikube profile list` output [#17735](https://github.com/kubernetes/minikube/pull/17735)
* CNI calico: support kubeadm.pod-network-cidr [#18233](https://github.com/kubernetes/minikube/pull/18233)
* CNI bridge: Ensure pod communications are allowed [#16143](https://github.com/kubernetes/minikube/pull/16143)

Bugs:
* Fix unescaped local host regex [#18617](https://github.com/kubernetes/minikube/pull/18617)
* Fix regex on validateNetwork to support special characters [#18158](https://github.com/kubernetes/minikube/pull/18158)

Version Upgrades:
* Bump Kubernetes version default: v1.30.0 and latest: v1.30.0 [#18669](https://github.com/kubernetes/minikube/pull/18669)
* Addon headlamp: Update headlamp-k8s/headlamp image from v0.23.0 to 0.23.1 [#18517](https://github.com/kubernetes/minikube/pull/18517)
* Addon inspektor-gadget: Update inspektor-gadget image from v0.26.0 to v0.27.0 [#18588](https://github.com/kubernetes/minikube/pull/18588)
* Addon istio-provisioner: Update istio/operator image from 1.21.0 to 1.21.1 [#18644](https://github.com/kubernetes/minikube/pull/18644)
* Addon metrics-server: Update metrics-server/metrics-server image from v0.7.0 to v0.7.1 [#18551](https://github.com/kubernetes/minikube/pull/18551)
* CNI: Update calico from v3.27.0 to v3.27.3 [#18206](https://github.com/kubernetes/minikube/pull/18206)
* CNI: Update flannel from v0.24.4 to v0.25.1 [#18641](https://github.com/kubernetes/minikube/pull/18641)
* Kicbase/ISO: Update buildkit from v0.13.0 to v0.13.1 [#18566](https://github.com/kubernetes/minikube/pull/18566)
* Kicbase/ISO: Update containerd from v1.7.14 to v1.7.15 [#18621](https://github.com/kubernetes/minikube/pull/18621)
* Kicbase/ISO: Update cri-dockerd from v0.3.3 to v0.3.12 [#18585](https://github.com/kubernetes/minikube/pull/18585)
* Kicbase/ISO: Update crun from 1.14 to 1.14.4 [#18610](https://github.com/kubernetes/minikube/pull/18610)
* Kicbase/ISO: Update docker from 25.0.4 to 26.0.1 [#18485](https://github.com/kubernetes/minikube/pull/18485) [#18649](https://github.com/kubernetes/minikube/pull/18649)
* Kicbase/ISO: Update nerdctl from 1.7.4 to 1.7.5 [#18634](https://github.com/kubernetes/minikube/pull/18634)
* Kicbase: Update nerdctld from 0.5.1 to 0.6.0 [#18647](https://github.com/kubernetes/minikube/pull/18647)

For a more detailed changelog, including changes occurring in pre-release versions, see [CHANGELOG.md](https://github.com/kubernetes/minikube/blob/master/CHANGELOG.md).

Thank you to our contributors for this release!

- Jan Klippel
- Jeff MAURY
- Jesse Hathaway
- Maxime Brunet
- Medya Ghazizadeh
- Paul Rey
- Predrag Rogic
- Skalador
- Steven Powell
- alessandrocapanna
- depthlending
- guangwu
- joaquimrocha
- nikitakurakin
- racequite
- shixiuguo
- skoenig
- sunyuxuan
- syxunion
- Товарищ программист

Thank you to our PR reviewers for this release!

- medyagh (5 comments)
- spowelljr (4 comments)
- Shubham82 (2 comments)

Thank you to our triage members for this release!

- afbjorklund (21 comments)
- T-Lakshmi (15 comments)
- Ritikaa96 (12 comments)
- kundan2707 (8 comments)
- medyagh (7 comments)

## Version 1.33.0-beta.0 - 2024-03-26

Features:
* Support multi-control plane - HA clusters `--ha` [#17909](https://github.com/kubernetes/minikube/pull/17909)
* Addon gvisor: Add arm64 support [#18063](https://github.com/kubernetes/minikube/pull/18063) [#18453](https://github.com/kubernetes/minikube/pull/18453)
* New Addon: YAKD - Kubernetes Dashboard addon [#17775](https://github.com/kubernetes/minikube/pull/17775)

Minor Improvements:
* Addon auto-pause: Remove memory leak & add configurable interval [#17936](https://github.com/kubernetes/minikube/pull/17936)
* image build: Add `docker.io/library` to image short names [#16214](https://github.com/kubernetes/minikube/pull/16214)
* cp: Create directory if not present [#17715](https://github.com/kubernetes/minikube/pull/17715)
* Move errors getting logs into log output itself [#18007](https://github.com/kubernetes/minikube/pull/18007)
* Add default sysctls to allow privileged ports with no capabilities [#18421](https://github.com/kubernetes/minikube/pull/18421)
* Include extended attributes in preload tarballs [#17829](https://github.com/kubernetes/minikube/pull/17829)
* Apply `kubeadm.applyNodeLabels` label to all nodes [#16416](https://github.com/kubernetes/minikube/pull/16416)
* Limit driver status check to 20s [#17553](https://github.com/kubernetes/minikube/pull/17553)
* Include journalctl logs if systemd service fails to start [#17659](https://github.com/kubernetes/minikube/pull/17659)
* ISO: Add CONFIG_DM_MULTIPATH [#18277](https://github.com/kubernetes/minikube/pull/18277)
* ISO: Add CONFIG_QFMT_V2 for arm64 [#17991](https://github.com/kubernetes/minikube/pull/17991)
* ISO: Add CONFIG_CEPH_FS for arm64 [#18213](https://github.com/kubernetes/minikube/pull/18213)
* ISO: Add CONFIG_BPF for arm64 [#17206](https://github.com/kubernetes/minikube/pull/17206)

Bugs:
* Fix "Failed to enable container runtime: sudo systemctl restart cri-docker" [#17907](https://github.com/kubernetes/minikube/pull/17907)
* Fix containerd redownloading existing images on start [#17671](https://github.com/kubernetes/minikube/pull/17671)
* Fix kvm2 not detecting containerd preload [#17658](https://github.com/kubernetes/minikube/pull/17658)
* Fix modifying Docker binfmt config [#17830](https://github.com/kubernetes/minikube/pull/17830)
* Fix auto-pause addon [#17866](https://github.com/kubernetes/minikube/pull/17866)
* Fix not using preload with overlayfs storage driver [#18333](https://github.com/kubernetes/minikube/pull/18333)
* Fix image repositories not allowing subdomains with numbers [#17496](https://github.com/kubernetes/minikube/pull/17496)
* Fix stopping cluster when using kvm2 with containerd [#17967](https://github.com/kubernetes/minikube/pull/17967)
* Fix starting more than one cluster on kvm2 arm64 [#18241](https://github.com/kubernetes/minikube/pull/18241)
* Fix starting kvm2 clusters using Linux on arm64 Mac [#18239](https://github.com/kubernetes/minikube/pull/18239)
* Fix displaying error when deleting non-existing cluster [#17713](https://github.com/kubernetes/minikube/pull/17713)
* Fix no-limit not being respected on restart [#17598](https://github.com/kubernetes/minikube/pull/17598)
* Fix not applying `kubeadm.applyNodeLabels` label to nodes added after inital start [#16416](https://github.com/kubernetes/minikube/pull/16416)
* Fix logs delimiter output [#17734](https://github.com/kubernetes/minikube/pull/17734)

Version Upgrades:
* Bump Kubernetes version default: v1.29.3 and latest: v1.30.0-beta.0 [#17786](https://github.com/kubernetes/minikube/pull/17786)
* Addon cloud-spanner: Update cloud-spanner-emulator/emulator image from 1.5.11 to 1.5.15 [#17595](https://github.com/kubernetes/minikube/pull/17595) [#17847](https://github.com/kubernetes/minikube/pull/17847) [#18165](https://github.com/kubernetes/minikube/pull/18165) [#18431](https://github.com/kubernetes/minikube/pull/18431)
* Addon gcp-auth: Update k8s-minikube/gcp-auth-webhook image from v0.1.0 to v0.1.2 [#18222](https://github.com/kubernetes/minikube/pull/18222) [#18384](https://github.com/kubernetes/minikube/pull/18384)
* Addon headlamp: Update headlamp-k8s/headlamp image from v0.20.1 to v0.23.0 [#17586](https://github.com/kubernetes/minikube/pull/17586) [#17846](https://github.com/kubernetes/minikube/pull/17846) [#18320](https://github.com/kubernetes/minikube/pull/18320)
* Addon ingress: Update ingress-nginx/controller image from v1.9.4 to v1.10.0 [#17848](https://github.com/kubernetes/minikube/pull/17848) [#18166](https://github.com/kubernetes/minikube/pull/18166) [#18284](https://github.com/kubernetes/minikube/pull/18284)
* Addon inspektor-gadget: Update inspektor-gadget/inspektor-gadget image from v0.22.0 to v0.26.0 [#17740](https://github.com/kubernetes/minikube/pull/17740) [#17885](https://github.com/kubernetes/minikube/pull/17885) [#18169](https://github.com/kubernetes/minikube/pull/18169) [#18358](https://github.com/kubernetes/minikube/pull/18358)
* Addon istio-provisioner: Update istio/operator image from 1.19.3 to 1.21.0 [#17651](https://github.com/kubernetes/minikube/pull/17651) [#17777](https://github.com/kubernetes/minikube/pull/17777) [#17957](https://github.com/kubernetes/minikube/pull/17957) [#18168](https://github.com/kubernetes/minikube/pull/18168) [#18429](https://github.com/kubernetes/minikube/pull/18429)
* Addon kong: Update kong image from 3.4.2 to 3.6.1 [#17605](https://github.com/kubernetes/minikube/pull/17605) [#18200](https://github.com/kubernetes/minikube/pull/18200) [#18350](https://github.com/kubernetes/minikube/pull/18350)
* Addon kubevirt: Update bitnami/kubectl image from 1.24.7 to 1.29.3 [#18170](https://github.com/kubernetes/minikube/pull/18170) [#18187](https://github.com/kubernetes/minikube/pull/18187) [#18427](https://github.com/kubernetes/minikube/pull/18427)
* Addon metrics-server: Update metrics-server/metrics-server image from v0.6.4 to v0.7.0 [#18051](https://github.com/kubernetes/minikube/pull/18051)
* Addon nvidia-device-plugin: Update nvidia/k8s-device-plugin image from v0.14.2 to v0.14.5 [#17623](https://github.com/kubernetes/minikube/pull/17623) [#18171](https://github.com/kubernetes/minikube/pull/18171) [#18283](https://github.com/kubernetes/minikube/pull/18283)
* Addon registry: Update k8s-minikube/kube-registry-proxy image from 0.0.5 to 0.0.6 [#18454](https://github.com/kubernetes/minikube/pull/18454)
* CNI: Update calico from v3.26.3 to v3.27.0 [#17644](https://github.com/kubernetes/minikube/pull/17644) [#17824](https://github.com/kubernetes/minikube/pull/17824)
* CNI: Update cilium from v1.12.3 to v1.15.1 [#18259](https://github.com/kubernetes/minikube/pull/18259)
* CNI: Update flannel from v0.22.3 to v0.24.4 [#17837](https://github.com/kubernetes/minikube/pull/17837) [#17975](https://github.com/kubernetes/minikube/pull/17975) [#18014](https://github.com/kubernetes/minikube/pull/18014) [#18500](https://github.com/kubernetes/minikube/pull/18500)
* CNI: Update kindnetd from v20230809-80a64d96 to v20240202-8f1494ea [#18167](https://github.com/kubernetes/minikube/pull/18167)
* Kicbase/ISO: Update buildkit from v0.12.3 to v0.13.0 [#17738](https://github.com/kubernetes/minikube/pull/17738) [#18375](https://github.com/kubernetes/minikube/pull/18375)
* Kicbase/ISO: Update cni-plugins from v1.3.0 to v1.4.1 [#17761](https://github.com/kubernetes/minikube/pull/17761) [#18375](https://github.com/kubernetes/minikube/pull/18375)
* Kicbase/ISO: Update containerd from v1.7.8 to v1.7.14 [#17634](https://github.com/kubernetes/minikube/pull/17634) [#17711](https://github.com/kubernetes/minikube/pull/17711) [#17765](https://github.com/kubernetes/minikube/pull/17765) [#18375](https://github.com/kubernetes/minikube/pull/18375)
* Kicbase/ISO: Update docker from 24.0.7 to 25.0.4 [#18375](https://github.com/kubernetes/minikube/pull/18375)
* Kicbase/ISO: Update Go from 1.21.3 to 1.22.1 [#17619](https://github.com/kubernetes/minikube/pull/17619) [#17760](https://github.com/kubernetes/minikube/pull/17760) [#17953](https://github.com/kubernetes/minikube/pull/17953) [#18197](https://github.com/kubernetes/minikube/pull/18197) [#18375](https://github.com/kubernetes/minikube/pull/18375)
* Kicbase/ISO: Update nerdctl from 1.6.2 to 1.7.4 [#17565](https://github.com/kubernetes/minikube/pull/17565) [#17703](https://github.com/kubernetes/minikube/pull/17703) [#17806](https://github.com/kubernetes/minikube/pull/17806) [#18375](https://github.com/kubernetes/minikube/pull/18375)
* Kicbase/ISO: Update runc from v1.1.9 to v1.1.12 [#17581](https://github.com/kubernetes/minikube/pull/17581) [#18020](https://github.com/kubernetes/minikube/pull/18020) [#18375](https://github.com/kubernetes/minikube/pull/18375)
* Kicbase: Update nerdctld from 0.2.0 to 0.5.1 [#17764](https://github.com/kubernetes/minikube/pull/17764) [#17857](https://github.com/kubernetes/minikube/pull/17857)
* Kicbase: Update ubuntu:jammy from 20231004 to 20240227 [#17719](https://github.com/kubernetes/minikube/pull/17719) [#17822](https://github.com/kubernetes/minikube/pull/17822) [#18244](https://github.com/kubernetes/minikube/pull/18244) [#18375](https://github.com/kubernetes/minikube/pull/18375)
* ISO: Update cri-o from v1.24.1 to v1.29.1 [#18020](https://github.com/kubernetes/minikube/pull/18020)

For a more detailed changelog, see [CHANGELOG.md](https://github.com/kubernetes/minikube/blob/master/CHANGELOG.md).

Thank you to our contributors for this release!

- Alberto Faria
- Anders F Björklund
- Blaine Gardner
- Camille Clayton
- Chase, Justin M
- Dani
- Eng Zer Jun
- Francis Laniel
- Jan Klippel
- Jeff MAURY
- Jin Li
- Jongwoo Han
- Marc Nuri
- Marcell Martini
- Marcus Dunn
- Mark Moretto
- Martin Jirku
- Medya Ghazizadeh
- Nir Soffer
- Predrag Rogic
- Pris Nasrat
- Raiden Shogun
- Sandipan Panda
- Sonu Kumar Singh
- Steven Powell
- Tarishi Jain
- Timothée Ravier
- Yuri Astrakhan
- andy
- chahatjaink
- coderrick
- joaquimrocha
- lixin18
- ljtian
- mahmut
- mattrobinsonsre
- prnvkv
- shixiuguo
- sunyuxuan
- sunyuxuna
- syxunion
- tianlj
- zdxgs
- zjx20

Thank you to our PR reviewers for this release!

- spowelljr (55 comments)
- medyagh (27 comments)
- afbjorklund (14 comments)
- liangyuanpeng (11 comments)
- prezha (4 comments)
- ComradeProgrammer (3 comments)
- acumino (2 comments)
- aiyijing (2 comments)
- Fenrur (1 comments)
- allenhaozi (1 comments)
- dharmit (1 comments)
- maximiliankolb (1 comments)
- neolit123 (1 comments)

Thank you to our triage members for this release!

- afbjorklund (70 comments)
- caerulescens (37 comments)
- T-Lakshmi (31 comments)
- spowelljr (22 comments)
- kundan2707 (20 comments)

Check out our [contributions leaderboard](https://minikube.sigs.k8s.io/docs/contrib/leaderboard/v1.33.0-beta.0/) for this release!

## Version 1.32.0 - 2023-11-07

Features:
* rootless: support `--container-runtime=docker` [#17520](https://github.com/kubernetes/minikube/pull/17520)

Minor Improvements:
* Install NVIDIA container toolkit during image build (offline support) [#17516](https://github.com/kubernetes/minikube/pull/17516)

Bugs:
* Fix no-limit option for config validation [#17530](https://github.com/kubernetes/minikube/pull/17530)

Version Upgrades:
* Addon ingress: Update ingress-nginx/controller image from v1.9.3 to v1.9.4 [#17525](https://github.com/kubernetes/minikube/pull/17525)
* Addon inspektor-gadget: Update inspektor-gadget image from v0.21.0 to v0.22.0 [#17550](https://github.com/kubernetes/minikube/pull/17550)
* Addon kong: Update kong/kubernetes-ingress-controller image from 2.9.3 to 2.12.0 [#17526](https://github.com/kubernetes/minikube/pull/17526)
* Addon nvidia-device-plugin: Update nvidia/k8s-device-plugin image from v0.14.1 to v0.14.2 [#17523](https://github.com/kubernetes/minikube/pull/17523)
* Kicbase/ISO: Update buildkit from v0.12.2 to v0.12.3 [#17486](https://github.com/kubernetes/minikube/pull/17486)
* Kicbase/ISO: Update containerd from v1.7.7 to v1.7.8 [#17527](https://github.com/kubernetes/minikube/pull/17527)
* Kicbase/ISO: Update docker from 24.0.6 to 24.0.7 [#17545](https://github.com/kubernetes/minikube/pull/17545)

For a more detailed changelog, including changes occurring in pre-release versions, see [CHANGELOG.md](https://github.com/kubernetes/minikube/blob/master/CHANGELOG.md).

Thank you to our contributors for this release!

- Akihiro Suda
- Christian Bergschneider
- Jeff MAURY
- Medya Ghazizadeh
- Raiden Shogun
- Steven Powell

Thank you to our PR reviewers for this release!

- medyagh (1 comments)
- r0b2g1t (1 comments)

Thank you to our triage members for this release!

- willsu (2 comments)
- afbjorklund (1 comments)
- ankur0904 (1 comments)
- ceelian (1 comments)
- idoly (1 comments)

## Version 1.32.0-beta.0 - 2023-10-27

Features:
* NVIDIA GPU support with new `--gpus=nvidia` flag for docker driver [#15927](https://github.com/kubernetes/minikube/pull/15927) [#17314](https://github.com/kubernetes/minikube/pull/17314) [#17488](https://github.com/kubernetes/minikube/pull/17488)
* New `kubeflow` addon [#17114](https://github.com/kubernetes/minikube/pull/17114)
* New `local-path-provisioner` addon [#15062](https://github.com/kubernetes/minikube/pull/15062)
* Kicbase: Add `no-limit` option to `--cpus` & `--memory` flags [#17491](https://github.com/kubernetes/minikube/pull/17491)

Minor Improvements:
* Hyper-V: Add memory validation for odd numbers [#17325](https://github.com/kubernetes/minikube/pull/17325)
* QEMU: Improve cpu type and IP detection [#17217](https://github.com/kubernetes/minikube/pull/17217)
* Mask http(s)_proxy password from startup output [#17116](https://github.com/kubernetes/minikube/pull/17116)
* `--delete-on-faliure` also recreates cluster for kubeadm failures [#16890](https://github.com/kubernetes/minikube/pull/16890)
* Addon auto-pause: Configure intervals using `--auto-pause-interval` [#17070](https://github.com/kubernetes/minikube/pull/17070)
* `--kubernetes-version` checks GitHub for version validation and improved error output for invalid versions [#16865](https://github.com/kubernetes/minikube/pull/16865)

Bugs:
* QEMU: Fix addons failing to enable [#17402](https://github.com/kubernetes/minikube/pull/17402)
* Fix downloading the wrong kubeadm images for k8s versions after minikube release [#17373](https://github.com/kubernetes/minikube/pull/17373)
* Fix enabling & disabling addons with non-existing cluster [#17324](https://github.com/kubernetes/minikube/pull/17324)
* Fix delete if container-runtime doesn't exist [#17347](https://github.com/kubernetes/minikube/pull/17347)
* Fix network not found not being detected on new Docker versions [#17323](https://github.com/kubernetes/minikube/pull/17323)
* Fix addon registry doesn't follow Minikube DNS domain name configuration (--dns-domain) [#15585](https://github.com/kubernetes/minikube/pull/15585)

Version Upgrades:
* Bump Kubernetes version default: v1.28.3 and latest: v1.28.3 [#17463](https://github.com/kubernetes/minikube/pull/17463)
* Addon cloud-spanner: Update cloud-spanner-emulator/emulator image from 1.5.9 to 1.5.11 [#17225](https://github.com/kubernetes/minikube/pull/17225) [#17259](https://github.com/kubernetes/minikube/pull/17259)
* Addon headlamp: Update headlamp-k8s/headlamp image from v0.19.0 to v0.20.1 [#17135](https://github.com/kubernetes/minikube/pull/17135) [#17365](https://github.com/kubernetes/minikube/pull/17365)
* Addon ingress: Update ingress-nginx/controller image from v1.8.1 to v1.9.3 [#17223](https://github.com/kubernetes/minikube/pull/17223) [#17297](https://github.com/kubernetes/minikube/pull/17297) [#17348](https://github.com/kubernetes/minikube/pull/17348) [#17421](https://github.com/kubernetes/minikube/pull/17421)
* Addon inspektor-gadget: Update inspektor-gadget image from v0.19.0 to v0.21.0 [#17176](https://github.com/kubernetes/minikube/pull/17176) [#17340](https://github.com/kubernetes/minikube/pull/17340)
* Addon istio-provisioner: Update istio/operator image from 1.12.2 to 1.19.3 [#17383](https://github.com/kubernetes/minikube/pull/17383) [#17436](https://github.com/kubernetes/minikube/pull/17436)
* Addon kong: Update kong image from 3.2 to 3.4.2 [#17485](https://github.com/kubernetes/minikube/pull/17485)
* Addon registry: Update registry image from 2.8.1 to 2.8.3 [#17382](https://github.com/kubernetes/minikube/pull/17382) [#17467](https://github.com/kubernetes/minikube/pull/17467)
* CNI: Update calico from v3.26.1 to v3.26.3 [#17363](https://github.com/kubernetes/minikube/pull/17363) [#17375](https://github.com/kubernetes/minikube/pull/17375)
* CNI: Update flannel from v0.22.1 to v0.22.3 [#17102](https://github.com/kubernetes/minikube/pull/17102) [#17263](https://github.com/kubernetes/minikube/pull/17263)
* CNI: Update kindnetd from v20230511-dc714da8 to v20230809-80a64d96 [#17233](https://github.com/kubernetes/minikube/pull/17233)
* Kicbase/ISO: Update buildkit from v0.11.6 to v0.12.2 [#17194](https://github.com/kubernetes/minikube/pull/17194)
* Kicbase/ISO: Update containerd from v1.7.3 to v1.7.7 [#17243](https://github.com/kubernetes/minikube/pull/17243) [#17466](https://github.com/kubernetes/minikube/pull/17466)
* Kicbase/ISO: Update crictl from v1.21.0 to v1.28.0 [#17240](https://github.com/kubernetes/minikube/pull/17240)
* Kicbase/ISO: Update docker from 24.0.4 to 24.0.6 [#17120](https://github.com/kubernetes/minikube/pull/17120) [#17207](https://github.com/kubernetes/minikube/pull/17207)
* Kicbase/ISO: Update nerdctl from 1.0.0 to 1.6.2 [#17145](https://github.com/kubernetes/minikube/pull/17145) [#17339](https://github.com/kubernetes/minikube/pull/17339) [#17434](https://github.com/kubernetes/minikube/pull/17434)
* Kicbase/ISO: Update runc from v1.1.7 to v1.1.9 [#17250](https://github.com/kubernetes/minikube/pull/17250)
* Kicbase: Bump ubuntu:jammy from 20230624 to 20231004 [#17086](https://github.com/kubernetes/minikube/pull/17086) [#17174](https://github.com/kubernetes/minikube/pull/17174) [#17345](https://github.com/kubernetes/minikube/pull/17345) [#17423](https://github.com/kubernetes/minikube/pull/17423)

For a more detailed changelog, see [CHANGELOG.md](https://github.com/kubernetes/minikube/blob/master/CHANGELOG.md).

Thank you to our contributors for this release!

- Anders F Björklund
- Dobes Vandermeer
- Emmanuel Chee-zaram Okeke
- Jeff MAURY
- Judah Nouriyelian
- Medya Ghazizadeh
- OneEpitome
- Piotr Resztak
- Predrag Rogic
- Raghavendra Talur
- Raiden Shogun
- Renato Moutinho
- Renato Silva
- Seongbin Hong
- Steven Powell
- Tristan Rice
- Wiktor Zając
- aiyijing
- jeremylinux-github
- joaquimrocha
- mahmut
- rogermm
- sunyuxuan
- tianlijun
- weidong
- Товарищ программист

Thank you to our PR reviewers for this release!

- medyagh (38 comments)
- spowelljr (19 comments)
- aiyijing (2 comments)
- Lyllt8 (1 comments)
- afbjorklund (1 comments)
- andresmmujica (1 comments)

Thank you to our triage members for this release!

- afbjorklund (32 comments)
- rmsilva1973 (27 comments)
- pnasrat (25 comments)
- spowelljr (21 comments)
- megazone23 (11 comments)

## Version 1.31.2 - 2023-08-16

docker-env Regression:
* Create `~/.ssh` directory if missing [#16934](https://github.com/kubernetes/minikube/pull/16934)
* Fix adding guest to `~/.ssh/known_hosts` when not needed [#17030](https://github.com/kubernetes/minikube/pull/17030)

Minor Improvements:
* Verify containerd storage separately from docker [#16972](https://github.com/kubernetes/minikube/pull/16972)

Version Upgrades:
* Bump Kubernetes version default: v1.27.4 and latest: v1.28.0-rc.1 [#17011](https://github.com/kubernetes/minikube/pull/17011) [#17051](https://github.com/kubernetes/minikube/pull/17051)
* Addon cloud-spanner: Update cloud-spanner-emulator/emulator image from 1.5.7 to 1.5.9 [#17017](https://github.com/kubernetes/minikube/pull/17017) [#17044](https://github.com/kubernetes/minikube/pull/17044)
* Addon headlamp: Update headlamp-k8s/headlamp image from v0.18.0 to v0.19.0 [#16992](https://github.com/kubernetes/minikube/pull/16992)
* Addon inspektor-gadget: Update inspektor-gadget image from v0.18.1 to v0.19.0 [#17016](https://github.com/kubernetes/minikube/pull/17016)
* Addon metrics-server: Update metrics-server/metrics-server image from v0.6.3 to v0.6.4 [#16969](https://github.com/kubernetes/minikube/pull/16969)
* CNI flannel: Update from v0.22.0 to v0.22.1 [#16968](https://github.com/kubernetes/minikube/pull/16968)

For a more detailed changelog, see [CHANGELOG.md](https://github.com/kubernetes/minikube/blob/master/CHANGELOG.md).

Thank you to our contributors for this release!

- Alex Serbul
- Anders F Björklund
- Jeff MAURY
- Medya Ghazizadeh
- Michelle Thompson
- Predrag Rogic
- Seth Rylan Gainey
- Steven Powell
- aiyijing
- joaquimrocha
- renyanda
- shixiuguo
- sunyuxuan
- Товарищ программист

Thank you to our PR reviewers for this release!

- medyagh (8 comments)
- spowelljr (2 comments)
- ComradeProgrammer (1 comments)
- Lyllt8 (1 comments)
- aiyijing (1 comments)

Thank you to our triage members for this release!

- afbjorklund (6 comments)
- vaibhav2107 (5 comments)
- kundan2707 (3 comments)
- spowelljr (3 comments)
- ao390 (2 comments)

Check out our [contributions leaderboard](https://minikube.sigs.k8s.io/docs/contrib/leaderboard/v1.31.2/) for this release!

## Version 1.31.1 - 2023-07-20

* cni: Fix regression in auto selection [#16912](https://github.com/kubernetes/minikube/pull/16912)

For a more detailed changelog, see [CHANGELOG.md](https://github.com/kubernetes/minikube/blob/master/CHANGELOG.md).

Thank you to our contributors for this release!

- Jeff MAURY
- Medya Ghazizadeh
- Steven Powell

Thank you to our triage members for this release!

- afbjorklund (5 comments)
- torenware (5 comments)
- mprimeaux (3 comments)
- prezha (3 comments)
- spowelljr (1 comments)

Check out our [contributions leaderboard](https://minikube.sigs.k8s.io/docs/contrib/leaderboard/v1.31.1/) for this release!

## Version 1.31.0 - 2023-07-18

Features:
* Add back VMware driver support [#16796](https://github.com/kubernetes/minikube/pull/16796)
* `docker-env` supports the containerd runtime (experimental) [#15452](https://github.com/kubernetes/minikube/pull/15452) [#16761](https://github.com/kubernetes/minikube/pull/16761)
* Automatically renew expired kubeadm certs [#16249](https://github.com/kubernetes/minikube/pull/16249)
* New addon inspektor-gadget [#15869](https://github.com/kubernetes/minikube/pull/15869)

Major Improvements:
* VM drivers: Fix all images getting removed on stop/start (40% start speedup) [#16655](https://github.com/kubernetes/minikube/pull/16655)
* Addon registry: Add support for all architectures [#16577](https://github.com/kubernetes/minikube/pull/16577)
* QEMU: Fix failing to interact with cluster after upgrading QEMU version [#16853](https://github.com/kubernetes/minikube/pull/16853)
* macOS/QEMU: Auto unblock bootpd from firewall if blocking socket_vmnet network [#16714](https://github.com/kubernetes/minikube/pull/16714) [#16789](https://github.com/kubernetes/minikube/pull/16789)
* `minikube cp` supports providing directory as a target [#15519](https://github.com/kubernetes/minikube/pull/15519)

Minor Improvements:
* Always use cni unless running with dockershim [#14780](https://github.com/kubernetes/minikube/pull/14780)
* none driver: Check for CNI plugins before starting cluster [#16419](https://github.com/kubernetes/minikube/pull/16419)
* QEMU: Add ability to create extra disks [#15887](https://github.com/kubernetes/minikube/pull/15887)
* --kubernetes-version: Assume latest patch version if not specified [#16569](https://github.com/kubernetes/minikube/pull/16569)
* audit: Set default max file size [#16543](https://github.com/kubernetes/minikube/pull/16543)
* service: Fail if no pods available [#15079](https://github.com/kubernetes/minikube/pull/15079)
* docker/podman driver: Use buildx for `image build` command [#16252](https://github.com/kubernetes/minikube/pull/16252)
* Addon gvisor: Simplify runtime configuration and use latest version [#14996](https://github.com/kubernetes/minikube/pull/14996)
* Add PowerShell code completion [#16232](https://github.com/kubernetes/minikube/pull/16232)
* build: Support DOS-style path for Dockerfile path [#15074](https://github.com/kubernetes/minikube/pull/15074)

Bugs:
* none driver: Fix `minikube start` not working without `sudo` [#16408](https://github.com/kubernetes/minikube/pull/16408)
* none driver: Fix `minikube image build` [#16386](https://github.com/kubernetes/minikube/pull/16386)
* Fix only allowing one global tunnel [#16839](https://github.com/kubernetes/minikube/pull/16839)
* Fix enabling addons when --no-kubernetes [#15003](https://github.com/kubernetes/minikube/pull/15003)
* Fix enabling addons on a paused cluster [#15868](https://github.com/kubernetes/minikube/pull/15868)
* Fix waiting for kicbase downloads on VM drivers [#16695](https://github.com/kubernetes/minikube/pull/16695)
* image list: Fix only outputting single tag of image with multiple tags [#16578](https://github.com/kubernetes/minikube/pull/16578)
* Addons: Fix cloud-spanner and headlamp incorrect file permissions [#16413](https://github.com/kubernetes/minikube/pull/16413)
* Fix csi-hostpath not allowing custom registry [#16395](https://github.com/kubernetes/minikube/pull/16395)
* Fix mount cleaning mechanism [#15782](https://github.com/kubernetes/minikube/pull/15782)
* Fix kubectl tab-completion and improve error messages [#14868](https://github.com/kubernetes/minikube/pull/14868
* Fix help text not being translated [#16850](https://github.com/kubernetes/minikube/pull/16850) [#16852](https://github.com/kubernetes/minikube/pull/16852)

New ISO Modules:
* Add BINFMT_MISC [#16712](https://github.com/kubernetes/minikube/pull/16712)
* Add BPF_SYSCALL to arm64 [#15164](https://github.com/kubernetes/minikube/pull/15164)
* Add GENEVE [#15665](https://github.com/kubernetes/minikube/pull/15665)
* add BLK_DEV_RBD & CEPH_LIB to arm64 [#16019](https://github.com/kubernetes/minikube/pull/16019)

Version Upgrades:
* Bump Kubernetes version default: v1.27.3 and latest: v1.27.3 [#16718](https://github.com/kubernetes/minikube/pull/16718)
* Addon cloud-spanner: Update cloud-spanner-emulator/emulator image from 1.5.2 to 1.5.7 [#16248](https://github.com/kubernetes/minikube/pull/16248) [#16352](https://github.com/kubernetes/minikube/pull/16352) [#16587](https://github.com/kubernetes/minikube/pull/16587) [#16652](https://github.com/kubernetes/minikube/pull/16652) [#16845](https://github.com/kubernetes/minikube/pull/16845)
* Addon gcp-auth: Update ingress-nginx/kube-webhook-certgen image from v20230312-helm-chart-4.5.2-28-g66a760794 to v20230407 [#16601](https://github.com/kubernetes/minikube/pull/16601)
* Addon gcp-auth: Update k8s-minikube/gcp-auth-webhook image from v0.0.14 to v0.1.0 [#16573](https://github.com/kubernetes/minikube/pull/16573)
* Addon headlamp: Update headlamp-k8s/headlamp image version from v0.16.0 to v0.18.0 [#16399](https://github.com/kubernetes/minikube/pull/16399) [#16540](https://github.com/kubernetes/minikube/pull/16540) [#16721](https://github.com/kubernetes/minikube/pull/16721)
* Addon ingress: Update ingress-nginx/controller image from v1.7.0 to v1.8.1 [#16601](https://github.com/kubernetes/minikube/pull/16601) [#16832](https://github.com/kubernetes/minikube/pull/16832)
* Addon ingress: Update ingress-nginx/kube-webhook-certgen image from v20230312-helm-chart-4.5.2-28-g66a760794 to v20230407 [#16601](https://github.com/kubernetes/minikube/pull/16601)
* Addon kong: Update kong image from 2.7 to 3.2 [#16424](https://github.com/kubernetes/minikube/pull/16424)
* Addon kong: Update kong/kubernetes-ingress-controller image from 2.1.1 to 2.9.3 [#16424](https://github.com/kubernetes/minikube/pull/16424)
* CNI calico: Update from v3.24.5 to v3.26.1 [#16144](https://github.com/kubernetes/minikube/pull/16144) [#16596](https://github.com/kubernetes/minikube/pull/16596) [#16732](https://github.com/kubernetes/minikube/pull/16732)
* CNI flannel: Update from v0.20.2 to v0.22.0 [#16074](https://github.com/kubernetes/minikube/pull/16074) [#16435](https://github.com/kubernetes/minikube/pull/16435) [#16597](https://github.com/kubernetes/minikube/pull/16597)
* CNI kindnet: Update from v20230330-48f316cd to v20230511-dc714da8 [#16488](https://github.com/kubernetes/minikube/pull/16488)
* Kicbase: Update base image from ubuntu:focal-20230308 to ubuntu:jammy-20230624 [#16069](https://github.com/kubernetes/minikube/pull/16069) [#16632](https://github.com/kubernetes/minikube/pull/16632) [#16731](https://github.com/kubernetes/minikube/pull/16731) [#16834](https://github.com/kubernetes/minikube/pull/16834)
* Kicbase/ISO: Update buildkit from v0.11.4 to v0.11.6 [#16426](https://github.com/kubernetes/minikube/pull/16426)
* Kicbase/ISO: Update cni-plugins from v0.8.5 to v1.3.0 [#16582](https://github.com/kubernetes/minikube/pull/16582)
* Kicbase/ISO: Update containerd from v1.7.0 to v1.7.1 [#16501](https://github.com/kubernetes/minikube/pull/16501)
* Kicbase/ISO: Update containerd from v1.7.1 to v1.7.2 [#16634](https://github.com/kubernetes/minikube/pull/16634)
* Kicbase/ISO: Update cri-dockerd from v0.3.1 to v0.3.3 [#16506](https://github.com/kubernetes/minikube/pull/16506) [#16703](https://github.com/kubernetes/minikube/pull/16703)
* Kicbase/ISO: Update docker from 20.10.23 to 24.0.4 [#16572](https://github.com/kubernetes/minikube/pull/16572) [#16612](https://github.com/kubernetes/minikube/pull/16612) [#16875](https://github.com/kubernetes/minikube/pull/16875)
* Kicbase/ISO: Update runc from v1.1.5 to v1.1.7 [#16417](https://github.com/kubernetes/minikube/pull/16417)

For a more detailed changelog, see [CHANGELOG.md](https://github.com/kubernetes/minikube/blob/master/CHANGELOG.md).

Thank you to our contributors for this release!

- AiYijing
- Aleksandr Chebotov
- Anders F Björklund
- Armel Soro
- Asbjørn Apeland
- Begula
- Blaine Gardner
- Bogdan Luca
- Fabricio Voznika
- Jeff MAURY
- Joe Bowbeer
- Juan Martín Loyola
- Judah Nouriyelian
- Kemal Akkoyun
- Max Cascone
- Medya Ghazizadeh
- Michele Sorcinelli
- Oldřich Jedlička
- Ricky Sadowski
- Sharran
- Steven Powell
- Terry Moschou
- Tongyao Si
- Vedant
- Viktor Gamov
- W. Duncan Fraser
- Yuiko Mouri
- aiyijing
- cui fliter
- guoguangwu
- himalayanZephyr
- joaquimrocha
- lixin18
- piljoong
- salasberryfin
- shixiuguo
- sunyuxuan
- syxunion
- tianlj
- tzzcfrank
- vgnshiyer
- winkelino
- x7upLime
- yolossn
- zhengtianbao
- Товарищ программист

Thank you to our PR reviewers for this release!

- spowelljr (180 comments)
- medyagh (64 comments)
- eiffel-fl (16 comments)
- afbjorklund (11 comments)
- aiyijing (9 comments)
- atoato88 (6 comments)
- BenTheElder (2 comments)
- travisn (2 comments)
- ComradeProgrammer (1 comments)
- Kimi450 (1 comments)
- alban (1 comments)
- mprimeaux (1 comments)
- shaneutt (1 comments)
- t-inu (1 comments)

Thank you to our triage members for this release!

- afbjorklund (30 comments)
- spowelljr (24 comments)
- kundan2707 (12 comments)
- mqasimsarfraz (6 comments)
- ShardulPrabhu (5 comments)

Check out our [contributions leaderboard](https://minikube.sigs.k8s.io/docs/contrib/leaderboard/v1.31.0/) for this release!

## Version 1.30.1 - 2023-04-04

* Docker driver: Fix incorrectly stating `Image was not built for the current minikube` [#16226](https://github.com/kubernetes/minikube/pull/16226)
* Mark VMware driver as unsupported  [#16233](https://github.com/kubernetes/minikube/pull/16233)

For a more detailed changelog, see [CHANGELOG.md](https://github.com/kubernetes/minikube/blob/master/CHANGELOG.md).

Thank you to our contributors for this release!

- Juan Martin Loyola
- Medya Ghazizadeh
- Steven Powell

Thank you to our PR reviewers for this release!

- medyagh (1 comments)

Thank you to our triage members for this release!

- afbjorklund (8 comments)
- spowelljr (6 comments)
- kundan2707 (2 comments)
- medyagh (1 comments)
- rafariossaa (1 comments)

Check out our [contributions leaderboard](https://minikube.sigs.k8s.io/docs/contrib/leaderboard/v1.30.1/) for this release!

## Version 1.30.0 - 2023-04-03

Features:
* Implement experimental QEMU on Windows [#15781](https://github.com/kubernetes/minikube/pull/15781)

Major Improvements:
* Ensure only one `minikube tunnel` instance runs at a time [#15834](https://github.com/kubernetes/minikube/pull/15834)
* Infer HyperKit HostIP as Gateway rather than hardcode to 192.168.64.1  [#15720](https://github.com/kubernetes/minikube/pull/15720)
* multi-node: Add support for volumes using CSI addon [#15829](https://github.com/kubernetes/minikube/pull/15829)

Minor Improvements:
* QEMU: Rename `user` network to `builtin` and update documentation [#15793](https://github.com/kubernetes/minikube/pull/15793)
* none driver: Look for cri-dockerd instead of hardcoding [#15784](https://github.com/kubernetes/minikube/pull/15784)
* Replace instances of `k8s.gcr.io` with `registry.k8s.io` [#16200](https://github.com/kubernetes/minikube/pull/16200)
* Handle CRI config of NetworkPlugin and PauseImage [#14703](https://github.com/kubernetes/minikube/pull/14703)
* Remove deprecated `container-runtime` flag from Kubernetes v1.24+ [#16124](https://github.com/kubernetes/minikube/pull/16124)
* none driver: Require crictl to be installed for Kubernetes v1.24+ [#16215](https://github.com/kubernetes/minikube/pull/16215)
* Add cri-dockerd logs to `minikube logs` output [#16149](https://github.com/kubernetes/minikube/pull/16149)
* Add ingress logs to `minikube logs` output [#15775](https://github.com/kubernetes/minikube/pull/15775)
* Add default cni logs to `minikbue logs` output [#15909](https://github.com/kubernetes/minikube/pull/15909)
* Add JSON output option to `miniikube service list` [#15831](https://github.com/kubernetes/minikube/pull/15831)
* Add kicbase download process to JSON output [#15685](https://github.com/kubernetes/minikube/pull/15685)
* Implement `--docs` for `minikube addons list -o json` [#15866](https://github.com/kubernetes/minikube/pull/15866)
* Implement `--skip-audit` flag and skip adding `profile` commands to audit log [#15872](https://github.com/kubernetes/minikube/pull/15872)
* Implement `--last-start-only` flag to `minikube logs` to only show last start logs  [#15770](https://github.com/kubernetes/minikube/pull/15770)

Bugs:
* Addon metallb: Fix failing to enable addon [#16056](https://github.com/kubernetes/minikube/pull/16056)
* Addon cloud-spanner: Fix failing to enable addon [#15743](https://github.com/kubernetes/minikube/pull/15743)
* Addon gcp-auth: Fix --refresh failing when existing cluster and minikube binary have differing image version [#15985](https://github.com/kubernetes/minikube/pull/15985)
* Fix numerous image related bugs when enabling addons [#15984](https://github.com/kubernetes/minikube/pull/15984)
* Fix some addons from erroring when trying to disable an already disabled addon [#16139](https://github.com/kubernetes/minikube/pull/16139)
* Fix panic if `docker version` returns exit code 0 with unexpected output [#15851](https://github.com/kubernetes/minikube/pull/15851)
* Fix `minikube service` not honoring `--wait` arg [#15735](https://github.com/kubernetes/minikube/pull/15735)
* Fix `minikube service` table format & hide unreachable URLs on Docker/Windows [#15911](https://github.com/kubernetes/minikube/pull/15911)
* Fix `minikube addons list` output showing incorrect status of default addons [#15762](https://github.com/kubernetes/minikube/pull/15762)
* Fix `minikube mount` printing an empty mount type [#15731](https://github.com/kubernetes/minikube/pull/15731)
* Fix bash completion for kubectl symlinked to minikube by not adding `--cluster` flag for the `kubectl __complete` subcommand [#15850](https://github.com/kubernetes/minikube/pull/15850)

Version Upgrades:
* Bump Kubernetes version default: v1.26.3 and latest: v1.27.0-rc.0 [#16181](https://github.com/kubernetes/minikube/pull/16181)
* Addon gcp-auth: Update ingress-nginx/kube-webhook-certgen image from v1.0 to v20230312-helm-chart-4.5.2-28-g66a760794 [#16199](https://github.com/kubernetes/minikube/pull/16199)
* Addon ingress: Update ingress-nginx/kube-webhook-certgen image from v20220916-gd32f8c343 to v20230312-helm-chart-4.5.2-28-g66a760794 [#16179](https://github.com/kubernetes/minikube/pull/16179)
* Addon ingress: Update ingress-nginx/controller image from v1.5.1 to v1.7.0 [#15882](https://github.com/kubernetes/minikube/pull/15882) [#16179](https://github.com/kubernetes/minikube/pull/16179)
* Addon cloud-spanner: Update cloud-spanner-emulator/emulator image from v1.5.0 to 1.5.2 [#15974](https://github.com/kubernetes/minikube/pull/15974) [#16142](https://github.com/kubernetes/minikube/pull/16142)
* Addon metrics-server: Update metrics-server/metrics-server image from v0.6.2 to v0.6.3 [#16136](https://github.com/kubernetes/minikube/pull/16136)
* Addon headlamp: Update headlamp-k8s/headlamp image from v0.14.1 to v0.16.0 [#15995](https://github.com/kubernetes/minikube/pull/15995) [#16065](https://github.com/kubernetes/minikube/pull/16065)
* Addon auto-pause: Update k8s-minikube/auto-pause-hook image from v0.0.3 to v0.0.4 [#16025](https://github.com/kubernetes/minikube/pull/16025)
* Addon gcp-auth: Update k8s-minikube/gcp-auth-webhook image from v0.0.13 to v0.0.14 [#16012](https://github.com/kubernetes/minikube/pull/16012)
* Kicbase/ISO: Update containerd from v1.6.15 to v1.7.0 [#15923](https://github.com/kubernetes/minikube/pull/15923) [#15973](https://github.com/kubernetes/minikube/pull/15973) [#16168](https://github.com/kubernetes/minikube/pull/16168)
* Kicbase/ISO: Update buildkit from v0.10.3 to v0.11.4 [#15728](https://github.com/kubernetes/minikube/pull/15728) [#16079](https://github.com/kubernetes/minikube/pull/16079)
* Kicbase/ISO: Update cri-dockerd from 0.3.0 to 0.3.1 [#15752](https://github.com/kubernetes/minikube/pull/15752)
* Kicbase: Update base image from ubuntu:focal-20221019 to ubuntu:focal-20230308 [#15768](https://github.com/kubernetes/minikube/pull/15768) [#15991](https://github.com/kubernetes/minikube/pull/15991) [#16068](https://github.com/kubernetes/minikube/pull/16068)
* ISO: Update runc from v1.1.4 to v1.1.5 [#16191](https://github.com/kubernetes/minikube/pull/16191)
* ISO: Update podman from v3.4.2 to v3.4.7 [#15565](https://github.com/kubernetes/minikube/pull/15565)
* CNI: Update kindnetd from v20221004-44d545d1 to v20230330-48f316cd [#15940](https://github.com/kubernetes/minikube/pull/15940) [#16207](https://github.com/kubernetes/minikube/pull/16207)

For a more detailed changelog, see [CHANGELOG.md](https://github.com/kubernetes/minikube/blob/master/CHANGELOG.md).

Thank you to our contributors for this release!

- Anders F Björklund
- Bart Van Bos
- Ben Krieger
- Denys Kondratenko
- Elizabeth Martin Campos
- Jeff MAURY
- Kundan Kumar
- Max Xu
- Maxime Brunet
- Medya Ghazizadeh
- Nick Mancari
- Om Saran
- Pablo Caderno
- Predrag Rogic
- Qasim Sarfraz
- S Santhosh Nagaraj
- Shubh Bapna
- Steven Powell
- Sudharsan Rangarajan
- Swastik Gour
- chncaption
- coffemakingtoaster
- joaquimrocha
- nickmancari
- shixiuguo
- sunyuxuan
- swastik959
- syxunion
- yolossn
- Товарищ программист

Thank you to our PR reviewers for this release!

- spowelljr (57 comments)
- medyagh (43 comments)
- neersighted (6 comments)
- shu-mutou (4 comments)
- afbjorklund (2 comments)
- akdean (1 comments)
- tstromberg (1 comments)

Thank you to our triage members for this release!

- afbjorklund (90 comments)
- spowelljr (25 comments)
- kundan2707 (20 comments)
- medyagh (9 comments)
- ComradeProgrammer (6 comments)

Check out our [contributions leaderboard](https://minikube.sigs.k8s.io/docs/contrib/leaderboard/v1.30.0/) for this release!

## Version 1.29.0 - 2023-01-26

Features:
* Bump QEMU driver priority from experimental to default [#15556](https://github.com/kubernetes/minikube/pull/15556)
* Ability to set static-ip for Docker driver [#15553](https://github.com/kubernetes/minikube/pull/15553)
* GCP-Auth Addon: automatically attach credentials to newly created namespaces [#15403](https://github.com/kubernetes/minikube/pull/15403)
* Allow forcing 1 CPU on Linux with docker and none driver [#15611](https://github.com/kubernetes/minikube/pull/15611) [#15610](https://github.com/kubernetes/minikube/pull/15610)

Major Improvements:
* Large improvements to cgroup detection and CNI and CRI configurations [#15463](https://github.com/kubernetes/minikube/pull/15463)
* Prevent redownloading kicbase when already downloaded [#15528](https://github.com/kubernetes/minikube/pull/15528)
* Warn when using an old ISO/Kicbase image [#15235](https://github.com/kubernetes/minikube/pull/15235)

Minor Improvements:
* Check brew install paths for socket_vmnet [#15701](https://github.com/kubernetes/minikube/pull/15701)
* Include gcp-auth logs in 'minikube logs' output [#15666](https://github.com/kubernetes/minikube/pull/15666)
* Use absolute path when calling crictl version [#15642](https://github.com/kubernetes/minikube/pull/15642)
* Add additional memory overhead for VirtualBox when `--memory=max` [#15317](https://github.com/kubernetes/minikube/pull/15317)
* Update Windows installer to create system-wide shortcut [#15405](https://github.com/kubernetes/minikube/pull/15405)
* Add `--subnet` validation [#15530](https://github.com/kubernetes/minikube/pull/15530)
* Warn users if using VirtualBox on macOS 13+ [#15624](https://github.com/kubernetes/minikube/pull/15624)
* Add groups check to SSH driver [#15513](https://github.com/kubernetes/minikube/pull/15513)
* Update references to deprecated beta.kubernetes.io [#15225](https://github.com/kubernetes/minikube/pull/15225)

Bugs:
* Fix possible race condition when enabling multiple addons [#15706](https://github.com/kubernetes/minikube/pull/15706)
* Fix cpus config field not supporting max value [#15479](https://github.com/kubernetes/minikube/pull/15479)
* Fix subnet checking failing if IPv6 network found [#15394](https://github.com/kubernetes/minikube/pull/15394)
* Fix Docker tunnel failing if too many SSH keys [#15560](https://github.com/kubernetes/minikube/pull/15560)
* Fix kubelet localStorageCapacityIsolation option [#15336](https://github.com/kubernetes/minikube/pull/15336)
* Fix setting snapshotter to unimplemented fuse-overlayfs [#15272](https://github.com/kubernetes/minikube/pull/15272)
* Remove progress bar for kic download with JSON output [#15482](https://github.com/kubernetes/minikube/pull/15482)

Version Upgrades:
* Bump default Kubernetes version from 1.25.3 to 1.26.1 [#15683](https://github.com/kubernetes/minikube/pull/15683)
* Addons: Update auto-pause from 0.0.2 to 0.0.3 [#15331](https://github.com/kubernetes/minikube/pull/15331)
* Addons: Update cloud-spanner from 1.4.6 to 1.5.0 [#15440](https://github.com/kubernetes/minikube/pull/15440) [#15667](https://github.com/kubernetes/minikube/pull/15667) [#15707](https://github.com/kubernetes/minikube/pull/15707)
* Addons: Update headlamp from 0.13.0 to 0.14.1 [#15401](https://github.com/kubernetes/minikube/pull/15401) [#15515](https://github.com/kubernetes/minikube/pull/15515)
* Addons: Update ingress from 1.2.1 to 1.5.1 [#15339](https://github.com/kubernetes/minikube/pull/15339)
* Addons: Update metrics-server from 0.6.1 to 0.6.2 [#15411](https://github.com/kubernetes/minikube/pull/15411)
* Addons: Update kubevirt from 1.17 to 1.24.7 [#15310](https://github.com/kubernetes/minikube/pull/15310)
* CNI: Update cilium from 1.9.9 to 1.12.3 [#15242](https://github.com/kubernetes/minikube/pull/15242)
* Kicbase: Update buildkit from 0.10.3 to v0.11.0 [#15630](https://github.com/kubernetes/minikube/pull/15630)
* Kicbase/ISO: Update containerd from 1.6.9 to 1.6.15 [#15541](https://github.com/kubernetes/minikube/pull/15541)
* Kicbase/ISO: Update cri-dockerd from 0.2.2 to 0.3.0 [#15541](https://github.com/kubernetes/minikube/pull/15541)
* Kicbase/ISO: Update docker from 20.10.20 to 20.10.23 [#15341](https://github.com/kubernetes/minikube/pull/15341) [#15541](https://github.com/kubernetes/minikube/pull/15541) [#15703](https://github.com/kubernetes/minikube/pull/15703)
* Update KVM-docker-machine amd64 base image from Ubuntu 16.04 to 20.04 [#15628](https://github.com/kubernetes/minikube/pull/15628)


For a more detailed changelog see [CHANGELOG.md](https://github.com/kubernetes/minikube/blob/master/CHANGELOG.md).

Thank you to our contributors for this release!

- Aarav Arora
- Akihiro Suda
- Anders F Björklund
- Andrew Stanton
- Carlos Santana
- Chris Kannon
- Eng Zer Jun
- Felipe Labbate
- Ian Stewart
- Jan Hutař
- Jeff MAURY
- Kaylen Dart
- Kush Mansingh
- Ludovic Maître
- Medya Ghazizadeh
- Olivier Lemasle
- Paco Xu
- Paul S. Schweigert
- Predrag Rogic
- Ronnel Santiago
- Sharif Elgamal
- Shubh Bapna
- Steven Powell
- Yuiko Mouri
- ckannon
- imjoseangel
- joaquimrocha
- jongwooo
- mardi2020
- shixiuguo
- Товарищ программист

Thank you to our PR reviewers for this release!

- spowelljr (61 comments)
- medyagh (41 comments)
- afbjorklund (7 comments)
- atoato88 (5 comments)
- t-inu (5 comments)
- mqasimsarfraz (4 comments)
- AkihiroSuda (2 comments)
- sharifelgamal (2 comments)
- profnandaa (1 comments)

Thank you to our triage members for this release!

- afbjorklund (98 comments)
- spowelljr (34 comments)
- medyagh (10 comments)
- kant777 (6 comments)
- kundan2707 (6 comments)

Check out our [contributions leaderboard](https://minikube.sigs.k8s.io/docs/contrib/leaderboard/v1.29.0/) for this release!

## Version 1.28.0 - 2022-11-04

**SECURITY WARNING:** Log4j CVEs were detected in an image the `efk` addon uses, if you don't use the `efk` addon no action is required. If you use the addon we recommend running `minikube addons disable efk` to terminate the vulnerable pods.
See [#15280](https://github.com/kubernetes/minikube/issues/15280) for more details.

Security:
* Prevent enabling `efk` addon due to containing Log4j CVE [#15281](https://github.com/kubernetes/minikube/pull/15281)

Features:
* Auto select network on QEMU [#15266](https://github.com/kubernetes/minikube/pull/15266)
* Implement mounting on QEMU with socket_vmnet [#15108](https://github.com/kubernetes/minikube/pull/15108)
* Added cloud-spanner emulator addon [#15160](https://github.com/kubernetes/minikube/pull/15160)
* Add `minikube license` command [#15158](https://github.com/kubernetes/minikube/pull/15158)

Minor Improvements:
* Allow port forwarding on Linux with Docker Desktop [#15126](https://github.com/kubernetes/minikube/pull/15126)
* Add back service to mount VirtualBox host directory into the guest. [#14784](https://github.com/kubernetes/minikube/pull/14784)
* ISO: Add FANOTIFY_ACCESS_PERMISSIONS to kernel configs [#15232](https://github.com/kubernetes/minikube/pull/15232)
* When enabling addon warn if addon has no associated Github username [#15081](https://github.com/kubernetes/minikube/pull/15081)

Bug Fixes:
* Fix detecting preload cache of size 0 as valid [#15256](https://github.com/kubernetes/minikube/pull/15256)
* Fix always writing to daemon by trimming `docker.io` from image name [#14956](https://github.com/kubernetes/minikube/pull/14956)
* Fix minikube tunnel repeated printout of status [#14933](https://github.com/kubernetes/minikube/pull/14933)

Version Upgrades:
* Upgrade Portainer addon to 2.15.1 & HTTPS access enabled [#15172](https://github.com/kubernetes/minikube/pull/15172)
* Upgrade Headlamp addon to 0.13.0 [#15186](https://github.com/kubernetes/minikube/pull/15186)
* ISO: Upgrade Docker from 20.10.18 to 20.10.20 [#15159](https://github.com/kubernetes/minikube/pull/15159)
* KIC: Upgrade base image from ubuntu:focal-20220826 to ubuntu:focal-20220922 [#15075](https://github.com/kubernetes/minikube/pull/15075)
* KCI: Upgrade base image from ubuntu:focal-20220922 to ubuntu:focal-20221019 [#15219](https://github.com/kubernetes/minikube/pull/15219)

For a more detailed changelog, including changes occurring in pre-release versions, see [CHANGELOG.md](https://github.com/kubernetes/minikube/blob/master/CHANGELOG.md).

Thank you to our contributors for this release!

- Chris Kannon
- Francis Laniel
- Jeff MAURY
- Jevon Tane
- Medya Ghazizadeh
- Nitin Agarwal
- Oldřich Jedlička
- Rahil Patel
- Steven Powell
- Tian
- Yue Yang
- joaquimrocha
- klaases
- shixiuguo

Thank you to our PR reviewers for this release!

- spowelljr (25 comments)
- medyagh (14 comments)

Thank you to our triage members for this release!

- RA489 (64 comments)
- klaases (39 comments)
- afbjorklund (23 comments)
- spowelljr (22 comments)
- medyagh (4 comments)

Check out our [contributions leaderboard](https://minikube.sigs.k8s.io/docs/contrib/leaderboard/v1.28.0/) for this release!

## Version 1.27.1 - 2022-10-07

Features (Experimental):
* QEMU Driver: Add support for dedicated network on macOS (socket_vmnet) [#14989](https://github.com/kubernetes/minikube/pull/14989)
* QEMU Driver: Add support minikube service and tunnel on macOS [#14989](https://github.com/kubernetes/minikube/pull/14989)

Minor Imprevements:
* Check if context is invalid during update-context command [#15032](https://github.com/kubernetes/minikube/pull/15032)
* Use SSH tunnel if user specifies bindAddress [#14951](https://github.com/kubernetes/minikube/pull/14951)
* Warn QEMU users if DNS issue detected [#15073](https://github.com/kubernetes/minikube/pull/15073)

Bug Fixes:
* Fix status command taking a long time on docker driver while paused [#15077](https://github.com/kubernetes/minikube/pull/15077)
* Fix not allowing passing only an exposed port to --ports [#15085](https://github.com/kubernetes/minikube/pull/15085)
* Fix `minikube dashboard` failing on macOS [#15037](https://github.com/kubernetes/minikube/pull/15037)
* Fix incorrect command in powershell command tip [#15012](https://github.com/kubernetes/minikube/pull/15012)

Version Upgrades:
* Bump Kubernetes version default: v1.25.2 and latest: v1.25.2 [#14995](https://github.com/kubernetes/minikube/pull/14995)
* Upgrade kubernetes dashboard from v2.6.0 to v2.7.0 [#15000](https://github.com/kubernetes/minikube/pull/15000)

For a more detailed changelog, including changes occurring in pre-release versions, see [CHANGELOG.md](https://github.com/kubernetes/minikube/blob/master/CHANGELOG.md).

Thank you to our contributors for this release!

- Anthony Nandaa
- Jeff MAURY
- Medya Ghazizadeh
- Rob Leland
- Steven Powell
- Yuiko Mouri
- cokia
- klaases
- ziyi-xie

Thank you to our PR reviewers for this release!

- eiffel-fl (9 comments)
- medyagh (6 comments)
- AkihiroSuda (2 comments)
- klaases (2 comments)
- t-inu (1 comments)

Thank you to our triage members for this release!

- klaases (31 comments)
- RA489 (30 comments)
- afbjorklund (17 comments)
- nikimanoledaki (7 comments)
- medyagh (3 comments)

Check out our [contributions leaderboard](https://minikube.sigs.k8s.io/docs/contrib/leaderboard/v1.27.1/) for this release!

## Version 1.27.0 - 2022-09-15

Kubernetes v1.25:
* Bump default Kubernetes version to v1.25.0 and resolve `/etc/resolv.conf` regression [#14848](https://github.com/kubernetes/minikube/pull/14848)
* Skip metallb PodSecurityPolicy object for kubernetes 1.25+ [#14903](https://github.com/kubernetes/minikube/pull/14903)
* The DefaultKubernetesRepo changed for 1.25.0 [#14768](https://github.com/kubernetes/minikube/pull/14768)

Minor Improvements:
* Add fscrypt kernel options [#14783](https://github.com/kubernetes/minikube/pull/14783)
* Output kubeadm logs [#14697](https://github.com/kubernetes/minikube/pull/14697)

Bug fixes:
* Fix QEMU delete errors [#14950](https://github.com/kubernetes/minikube/pull/14950)
* Fix containerd configuration issue with insecure registries [#14482](https://github.com/kubernetes/minikube/pull/14482)
* Fix registry when custom images provided [#14690](https://github.com/kubernetes/minikube/pull/14690)

Version Upgrades:
* ISO: Update Docker from 20.10.17 to 20.10.18 [#14935](https://github.com/kubernetes/minikube/pull/14935)
* Update kicbase base image to Ubuntu:focal-20220826 [#14904](https://github.com/kubernetes/minikube/pull/14904)
* Update registry addon image from 2.7.1 to 2.8.1 [#14886](https://github.com/kubernetes/minikube/pull/14886)
* Update gcp-auth-webhook addon from v0.0.10 to v0.0.11 [#14847](https://github.com/kubernetes/minikube/pull/14847)
* Update Headlamp addon image from v0.9.0 to v0.11.1 [#14802](https://github.com/kubernetes/minikube/pull/14802)

For a more detailed changelog, including changes occurring in pre-release versions, see [CHANGELOG.md](https://github.com/kubernetes/minikube/blob/master/CHANGELOG.md).

Thank you to our contributors for this release!

- Abirdcfly
- Alex
- Anders F Björklund
- Andrew Hamilton
- Jeff MAURY
- Jānis Bebrītis
- Marcel Lauhoff
- Medya Ghazizadeh
- Renato Costa
- Santhosh Nagaraj S
- Siddhant Khisty
- Steven Powell
- Yuiko Mouri
- klaases
- mtardy
- shaunmayo
- shixiuguo

Thank you to our PR reviewers for this release!

- spowelljr (23 comments)
- medyagh (6 comments)
- klaases (5 comments)
- vbezhenar (2 comments)
- nixpanic (1 comments)
- reylejano (1 comments)
- t-inu (1 comments)

Thank you to our triage members for this release!

- afbjorklund (76 comments)
- klaases (58 comments)
- RA489 (38 comments)
- spowelljr (16 comments)
- eiffel-fl (10 comments)

Check out our [contributions leaderboard](https://minikube.sigs.k8s.io/docs/contrib/leaderboard/v1.27.0/) for this release!

## Version 1.26.1 - 2022-08-02

Minor Improvements:
* Check for cri-dockerd & dockerd runtimes when using none-driver on Kubernetes 1.24+  [#14555](https://github.com/kubernetes/minikube/pull/14555)
* Add solution message for when `cri-docker` is missing [#14483](https://github.com/kubernetes/minikube/pull/14483)
* Limit number of audit entries [#14695](https://github.com/kubernetes/minikube/pull/14695)
* Optimize audit logging [#14596](https://github.com/kubernetes/minikube/pull/14596)
* Show the container runtime when running without kubernetes #13432  [#14200](https://github.com/kubernetes/minikube/pull/14200)
* Add warning when enabling thrid-party addons [#14499](https://github.com/kubernetes/minikube/pull/14499)

Bug fixes:
* Fix url index out of range error in service [#14658](https://github.com/kubernetes/minikube/pull/14658)
* Fix incorrect user and profile in audit logging [#14562](https://github.com/kubernetes/minikube/pull/14562)
* Fix overwriting err for OCI "minikube start" [#14506](https://github.com/kubernetes/minikube/pull/14506)
* Fix panic when environment variables are empty [#14415](https://github.com/kubernetes/minikube/pull/14415)

Version Upgrades:
* Bump Kubernetes version default: v1.24.3 and latest: v1.24.3 [#14606](https://github.com/kubernetes/minikube/pull/14606)
* ISO: Update Docker from 20.10.16 to 20.10.17 [#14534](https://github.com/kubernetes/minikube/pull/14534)
* ISO/Kicbase: Update cri-o from v1.22.3 to v1.24.1 [#14420](https://github.com/kubernetes/minikube/pull/14420)
* ISO: Update conmon from v2.0.24 to v2.1.2 [#14545](https://github.com/kubernetes/minikube/pull/14545)
* Update gcp-auth-webhook from v0.0.9 to v0.0.10 [#14670](https://github.com/kubernetes/minikube/pull/14670)
* ISO/Kicbase: Update base images [#14481](https://github.com/kubernetes/minikube/pull/14481)

For a more detailed changelog, including changes occurring in pre-release versions, see [CHANGELOG.md](https://github.com/kubernetes/minikube/blob/master/CHANGELOG.md).

Thank you to our contributors for this release!

- Akihiro Suda
- Akira Yoshiyama
- Bradley S
- Christoph "criztovyl" Schulz
- Gimb0
- HarshCasper
- Jeff MAURY
- Medya Ghazizadeh
- Niels de Vos
- Paul S. Schweigert
- Santhosh Nagaraj S
- Steven Powell
- Tobias Pfandzelter
- anoop142
- inifares23lab
- klaases
- peizhouyu
- zhouguowei
- 吴梓铭
- 李龙峰

Thank you to our PR reviewers for this release!

- spowelljr (50 comments)
- medyagh (9 comments)
- atoato88 (3 comments)
- klaases (2 comments)
- afbjorklund (1 comments)

Thank you to our triage members for this release!

- afbjorklund (75 comments)
- RA489 (56 comments)
- klaases (32 comments)
- spowelljr (27 comments)
- medyagh (13 comments)

Check out our [contributions leaderboard](https://minikube.sigs.k8s.io/docs/contrib/leaderboard/v1.26.1/) for this release!

## Version 1.26.0 - 2022-06-22

Features:
* Add `headlamp` addon [#14315](https://github.com/kubernetes/minikube/pull/14315)
* Add `InAccel FPGA Operator` addon [#12995](https://github.com/kubernetes/minikube/pull/12995)

QEMU:
* Only set highmem=off for darwin if qemu version is below 7.0 or memory is below 3GB [#14291](https://github.com/kubernetes/minikube/pull/14291)
* Define qemu as a qemu2 driver alias [#14284](https://github.com/kubernetes/minikube/pull/14284)
* Allow users to supply custom QEMU firmware path [#14283](https://github.com/kubernetes/minikube/pull/14283)

Minor Improvements:
* Add eBPF related kernel options [#14316](https://github.com/kubernetes/minikube/pull/14316)
* Add bind address flag for `minikube tunnel` [#14245](https://github.com/kubernetes/minikube/pull/14245)
* Add active column for `minikube profile list` [#14079](https://github.com/kubernetes/minikube/pull/14079)
* Add documentation URL to the addon list table [#14123](https://github.com/kubernetes/minikube/pull/14123)
* `minikube config defaults kubernetes-version` lists all currently supported Kubernetes versions [#13775](https://github.com/kubernetes/minikube/pull/13775)
* Support starting minikube with the Podman driver on NixOS systems [#12739](https://github.com/kubernetes/minikube/pull/12739)

Bug Fixes:
* Fix terminated commands not writing to audit log [#13307](https://github.com/kubernetes/minikube/pull/13307)
* Fix Podman port mapping publish on macOS [#14290](https://github.com/kubernetes/minikube/pull/14290)
* Fix `minikube delete` deleting networks from other profiles [#14279](https://github.com/kubernetes/minikube/pull/14279)

Version Upgrades:
* Bump Kubernetes version default: v1.24.1 and latest: v1.24.1 [#14197](https://github.com/kubernetes/minikube/pull/14197)
* ISO: Upgrade Docker from 20.10.14 to 20.10.16 [#14153](https://github.com/kubernetes/minikube/pull/14153)
* ISO: Upgrade kernel from 4.19.235 to 5.10.57 [#12707](https://github.com/kubernetes/minikube/pull/12707)
* Upgrade Dashboard addon from v2.5.1 to v2.6.0 & MetricsScraper from v1.0.7 to v1.0.8 [#14269](https://github.com/kubernetes/minikube/pull/14269)
* Upgrade gcp-auth-webhook from v0.0.8 to v0.0.9 [#14372](https://github.com/kubernetes/minikube/pull/14372)
* Upgrade nginx image from v1.2.0 to v1.2.1 [#14317](https://github.com/kubernetes/minikube/pull/14317)

**Important Changes in Pre-Release Versions**
Features:
* Add configure option to registry-aliases addon [#13912](https://github.com/kubernetes/minikube/pull/13912)
* Add support for building aarch64 ISO [#13762](https://github.com/kubernetes/minikube/pull/13762)
* Support rootless Podman driver (Usage: `minikube config set rootless true`) [#13829](https://github.com/kubernetes/minikube/pull/13829)

QEMU:
* Add support for the QEMU driver [#13639](https://github.com/kubernetes/minikube/pull/13639)
* Fix qemu firmware path locations [#14182](https://github.com/kubernetes/minikube/pull/14182)
* Re-establish apiserver tunnel on restart  [#14183](https://github.com/kubernetes/minikube/pull/14183)

For a more detailed changelog, including changes occurring in pre-release versions, see [CHANGELOG.md](https://github.com/kubernetes/minikube/blob/master/CHANGELOG.md).

Thank you to our contributors for this release!

- Alex Andrews
- Anders F Björklund
- Elias Koromilas
- Francis Laniel
- Giildo
- Harsh Vardhan
- Jack Zhang
- Jeff MAURY
- Kevin Grigorenko
- Kian-Meng Ang
- Leonardo Grasso
- Medya Ghazizadeh
- Nikhil Sharma
- Nils Fahldieck
- Pablo Caderno
- Peter Becich
- Predrag Rogic
- Santhosh Nagaraj S
- Sharif Elgamal
- Steven Powell
- Toshiaki Inukai
- klaases
- lakshkeswani
- layakdev
- lilongfeng
- simonren-tes
- ziyi-xie
- 李龙峰

Thank you to our PR reviewers for this release!

- spowelljr (76 comments)
- sharifelgamal (11 comments)
- medyagh (8 comments)
- afbjorklund (6 comments)
- kakkoyun (2 comments)
- knrt10 (2 comments)
- mprimeaux (2 comments)
- shu-mutou (2 comments)
- javierhonduco (1 comments)
- nburlett (1 comments)

Thank you to our triage members for this release!

- spowelljr (39 comments)
- RA489 (30 comments)
- sharifelgamal (27 comments)
- afbjorklund (14 comments)
- klaases (14 comments)

Check out our [contributions leaderboard](https://minikube.sigs.k8s.io/docs/contrib/leaderboard/v1.26.0/) for this release!

## Version 1.26.0-beta.1 - 2022-05-17

QEMU driver enhancements:
* fix qemu firmware path locations [#14182](https://github.com/kubernetes/minikube/pull/14182)
* re-establish apiserver tunnel on restart  [#14183](https://github.com/kubernetes/minikube/pull/14183)

Features:
* Add configure option to registry-aliases addon [#13912](https://github.com/kubernetes/minikube/pull/13912)

For a more detailed changelog, including changes occurring in pre-release versions, see [CHANGELOG.md](https://github.com/kubernetes/minikube/blob/master/CHANGELOG.md).

Thank you to our contributors for this release!

- Jack Zhang
- Pablo Caderno
- Sharif Elgamal
- Steven Powell
- Yuki Okushi
- loftkun

Thank you to our PR reviewers for this release!

- spowelljr (20 comments)
- afbjorklund (1 comments)
- sharifelgamal (1 comments)

Thank you to our triage members for this release!

- afbjorklund (4 comments)
- spowelljr (4 comments)
- Al4DIN (1 comments)
- Gimb0 (1 comments)
- Neandril (1 comments)

## Version 1.26.0-beta.0 - 2022-05-13

Featues:
* Add support for the QEMU driver [#13639](https://github.com/kubernetes/minikube/pull/13639)
* Add support for building aarch64 ISO [#13762](https://github.com/kubernetes/minikube/pull/13762)
* Support rootless Podman driver (Usage: `minikube config set rootless true`) [#13829](https://github.com/kubernetes/minikube/pull/13829)

Minor Improvements:
* Add JSON output to `minikube delete` [#13979](https://github.com/kubernetes/minikube/pull/13979)
* Add `--audit` flag to `minikube logs` command [#13991](https://github.com/kubernetes/minikube/pull/13991)
* Add `--disable-metrics` flag [#13802](https://github.com/kubernetes/minikube/pull/13802)
* Get latest valid tag for each image during caching [#14006](https://github.com/kubernetes/minikube/pull/14006)
* Remove docker requirement for none driver [#13885](https://github.com/kubernetes/minikube/pull/13885)
* Add 'subnet' flag for docker/podman driver [#13730](https://github.com/kubernetes/minikube/pull/13730)
* Don't write logs that contain environment variables [#13877](https://github.com/kubernetes/minikube/pull/13877)
* Implemented minimum and recommended Docker versions [#13842](https://github.com/kubernetes/minikube/pull/13842)

Bug Fixes:
* Fix "Your cgroup does not allow setting memory" [#14115](https://github.com/kubernetes/minikube/pull/14115)
* Fix nvidia-gpu with kvm-driver [#13972](https://github.com/kubernetes/minikube/pull/13972)
* Fix `minikube delete` for Podman v4 [#13881](https://github.com/kubernetes/minikube/pull/13881)
* Fix pre command flags [#13995](https://github.com/kubernetes/minikube/pull/13995)
* Fix logging when JSON output selected [#13955](https://github.com/kubernetes/minikube/pull/13955)
* Fix port validation error on specifying tcp/udp or range of ports. [#13812](https://github.com/kubernetes/minikube/pull/13812)
* Fix not downloading kic for offline mode [#13910](https://github.com/kubernetes/minikube/pull/13910)
* Fix trying to pause multiple containers with runc [#13783](https://github.com/kubernetes/minikube/pull/13783)
* Fix `minikube service` docker/port-forward issues [#13756](https://github.com/kubernetes/minikube/pull/13756)

Version Upgrades:
* Upgrade Kubernetes default: v1.23.6 and latest: v1.23.6 [#14144](https://github.com/kubernetes/minikube/pull/14144)
* ISO/KIC: Upgrade buildkit from 0.9.0 to 0.10.3 [#13791](https://github.com/kubernetes/minikube/pull/13791)
* ISO: Upgrade Docker from 20.10.12 to 20.10.14 [#13860](https://github.com/kubernetes/minikube/pull/13860)
* ISO: Upgrade crio from 1.22.1 to 1.22.3 [#13800](https://github.com/kubernetes/minikube/pull/13800)
* ISO: Upgrade buildroot from 2021.02.4 to 2021.02.12 [#13814](https://github.com/kubernetes/minikube/pull/13814)
* Upgrade nginx image from 1.1.1 to 1.2.0 [#14028](https://github.com/kubernetes/minikube/pull/14028)
* ISO: Upgrade falco-module from 0.24.0 to 0.31.1 [#13659](https://github.com/kubernetes/minikube/pull/13659)
* Upgrade kubernetes dashboard from 2.3.1 to 2.5.1 [#13741](https://github.com/kubernetes/minikube/pull/13741)
* KIC: Upgrade kicbase base image from 20210401 to 20220316 [#13815](https://github.com/kubernetes/minikube/pull/13815)
* ISO: Upgrade Podman from 2.2.1 to 3.4.2 [#13126](https://github.com/kubernetes/minikube/pull/13126)
* ISO: Add packaging for crun [#11679](https://github.com/kubernetes/minikube/pull/11679)

For a more detailed changelog, including changes occurring in pre-release versions, see [CHANGELOG.md](https://github.com/kubernetes/minikube/blob/master/CHANGELOG.md).

Thank you to our contributors for this release!

- Akihiro Suda
- Akira Yoshiyama
- Anders F Björklund
- Ashwin Somasundara
- Carlos Eduardo Arango Gutierrez
- Daniel Petri
- F1ko
- Filip Nikolic
- Ileriayo Adebiyi
- Jeff MAURY
- Jin Zhang
- Medya Ghazizadeh
- Nikhil2001
- Pablo Caderno
- Piotr Resztak
- Predrag Rogic
- Sean Wei
- Sharif Elgamal
- Steven Powell
- Tomohito YABU
- Toshiaki Inukai
- betaboon
- ckannon
- edwinwalela
- klaases
- naveensrinivasan
- staticdev
- ziyi-xie

Thank you to our PR reviewers for this release!

- spowelljr (55 comments)
- medyagh (39 comments)
- afbjorklund (14 comments)
- klaases (14 comments)
- jesperpedersen (9 comments)
- sharifelgamal (6 comments)
- atoato88 (3 comments)
- jepio (3 comments)
- mprimeaux (2 comments)
- shu-mutou (2 comments)
- t-inu (2 comments)
- AkihiroSuda (1 comments)

Thank you to our triage members for this release!

- afbjorklund (52 comments)
- klaases (39 comments)
- RA489 (28 comments)
- spowelljr (24 comments)
- zhan9san (24 comments)

## Version 1.25.2 - 2022-02-23

Features:
* [Addon] Kong Ingress Controller [#13326](https://github.com/kubernetes/minikube/pull/13326)
* add arch to binary and image cache paths [#13539](https://github.com/kubernetes/minikube/pull/13539)
* Adds 'minikube service --all' feature to allow forwarding all services in a namespace [#13367](https://github.com/kubernetes/minikube/pull/13367)
* Make the default container runtime dynamic [#13251](https://github.com/kubernetes/minikube/pull/13251)
* Add `--disable-optimizations` flag [#13340](https://github.com/kubernetes/minikube/pull/13340)

Bug Fixes:
* Fix losing cluster on restart [#13506](https://github.com/kubernetes/minikube/pull/13506)
* Using Get-CmiInstance to detect Hyper-V availability [#13596](https://github.com/kubernetes/minikube/pull/13596)
* Fixes validation on image repository URL when it contains port but no scheme [#13053](https://github.com/kubernetes/minikube/pull/13053)
* Fixed SIGSEGV in kubectl when k8s not running [#13631](https://github.com/kubernetes/minikube/pull/13631)
* Fix hard coded docker driver in minikube service command [#13514](https://github.com/kubernetes/minikube/pull/13514)
* Fix hard coded docker driver in minikube tunnel command [#13444](https://github.com/kubernetes/minikube/pull/13444)
* Fix IstioOperator CustomResourceDefinition for istio-provisioner addon [#13024](https://github.com/kubernetes/minikube/pull/13024)
* fix ingress (also for multinode clusters) [#13439](https://github.com/kubernetes/minikube/pull/13439)
* Add exit message for too new Kubernetes version [#13354](https://github.com/kubernetes/minikube/pull/13354)
* drivers/kvm: Use ARP for retrieving interface ip addresses [#13482](https://github.com/kubernetes/minikube/pull/13482)
* kubeadm: allow skipping kube-proxy addon on restart [#13538](https://github.com/kubernetes/minikube/pull/13538)
* configure container runtimes for clusters without Kubernetes too [#13442](https://github.com/kubernetes/minikube/pull/13442)

Version Upgrades:
* KIC: Upgrade cri-dockerd [#13302](https://github.com/kubernetes/minikube/pull/13302)
* upgrade libvirt to "8th gen" [#13440](https://github.com/kubernetes/minikube/pull/13440)
* Upgrade cri-dockerd to fix the socket path [#13563](https://github.com/kubernetes/minikube/pull/13563)
* ISO: Add packaging for cri-dockerd [#13191](https://github.com/kubernetes/minikube/pull/13191)


For a more detailed changelog, including changes occurring in pre-release versions, see [CHANGELOG.md](https://github.com/kubernetes/minikube/blob/master/CHANGELOG.md).

Thank you to our contributors for this release!

- Akira Yoshiyama
- Anders F Björklund
- Anoop C S
- Balakumaran GR
- Carlos Santana
- Chris Tomkins
- Daniel Helfand
- Jeff MAURY
- Jin Zhang
- Medya Ghazizadeh
- Nikhil Sharma
- Olivier Bouchoms
- Piotr Resztak
- Predrag Rogic
- Sahan Serasinghe
- Sharif Elgamal
- Steven Powell
- Tiago Alves
- Todd MacIntyre
- Viktor Gamov
- ckannon
- klaases
- nishipy
- pedrothome1

Thank you to our PR reviewers for this release!

- medyagh (23 comments)
- afbjorklund (10 comments)
- sharifelgamal (10 comments)
- spowelljr (10 comments)
- t-inu (6 comments)
- klaases (4 comments)
- s-kawamura-w664 (3 comments)
- atoato88 (2 comments)
- alexbaeza (1 comments)
- csantanapr (1 comments)
- totollygeek (1 comments)

Thank you to our triage members for this release!

- RA489 (31 comments)
- afbjorklund (27 comments)
- spowelljr (25 comments)
- klaases (13 comments)
- sharifelgamal (12 comments)

Check out our [contributions leaderboard](https://minikube.sigs.k8s.io/docs/contrib/leaderboard/v1.25.2/) for this release!

## Version 1.25.1 - 2022-01-20

* Resolved regression breaking `minikube start` with hyperkit driver [#13418](https://github.com/kubernetes/minikube/pull/13418)

For a more detailed changelog, including changes occurring in pre-release versions, see [CHANGELOG.md](https://github.com/kubernetes/minikube/blob/master/CHANGELOG.md).

Thank you to our contributors for this release!

- Medya Ghazizadeh
- Sharif Elgamal
- Steven Powell

Thank you to our triage members for this release!

- klaases (13 comments)
- RA489 (12 comments)
- spowelljr (7 comments)
- afbjorklund (6 comments)
- sharifelgamal (2 comments)

## Version 1.25.0 - 2022-01-18

Features:
* New flag "--binary-mirror" to override mirror URL downloading (kubectl, kubelet, & kubeadm) [#12804](https://github.com/kubernetes/minikube/pull/12804)
* Add format flag to the `image ls` command [#12996](https://github.com/kubernetes/minikube/pull/12996)
* Add all mount flags to start command [#12930](https://github.com/kubernetes/minikube/pull/12930)
* Auto set config to support btrfs storage driver [#12990](https://github.com/kubernetes/minikube/pull/12990)
* Support CRI-O runtime with Rootless Docker driver (`--driver=docker --container-runtime=cri-o`) [#12900](https://github.com/kubernetes/minikube/pull/12900)
* Allow custom cert for ingress to be overwritten [#12897](https://github.com/kubernetes/minikube/pull/12897)
* Allow ppc64le & armv7 with Docker driver [#13124](https://github.com/kubernetes/minikube/pull/13124)

Minor Improvements:
* Support DOCKER_HOST not being numeric IP [#13300](https://github.com/kubernetes/minikube/pull/13300)
* Support mounting with the --no-kubernetes flag [#13144](https://github.com/kubernetes/minikube/pull/13144)
* Support changing apiserver-ips when restarting minikube [#12692](https://github.com/kubernetes/minikube/pull/12692)

Bug fixes:
* Fix ingress for k8s v1.19 [#13173](https://github.com/kubernetes/minikube/pull/13173)
* Fix mounting with VMware #12426 [#13000](https://github.com/kubernetes/minikube/pull/13000)
* Fix `Bad file descriptor` on mount [#13013](https://github.com/kubernetes/minikube/pull/13013)
* Fix `docker-env` with new PowerShell versions [#12870](https://github.com/kubernetes/minikube/pull/12870)

Version Upgrades:
* Upgrade Docker, from v20.10.8 to v20.10.11
* Upgrade containerd, from v1.4.9 to v1.4.12
* Upgrade cri-o from v1.22.0 to v1.22.1 [#13059](https://github.com/kubernetes/minikube/pull/13059)
* Update gcp-auth-webhook image to v0.0.8 [#13185](https://github.com/kubernetes/minikube/pull/13185)

Deprecation:
* mount: Remove `--mode` flag [#13162](https://github.com/kubernetes/minikube/pull/13162)

For a more detailed changelog, including changes occurring in pre-release versions, see [CHANGELOG.md](https://github.com/kubernetes/minikube/blob/master/CHANGELOG.md).

Thank you to our contributors for this release!

- Akihiro Suda
- Akira Yoshiyama
- Anders F Björklund
- Ashwin901
- Carl Chesser
- Daehyeok Mun
- Davanum Srinivas
- Dimitris Aragiorgis
- Emilano Vazquez
- Eugene Kalinin
- Frank Schwichtenberg
- James Yin
- Jan Klippel
- Jeff MAURY
- Joey Klaas
- Marcus Puckett
- Medya Ghazizadeh
- Nikhil Sharma
- Nikolay Nikolaev
- Oleksii Prudkyi
- Pablo Caderno
- Piotr Resztak
- Predrag Rogic
- Rahil Patel
- Sergio Galvan
- Sharif Elgamal
- Steven Powell
- Tian Yang
- Toshiaki Inukai
- Vishal Jain
- Zvi Cahana
- gamba47
- rahil-p
- srikrishnabh93@gmail.com

Thank you to our PR reviewers for this release!

- spowelljr (65 comments)
- medyagh (64 comments)
- t-inu (46 comments)
- atoato88 (39 comments)
- sharifelgamal (39 comments)
- klaases (17 comments)
- afbjorklund (8 comments)
- s-kawamura-w664 (8 comments)
- yosshy (6 comments)
- neolit123 (3 comments)
- AkihiroSuda (1 comments)
- dims (1 comments)
- dobegor (1 comments)
- dytyniuk (1 comments)
- inductor (1 comments)
- rmohr (1 comments)

Thank you to our triage members for this release!

- spowelljr (48 comments)
- afbjorklund (44 comments)
- RA489 (37 comments)
- medyagh (33 comments)
- sharifelgamal (25 comments)

## Version 1.24.0 - 2021-11-04

Features:
* Add --no-kubernetes flag  to start minikube without kubernetes [#12848](https://github.com/kubernetes/minikube/pull/12848)
* `minikube addons list` shows addons if cluster does not exist [#12837](https://github.com/kubernetes/minikube/pull/12837)

Bug fixes:
* virtualbox: change default `host-only-cidr` [#12811](https://github.com/kubernetes/minikube/pull/12811)
* fix zsh completion [#12841](https://github.com/kubernetes/minikube/pull/12841)
* Fix starting on Windows with VMware driver on non `C:` drive [#12819](https://github.com/kubernetes/minikube/pull/12819)

For a more detailed changelog, including changes occurring in pre-release versions, see [CHANGELOG.md](https://github.com/kubernetes/minikube/blob/master/CHANGELOG.md).

Thank you to our contributors for this release!

- Akira Yoshiyama
- Keyhoh
- Medya Ghazizadeh
- Nicolas Busseneau
- Sharif Elgamal
- Steven Powell
- Toshiaki Inukai

Thank you to our PR reviewers for this release!

- spowelljr (11 comments)
- sharifelgamal (10 comments)
- afbjorklund (6 comments)
- atoato88 (5 comments)
- medyagh (3 comments)
- yosshy (1 comments)

Thank you to our triage members for this release!

- sharifelgamal (13 comments)
- afbjorklund (9 comments)
- spowelljr (6 comments)
- medyagh (3 comments)
- Sarathgiggso (2 comments)

## Version 1.24.0-beta.0 - 2021-10-28

Features:
* Allow running podman as experimental driver in Windows & macOS [#12579](https://github.com/kubernetes/minikube/pull/12579)
* Add Aliyun (China) mirror for preload images and K8s release binaries [#12578](https://github.com/kubernetes/minikube/pull/12578)

Minor Improvements:
* certs: Renew minikube certs if expired [#12534](https://github.com/kubernetes/minikube/pull/12534)
* mount: Persist mount settings after stop start [#12719](https://github.com/kubernetes/minikube/pull/12719)
* cri-o: Implement --force-systemd into cri-o [#12553](https://github.com/kubernetes/minikube/pull/12553)
* tunnel: Use new bridge interface name on OSX Monterey [#12799](https://github.com/kubernetes/minikube/pull/12799)
* Added port validation [#12233](https://github.com/kubernetes/minikube/pull/12233)
* buildkit: Start the daemon on demand (socket-activated) [#12081](https://github.com/kubernetes/minikube/pull/12081)

Bug Fixes:
* ingress: Restore ingress & ingress-dns backwards compatibility for k8s < v1.19 [#12794](https://github.com/kubernetes/minikube/pull/12794)
* gcp-auth: Fix disabling addon [#12779](https://github.com/kubernetes/minikube/pull/12779)
* podman: Fix network inspect index check [#12756](https://github.com/kubernetes/minikube/pull/12756)
* cilium: Fix Ipv4 cidr [#12587](https://github.com/kubernetes/minikube/pull/12587)
* mount: Fix mounting on non-default profile [#12711](https://github.com/kubernetes/minikube/pull/12711)
* podman: Match the lower case of the podman error message [#12685](https://github.com/kubernetes/minikube/pull/12685)
* ssh: Fix using tilde in ssh-key path [#12672](https://github.com/kubernetes/minikube/pull/12672)
* podman: Fix network not getting deleted [#12627](https://github.com/kubernetes/minikube/pull/12627)
* zsh: Fix completion [#12420](https://github.com/kubernetes/minikube/pull/12420)
* windows wsl2: Fix invoking kubeadm failing when spaces in PATH for none driver [#12617](https://github.com/kubernetes/minikube/pull/12617)
* image build: Only build on control plane by default [#12149](https://github.com/kubernetes/minikube/pull/12149)
* mount: Fix `minikube stop` on Windows VMs taking 9 minutes when mounted [#12716](https://github.com/kubernetes/minikube/pull/12716)

Version Upgrades:
* ingres controller: Update to v1/1.0.4 and v1beta1/0.49.3 [#12702](https://github.com/kubernetes/minikube/pull/12702)
* minikube-ingress-dns: Update image to 0.0.2 [#12730](https://github.com/kubernetes/minikube/pull/12730)
* helm-tiller: Update image to v2.17.0 [#12641](https://github.com/kubernetes/minikube/pull/12641)

For a more detailed changelog, including changes occurring in pre-release versions, see [CHANGELOG.md](https://github.com/kubernetes/minikube/blob/master/CHANGELOG.md).

Thank you to our contributors for this release!

- Akira Yoshiyama
- Alexandre Garnier
- Anders F Björklund
- Aniruddha Amit Dutta
- Avinash Upadhyaya
- Cameron Brunner
- Carlos Santana
- Claudiu Belu
- Gio Gutierrez
- Jeff MAURY
- KallyDev
- Keyhoh
- Kumar Shivendu
- Li Yi
- Marc Velasco
- Marcus Watkins
- Medya Ghazizadeh
- Michael Cade
- Pablo Caderno
- Peixuan Ding
- Piotr Resztak
- Predrag Rogic
- RA489
- Sharif Elgamal
- Steven Powell
- Taylor Steil
- Wei Luo
- phbits
- yxxhero

Thank you to our PR reviewers for this release!

- spowelljr (27 comments)
- medyagh (22 comments)
- t-inu (20 comments)
- sharifelgamal (9 comments)
- atoato88 (6 comments)
- rikatz (5 comments)
- YuikoTakada (1 comments)
- tstromberg (1 comments)

Thank you to our triage members for this release!

- spowelljr (37 comments)
- afbjorklund (34 comments)
- RA489 (30 comments)
- medyagh (29 comments)
- sharifelgamal (29 comments)

## Version 1.23.2 - 2021-09-21

Fix crio regression:
* Roll back default crio cgroup to systemd [#12533](https://github.com/kubernetes/minikube/pull/12533)
* Fix template typo [#12532](https://github.com/kubernetes/minikube/pull/12532)

For a more detailed changelog, including changes occurring in pre-release versions, see [CHANGELOG.md](https://github.com/kubernetes/minikube/blob/master/CHANGELOG.md).

Thank you to our contributors for this release!

- Jeff MAURY
- Lakshya Gupta
- Medya Ghazizadeh
- Sharif Elgamal
- Steven Powell

Thank you to our PR reviewers for this release!

- medyagh (2 comments)
- sharifelgamal (1 comments)

Thank you to our triage members for this release!

- afbjorklund (12 comments)
- yxxhero (10 comments)
- medyagh (7 comments)
- spowelljr (4 comments)
- dilyanpalauzov (2 comments)


## Version 1.23.1 - 2021-09-17

Minor Improvements:
* Add crun version to `minikube version --components` [#12381](https://github.com/kubernetes/minikube/pull/12381)

Bug Fixes:
* ingress addon: fix regression from v1.23.0 [#12443](https://github.com/kubernetes/minikube/pull/12443)
* ingress addon: fix role resource's referenced configmap [#12446](https://github.com/kubernetes/minikube/pull/12446)
* ingress-dns addon: fix regression from v1.23.0 [#12476](https://github.com/kubernetes/minikube/pull/12476)
* gcp-auth addon: delete image pull secrets on addon disable [#12473](https://github.com/kubernetes/minikube/pull/12473)
* gcp-auth addon: create pull secret even if creds JSON is nil [#12461](https://github.com/kubernetes/minikube/pull/12461)
* gcp-auth addon: fix refreshing pull secret [#12497](https://github.com/kubernetes/minikube/pull/12497)
* metallb addon: ask user for config values even if already set [#12437](https://github.com/kubernetes/minikube/pull/12437)
* ambassador addon: warn on enable that addon no longer works [#12474](https://github.com/kubernetes/minikube/pull/12474)
* dashboard addon: fix sha for metrics-scraper [#12496](https://github.com/kubernetes/minikube/pull/12496)
* windows installer: remove quotes from incorrect fields [#12430](https://github.com/kubernetes/minikube/pull/12430)
* strip namespace from images from aliyun registry [#11785](https://github.com/kubernetes/minikube/pull/11785)

Version Upgrades:
* Bump cri-o from v1.20.0 to 1.22.0 [#12425](https://github.com/kubernetes/minikube/pull/12425)
* Bump dashboard from v2.1.0 to v2.3.1 and metrics-scraper from v1.0.4 to v1.0.7 [#12475](https://github.com/kubernetes/minikube/pull/12475)

For a more detailed changelog, including changes occurring in pre-release versions, see [CHANGELOG.md](https://github.com/kubernetes/minikube/blob/master/CHANGELOG.md).

Thank you to our contributors for this release!

- Brian Li
- Brian de Alwis
- Hiroya Onoe
- Jayesh Srivastava
- Jeff MAURY
- Joel Jeremy Marquez
- Leif Ringstad
- Medya Ghazizadeh
- Sharif Elgamal
- Steven Powell
- Toshiaki Inukai

Thank you to our PR reviewers for this release!

- medyagh (9 comments)
- spowelljr (2 comments)
- afbjorklund (1 comments)

Thank you to our triage members for this release!

- spowelljr (17 comments)
- afbjorklund (16 comments)
- sharifelgamal (16 comments)
- RA489 (15 comments)
- medyagh (14 comments)

## Version 1.23.0 - 2021-09-03

Features:
* Support Rootless Docker [#12359](https://github.com/kubernetes/minikube/pull/12359)
* Add support for tcsh in docker-env subcommand [#12332](https://github.com/kubernetes/minikube/pull/12332)
* Support Ingress on MacOS, driver docker [#12089](https://github.com/kubernetes/minikube/pull/12089)
* Add support for linux/s390x on docker/podman drivers [#12079](https://github.com/kubernetes/minikube/pull/12079)
* Add configurable port for minikube mount [#11979](https://github.com/kubernetes/minikube/pull/11979)
* Add method for customized box output [#11709](https://github.com/kubernetes/minikube/pull/11709)
* Add addon support for portainer [#11933](https://github.com/kubernetes/minikube/pull/11933)
* minikube start --image-repository will now accept URLs with port [#11585](https://github.com/kubernetes/minikube/pull/11585)
* Add ability to create extra disks for hyperkit driver [#11483](https://github.com/kubernetes/minikube/pull/11483)
* Add ability to create extra disks for kvm2 driver [#12351](https://github.com/kubernetes/minikube/pull/12351)

minikube image:
* Add `minikube image` commands for pull and tag and push [#12326](https://github.com/kubernetes/minikube/pull/12326)
* new `image save` command [#12162](https://github.com/kubernetes/minikube/pull/12162)
* Auto start buildkit daemon on `image build` for containerd [#12076](https://github.com/kubernetes/minikube/pull/12076)

Bug fixes:
* Select WSL VM IP when performing mounting [#12319](https://github.com/kubernetes/minikube/pull/12319)
* Fix minikube restart on Cloud Shell [#12237](https://github.com/kubernetes/minikube/pull/12237)
* pause each container separately [#12318](https://github.com/kubernetes/minikube/pull/12318)
* Add output parameter to the docker-env none shell [#12263](https://github.com/kubernetes/minikube/pull/12263)
* Clean up ssh tunnels during exit. [#11745](https://github.com/kubernetes/minikube/pull/11745)
* Fix loading an image from tar failing on existing delete [#12143](https://github.com/kubernetes/minikube/pull/12143)
* configure gcp-auth addon pull secret to work with all GCR and AR mirrors [#12106](https://github.com/kubernetes/minikube/pull/12106)
* Fix the error output of minikube version --components command [#12085](https://github.com/kubernetes/minikube/pull/12085)
* Added restart command after setting crio options [#11968](https://github.com/kubernetes/minikube/pull/11968)
* Don't set conntrack parameters in kube-proxy [#11957](https://github.com/kubernetes/minikube/pull/11957)
* Fix kvm2 driver arm64 deb package [#11937](https://github.com/kubernetes/minikube/pull/11937)
* Allow to set the dashboard proxyfied port [#11553](https://github.com/kubernetes/minikube/pull/11553)

Version Upgrades:
* bump golang version to 1.17 [#12378](https://github.com/kubernetes/minikube/pull/12378)
* Bump default Kubernetes version to v1.22.1 and update addons to with new API (ingress, gcpauth, olm and cilium) [#12325](https://github.com/kubernetes/minikube/pull/12325)
* Add kubeadm image versions for kubernetes 1.22 [#12331](https://github.com/kubernetes/minikube/pull/12331)
* bump calico to v3.20 and move away from v1beta apis [#12230](https://github.com/kubernetes/minikube/pull/12230)
* Upgrade Buildroot to 2021.02 LTS with Linux 4.19 [#12268](https://github.com/kubernetes/minikube/pull/12268)
* Upgrade buildkit from 0.8.2 to 0.9.0 [#12032](https://github.com/kubernetes/minikube/pull/12032)
* ISO: Upgrade Docker, from 20.10.6 to 20.10.8 [#12122](https://github.com/kubernetes/minikube/pull/12122)
* ISO: Upgrade crictl (from cri-tools) to v1.21.0 [#12129](https://github.com/kubernetes/minikube/pull/12129)

Thank you to our contributors for this release!

- Akihiro Suda
- Alexandre Garnier
- Anders F Björklund
- Andriy Dzikh
- Blaine Gardner
- Devdutt Shenoi
- Ilya Zuyev
- Jack Zhang
- Jeff MAURY
- Joel Klint
- Julien Breux
- Leopold Schabel
- Matt Dainty
- Medya Ghazizadeh
- Pablo Caderno
- Parthvi Vala
- Peixuan Ding
- Predrag Rogic
- Raghavendra Talur
- Rajwinder Mahal
- Sharif Elgamal
- Steven Powell
- Tejal Desai
- Vishal Jain
- Zhang Shihe
- amit dixit
- balasu
- dmpe
- jayonlau
- m-aciek
- rajdevworks
- なつき

Thank you to our PR reviewers for this release!

- medyagh (68 comments)
- sharifelgamal (26 comments)
- afbjorklund (22 comments)
- spowelljr (15 comments)
- andriyDev (7 comments)
- mikebrow (7 comments)
- iliadmitriev (2 comments)
- ilya-zuyev (2 comments)
- azhao155 (1 comments)
- briandealwis (1 comments)
- ncresswell (1 comments)
- shahiddev (1 comments)

Thank you to our triage members for this release!

- afbjorklund (47 comments)
- RA489 (36 comments)
- sharifelgamal (32 comments)
- spowelljr (28 comments)
- medyagh (20 comments)

Check out our [contributions leaderboard](https://minikube.sigs.k8s.io/docs/contrib/leaderboard/v1.23.0/) for this release!

## Version 1.22.0 - 2021-07-07

Features:
* `minikube version`: add `--components` flag to list all included software [#11843](https://github.com/kubernetes/minikube/pull/11843)

Minor Improvements:
* auto-pause: add support for other container runtimes [#11834](https://github.com/kubernetes/minikube/pull/11834)
* windows: support renaming binary to `kubectl.exe` and running as kubectl [#11819](https://github.com/kubernetes/minikube/pull/11819)

Bugs:
* Fix "kubelet Default-Start contains no runlevels" error [#11815](https://github.com/kubernetes/minikube/pull/11815)

Version Upgrades:
* bump default kubernetes version to v1.21.2 & newest kubernetes version to v1.22.0-beta.0 [#11901](https://github.com/kubernetes/minikube/pull/11901)

For a more detailed changelog, including changes occurring in pre-release versions, see [CHANGELOG.md](https://github.com/kubernetes/minikube/blob/master/CHANGELOG.md).

Thank you to our contributors for this release!

- Anders F Björklund
- Andriy Dzikh
- Dakshraj Sharma
- Ilya Zuyev
- Jeff MAURY
- Maxime Kjaer
- Medya Ghazizadeh
- Rajwinder Mahal
- Sharif Elgamal
- Steven Powell

Thank you to our PR reviewers for this release!

- medyagh (27 comments)
- sharifelgamal (10 comments)
- andriyDev (5 comments)
- spowelljr (4 comments)
- ilya-zuyev (3 comments)

Thank you to our triage members for this release!

- medyagh (16 comments)
- spowelljr (7 comments)
- afbjorklund (4 comments)
- mahalrs (4 comments)
- sharifelgamal (3 comments)

## Version 1.22.0-beta.0 - 2021-06-28

Features:

* auto-pause addon: add support for arm64 [#11743](https://github.com/kubernetes/minikube/pull/11743)
* `addon list`: add info on each addon's maintainer  [#11753](https://github.com/kubernetes/minikube/pull/11753)
* add ability to pass max to `--cpu` and `--memory` flags [#11692](https://github.com/kubernetes/minikube/pull/11692)

Bugs:

* Fix `--base-image` caching for images specified by name:tag [#11603](https://github.com/kubernetes/minikube/pull/11603)
* Fix embed-certs global config [#11576](https://github.com/kubernetes/minikube/pull/11576)
* Fix a download link to use arm64 instead of amd64 [#11653](https://github.com/kubernetes/minikube/pull/11653)
* fix downloading duplicate base image [#11690](https://github.com/kubernetes/minikube/pull/11690)
* fix multi-node losing track of nodes after second restart [#11731](https://github.com/kubernetes/minikube/pull/11731)
* gcp-auth: do not override existing environment variables in pods [#11665](https://github.com/kubernetes/minikube/pull/11665)

Minor improvements:

* Allow running amd64 binary on M1 [#11674](https://github.com/kubernetes/minikube/pull/11674)
* improve containerd experience on cgroup v2 [#11632](https://github.com/kubernetes/minikube/pull/11632)
* Improve French locale [#11728](https://github.com/kubernetes/minikube/pull/11728)
* Fix UI error for stopping systemd service [#11667](https://github.com/kubernetes/minikube/pull/11667)
* international languages: allow using LC_ALL env to set local language for windows [#11721](https://github.com/kubernetes/minikube/pull/11721)
* Change registery_mirror to registery-mirror [#11678](https://github.com/kubernetes/minikube/pull/11678)

Version Upgrades:

* ISO: Upgrade podman to 3.1.2 [#11704](https://github.com/kubernetes/minikube/pull/11704)
* Upgrade Buildroot to 2021.02 LTS with Linux 4.19 [#11688](https://github.com/kubernetes/minikube/pull/11688)

For a more detailed changelog, including changes occurring in pre-release versions, see [CHANGELOG.md](https://github.com/kubernetes/minikube/blob/master/CHANGELOG.md).

Thank you to our contributors for this release!

- Anders F Björklund
- Andriy Dzikh
- Daehyeok Mun
- Dongjoon Hyun
- Felipe Crescencio de Oliveira
- Ilya Zuyev
- JacekDuszenko
- Jeff MAURY
- Medya Ghazizadeh
- Peixuan Ding
- RA489
- Sharif Elgamal
- Steven Powell
- Vishal Jain
- zhangdb-git

Thank you to our PR reviewers for this release!

- medyagh (63 comments)
- sharifelgamal (9 comments)
- ilya-zuyev (6 comments)
- andriyDev (3 comments)
- spowelljr (3 comments)
- afbjorklund (1 comments)
- prezha (1 comments)
- tharun208 (1 comments)

Thank you to our triage members for this release!

## Version 1.21.0 - 2021-06-10
* add more polish translations [#11587](https://github.com/kubernetes/minikube/pull/11587)
* Modify MetricsServer to use v1 api version (instead of v1beta1). [#11584](https://github.com/kubernetes/minikube/pull/11584)

For a more detailed changelog, including changes occurring in pre-release versions, see [CHANGELOG.md](https://github.com/kubernetes/minikube/blob/master/CHANGELOG.md).

Thank you to our contributors for this release!

- Andriy Dzikh
- Ilya Zuyev
- JacekDuszenko
- Medya Ghazizadeh
- Sharif Elgamal
- Steven Powell

Thank you to our PR reviewers for this release!

- spowelljr (11 comments)
- medyagh (2 comments)
- sharifelgamal (2 comments)
- andriyDev (1 comments)

Thank you to our triage members for this release!

- RA489 (12 comments)
- andriyDev (10 comments)
- sharifelgamal (10 comments)
- JacekDuszenko (7 comments)
- spowelljr (5 comments)

Check out our [contributions leaderboard](https://minikube.sigs.k8s.io/docs/contrib/leaderboard/v1.21.0/) for this release!

## Version 1.21.0-beta.0 - 2021-06-02
Features:
* Support setting addons from environmental variables [#11469](https://github.com/kubernetes/minikube/pull/11469)
* Add "resume" as an alias for "unpause" [#11431](https://github.com/kubernetes/minikube/pull/11431)
* Implement target node option for `cp` command [#11304](https://github.com/kubernetes/minikube/pull/11304)

Bugs:
* Fix delete command for paused kic driver with containerd/crio runtime [#11504](https://github.com/kubernetes/minikube/pull/11504)
* kicbase: try image without sha before failing [#11559](https://github.com/kubernetes/minikube/pull/11559)
* bug: return error on invalid function name in extract.TranslatableStrings [#11454](https://github.com/kubernetes/minikube/pull/11454)
* Prevent downloading duplicate binaries already present in preload [#11461](https://github.com/kubernetes/minikube/pull/11461)
* gcp-auth addon: do not reapply gcp-auth yamls on minikube restart [#11486](https://github.com/kubernetes/minikube/pull/11486)
* Disable Non-Active Containers Runtimes [#11516](https://github.com/kubernetes/minikube/pull/11516)
* Persist custom addon image/registry settings. [#11432](https://github.com/kubernetes/minikube/pull/11432)
* Fix auto-pause on VMs (detect right control-plane IP) [#11438](https://github.com/kubernetes/minikube/pull/11438)

Version Upgrades:
* bump default k8s version to v1.20.7 and newest to v1.22.0-alpha.2 [#11525](https://github.com/kubernetes/minikube/pull/11525)
* containerd: upgrade `io.containerd.runtime.v1.linux` to `io.containerd.runc.v2` (suppot cgroup v2) [#11325](https://github.com/kubernetes/minikube/pull/11325)
* metallb-addon: Update metallb from 0.8.2 to 0.9.6 [#11410](https://github.com/kubernetes/minikube/pull/11410)

For a more detailed changelog, including changes occurring in pre-release versions, see [CHANGELOG.md](https://github.com/kubernetes/minikube/blob/master/CHANGELOG.md).

Thank you to our contributors for this release!

- Akihiro Suda
- Alessandro Lenzen
- Anders F Björklund
- Andriy Dzikh
- Brian de Alwis
- Claudia J. Kang
- Daehyeok Mun
- Emma
- Evan Anderson
- Evan Baker
- Garen Torikian
- Ilya Zuyev
- Jasmine Hegman
- Kent Iso
- KushagraIndurkhya
- Li Zhijian
- Medya Ghazizadeh
- Peixuan Ding
- Predrag Rogic
- Sharif Elgamal
- Steven Powell
- TAKAHASHI Shuuji
- Thomas Güttler
- Tomasz Janiszewski
- Utkarsh Srivastava
- VigoTheHacker
- hex0punk

Thank you to our PR reviewers for this release!

- medyagh (129 comments)
- ilya-zuyev (20 comments)
- afbjorklund (10 comments)
- spowelljr (9 comments)
- sharifelgamal (5 comments)
- AkihiroSuda (1 comments)
- andriyDev (1 comments)

Thank you to our triage members for this release!

- afbjorklund (34 comments)
- medyagh (32 comments)
- andriyDev (14 comments)
- dinever (13 comments)
- ilya-zuyev (11 comments)


## Version 1.20.0 - 2021-05-06

Feature:
* Add --file flag to 'minikube logs' to automatically put logs into a file. [#11240](https://github.com/kubernetes/minikube/pull/11240)

Minor Improvements:
* Batch logs output to speedup `minikube logs` command [#11274](https://github.com/kubernetes/minikube/pull/11274)
* warn about performance for certain versions of kubernetes [#11217](https://github.com/kubernetes/minikube/pull/11217)

Version Upgrades:
* Update olm addon to v0.17.0 [#10947](https://github.com/kubernetes/minikube/pull/10947)
* Update newest supported Kubernetes version to v1.22.0-alpha.1 [#11287](https://github.com/kubernetes/minikube/pull/11287)

For a more detailed changelog, including changes occurring in pre-release versions, see [CHANGELOG.md](https://github.com/kubernetes/minikube/blob/master/CHANGELOG.md).

Thank you to our contributors for this release!

- Anders F Björklund
- Andriy Dzikh
- Daehyeok Mun
- Ilya Zuyev
- Medya Ghazizadeh
- Predrag Rogic
- Sharif Elgamal
- Steven Powell
- Tomas Kral
- Yanshu
- zhangshj


## Version 1.20.0-beta.0 - 2021-04-30

Features:

* New command: `build` to build images using minikube [#11164](https://github.com/kubernetes/minikube/pull/11164)
* New command 'image pull': allow to load remote images directly without cache [#11127](https://github.com/kubernetes/minikube/pull/11127)
* Add feature to opt-in to get notifications for beta releases [#11169](https://github.com/kubernetes/minikube/pull/11169)
* UI: Add log file to GitHub issue output [#11158](https://github.com/kubernetes/minikube/pull/11158)

Bug Fixes:

* Ingress Addon: fix bug which the networking.k8s.io/v1 ingress is always rejected [#11189](https://github.com/kubernetes/minikube/pull/11189)
* Improve how cni and cruntimes work together [#11185, #11209](https://github.com/kubernetes/minikube/pull/11209, https://github.com/kubernetes/minikube/pull/11185)
* Docker driver: support docker installed by Snap Package Manager [#11088](https://github.com/kubernetes/minikube/pull/11088)
* Change 'minikube version --short' to only print the version without a prompt. [#11167](https://github.com/kubernetes/minikube/pull/11167)


Thank you to our contributors for this release!

- Anders F Björklund
- Andriy Dzikh
- Ed Vinyard
- Hu Shuai
- Ilya Zuyev
- Kenta Iso
- Medya Ghazizadeh
- Michael Captain
- Predrag Rogic
- Sharif Elgamal
- Steven Powell
- Tobias Klauser
- csiepka
- hiroygo
- 李龙峰


## Version 1.19.0 - 2021-04-09

* allow Auto-Pause addon on VMs [#11019](https://github.com/kubernetes/minikube/pull/11019)
* Do not allow running darwin/amd64 minikube binary on darwin/arm64 systems [#11024](https://github.com/kubernetes/minikube/pull/11024)
* Respect memory being set in the minikube config [#11014](https://github.com/kubernetes/minikube/pull/11014)
* new command image ls to list images in a cluster [#11007](https://github.com/kubernetes/minikube/pull/11007)

Thank you to our contributors for this release!

- Anders F Björklund
- Cookie Wang
- Ilya Zuyev
- Medya Ghazizadeh
- Predrag Rogic
- Sharif Elgamal
- Steven Powell
- 李龙峰

## Version 1.19.0-beta.0 - 2021-04-05

Features:

* add `minikube image rm` command [#10924](https://github.com/kubernetes/minikube/pull/10924)
* GCP-Auth addon: Add support for GCR creds [#10853](https://github.com/kubernetes/minikube/pull/10853)
* new command: `minikube cp` to copy files into minikube [#10198](https://github.com/kubernetes/minikube/pull/10198)
* new flag "--listen-address" for docker and podman driver [#10653](https://github.com/kubernetes/minikube/pull/10653)
* iso: enable Network Block Device support [#10217](https://github.com/kubernetes/minikube/pull/10217)

Minor Improvements:

* auto-pause: initialize the pause state from the current state [#10958](https://github.com/kubernetes/minikube/pull/10958)
* iso: make sure to capture failures through pipes [#10974](https://github.com/kubernetes/minikube/pull/10974)
* Avoid logging 'kubeconfig endpoint' error when cluster is 'starting' [#10968](https://github.com/kubernetes/minikube/pull/10968)
* docker-env: improve detecting powershell if SSHed from linux [#10722](https://github.com/kubernetes/minikube/pull/10722)
* kvm2 driver: add dedicated network & static ip [#10792](https://github.com/kubernetes/minikube/pull/10792)
* Replace glog with klog [#10955](https://github.com/kubernetes/minikube/pull/10955)
* retry kapi.ScaleDeployment on failure [#10938](https://github.com/kubernetes/minikube/pull/10938)
* Auto-pause handle internal kubernetes requests [#10823](https://github.com/kubernetes/minikube/pull/10823)
* add additional options to avoid node drain or delete getting stuck [#10926](https://github.com/kubernetes/minikube/pull/10926)
* cache add: improved error message when image does not exist [#10811](https://github.com/kubernetes/minikube/pull/10811)
* Image load: Allow loading local images from tar or cache [#10807](https://github.com/kubernetes/minikube/pull/10807)
* status: Omit `timeToStop` if nonexistent [#10906](https://github.com/kubernetes/minikube/pull/10906)
* arm64: Fix incorrect image arch in the manifest for etcd and other kube images [#10642](https://github.com/kubernetes/minikube/pull/10642)
* add docker-env and podman-env to minikube status [#10872](https://github.com/kubernetes/minikube/pull/10872)
* adding new exit code word for when runtime not running  [#10364](https://github.com/kubernetes/minikube/pull/10364)
* Generate one log file per minikube command [#10425](https://github.com/kubernetes/minikube/pull/10425)
* bridge cni: Make sure to create the directory for cni config [#10868](https://github.com/kubernetes/minikube/pull/10868)
* docker-env: Add the daemon host address as Alternate Name in apiserver.crt if it's not an IP [#10873](https://github.com/kubernetes/minikube/pull/10873)
* Add solution message if Docker is rootless [#10878](https://github.com/kubernetes/minikube/pull/10878)
* Add a red box around docker desktop registry port [#10818](https://github.com/kubernetes/minikube/pull/10818)
* new flag --ssh for `minikube kubectl` to allow running it over the ssh connection [#10844](https://github.com/kubernetes/minikube/pull/10844)
* UI: Add progressbar when downloading kic base image [#10887](https://github.com/kubernetes/minikube/pull/10887)
* Show last start and audit logs on `minikube logs` if minikube not running [#10839](https://github.com/kubernetes/minikube/pull/10839)
* unique error codes for KVM network and docker ip conflict [#10841](https://github.com/kubernetes/minikube/pull/10841)
* Unique error code for no disk space [#10837](https://github.com/kubernetes/minikube/pull/10837)
* Add rpm and deb packaging for ppc64le and s390x [#10824](https://github.com/kubernetes/minikube/pull/10824)
* Provide unique error code (GUEST_KIC_CP_PUBKEY) for not copyable cert for kic [#10834](https://github.com/kubernetes/minikube/pull/10834)
* minikube kubectl: The --cluster flags should be prepended [#10793](https://github.com/kubernetes/minikube/pull/10793)
* ui: break down usage for no profile found [#10800](https://github.com/kubernetes/minikube/pull/10800)
* Enable portmap for the default cni bridge [#10782](https://github.com/kubernetes/minikube/pull/10782)
* install losetup from util-linux in the ISO to enable support for VolumeMode=Block PVCs [#10704](https://github.com/kubernetes/minikube/pull/10704)
* auto-detect gce and do not enable gcp auth addon [#10730](https://github.com/kubernetes/minikube/pull/10730)
* add validations --image-repository inputs [#10760](https://github.com/kubernetes/minikube/pull/10760)
* docker-env & podman-env: silent output when talking to a shell [#10763](https://github.com/kubernetes/minikube/pull/10763)
* The cluster doesn't have to be healthy for kubectl [#10732](https://github.com/kubernetes/minikube/pull/10732)
* Need to exit if unable to cache kubectl [#10734](https://github.com/kubernetes/minikube/pull/10734)
* increase wait for docker starting on windows [#10765](https://github.com/kubernetes/minikube/pull/10765)
* Correct spelling in --insecure-registry validation error message [#10735](https://github.com/kubernetes/minikube/pull/10735)
* kvm: provide solution if user doesn't belong to libvirt group [#10712](https://github.com/kubernetes/minikube/pull/10712)
* CoreDNS early scale down to 1 replica [#10656](https://github.com/kubernetes/minikube/pull/10656)
* Wait for crictl version after the socket is up [#10705](https://github.com/kubernetes/minikube/pull/10705)

Bug Fixes:

* Fix CNI issue related to picking up wrong CNI   [#10985](https://github.com/kubernetes/minikube/pull/10985)
* Improve validation for extra-config. [#10886](https://github.com/kubernetes/minikube/pull/10886)
* Fix the failure of `minikube mount` in case of KVM2 [#10733](https://github.com/kubernetes/minikube/pull/10733)
* Fix/minikube status for scheduled stop [#10911](https://github.com/kubernetes/minikube/pull/10911)
* create network: use locks and reservations to solve race condition [#10858](https://github.com/kubernetes/minikube/pull/10858)
* fix driver.IndexFromMachineName() [#10821](https://github.com/kubernetes/minikube/pull/10821)
* multinode cluster: fix waits and joins [#10758](https://github.com/kubernetes/minikube/pull/10758)
* hyperkit: fix hyperkit-vpnkit-sock setting [#10631](https://github.com/kubernetes/minikube/pull/10631)

Version changes:

* BuildKit 0.8.2 [#10648](https://github.com/kubernetes/minikube/pull/10648)
* ISO Upgrade Docker, from 20.10.3 to 20.10.4 [#10647](https://github.com/kubernetes/minikube/pull/10647)
* Addon: bump csi-hostpath-driver to v1.6.0 [#10798](https://github.com/kubernetes/minikube/pull/10798)
* Upgrade ingress addon files according to upstream(ingress-nginx v0.44.0) [#10879](https://github.com/kubernetes/minikube/pull/10879)
* addon: Upgrade VolumeSnapshot to GA(v1) [#10654](https://github.com/kubernetes/minikube/pull/10654)

Thank you to our contributors for this release!

- Anders F Björklund
- Andrew Stanton
- BLasan
- Daehyeok Mun
- Federico Gallo
- Ilya Zuyev
- Kent Iso
- Madhav Jivrajani
- Medya Ghazizadeh
- Niels de Vos
- Patrik Freij
- Prasanna Kumar Kalever
- Predrag Rogic
- Sharif Elgamal
- Steven Powell
- Szabolcs Dombi
- Tharun
- Thomas Strömberg
- Tom Di Nunzio
- Vishal Jain
- Yanshu Zhao
- alonyb
- anencore94
- bharathkkb
- dependabot[bot]
- hetong07
- ely
- maoyangLiu
- tripolkaandrey
- yxxhero
- zhangshj
- 李龙峰

## Version 1.18.1 - 2021-03-04

Features:

* kvm2 driver: Add flag --kvm-numa-count" support topology-manager simulate numa  [#10471](https://github.com/kubernetes/minikube/pull/10471)

Minor Improvements:

* Spanish translations [#10687](https://github.com/kubernetes/minikube/pull/10687)
* Change podman priority to default on Linux [#10458](https://github.com/kubernetes/minikube/pull/10458)

Bug Fixes:

* Remove WSLENV empty check from IsMicrosoftWSL [#10711](https://github.com/kubernetes/minikube/pull/10711)
* Added WaitGroups to prevent stderr/stdout from being empty in error logs [#10694](https://github.com/kubernetes/minikube/pull/10694)

Version changes:

* Restore kube-cross build image and bump go to version 1.16 [#10691](https://github.com/kubernetes/minikube/pull/10691)
* Bump github.com/spf13/viper from 1.7.0 to 1.7.1 [#10658](https://github.com/kubernetes/minikube/pull/10658)

Thank you to our contributors for this release!

- Anders F Björklund
- Emanuel
- Ilya Zuyev
- Medya Ghazizadeh
- Sharif Elgamal
- Steven Powell
- phantooom


## Version 1.18.0 - 2021-03-01

Bug Fixes:

* fix: metric server serve on all ipv4 interfaces [#10613](https://github.com/kubernetes/minikube/pull/10613)

Thank you to our contributors for this release!

- Medya Ghazizadeh
- Predrag Rogic
- Priya Wadhwa
- Sharif Elgamal
- liuwei10

## Version 1.18.0-beta.0 - 2021-02-26

Features:

* Auto-pause addon: automatically pause Kubernetes when not in use [#10427](https://github.com/kubernetes/minikube/pull/10427)
* GCP Auth addon: bump to v0.0.4 for multiarch [#10361](https://github.com/kubernetes/minikube/pull/10361)
* Add new command: image load [#10366](https://github.com/kubernetes/minikube/pull/10366)
* Add faster `profile list` command with -l or --light option. [#10380](https://github.com/kubernetes/minikube/pull/10380)
* Add last start logs to 'minikube logs' output [#10465](https://github.com/kubernetes/minikube/pull/10465)
* Introduce alias 'native' for 'none' driver [#10540](https://github.com/kubernetes/minikube/pull/10540)
* Add audit logs to 'minikube logs' output [#10350](https://github.com/kubernetes/minikube/pull/10350)
* Allow setting custom images for addons [#10111](https://github.com/kubernetes/minikube/pull/10111)

Minor Improvements:

* Deb package: make sure to update the package metadata [#10420](https://github.com/kubernetes/minikube/pull/10420)
* Improve the error message of setting cgroup memory limit. [#10575](https://github.com/kubernetes/minikube/pull/10575)
* SSH driver: Don't select Discouraged or Obsolete by default [#10554](https://github.com/kubernetes/minikube/pull/10554)
* drop support for github packages for kicbase [#10582](https://github.com/kubernetes/minikube/pull/10582)
* disable minikube-scheduled-stop.service until a user schedules a stop [#10548](https://github.com/kubernetes/minikube/pull/10548)
* docker/podman: add crun for running on cgroups v2 [#10426](https://github.com/kubernetes/minikube/pull/10426)
* Specify mount point for cri-o config [#10528](https://github.com/kubernetes/minikube/pull/10528)
* Esnure addon integrity by adding Image SHA [#10527](https://github.com/kubernetes/minikube/pull/10527)
* improve kvm network delete/cleanup [#10479](https://github.com/kubernetes/minikube/pull/10479)
* docker/podman: avoid creating overlapping networks with other tools (KVM,...) [#10439](https://github.com/kubernetes/minikube/pull/10439)
* Improve insecure registry validation [#10493](https://github.com/kubernetes/minikube/pull/10493)
* Stop using --memory-swap if it is not available [#10507](https://github.com/kubernetes/minikube/pull/10507)
* UI: do not send image repo info to stderr [#10462](https://github.com/kubernetes/minikube/pull/10462)
* add new extra component to --wait=all to validate a healthy cluster [#10424](https://github.com/kubernetes/minikube/pull/10424)
* Add condition to check --cpus count with available cpu count [#10388](https://github.com/kubernetes/minikube/pull/10388)
* Disable all drivers except "docker" and "ssh" on darwin/arm64 [#10452](https://github.com/kubernetes/minikube/pull/10452)
* Podman: explicitly remove podman volume and network [#10435](https://github.com/kubernetes/minikube/pull/10435)
* Disallow running windows binary (.exe) inside WSL [#10354](https://github.com/kubernetes/minikube/pull/10354)
* adding insecure registry support to containerd runtime [#10385](https://github.com/kubernetes/minikube/pull/10385)
* Docker driver: support ancient versions of docker [#10369](https://github.com/kubernetes/minikube/pull/10369)

Bug Fixes:

* cgroup v2: skip setting --memory limits when not configurable. [#10512](https://github.com/kubernetes/minikube/pull/10512)
* metallb addon: fix configuration  empty load balancing IP range [#10395](https://github.com/kubernetes/minikube/pull/10395)
* Fixed bug where tmp dir incorrectly set for Snap package manager [#10372](https://github.com/kubernetes/minikube/pull/10372)
* Fixed audit.json error when `delete --purge` ran [#10586](https://github.com/kubernetes/minikube/pull/10586)
* Fix exit message for insufficient memory [#10553](https://github.com/kubernetes/minikube/pull/10553)
* Fix minikube kubectl context switching [#10535](https://github.com/kubernetes/minikube/pull/10535)
* SSH driver: Make sure that ssh driver gets an ip address [#10309](https://github.com/kubernetes/minikube/pull/10309)
* Fix profile list when there are multi node clusters  [#9996](https://github.com/kubernetes/minikube/pull/9996)
* Don't allow profile names that conflict with a multi-node name [#10119](https://github.com/kubernetes/minikube/pull/10119)

Version changes:

* Buildroot 2020.02.10 [#10348](https://github.com/kubernetes/minikube/pull/10348)
* Change from crio-1.19 to crio-1.20 in kicbase [#10477](https://github.com/kubernetes/minikube/pull/10477)
* bump oldest kubernetes version to v1.14.0, bump default kubernetes version [#10531](https://github.com/kubernetes/minikube/pull/10531)
* Upgrade crio to 1.20.0 [#10476](https://github.com/kubernetes/minikube/pull/10476)
* Upgrade Docker, from 20.10.2 to 20.10.3 [#10417](https://github.com/kubernetes/minikube/pull/10417)

---

Thank you to our contributors for this release!

- Anders F Björklund
- BLasan
- Daehyeok Mun
- Federico Gallo
- Hari Udhayakumar
- Ilya Zuyev
- Jiefeng He
- John Losito
- Kent Iso
- Ling Samuel
- Maikel
- Medya Ghazizadeh
- Michael Henkel
- Moshi Binyamini
- Pablo Caderno
- Predrag Rogic
- Priya Wadhwa
- Sadlil
- Sebastian Madejski
- Sharif Elgamal
- Steven Powell
- Thomas Strömberg
- Yanshu Zhao
- alonyb
- ashwanth1109
- hetong07
- liuwei10
- vlad doster

## Version 1.17.1 - 2020-01-28

Features:

* Add new flag --user and to log executed commands [#10106](https://github.com/kubernetes/minikube/pull/10106)
* Unhide --schedule flag for scheduled stop [#10274](https://github.com/kubernetes/minikube/pull/10274)

Bugs:

* fixing debian and arch concurrent multiarch builds [#9998](https://github.com/kubernetes/minikube/pull/9998)
* configure the crictl yaml file to avoid the warning [#10221](https://github.com/kubernetes/minikube/pull/10221)


Thank you to our contributors for this release!

- Anders F Björklund
- BLasan
- Ilya Zuyev
- Jiefeng He
- Jorropo
- Medya Ghazizadeh
- Niels de Vos
- Priya Wadhwa
- Sharif Elgamal
- Steven Powell
- Thomas Strömberg
- andrzejsydor


## Version 1.17.0 - 2020-01-22

Features:

* Add multi-arch (arm64) support for docker/podman drivers [#9969](https://github.com/kubernetes/minikube/pull/9969)
* Add new driver "SSH" to bootstrap generic minkube clusters over ssh [#10099](https://github.com/kubernetes/minikube/pull/10099)
* Annotate Kubeconfig with  'Extension' to identify contexts/clusters created by minikube [#10126](https://github.com/kubernetes/minikube/pull/10126)
* Add support for systemd cgroup to containerd runtime [#10100](https://github.com/kubernetes/minikube/pull/10100)
* Add --network flag to select docker network to run with docker driver [#9538](https://github.com/kubernetes/minikube/pull/9538)



Minor Improvements:

* Improve exit codes by splitting PROVIDER_DOCKER_ERROR into more specific reason codes [#10212](https://github.com/kubernetes/minikube/pull/10212)
* Improve warning about the suggested memory size [#10187](https://github.com/kubernetes/minikube/pull/10187)
* Remove systemd dependency from none driver [#10112](https://github.com/kubernetes/minikube/pull/10112)
* Delete the existing cluster if guest driver mismatch [#10084](https://github.com/kubernetes/minikube/pull/10084)
* Remove obsolete 'vmwarefusion' driver, add friendly message [#9958](https://github.com/kubernetes/minikube/pull/9958)
* UI: Add a spinner for `creating container` step [#10024](https://github.com/kubernetes/minikube/pull/10024)
* Added validation for --insecure-registry values [#9977](https://github.com/kubernetes/minikube/pull/9977)


Bug Fixes:

* Snap package manager: fix cert copy issue    [#10042](https://github.com/kubernetes/minikube/pull/10042)
* Ignore non-socks5 ALL_PROXY env var when checking docker status [#10109](https://github.com/kubernetes/minikube/pull/10109)
* Docker-env: avoid race condition in bootstrap certs for parallel runs [#10118](https://github.com/kubernetes/minikube/pull/10118)
* Fix 'profile list' for multi-node clusters  [#9955](https://github.com/kubernetes/minikube/pull/9955)
* Change metrics-server pull policy to IfNotPresent [#10096](https://github.com/kubernetes/minikube/pull/10096)
* Podman driver: Handle installations without default bridge [#10092](https://github.com/kubernetes/minikube/pull/10092)
* Fix docker inspect network go template for network which doesn't have MTU [#10053](https://github.com/kubernetes/minikube/pull/10053)
* Docker/Podman: add control-plane to NO_PROXY [#10046](https://github.com/kubernetes/minikube/pull/10046)
* "cache add": fix command error when not specifying :latest tag  [#10058](https://github.com/kubernetes/minikube/pull/10058)
* Networking: Fix ClusterDomain value in kubeadm KubeletConfiguration [#10049](https://github.com/kubernetes/minikube/pull/10049)
* Fix typo in the csi-hostpath-driver addon name [#10034](https://github.com/kubernetes/minikube/pull/10034)


Upgrades:

* bump default Kubernetes version to v1.20.2 and add v1.20.3-rc.0 [#10194](https://github.com/kubernetes/minikube/pull/10194)
* Upgrade Docker, from 20.10.1 to 20.10.2 [#10154](https://github.com/kubernetes/minikube/pull/10154)
* ISO: Added sch_htb, cls_fw, cls_matchall, act_connmark and ifb kernel modules [#10048](https://github.com/kubernetes/minikube/pull/10048)
* ISO: add XFS_QUOTA support to guest vm [#9999](https://github.com/kubernetes/minikube/pull/9999)

Thank you to our contributors for this release!

- AUT0R3V
- Amar Tumballi
- Anders F Björklund
- Daehyeok Mun
- Eric Briand
- Ilya Zuyev
- Ivan Milchev
- Jituri, Pranav
- Laurent VERDOÏA
- Ling Samuel
- Medya Ghazizadeh
- Oliver Radwell
- Pablo Caderno
- Priya Wadhwa
- Sadlil
- Sharif Elgamal
- Steven Powell
- Thomas Strömberg
- Yanshu Zhao
- alonyb
- anencore94
- cxsu
- zouyu

## Version 1.16.0 - 2020-12-17

* fix ip node retrieve for none driver [#9986](https://github.com/kubernetes/minikube/pull/9986)
* remove experimental warning for multinode [#9987](https://github.com/kubernetes/minikube/pull/9987)
* Enable Ingress Addon for Docker Windows [#9761](https://github.com/kubernetes/minikube/pull/9761)
* Bump preload to v8 to include updated dashboard [#9984](https://github.com/kubernetes/minikube/pull/9984)
* Add ssh-host command for getting the ssh host keys [#9630](https://github.com/kubernetes/minikube/pull/9630)
* Added sub-step logging to adm init step on start [#9904](https://github.com/kubernetes/minikube/pull/9904)
* Add --node option for command `ip` and `ssh-key` [#9873](https://github.com/kubernetes/minikube/pull/9873)
* Upgrade Docker, from 20.10.0 to 20.10.1 [#9966](https://github.com/kubernetes/minikube/pull/9966)
* Upgrade kubernetes dashboard to v2.1.0 for 1.20 [#9963](https://github.com/kubernetes/minikube/pull/9963)
* Upgrade buildkit from 0.8.0 to 0.8.1 [#9967](https://github.com/kubernetes/minikube/pull/9967)

Thank you to our contributors for this release!

- Anders F Björklund
- Jituri, Pranav
- Ling Samuel
- Sharif Elgamal
- Steven Powell
- Thomas Strömberg
- priyawadhwa
- wangxy518

## Version 1.16.0-beta.0 - 2020-12-14

Features:
* Add persistent storage for /var/lib/buildkit [#9948](https://github.com/kubernetes/minikube/pull/9948)
* start: Support comma-delimited --addons [#9957](https://github.com/kubernetes/minikube/pull/9957)
* added statusName for kubeconfig [#9888](https://github.com/kubernetes/minikube/pull/9888)
* Add spinner at preparing Kubernetes... [#9855](https://github.com/kubernetes/minikube/pull/9855)
* Make none driver work as regular user (use sudo on demand) [#9379](https://github.com/kubernetes/minikube/pull/9379)
* Display ScheduledStop status in minikube status [#9793](https://github.com/kubernetes/minikube/pull/9793)
* Add support for restoring existing podman env [#9801](https://github.com/kubernetes/minikube/pull/9801)
* Add linux packages for the arm64 architecture [#9859](https://github.com/kubernetes/minikube/pull/9859)
* Ability to use a custom TLS certificate with the Ingress [#9797](https://github.com/kubernetes/minikube/pull/9797)
* Add private network implementation for podman [#9716](https://github.com/kubernetes/minikube/pull/9716)
* Add --cancel-scheduled flag to cancel all existing scheduled stops [#9774](https://github.com/kubernetes/minikube/pull/9774)
* Add OpenTelemetry tracing to minikube [#9723](https://github.com/kubernetes/minikube/pull/9723)
* Implement scheduled stop on windows [#9689](https://github.com/kubernetes/minikube/pull/9689)
* Support non-default docker endpoints [#9510](https://github.com/kubernetes/minikube/pull/9510)

Bug Fixes:
* wsl2: log warning if br_netfilter cannot be enabled rather than fatally exit [#9932](https://github.com/kubernetes/minikube/pull/9932)
* Fix podman network inspect format and error [#9866](https://github.com/kubernetes/minikube/pull/9866)
* Fix multi node two pods getting same IP and nodespec not having PodCIDR [#9875](https://github.com/kubernetes/minikube/pull/9875)
* Fix `node start` master node [#9833](https://github.com/kubernetes/minikube/pull/9833)
* Optionally use ssh for docker-env instead of tcp [#9548](https://github.com/kubernetes/minikube/pull/9548)
* Fix --extra-config when starting an existing cluster [#9634](https://github.com/kubernetes/minikube/pull/9634)
* fix unable to set memory in config [#9789](https://github.com/kubernetes/minikube/pull/9789)
* Set 'currentstep' for PullingBaseImage json event [#9844](https://github.com/kubernetes/minikube/pull/9844)
* Fix missing InitialSetup in `node start` [#9832](https://github.com/kubernetes/minikube/pull/9832)
* fix base image when using with custom image repository [#9791](https://github.com/kubernetes/minikube/pull/9791)
* add Restart=on-failure for inner docker systemd service [#9775](https://github.com/kubernetes/minikube/pull/9775)
* Add number of nodes for cluster in `minikube profile list` [#9702](https://github.com/kubernetes/minikube/pull/9702)
* Do not auto-select Hyper-V driver if session has no privilege [#9588](https://github.com/kubernetes/minikube/pull/9588)
* Fix registry-creds addon failure with ImageRepository [#9733](https://github.com/kubernetes/minikube/pull/9733)

Upgrades:
* Upgrade buildkit from 0.7.2 to 0.8.0 [#9940](https://github.com/kubernetes/minikube/pull/9940)
* Upgrade crio.conf to version v1.19.0 [#9917](https://github.com/kubernetes/minikube/pull/9917)
* Update the containerd configuration to v2 [#9915](https://github.com/kubernetes/minikube/pull/9915)
* update default kubernetes version to 1.20.0 [#9897](https://github.com/kubernetes/minikube/pull/9897)
* Upgrade CRI-O, from 1.18.4 to 1.19.0 [#9902](https://github.com/kubernetes/minikube/pull/9902)
* Update crictl to v1.19.0 [#9901](https://github.com/kubernetes/minikube/pull/9901)
* ISO: Upgrade Podman, from 2.2.0 to 2.2.1 [#9896](https://github.com/kubernetes/minikube/pull/9896)
* Upgrade go version to 1.15.5 [#9899](https://github.com/kubernetes/minikube/pull/9899)
* Upgrade Docker, from 19.03.14 to 20.10.0 [#9895](https://github.com/kubernetes/minikube/pull/9895)
* ISO: Upgrade podman to version 2.2.0 and remove varlink [#9635](https://github.com/kubernetes/minikube/pull/9635)
* KIC: Upgrade podman to version 2.2.0 and remove varlink [#9636](https://github.com/kubernetes/minikube/pull/9636)
* Upgrade kicbase to ubuntu:focal-20201106 [#9863](https://github.com/kubernetes/minikube/pull/9863)
* Upgrade Docker, from 19.03.13 to 19.03.14 [#9861](https://github.com/kubernetes/minikube/pull/9861)
* Buildroot 2020.02.8 [#9862](https://github.com/kubernetes/minikube/pull/9862)
* Update crictl to v1.18.0 [#9867](https://github.com/kubernetes/minikube/pull/9867)
* bump storage provisioner to multi arch [#9822](https://github.com/kubernetes/minikube/pull/9822)

Thank you to our contributors for this release!

- AUT0R3V
- Anders Björklund
- Andrea Spadaccini
- Brian Li
- Daehyeok Mun
- Ilya Zuyev
- Jeroen Dekkers
- Jituri, Pranav
- Ling Samuel
- Martin Schimandl
- Medya Ghazizadeh
- Parthvi Vala
- Peyton Duncan
- Predrag Rogic
- Priya Wadhwa
- Ruben Baez
- Sadlil
- Sharif Elgamal
- Stefan Lobbenmeier
- Steven Powell
- Tharun
- Thomas Strömberg
- Tpk
- Yehiyam Livneh
- edtrist
- msedzins

## Version 1.15.1 - 2020-11-16

Features:

* Add Support for driver name alias [#9672](https://github.com/kubernetes/minikube/pull/9672)

Bug fixes:

* less verbose language selector [#9715](https://github.com/kubernetes/minikube/pull/9715)

Thank you to our contributors for this release!

- Ben Leggett
- Medya Ghazizadeh
- Priya Wadhwa
- Sadlil
- Sharif Elgamal
- Vasilyev, Viacheslav

## Version 1.15.0 - 2020-11-13

Features:

* Add support for latest kubernetes version v1.20.0-beta.1  [#9693](https://github.com/kubernetes/minikube/pull/9693)
* Implement schedule stop for unix [#9503](https://github.com/kubernetes/minikube/pull/9503)
* New flag --watch flag for minikube status with optional interval duration [#9487](https://github.com/kubernetes/minikube/pull/9487)
* New flag  --namespace for activating non default kubeconfig context [#9506](https://github.com/kubernetes/minikube/pull/9506)
* Add JSON output to stop, pause and unpause [#9576](https://github.com/kubernetes/minikube/pull/9576)
* Add support for podman v2 to podman-env command [#9535](https://github.com/kubernetes/minikube/pull/9535)
* Support ImageRepository for addons [#9551](https://github.com/kubernetes/minikube/pull/9551)

Bug Fixes:

* implement "--log_file" and "--log_dir" for klog [#9592](https://github.com/kubernetes/minikube/pull/9592)
* GCP Auth Addon: support special location for cloud shell [#9674](https://github.com/kubernetes/minikube/pull/9674)
* Enable TCP Path MTU Discovery when an ICMP black hole is detected [#9537](https://github.com/kubernetes/minikube/pull/9537)
* Remove hard-coded list of valid cgroupfs mountpoints to bind mount [#9508](https://github.com/kubernetes/minikube/pull/9508)
* Improve parsing of start flag apiserver-names [#9385](https://github.com/kubernetes/minikube/pull/9385)
* kvm: recover from minikube-net network left over failures  [#9641](https://github.com/kubernetes/minikube/pull/9641)
* fix help flag 'pflag: help requested' error [#9614](https://github.com/kubernetes/minikube/pull/9614)
* Update "parallels" driver library and make this driver built into minikube [#9517](https://github.com/kubernetes/minikube/pull/9517)

Minor Improvements:

* Upgrade crio to 1.18.4 [#9628](https://github.com/kubernetes/minikube/pull/9628)
* Update ingress-nginx image to v0.40.2 [#9445](https://github.com/kubernetes/minikube/pull/9445)
* Improving log message when profile not found [#9613](https://github.com/kubernetes/minikube/pull/9613)
* Upgrade buildroot and kernel minor version [#9523](https://github.com/kubernetes/minikube/pull/9523)

Thank you to our contributors for this release!

- Anders F Björklund
- Evgeny Shmarnev
- Ma Xinjian
- Manuel Alejandro de Brito Fontes
- Medya Ghazizadeh
- Michael Ryan Dempsey
- Mikhail Zholobov
- Peter Lin
- Predrag Rogic
- Priya Wadhwa
- Sharif Elgamal
- Thomas Strömberg
- Yehiyam Livneh
- prezha
- vinu2003
- zouyu

## Version 1.14.2 - 2020-10-27

Bug Fixes:

* fix "profile list" timing out when cluster stopped. [#9557](https://github.com/kubernetes/minikube/pull/9557)

Thank you to our contributors for this release!

- Medya Ghazizadeh
- Sharif Elgamal
- Thomas Strömberg

## Version 1.14.1 - 2020-10-23

Features:

* new --wait flag component "kubelet" [#9459](https://github.com/kubernetes/minikube/pull/9459)

Bug Fixes:

* docker: When creating networks, use MTU of built-in bridge network [#9530](https://github.com/kubernetes/minikube/pull/9530)
* multinode: ensure worker node join control plane on restart [#9476](https://github.com/kubernetes/minikube/pull/9476)
* Fix "--native-ssh" flag for "minikube ssh" [#9417](https://github.com/kubernetes/minikube/pull/9417)
* Fix parallels driver initialization [#9494](https://github.com/kubernetes/minikube/pull/9494)

Minor Improvements:

* Omit error message if 100-crio-bridge.conf has already been disabled [#9505](https://github.com/kubernetes/minikube/pull/9505)
* avoid re-downloading hyperkit driver [#9365](https://github.com/kubernetes/minikube/pull/9365)
* improve gcp-auth addon failure policy [#9408](https://github.com/kubernetes/minikube/pull/9408)
* Added deprecation warning for --network-plugin=cni [#9368](https://github.com/kubernetes/minikube/pull/9368)
* Update warning message for local proxy. [#9490](https://github.com/kubernetes/minikube/pull/9490)
* bump helm-tiller addon to v2.16.12 [#9444](https://github.com/kubernetes/minikube/pull/9444)
* bump version for ingress dns addon [#9435](https://github.com/kubernetes/minikube/pull/9435)

Thank you to our contributors for this release!

- Anders F Björklund
- Dale Hamel
- GRXself
- Ilya Zuyev
- Josh Woodcock
- Joshua Mühlfort
- Kenta Iso
- Medya Ghazizadeh
- Mikhail Zholobov
- Nick Kubala
- Pablo Caderno
- Predrag Rogic
- Priya Modali
- Priya Wadhwa
- Sharif Elgamal
- Thomas Strömberg
- heyf

## Version 1.14.0 - 2020-10-08

Features:

* Delete context when stopped [#9414](https://github.com/kubernetes/minikube/pull/9414)
* New flag "--ports" to expose ports for docker & podman drivers [#9404](https://github.com/kubernetes/minikube/pull/9404)

Bug fixes and minor improvements:

* Ingress addon: fix the controller name [#9413](https://github.com/kubernetes/minikube/pull/9413)
* docker/podman drivers: no panic when updating mount-string with no configuration  [#9412](https://github.com/kubernetes/minikube/pull/9412)
* Improve solution message when there is no space left on device [#9316](https://github.com/kubernetes/minikube/pull/9316)

* To see more changes checkout the last beta release notes [1.14.0-beta.0](https://github.com/kubernetes/minikube/releases/tag/v1.14.0-beta.0).

Thank you to our contributors for this release.

- Anders F Björklund
- Asare Worae
- Medya Ghazizadeh
- Prajilesh N
- Predrag Rogic
- Priya Wadhwa
- Thomas Strömberg
- ToonvanStrijp

## Version 1.14.0-beta.0 - 2020-10-06

Features:

* add dedicated network for docker driver [#9294](https://github.com/kubernetes/minikube/pull/9294)
* Make sure gcp-auth addon can be enabled on startup [#9318](https://github.com/kubernetes/minikube/pull/9318)

Bug fixes:

* Fix minikube status bug when cluster is paused [#9383](https://github.com/kubernetes/minikube/pull/9383)
* don't allow profile name to be less than 2 characters [#9367](https://github.com/kubernetes/minikube/pull/9367)
* fix: "profile list" shows paused clusters as "Running" [#8978](https://github.com/kubernetes/minikube/pull/8978)
* Fix error in unittest, as pointed out by warning [#9345](https://github.com/kubernetes/minikube/pull/9345)

Improvements:

* update kicbase image to ubuntu-based [#9353](https://github.com/kubernetes/minikube/pull/9353)

Thank you to our contributors for this release!

- Anders F Björklund
- Bob Killen
- Daniel Weibel
- Dominik Braun
- Ilya Zuyev
- JJ Asghar
- Jituri, Pranav
- Medya Ghazizadeh
- Michael Ryan Dempsey
- Predrag Rogic
- Priya Wadhwa
- Sharif Elgamal
- Tacio Costa
- Thomas Strömberg
- Till Hoffmann
- loftkun
- programistka
- zhanwang

## Version 1.13.1 - 2020-09-18

* Update Default Kubernetes Version to v1.19.2 [#9265](https://github.com/kubernetes/minikube/pull/9265)
* fix mounting for docker driver in windows [#9263](https://github.com/kubernetes/minikube/pull/9263)
* CSI Hostpath Driver & VolumeSnapshots addons [#8461](https://github.com/kubernetes/minikube/pull/8461)
* docker/podman drivers: Make sure CFS_BANDWIDTH is available for --cpus [#9255](https://github.com/kubernetes/minikube/pull/9255)
* Fix ForwardedPort for podman version 2.0.1 and up [#9237](https://github.com/kubernetes/minikube/pull/9237)
* Avoid setting time for memory assets [#9256](https://github.com/kubernetes/minikube/pull/9256)
* point to newest gcp-auth-webhook version [#9199](https://github.com/kubernetes/minikube/pull/9199)
* Set preload=false if not using overlay2 as docker storage driver [#8831](https://github.com/kubernetes/minikube/pull/8831)
* Upgrade crio to 1.17.3 [#8922](https://github.com/kubernetes/minikube/pull/8922)
* Add Docker Desktop instructions if memory is greater than minimum but less than recommended [#9181](https://github.com/kubernetes/minikube/pull/9181)
* Update minimum memory constants to use MiB instead of MB [#9180](https://github.com/kubernetes/minikube/pull/9180)

Thank you to our contributors for this release!

- Anders F Björklund
- Dean Coakley
- Julien Breux
- Li Zhijian
- Medya Ghazizadeh
- Priya Wadhwa
- Sharif Elgamal
- Thomas Strömberg
- Zadjad Rezai
- jjanik

## Version 1.13.0 - 2020-09-03

Features:

* Update default Kubernetes version to v1.19.0 🎉 [#9050](https://github.com/kubernetes/minikube/pull/9050)
* start: Support for mounting host volumes on start with docker driver [#8159](https://github.com/kubernetes/minikube/pull/8159)
* start: Add a machine readable reason to all error paths [#9126](https://github.com/kubernetes/minikube/pull/9126)
* stop: add --keep-context-active flag [#9044](https://github.com/kubernetes/minikube/pull/9044)
* kubectl: Invoke kubectl if minikube binary is named 'kubectl' [#8872](https://github.com/kubernetes/minikube/pull/8872)

Bug fixes:

* docker: Choose the appropriate bridge interface when multiple exist [#9062](https://github.com/kubernetes/minikube/pull/9062)
* cache: Fix "cache add" for local images by cherry-picking go-containerregistry fix [#9160](https://github.com/kubernetes/minikube/pull/9160)
* update-context: Fix nil pointer dereference [#8989](https://github.com/kubernetes/minikube/pull/8989)
* start: Fix --extra-config for scheduler/controllerManager by removing hardcoded values [#9136](https://github.com/kubernetes/minikube/pull/9136)
* start: Fix --memory flag parsing in minikube start [#9033](https://github.com/kubernetes/minikube/pull/9033)
* start: Improve overlay module check (behavior & UX) [#9163](https://github.com/kubernetes/minikube/pull/9163)
* gcp-auth addon: trim whitespace when setting gcp project id [#9164](https://github.com/kubernetes/minikube/pull/9164)
* cni: Allow flannel CNI to work with kicbase by fixing IP conflict [#9046](https://github.com/kubernetes/minikube/pull/9046)
* cni: fix multiple node calico-node not ready [#9019](https://github.com/kubernetes/minikube/pull/9019)
* kic: Retry fix_cgroup on failure [#8974](https://github.com/kubernetes/minikube/pull/8974)
* json: fix type for kubectl version mismatch to warning [#9157](https://github.com/kubernetes/minikube/pull/9157)
* json: fix type for latest minikube availability message [#9109](https://github.com/kubernetes/minikube/pull/9109)
* addon-manager: Add namespace to persistent volume path [#9128](https://github.com/kubernetes/minikube/pull/9128)
* ssh: respect native-ssh flag [#8907](https://github.com/kubernetes/minikube/pull/8907)

Improvements:

* kic: Disable swap in Docker & podman containers [#9149](https://github.com/kubernetes/minikube/pull/9149)
* kic: prioritize /etc/hosts over dns [#9029](https://github.com/kubernetes/minikube/pull/9029)
* start: Repair kubecontext before checking cluster health [#9143](https://github.com/kubernetes/minikube/pull/9143)
* start: Don't enable kubelet until after kubeadm generates config [#9111](https://github.com/kubernetes/minikube/pull/9111)
* start: Add -o shorthand option for --output [#9097](https://github.com/kubernetes/minikube/pull/9097)
* ux: Add MINIKUBE_IN_STYLE auto-detection for Windows terminal [#9127](https://github.com/kubernetes/minikube/pull/9127)
* ux: Warn if /var disk space is full and add a solution message [#9028](https://github.com/kubernetes/minikube/pull/9028)
* iso Upgrade falco-module to version 0.24.0 [#9068](https://github.com/kubernetes/minikube/pull/9068)
* status: `minikube status` should display InsufficientStorage status  [#9034](https://github.com/kubernetes/minikube/pull/9034)
* perf: set proxy-refresh-interval=70000 for etcd to improve CPU overhead [#8850](https://github.com/kubernetes/minikube/pull/8850)
* json: buffer download progress every second [#9099](https://github.com/kubernetes/minikube/pull/9099)
* localization: Fix typos in pl translation [#9168](https://github.com/kubernetes/minikube/pull/9168)
* dashboard: Update dashboard to v2.0.3 [#9129](https://github.com/kubernetes/minikube/pull/9129)

Thank you to our many wonderful contributors for this release!

- AlexanderChen1989
- Ambor
- Anders F Björklund
- Anshul Sirur
- Asare Worae
- Chang-Woo Rhee
- Evgeny Shmarnev
- Jose Donizetti
- Kazuki Suda
- Li Zhijian
- Marcin Niemira
- Markus Frosch
- Medya Ghazizadeh
- Pablo Caderno
- Priya Wadhwa
- Sharif Elgamal
- Thomas Strömberg
- anencore94
- mckrl
- ollipa
- staticdev
- vinu2003
- zhanwang

## Version 1.12.3 - 2020-08-12

Features:

* Make waiting for Host configurable via --wait-timeout flag [#8948](https://github.com/kubernetes/minikube/pull/8948)

Bug Fixes:

* Ignore localhost proxy started with scheme. [#8885](https://github.com/kubernetes/minikube/pull/8885)
* Improve error handling for validating memory limits [#8959](https://github.com/kubernetes/minikube/pull/8959)
* Skip validations if --force is supplied [#8969](https://github.com/kubernetes/minikube/pull/8969)
* Fix handling of parseIP error [#8820](https://github.com/kubernetes/minikube/pull/8820)

Improvements:

* GCP Auth Addon: Exit with better error messages [#8932](https://github.com/kubernetes/minikube/pull/8932)
* Add warning for ingress addon enabled with driver of none [#8870](https://github.com/kubernetes/minikube/pull/8870)
* Update Japanese translation [#8967](https://github.com/kubernetes/minikube/pull/8967)
* Fix for a few typos in polish translations [#8950](https://github.com/kubernetes/minikube/pull/8950)

Thank you to our contributors for this release!

- Anders F Björklund
- Andrej Guran
- Chris Paika
- Dean Coakley
- Evgeny Shmarnev
- Ling Samuel
- Ma Xinjian
- Marcin Niemira
- Medya Ghazizadeh
- Pablo Caderno
- Priya Wadhwa
- RA489
- Sharif Elgamal
- TAKAHASHI Shuuji
- Thomas Strömberg
- inductor
- priyawadhwa
- programistka

## Version 1.12.2 - 2020-08-03

Features:

* New Addon: Automated GCP Credentials [#8682](https://github.com/kubernetes/minikube/pull/8682)
* status: Add experimental cluster JSON status with state transition support [#8868](https://github.com/kubernetes/minikube/pull/8868)
* Add support for Error type to JSON output [#8796](https://github.com/kubernetes/minikube/pull/8796)
* Implement Warning type for JSON output [#8793](https://github.com/kubernetes/minikube/pull/8793)
* Add stopping as a possible state in deleting, change errorf to warningf [#8896](https://github.com/kubernetes/minikube/pull/8896)
* Use preloaded tarball for cri-o container runtime [#8588](https://github.com/kubernetes/minikube/pull/8588)
* Add SCH_PRIO, SCH_SFQ and CLS_BASIC kernel module to add filter on traffic control [#8670](https://github.com/kubernetes/minikube/pull/8670)

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

Improvements:

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

- Szabolcs Dombi
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

Bug fixes:

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
* Add default CNI network for running with podman [#7754](https://github.com/kubernetes/minikube/pull/7754)
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
