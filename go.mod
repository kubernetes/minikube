module k8s.io/minikube

go 1.12

require (
	github.com/Parallels/docker-machine-parallels v1.3.0
	github.com/Sirupsen/logrus v0.0.0-20170822132746-89742aefa4b2 // indirect
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/blang/semver v3.5.0+incompatible
	github.com/c4milo/gotoolkit v0.0.0-20170318115440-bcc06269efa9 // indirect
	github.com/cloudfoundry-attic/jibber_jabber v0.0.0-20151120183258-bcc4c8345a21
	github.com/cloudfoundry/jibber_jabber v0.0.0-20151120183258-bcc4c8345a21 // indirect
	github.com/docker/docker v1.13.1 // indirect
	github.com/docker/go-units v0.3.3
	github.com/docker/machine v0.16.1-0.20190718054102-a555e4f7a8f5
	github.com/elazarl/goproxy v0.0.0-20190421051319-9d40249d3c2f
	github.com/elazarl/goproxy/ext v0.0.0-20190421051319-9d40249d3c2f // indirect
	github.com/fatih/color v1.7.0 // indirect
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/golang-collections/collections v0.0.0-20130729185459-604e922904d3
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/google/go-cmp v0.3.0
	github.com/google/go-containerregistry v0.0.0-20190318164241-019cdfc6adf9
	github.com/google/go-github/v25 v25.0.2
	github.com/gorilla/mux v1.7.1 // indirect
	github.com/hashicorp/errwrap v0.0.0-20141028054710-7554cd9344ce // indirect
	github.com/hashicorp/go-multierror v0.0.0-20160811015721-8c5f0ad93604 // indirect
	github.com/hashicorp/go-retryablehttp v0.5.4
	github.com/hashicorp/go-version v1.1.0 // indirect
	github.com/hooklift/assert v0.0.0-20170704181755-9d1defd6d214 // indirect
	github.com/hooklift/iso9660 v0.0.0-20170318115843-1cf07e5970d8
	github.com/intel-go/cpuid v0.0.0-20181003105527-1a4a6f06a1c6 // indirect
	github.com/jimmidyson/go-download v0.0.0-20161028105827-7f9a90c8c95b
	github.com/johanneswuerbach/nfsexports v0.0.0-20181204082207-1aa528dcb345
	github.com/libvirt/libvirt-go v3.4.0+incompatible
	github.com/machine-drivers/docker-machine-driver-vmware v0.1.1
	github.com/mattn/go-colorable v0.1.1 // indirect
	github.com/mattn/go-isatty v0.0.5
	github.com/mattn/go-runewidth v0.0.0-20161012013512-737072b4e32b // indirect
	github.com/mitchellh/go-ps v0.0.0-20170309133038-4fdf99ab2936
	github.com/moby/hyperkit v0.0.0-20171020124204-a12cd7250bcd
	github.com/olekukonko/tablewriter v0.0.0-20160923125401-bdcc175572fd
	github.com/pborman/uuid v1.2.0
	github.com/phayes/freeport v0.0.0-20180830031419-95f893ade6f2
	github.com/pkg/browser v0.0.0-20160118053552-9302be274faa
	github.com/pkg/errors v0.8.0
	github.com/pkg/profile v0.0.0-20161223203901-3a8809bd8a80
	github.com/pmezard/go-difflib v1.0.0
	github.com/samalba/dockerclient v0.0.0-20160414174713-91d7393ff859 // indirect
	github.com/shirou/gopsutil v2.18.12+incompatible
	github.com/sirupsen/logrus v1.4.1 // indirect
	github.com/spf13/cobra v0.0.4
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.3.2
	github.com/xeipuuv/gojsonpointer v0.0.0-20151027082146-e0fe6f683076 // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20150808065054-e02fc20de94c // indirect
	github.com/xeipuuv/gojsonschema v0.0.0-20160623135812-c539bca196be
	github.com/zchee/go-vmnet v0.0.0-20161021174912-97ebf9174097
	golang.org/x/crypto v0.0.0-20190611184440-5c40567a22f8
	golang.org/x/oauth2 v0.0.0-20190402181905-9f3314589c9a
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
	golang.org/x/sys v0.0.0-20190616124812-15dcb6c0061f
	golang.org/x/text v0.3.2
	gopkg.in/airbrake/gobrake.v2 v2.0.9 // indirect
	gopkg.in/cheggaaa/pb.v1 v1.0.6 // indirect
	gopkg.in/gemnasium/logrus-airbrake-hook.v2 v2.1.2 // indirect
	k8s.io/api v0.0.0
	k8s.io/apimachinery v0.0.0
	k8s.io/client-go v0.0.0
	k8s.io/kubernetes v1.15.0
	sigs.k8s.io/sig-storage-lib-external-provisioner v4.0.0+incompatible
)

replace (
	k8s.io/api => k8s.io/kubernetes/staging/src/k8s.io/api v0.0.0-20190623232353-8c3b7d7679cc
	k8s.io/apiextensions-apiserver => k8s.io/kubernetes/staging/src/k8s.io/apiextensions-apiserver v0.0.0-20190623232353-8c3b7d7679cc
	k8s.io/apimachinery => k8s.io/kubernetes/staging/src/k8s.io/apimachinery v0.0.0-20190623232353-8c3b7d7679cc
	k8s.io/apiserver => k8s.io/kubernetes/staging/src/k8s.io/apiserver v0.0.0-20190623232353-8c3b7d7679cc
	k8s.io/cli-runtime => k8s.io/kubernetes/staging/src/k8s.io/cli-runtime v0.0.0-20190623232353-8c3b7d7679cc
	k8s.io/client-go => k8s.io/kubernetes/staging/src/k8s.io/client-go v0.0.0-20190623232353-8c3b7d7679cc
	k8s.io/cloud-provider => k8s.io/kubernetes/staging/src/k8s.io/cloud-provider v0.0.0-20190623232353-8c3b7d7679cc
	k8s.io/cluster-bootstrap => k8s.io/kubernetes/staging/src/k8s.io/cluster-bootstrap v0.0.0-20190623232353-8c3b7d7679cc
	k8s.io/code-generator => k8s.io/kubernetes/staging/src/k8s.io/code-generator v0.0.0-20190623232353-8c3b7d7679cc
	k8s.io/component-base => k8s.io/kubernetes/staging/src/k8s.io/component-base v0.0.0-20190623232353-8c3b7d7679cc
	k8s.io/cri-api => k8s.io/kubernetes/staging/src/k8s.io/cri-api v0.0.0-20190623232353-8c3b7d7679cc
	k8s.io/csi-translation-lib => k8s.io/kubernetes/staging/src/k8s.io/csi-translation-lib v0.0.0-20190623232353-8c3b7d7679cc
	k8s.io/kube-aggregator => k8s.io/kubernetes/staging/src/k8s.io/kube-aggregator v0.0.0-20190623232353-8c3b7d7679cc
	k8s.io/kube-controller-manager => k8s.io/kubernetes/staging/src/k8s.io/kube-controller-manager v0.0.0-20190623232353-8c3b7d7679cc
	k8s.io/kube-proxy => k8s.io/kubernetes/staging/src/k8s.io/kube-proxy v0.0.0-20190623232353-8c3b7d7679cc
	k8s.io/kube-scheduler => k8s.io/kubernetes/staging/src/k8s.io/kube-scheduler v0.0.0-20190623232353-8c3b7d7679cc
	k8s.io/kubectl => k8s.io/kubernetes/staging/src/k8s.io/kubectl v0.0.0-20190623232353-8c3b7d7679cc
	k8s.io/kubelet => k8s.io/kubernetes/staging/src/k8s.io/kubelet v0.0.0-20190623232353-8c3b7d7679cc
	k8s.io/legacy-cloud-providers => k8s.io/kubernetes/staging/src/k8s.io/legacy-cloud-providers v0.0.0-20190623232353-8c3b7d7679cc
	k8s.io/metrics => k8s.io/kubernetes/staging/src/k8s.io/metrics v0.0.0-20190623232353-8c3b7d7679cc
	k8s.io/node-api => k8s.io/kubernetes/staging/src/k8s.io/node-api v0.0.0-20190623232353-8c3b7d7679cc
	k8s.io/sample-apiserver => k8s.io/kubernetes/staging/src/k8s.io/sample-apiserver v0.0.0-20190623232353-8c3b7d7679cc
	k8s.io/sample-cli-plugin => k8s.io/kubernetes/staging/src/k8s.io/sample-cli-plugin v0.0.0-20190623232353-8c3b7d7679cc
	k8s.io/sample-controller => k8s.io/kubernetes/staging/src/k8s.io/sample-controller v0.0.0-20190623232353-8c3b7d7679cc
)
