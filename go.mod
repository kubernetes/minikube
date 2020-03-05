module k8s.io/minikube

go 1.13

require (
	cloud.google.com/go v0.45.1
	github.com/BurntSushi/toml v0.3.1
	github.com/Microsoft/go-winio v0.4.15-0.20190919025122-fc70bd9a86b5
	github.com/Microsoft/hcsshim v0.8.7
	github.com/Parallels/docker-machine-parallels v1.3.0
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/blang/semver v3.5.0+incompatible
	github.com/c4milo/gotoolkit v0.0.0-20170318115440-bcc06269efa9 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/checkpoint-restore/go-criu v0.0.0-20190109184317-bdb7599cd87b
	github.com/cheggaaa/pb/v3 v3.0.1
	github.com/cilium/ebpf v0.0.0-20200303152956-35c8d6cff6e0
	github.com/cloudfoundry-attic/jibber_jabber v0.0.0-20151120183258-bcc4c8345a21
	github.com/cloudfoundry/jibber_jabber v0.0.0-20151120183258-bcc4c8345a21 // indirect
	github.com/containerd/aufs v0.0.0-20200106064538-76944a95669d
	github.com/containerd/btrfs v0.0.0-20200117014249-153935315f4a
	github.com/containerd/cgroups v0.0.0-20200226104544-44306b6a1d46
	github.com/containerd/console v0.0.0-20191206165004-02ecf6a7291e
	github.com/containerd/containerd v1.3.1-0.20191213020239-082f7e3aed57
	github.com/containerd/continuity v0.0.0-20200228182428-0f16d7a0959c
	github.com/containerd/cri v1.11.1
	github.com/containerd/fifo v0.0.0-20191213151349-ff969a566b00
	github.com/containerd/go-cni v0.0.0-20200107172653-c154a49e2c75 // indirect
	github.com/containerd/go-runc v0.0.0-20200220073739-7016d3ce2328
	github.com/containerd/ttrpc v0.0.0-20200121165050-0be804eadb15
	github.com/containerd/typeurl v0.0.0-20190911142611-5eb25027c9fd
	github.com/containerd/zfs v0.0.0-20200115132605-fdbd9435326f
	github.com/containernetworking/cni v0.7.1
	github.com/containernetworking/plugins v0.8.5
	github.com/containers/conmon v2.0.11+incompatible
	github.com/coreos/go-iptables v0.4.5
	github.com/coreos/go-systemd v0.0.0-20190321100706-95778dfbb74e
	github.com/cyphar/filepath-securejoin v0.2.2
	github.com/d2g/dhcp4 v0.0.0-20170904100407-a1d1b6c41b1c
	github.com/d2g/dhcp4client v1.0.0
	github.com/docker/cli v0.0.0-20200303162255-7d407207c304 // indirect
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v1.13.1
	github.com/docker/go-events v0.0.0-20190806004212-e31b211e4f1c
	github.com/docker/go-metrics v0.0.1
	github.com/docker/go-units v0.4.0
	github.com/docker/machine v0.7.1-0.20190718054102-a555e4f7a8f5 // version is 0.7.1 to pin to a555e4f7a8f5
	github.com/elazarl/goproxy v0.0.0-20190421051319-9d40249d3c2f
	github.com/elazarl/goproxy/ext v0.0.0-20190421051319-9d40249d3c2f // indirect
	github.com/evanphx/json-patch v4.5.0+incompatible // indirect
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/godbus/dbus v0.0.0-20190422162347-ade71ed3457e
	github.com/gogo/googleapis v1.3.2
	github.com/gogo/protobuf v1.3.1
	github.com/golang-collections/collections v0.0.0-20130729185459-604e922904d3
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/protobuf v1.3.2
	github.com/google/go-cmp v0.3.1
	github.com/google/go-containerregistry v0.0.0-20200131185320-aec8da010de2
	github.com/google/pprof v0.0.0-20190515194954-54271f7e092f
	github.com/googleapis/gnostic v0.3.0 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/hashicorp/go-getter v1.4.0
	github.com/hashicorp/go-retryablehttp v0.5.4
	github.com/hooklift/assert v0.0.0-20170704181755-9d1defd6d214 // indirect
	github.com/hooklift/iso9660 v0.0.0-20170318115843-1cf07e5970d8
	github.com/ianlancetaylor/demangle v0.0.0-20181102032728-5e5cf60278f6 // indirect
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/intel-go/cpuid v0.0.0-20181003105527-1a4a6f06a1c6 // indirect
	github.com/j-keck/arping v1.0.0
	github.com/johanneswuerbach/nfsexports v0.0.0-20181204082207-1aa528dcb345
	github.com/juju/clock v0.0.0-20190205081909-9c5c9712527c
	github.com/juju/errors v0.0.0-20190806202954-0232dcc7464d // indirect
	github.com/juju/loggo v0.0.0-20190526231331-6e530bcce5d8 // indirect
	github.com/juju/mutex v0.0.0-20180619145857-d21b13acf4bf
	github.com/juju/retry v0.0.0-20180821225755-9058e192b216 // indirect
	github.com/juju/testing v0.0.0-20190723135506-ce30eb24acd2 // indirect
	github.com/juju/utils v0.0.0-20180820210520-bf9cc5bdd62d // indirect
	github.com/juju/version v0.0.0-20180108022336-b64dbd566305 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51
	github.com/libvirt/libvirt-go v3.4.0+incompatible
	github.com/machine-drivers/docker-machine-driver-vmware v0.1.1
	github.com/mattn/go-isatty v0.0.9
	github.com/mattn/go-shellwords v1.0.5
	github.com/mitchellh/go-ps v0.0.0-20170309133038-4fdf99ab2936
	github.com/moby/hyperkit v0.0.0-20171020124204-a12cd7250bcd
	github.com/mrunalp/fileutils v0.0.0-20171103030105-7d4729fb3618
	github.com/olekukonko/tablewriter v0.0.0-20170122224234-a0225b3f23b5
	github.com/onsi/ginkgo v1.10.3
	github.com/onsi/gomega v1.7.1
	github.com/opencontainers/go-digest v1.0.0-rc1
	github.com/opencontainers/image-spec v1.0.1
	github.com/opencontainers/runc v1.0.0-rc9
	github.com/opencontainers/runtime-spec v1.0.1
	github.com/opencontainers/selinux v1.3.1-0.20190929122143-5215b1806f52
	github.com/otiai10/copy v1.0.2
	github.com/pborman/uuid v1.2.0
	github.com/phayes/freeport v0.0.0-20180830031419-95f893ade6f2
	github.com/pkg/browser v0.0.0-20160118053552-9302be274faa
	github.com/pkg/errors v0.9.1
	github.com/pkg/profile v0.0.0-20161223203901-3a8809bd8a80
	github.com/pmezard/go-difflib v1.0.0
	github.com/prometheus/client_golang v1.1.0
	github.com/samalba/dockerclient v0.0.0-20160414174713-91d7393ff859 // indirect
	github.com/seccomp/libseccomp-golang v0.9.1
	github.com/shirou/gopsutil v2.18.12+incompatible
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.3.2
	github.com/syndtr/gocapability v0.0.0-20180916011248-d98352740cb2
	github.com/tchap/go-patricia v2.3.0+incompatible // indirect
	github.com/urfave/cli v1.22.2
	github.com/vishvananda/netlink v1.0.0
	github.com/xeipuuv/gojsonschema v0.0.0-20180618132009-1d523034197f
	github.com/zchee/go-vmnet v0.0.0-20161021174912-97ebf9174097
	go.etcd.io/bbolt v1.3.3
	golang.org/x/arch v0.0.0-20191126211547-368ea8f32fff
	golang.org/x/build v0.0.0-20190927031335-2835ba2e683f
	golang.org/x/crypto v0.0.0-20200302210943-78000ba7a073
	golang.org/x/net v0.0.0-20200301022130-244492dfa37a
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	golang.org/x/sys v0.0.0-20200124204421-9fbb57f87de9
	golang.org/x/text v0.3.2
	google.golang.org/api v0.9.0
	google.golang.org/grpc v1.26.0
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22 // indirect
	gotest.tools v2.2.0+incompatible
	gotest.tools/v3 v3.0.2 // indirect
	k8s.io/api v0.17.3
	k8s.io/apimachinery v0.17.3
	k8s.io/client-go v0.17.3
	k8s.io/kubectl v0.0.0
	k8s.io/kubernetes v1.17.3
	k8s.io/utils v0.0.0-20200229041039-0a110f9eb7ab // indirect
	sigs.k8s.io/sig-storage-lib-external-provisioner v4.0.0+incompatible
)

replace (
	git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999
	github.com/docker/docker => github.com/docker/docker v1.4.2-0.20190924003213-a8608b5b67c7
	github.com/docker/machine => github.com/machine-drivers/machine v0.7.1-0.20191109154235-b39d5b50de51
	github.com/hashicorp/go-getter => github.com/afbjorklund/go-getter v1.4.1-0.20190910175809-eb9f6c26742c
	github.com/samalba/dockerclient => github.com/sayboras/dockerclient v0.0.0-20191231050035-015626177a97
	k8s.io/api => k8s.io/api v0.17.3
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.3
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.3
	k8s.io/apiserver => k8s.io/apiserver v0.17.3
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.17.3
	k8s.io/client-go => k8s.io/client-go v0.17.3
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.17.3
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.17.3
	k8s.io/code-generator => k8s.io/code-generator v0.17.3
	k8s.io/component-base => k8s.io/component-base v0.17.3
	k8s.io/cri-api => k8s.io/cri-api v0.17.3
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.17.3
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.17.3
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.17.3
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.17.3
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.17.3
	k8s.io/kubectl => k8s.io/kubectl v0.17.3
	k8s.io/kubelet => k8s.io/kubelet v0.17.3
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.17.3
	k8s.io/metrics => k8s.io/metrics v0.17.3
	k8s.io/node-api => k8s.io/node-api v0.17.3
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.17.3
	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.17.3
	k8s.io/sample-controller => k8s.io/sample-controller v0.17.3
)
