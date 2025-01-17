## Version 1.35.0 - 2025-01-15

Features:
* Add support for AMD GPUs via --gpus=amd [#19749](https://github.com/kubernetes/minikube/pull/19749)
* Support latest Kubernetes v1.32.0 [#20091](https://github.com/kubernetes/minikube/pull/20091)
* Adds support for kubeadm.k8s.io/v1beta4 available since k8s v1.31 [#19790](https://github.com/kubernetes/minikube/pull/19790)
* Download kicbase from github assets as a fail over option  [#19464](https://github.com/kubernetes/minikube/pull/19464)

Improvements:
* Merge nvidia-gpu-device-plugin and nvidia-device-plugin. [#19545](https://github.com/kubernetes/minikube/pull/19545)
* cilium: remove appArmorProfile for k8s<v1.30.0 [#19888](https://github.com/kubernetes/minikube/pull/19888)
* auto-pause: restart service after configuration [#19900](https://github.com/kubernetes/minikube/pull/19900)
* Revert "Change MINIKUBE_HOME logic" [#20045](https://github.com/kubernetes/minikube/pull/20045)
* don't pollute minikube profile list with errors if exitcode is absent [#19728](https://github.com/kubernetes/minikube/pull/19728)
* unified minikube cluster status query [#18998](https://github.com/kubernetes/minikube/pull/18998)
* Vfkit driver: fix TestMachineType failing on macOS [#19726](https://github.com/kubernetes/minikube/pull/19726)
* No more arch restriction on nerdctld [#19730](https://github.com/kubernetes/minikube/pull/19730)
* remove helm-tiller addon [#19636](https://github.com/kubernetes/minikube/pull/19636)
* More robust MAC address matching [#19750](https://github.com/kubernetes/minikube/pull/19750)
* Add instructions to resolve docker context error [#19197](https://github.com/kubernetes/minikube/pull/19197)

Bug fixes:
* fix --wait's failure to work on coredns pods [#19748](https://github.com/kubernetes/minikube/pull/19748)
* Fix panic when no services in namespace with --all specified [#19957](https://github.com/kubernetes/minikube/pull/19957)
* fix timeout when stopping KVM machine with CRI-O container runtime [#19758](https://github.com/kubernetes/minikube/pull/19758)
* Fix long lines in lastStart.txt not outputting in log outputs [#19740](https://github.com/kubernetes/minikube/pull/19740)
* Fix wrongly detecting kicbase arch as incorrect [#19664](https://github.com/kubernetes/minikube/pull/19664)

Breaking Changes:
* skip building kvm2-arm64 till 19959 is resolved [#20062](https://github.com/kubernetes/minikube/pull/20062)
* remove arm64 kvm [#19985](https://github.com/kubernetes/minikube/pull/19985)


Languages:
* Add more Chinese translations [#19490](https://github.com/kubernetes/minikube/pull/19490) [#19508](https://github.com/kubernetes/minikube/pull/19508) [#19962](https://github.com/kubernetes/minikube/pull/19962)  [#19718](https://github.com/kubernetes/minikube/pull/19718)  [#19772](https://github.com/kubernetes/minikube/pull/19772)
* Improve french translation [#19654](https://github.com/kubernetes/minikube/pull/19654) [#19978](https://github.com/kubernetes/minikube/pull/19978)




Version Updates:
* CNI: Update flannel from v0.25.6 to v0.25.7 [#19761](https://github.com/kubernetes/minikube/pull/19761)
* CNI: Update flannel to v0.26.2 [#20107](https://github.com/kubernetes/minikube/pull/20107)
* CNI: Update cilium from v1.16.1 to v1.16.2 [#19734](https://github.com/kubernetes/minikube/pull/19734)
* CNI: Update cilium to v1.16.5 [#20148](https://github.com/kubernetes/minikube/pull/20148)
* CNI: Update calico to v3.29.0 [#19884](https://github.com/kubernetes/minikube/pull/19884)
* CNI: Update cilium from v1.16.2 to v1.16.3 [#19823](https://github.com/kubernetes/minikube/pull/19823)
* CNI: Update kindnetd from v20240813-c6f155d6 to v20241007-36f62932 [#19780](https://github.com/kubernetes/minikube/pull/19780)
* CNI: Update kindnetd from v20241007-36f62932 to v20241023-a345ebe4 [#19865](https://github.com/kubernetes/minikube/pull/19865)
* CNI: Update calico from v3.28.1 to v3.28.2 [#19667](https://github.com/kubernetes/minikube/pull/19667)
* Update golang to 1.23.3 [#20065](https://github.com/kubernetes/minikube/pull/20065)
* CNI: Update calico from v3.29.0 to v3.29.1 [#20052](https://github.com/kubernetes/minikube/pull/20052)
* CNI: Update kindnetd from v20241023-a345ebe4 to v20241108-5c6d2daf [#20051](https://github.com/kubernetes/minikube/pull/20051)
* Addon istio-provisioner: Update istio/operator image from 1.23.1 to 1.23.2 [#19678](https://github.com/kubernetes/minikube/pull/19678)
* Addon cloud-spanner: Update cloud-spanner-emulator/emulator image from 1.5.23 to 1.5.24 [#19679](https://github.com/kubernetes/minikube/pull/19679)
* Addon gcp-auth: Update k8s-minikube/gcp-auth-webhook image from v0.1.2 to v0.1.3 [#19787](https://github.com/kubernetes/minikube/pull/19787)
* Addon ingress: Update ingress-nginx/controller image from v1.11.2 to v1.11.3 [#19781](https://github.com/kubernetes/minikube/pull/19781)
* Addon registry: Update kube-registry-proxy image from 0.0.7 to 0.0.8 [#19782](https://github.com/kubernetes/minikube/pull/19782)
* addon gvisor: Update gvisor-addon image from v0.0.1 to v0.0.2 [#19776](https://github.com/kubernetes/minikube/pull/19776)
* Addon istio-provisioner: Update istio/operator image from 1.23.0 to 1.23.1 [#19629](https://github.com/kubernetes/minikube/pull/19629)
* Addon inspektor-gadget: Update inspektor-gadget image from v0.35.0 to v0.36.0 [#20205](https://github.com/kubernetes/minikube/pull/20205)
* Addon inspektor-gadget: Update inspektor-gadget image from v0.33.0 to v0.35.0 [#20033](https://github.com/kubernetes/minikube/pull/20033)
* Addon registry: Update registry image from 2.8.3 to 2.8.3 [#20035](https://github.com/kubernetes/minikube/pull/20035)
* Addon kubevirt: Update bitnami/kubectl image from 1.31.1 to 1.31.1 [#19763](https://github.com/kubernetes/minikube/pull/19763)
* Addon kubevirt: Update bitnami/kubectl image from 1.31.2 to 1.31.3 [#20028](https://github.com/kubernetes/minikube/pull/20028)
* Addon kubevirt: Update bitnami/kubectl image from 1.31.3 to 1.31.3 [#20068](https://github.com/kubernetes/minikube/pull/20068)
* Addon kubevirt: Update bitnami/kubectl image from 1.31.2 to 1.31.2 [#19937](https://github.com/kubernetes/minikube/pull/19937)
* Addon kubevirt: Update bitnami/kubectl image from 1.31.1 to 1.31.1 [#19690](https://github.com/kubernetes/minikube/pull/19690)
* Addon headlamp: Update headlamp-k8s/headlamp image from v0.25.0 to v0.25.1 [#19570](https://github.com/kubernetes/minikube/pull/19570)
* Addon kong: Update kong image from 3.7.1 to 3.8.0 [#19651](https://github.com/kubernetes/minikube/pull/19651)
* Addon kong: Update kong/kubernetes-ingress-controller image from 2.12.0 to 3.3.1 [#18424](https://github.com/kubernetes/minikube/pull/18424)
* Addon kubevirt: Update bitnami/kubectl image from 1.31.0 to 1.31.1 [#19652](https://github.com/kubernetes/minikube/pull/19652)
* Addon Volcano: Update volcano images from v1.9.0 to v1.10.0 [#19700](https://github.com/kubernetes/minikube/pull/19700)
* Addon kong: Update kong image from 3.8.0 to 3.8.0 [#19689](https://github.com/kubernetes/minikube/pull/19689)
* Addon ingress: Update ingress-nginx/controller image from v1.11.3 to v1.12.0-beta.0 [#19824](https://github.com/kubernetes/minikube/pull/19824)
* Addon inspektor-gadget: Update inspektor-gadget image from v0.32.0 to v0.33.0 [#19764](https://github.com/kubernetes/minikube/pull/19764)
* Addons registry: Update kube-registry-proxy from 0.0.6 to 0.0.7 [#19711](https://github.com/kubernetes/minikube/pull/19711)
* Kicbase: Bump ubuntu:jammy from 20240808 to 20240911.1 [#19662](https://github.com/kubernetes/minikube/pull/19662)
* Kicbase/ISO: Update nerdctl from 1.7.6 to 1.7.7 [#19649](https://github.com/kubernetes/minikube/pull/19649)
* Kicbase/ISO: Update cni-plugins from v1.6.1 to v1.6.2 [#20236](https://github.com/kubernetes/minikube/pull/20236)
* Kicbase/ISO: Update buildkit from v0.16.0 to v0.18.1 [#20089](https://github.com/kubernetes/minikube/pull/20089)
* Kicbase/ISO: Update crun from 1.16.1 to 1.17 [#19640](https://github.com/kubernetes/minikube/pull/19640)
* Kicbase/ISO: Update crun from 1.18.2 to 1.19 [#20083](https://github.com/kubernetes/minikube/pull/20083)
* Kicbase/ISO: Update crun from 1.18 to 1.18.2 [#19917](https://github.com/kubernetes/minikube/pull/19917)
* Kicbase/ISO: Update dependency versions [#20090](https://github.com/kubernetes/minikube/pull/20090)
* Kicbase/ISO: Update runc from v1.1.13 to v1.1.14 [#19598](https://github.com/kubernetes/minikube/pull/19598)
* Kicbase/ISO: Update runc from v1.1.14 to v1.1.15 [#19774](https://github.com/kubernetes/minikube/pull/19774)
* Kicbase/ISO: Update buildkit from v0.15.2 to v0.16.0 [#19644](https://github.com/kubernetes/minikube/pull/19644)
* Kicbase/ISO: Update containerd from v1.7.21 to v1.7.22 [#19643](https://github.com/kubernetes/minikube/pull/19643)
* Kicbase/ISO: Update docker from 27.2.0 to 27.2.1 [#19616](https://github.com/kubernetes/minikube/pull/19616)
* Kicbase/ISO: Update docker from 27.2.1 to 27.3.0 [#19672](https://github.com/kubernetes/minikube/pull/19672)
* Kicbase/ISO: Update docker from 27.3.0 to 27.3.1 [#19696](https://github.com/kubernetes/minikube/pull/19696)
* Kicbase/ISO: Update containerd from v1.7.22 to v1.7.23 [#19806](https://github.com/kubernetes/minikube/pull/19806)
* Kicbase/ISO: Update crun from 1.17 to 1.18 [#19883](https://github.com/kubernetes/minikube/pull/19883)
* Kicbase/ISO: Update cni-plugins from v1.5.1 to v1.6.0 [#19872](https://github.com/kubernetes/minikube/pull/19872)
* HA (multi-control plane): Update kube-vip from v0.8.0 to v0.8.3 [#19736](https://github.com/kubernetes/minikube/pull/19736)
* HA (multi-control plane): Update kube-vip from v0.8.3 to v0.8.4 [#19800](https://github.com/kubernetes/minikube/pull/19800)
19890)
* HA (multi-control plane): Update kube-vip from v0.8.4 to v0.8.5 [#19906](https://github.com/kubernetes/minikube/pull/19906)
* HA (multi-control plane): Update kube-vip from v0.8.5 to v0.8.6 [#19910](https://github.com/kubernetes/minikube/pull/19910)
* Addon istio-provisioner: Update istio/operator image from 1.23.2 to 1.23.3 [#19876](https://github.com/kubernetes/minikube/pull/19876)
* Addon kubevirt: Update bitnami/kubectl image from 1.31.1 to 1.31.2 [#19875](https://github.com/kubernetes/minikube/pull/19875)
* Update go from 1.23.0 to 1.23.1 [#19756](https://github.com/kubernetes/minikube/pull/19756)
* Update go from 1.23.1 to 1.23.2 [#19868](https://github.com/kubernetes/minikube/pull/19868)
* kicbase: Update nvidia packages [#19738](https://github.com/kubernetes/minikube/pull/19738)


For a more detailed changelog, including changes occurring in pre-release versions, see [CHANGELOG.md](https://github.com/kubernetes/minikube/blob/master/CHANGELOG.md).

Thank you to our contributors for this release!

- Anders F Björklund
- Bingtan Lu
- Fredrik Holmqvist
- Hello World
- Jeff MAURY
- Kubernetes Prow Robot
- Matt L
- Medya Gh
- Medya Ghazizadeh
- Nir Soffer
- Predrag Rogic
- Qasim Sarfraz
- Ramachandran C
- Sarath Kumar
- Steven Powell
- Sylvester Carolan
- Szymon Nadbrzeżny
- Tyler Auerbeck
- cuiyourong
- dependabot[bot]
- fbyrne
- joaquimrocha
- minikube-bot
- shixiuguo
- syxunion
- tianlj
- zdxgs
- 錦南路之花

Thank you to our PR reviewers for this release!
