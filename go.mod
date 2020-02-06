module k8s.io/minikube

go 1.13

require (
	github.com/Parallels/docker-machine-parallels v1.3.0
	github.com/Sirupsen/logrus v0.0.0-20170822132746-89742aefa4b2 // indirect
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/blang/semver v3.5.0+incompatible
	github.com/c4milo/gotoolkit v0.0.0-20170318115440-bcc06269efa9 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/cheggaaa/pb/v3 v3.0.1
	github.com/cloudfoundry-attic/jibber_jabber v0.0.0-20151120183258-bcc4c8345a21
	github.com/cloudfoundry/jibber_jabber v0.0.0-20151120183258-bcc4c8345a21 // indirect
	github.com/docker/docker v1.13.1
	github.com/docker/go-units v0.4.0
	github.com/docker/machine v0.7.1-0.20190718054102-a555e4f7a8f5 // version is 0.7.1 to pin to a555e4f7a8f5
	github.com/elazarl/goproxy v0.0.0-20190421051319-9d40249d3c2f
	github.com/elazarl/goproxy/ext v0.0.0-20190421051319-9d40249d3c2f // indirect
	github.com/evanphx/json-patch v4.5.0+incompatible // indirect
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/golang-collections/collections v0.0.0-20130729185459-604e922904d3
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/google/go-cmp v0.3.0
	github.com/google/go-containerregistry v0.0.0-20180731221751-697ee0b3d46e
	github.com/googleapis/gnostic v0.3.0 // indirect
	github.com/gorilla/mux v1.7.3 // indirect
	github.com/hashicorp/errwrap v0.0.0-20141028054710-7554cd9344ce // indirect
	github.com/hashicorp/go-getter v1.4.0
	github.com/hashicorp/go-multierror v0.0.0-20160811015721-8c5f0ad93604 // indirect
	github.com/hashicorp/go-retryablehttp v0.5.4
	github.com/hooklift/assert v0.0.0-20170704181755-9d1defd6d214 // indirect
	github.com/hooklift/iso9660 v0.0.0-20170318115843-1cf07e5970d8
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/intel-go/cpuid v0.0.0-20181003105527-1a4a6f06a1c6 // indirect
	github.com/jimmidyson/go-download v0.0.0-20161028105827-7f9a90c8c95b
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
	github.com/mitchellh/go-ps v0.0.0-20170309133038-4fdf99ab2936
	github.com/moby/hyperkit v0.0.0-20171020124204-a12cd7250bcd
	github.com/olekukonko/tablewriter v0.0.0-20170122224234-a0225b3f23b5
	github.com/otiai10/copy v1.0.2
	github.com/pborman/uuid v1.2.0
	github.com/phayes/freeport v0.0.0-20180830031419-95f893ade6f2
	github.com/pkg/browser v0.0.0-20160118053552-9302be274faa
	github.com/pkg/errors v0.8.1
	github.com/pkg/profile v0.0.0-20161223203901-3a8809bd8a80
	github.com/pmezard/go-difflib v1.0.0
	github.com/samalba/dockerclient v0.0.0-20160414174713-91d7393ff859 // indirect
	github.com/shirou/gopsutil v2.18.12+incompatible
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.3.2
	github.com/xeipuuv/gojsonpointer v0.0.0-20151027082146-e0fe6f683076 // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20150808065054-e02fc20de94c // indirect
	github.com/xeipuuv/gojsonschema v0.0.0-20160623135812-c539bca196be
	github.com/zchee/go-vmnet v0.0.0-20161021174912-97ebf9174097
	golang.org/x/build v0.0.0-20190927031335-2835ba2e683f
	golang.org/x/crypto v0.0.0-20190820162420-60c769a6c586
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	golang.org/x/sys v0.0.0-20190826190057-c7b8b68b1456
	golang.org/x/text v0.3.2
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22 // indirect
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/kubectl v0.0.0
	k8s.io/kubernetes v1.17.2
	k8s.io/utils v0.0.0-20200122174043-1e243dd1a584 // indirect
	sigs.k8s.io/sig-storage-lib-external-provisioner v4.0.0+incompatible
)

replace (
	git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999
	github.com/docker/machine => github.com/machine-drivers/machine v0.7.1-0.20191109154235-b39d5b50de51
	github.com/hashicorp/go-getter => github.com/afbjorklund/go-getter v1.4.1-0.20190910175809-eb9f6c26742c
	k8s.io/api => k8s.io/api v0.17.2
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.2
	k8s.io/apiserver => k8s.io/apiserver v0.17.2
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.17.2
	k8s.io/client-go => k8s.io/client-go v0.17.2
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.17.2
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.17.2
	k8s.io/code-generator => k8s.io/code-generator v0.17.2
	k8s.io/component-base => k8s.io/component-base v0.17.2
	k8s.io/cri-api => k8s.io/cri-api v0.17.2
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.17.2
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.17.2
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.17.2
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.17.2
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.17.2
	k8s.io/kubectl => k8s.io/kubectl v0.17.2
	k8s.io/kubelet => k8s.io/kubelet v0.17.2
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.17.2
	k8s.io/metrics => k8s.io/metrics v0.17.2
	k8s.io/node-api => k8s.io/node-api v0.17.2
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.17.2
	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.17.2
	k8s.io/sample-controller => k8s.io/sample-controller v0.17.2
)
