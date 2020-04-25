module k8s.io/minikube

go 1.13

require (
	cloud.google.com/go v0.45.1
	github.com/Microsoft/go-winio v0.4.15-0.20190919025122-fc70bd9a86b5 // indirect
	github.com/Parallels/docker-machine-parallels v1.3.0
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/blang/semver v3.5.0+incompatible
	github.com/c4milo/gotoolkit v0.0.0-20170318115440-bcc06269efa9 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/cheggaaa/pb/v3 v3.0.1
	github.com/cloudfoundry-attic/jibber_jabber v0.0.0-20151120183258-bcc4c8345a21
	github.com/cloudfoundry/jibber_jabber v0.0.0-20151120183258-bcc4c8345a21 // indirect
	github.com/containerd/containerd v1.3.1-0.20191213020239-082f7e3aed57 // indirect
	github.com/docker/cli v0.0.0-20200303162255-7d407207c304 // indirect
	github.com/docker/docker v1.13.1
	github.com/docker/go-units v0.4.0
	github.com/docker/machine v0.7.1-0.20190902101342-b170508bf44c // v0.16.2^
	github.com/elazarl/goproxy v0.0.0-20190421051319-9d40249d3c2f
	github.com/elazarl/goproxy/ext v0.0.0-20190421051319-9d40249d3c2f // indirect
	github.com/evanphx/json-patch v4.5.0+incompatible // indirect
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang-collections/collections v0.0.0-20130729185459-604e922904d3
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/google/go-cmp v0.3.2-0.20191028172631-481baca67f93
	github.com/google/go-containerregistry v0.0.0-20200131185320-aec8da010de2
	github.com/google/go-github v17.0.0+incompatible
	github.com/google/martian v2.1.1-0.20190517191504-25dcb96d9e51+incompatible // indirect
	github.com/googleapis/gnostic v0.3.0 // indirect
	github.com/hashicorp/go-getter v1.4.0
	github.com/hashicorp/go-retryablehttp v0.5.4
	github.com/hooklift/assert v0.0.0-20170704181755-9d1defd6d214 // indirect
	github.com/hooklift/iso9660 v0.0.0-20170318115843-1cf07e5970d8
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/intel-go/cpuid v0.0.0-20181003105527-1a4a6f06a1c6 // indirect
	github.com/johanneswuerbach/nfsexports v0.0.0-20200318065542-c48c3734757f
	github.com/juju/clock v0.0.0-20190205081909-9c5c9712527c
	github.com/juju/errors v0.0.0-20190806202954-0232dcc7464d // indirect
	github.com/juju/loggo v0.0.0-20190526231331-6e530bcce5d8 // indirect
	github.com/juju/mutex v0.0.0-20180619145857-d21b13acf4bf
	github.com/juju/retry v0.0.0-20180821225755-9058e192b216 // indirect
	github.com/juju/testing v0.0.0-20190723135506-ce30eb24acd2 // indirect
	github.com/juju/utils v0.0.0-20180820210520-bf9cc5bdd62d // indirect
	github.com/juju/version v0.0.0-20180108022336-b64dbd566305 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51
	github.com/kr/pretty v0.2.0 // indirect
	github.com/libvirt/libvirt-go v3.4.0+incompatible
	github.com/machine-drivers/docker-machine-driver-vmware v0.1.1
	github.com/mattn/go-isatty v0.0.11
	github.com/mitchellh/go-ps v0.0.0-20170309133038-4fdf99ab2936
	github.com/moby/hyperkit v0.0.0-20171020124204-a12cd7250bcd
	github.com/olekukonko/tablewriter v0.0.4
	github.com/onsi/ginkgo v1.10.3 // indirect
	github.com/onsi/gomega v1.7.1 // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1
	github.com/otiai10/copy v1.0.2
	github.com/pborman/uuid v1.2.0
	github.com/pelletier/go-toml v1.6.0 // indirect
	github.com/phayes/freeport v0.0.0-20180830031419-95f893ade6f2
	github.com/pkg/browser v0.0.0-20160118053552-9302be274faa
	github.com/pkg/errors v0.9.1
	github.com/pkg/profile v0.0.0-20161223203901-3a8809bd8a80
	github.com/pmezard/go-difflib v1.0.0
	github.com/prometheus/client_golang v1.1.0 // indirect
	github.com/prometheus/procfs v0.0.5 // indirect
	github.com/russross/blackfriday v1.5.3-0.20200218234912-41c5fccfd6f6 // indirect
	github.com/samalba/dockerclient v0.0.0-20160414174713-91d7393ff859 // indirect
	github.com/shirou/gopsutil v2.18.12+incompatible
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.6.1
	github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v0.0.0-20180618132009-1d523034197f
	github.com/zchee/go-vmnet v0.0.0-20161021174912-97ebf9174097
	golang.org/x/build v0.0.0-20190927031335-2835ba2e683f
	golang.org/x/crypto v0.0.0-20200302210943-78000ba7a073
	golang.org/x/net v0.0.0-20200301022130-244492dfa37a // indirect
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	golang.org/x/sys v0.0.0-20200124204421-9fbb57f87de9
	golang.org/x/text v0.3.2
	google.golang.org/api v0.9.0
	google.golang.org/genproto v0.0.0-20200117163144-32f20d992d24 // indirect
	google.golang.org/grpc v1.26.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/ini.v1 v1.51.1 // indirect
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22 // indirect
	gopkg.in/yaml.v2 v2.2.8
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
	github.com/docker/machine => github.com/machine-drivers/machine v0.7.1-0.20200323212942-41eb826190d8
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
