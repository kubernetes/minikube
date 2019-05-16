module k8s.io/minikube

go 1.12

require (
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/Parallels/docker-machine-parallels v1.3.0
	github.com/Sirupsen/logrus v0.0.0-20170822132746-89742aefa4b2 // indirect
	github.com/blang/semver v3.5.0+incompatible
	github.com/c4milo/gotoolkit v0.0.0-20170318115440-bcc06269efa9 // indirect
	github.com/cloudfoundry-attic/jibber_jabber v0.0.0-20151120183258-bcc4c8345a21
	github.com/cloudfoundry/jibber_jabber v0.0.0-20151120183258-bcc4c8345a21 // indirect
	github.com/cpuguy83/go-md2man v1.0.4 // indirect
	github.com/docker/docker v1.13.1 // indirect
	github.com/docker/go-units v0.0.0-20170127094116-9e638d38cf69
	github.com/docker/machine v0.16.1
	github.com/elazarl/goproxy v0.0.0-20190421051319-9d40249d3c2f
	github.com/elazarl/goproxy/ext v0.0.0-20190421051319-9d40249d3c2f // indirect
	github.com/fatih/color v1.7.0 // indirect
	github.com/ghodss/yaml v0.0.0-20150909031657-73d445a93680 // indirect
	github.com/gogo/protobuf v0.0.0-20170330071051-c0656edd0d9e // indirect
	github.com/golang/glog v0.0.0-20141105023935-44145f04b68c
	github.com/golang/groupcache v0.0.0-20160516000752-02826c3e7903 // indirect
	github.com/google/btree v1.0.0 // indirect
	github.com/google/go-cmp v0.2.0
	github.com/google/go-containerregistry v0.0.0-20190318164241-019cdfc6adf9
	github.com/google/go-github/v25 v25.0.2
	github.com/google/gofuzz v0.0.0-20161122191042-44d81051d367 // indirect
	github.com/googleapis/gnostic v0.0.0-20170729233727-0c5108395e2d // indirect
	github.com/gorilla/mux v1.7.1 // indirect
	github.com/gregjones/httpcache v0.0.0-20180305231024-9cad4c3443a7 // indirect
	github.com/hashicorp/errwrap v0.0.0-20141028054710-7554cd9344ce // indirect
	github.com/hashicorp/go-multierror v0.0.0-20160811015721-8c5f0ad93604 // indirect
	github.com/hashicorp/go-version v1.1.0 // indirect
	github.com/hashicorp/golang-lru v0.0.0-20160207214719-a0d98a5f2880 // indirect
	github.com/hashicorp/hcl v0.0.0-20160711231752-d8c773c4cba1 // indirect
	github.com/hooklift/assert v0.0.0-20170704181755-9d1defd6d214 // indirect
	github.com/hooklift/iso9660 v0.0.0-20170318115843-1cf07e5970d8
	github.com/imdario/mergo v0.0.0-20141206190957-6633656539c1 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/intel-go/cpuid v0.0.0-20181003105527-1a4a6f06a1c6 // indirect
	github.com/jimmidyson/go-download v0.0.0-20161028105827-7f9a90c8c95b
	github.com/johanneswuerbach/nfsexports v0.0.0-20181204082207-1aa528dcb345
	github.com/json-iterator/go v1.1.5 // indirect
	github.com/kr/fs v0.0.0-20131111012553-2788f0dbd169 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/libvirt/libvirt-go v3.4.0+incompatible
	github.com/machine-drivers/docker-machine-driver-vmware v0.1.1
	github.com/magiconair/properties v0.0.0-20160816085511-61b492c03cf4 // indirect
	github.com/mattn/go-colorable v0.1.1 // indirect
	github.com/mattn/go-isatty v0.0.5
	github.com/mattn/go-runewidth v0.0.0-20161012013512-737072b4e32b // indirect
	github.com/mitchellh/go-ps v0.0.0-20170309133038-4fdf99ab2936
	github.com/mitchellh/mapstructure v0.0.0-20170307201123-53818660ed49 // indirect
	github.com/moby/hyperkit v0.0.0-20171020124204-a12cd7250bcd
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v0.0.0-20180701023420-4b7aa43c6742 // indirect
	github.com/olekukonko/tablewriter v0.0.0-20160923125401-bdcc175572fd
	github.com/onsi/ginkgo v1.8.0 // indirect
	github.com/onsi/gomega v1.5.0 // indirect
	github.com/pborman/uuid v0.0.0-20150603214016-ca53cad383ca
	github.com/pelletier/go-buffruneio v0.1.0 // indirect
	github.com/pelletier/go-toml v0.0.0-20160822122712-0049ab3dc4c4 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/phayes/freeport v0.0.0-20180830031419-95f893ade6f2
	github.com/pkg/browser v0.0.0-20160118053552-9302be274faa
	github.com/pkg/errors v0.8.0
	github.com/pkg/profile v0.0.0-20161223203901-3a8809bd8a80
	github.com/pkg/sftp v0.0.0-20160930220758-4d0e916071f6 // indirect
	github.com/pmezard/go-difflib v1.0.0
	github.com/r2d4/external-storage v0.0.0-20171222174501-8c0e8605dc7b
	github.com/russross/blackfriday v0.0.0-20151117072312-300106c228d5 // indirect
	github.com/samalba/dockerclient v0.0.0-20160414174713-91d7393ff859 // indirect
	github.com/shurcooL/sanitized_anchor_name v0.0.0-20151028001915-10ef21a441db // indirect
	github.com/sirupsen/logrus v1.4.1
	github.com/spf13/afero v0.0.0-20160816080757-b28a7effac97 // indirect
	github.com/spf13/cast v0.0.0-20160730092037-e31f36ffc91a // indirect
	github.com/spf13/cobra v0.0.0-20180228053838-6644d46b81fa
	github.com/spf13/jwalterweatherman v0.0.0-20160311093646-33c24e77fb80 // indirect
	github.com/spf13/pflag v1.0.1
	github.com/spf13/viper v1.0.0
	github.com/stretchr/testify v1.3.0 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20151027082146-e0fe6f683076 // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20150808065054-e02fc20de94c // indirect
	github.com/xeipuuv/gojsonschema v0.0.0-20160623135812-c539bca196be
	github.com/zchee/go-vmnet v0.0.0-20161021174912-97ebf9174097
	golang.org/x/crypto v0.0.0-20190308221718-c2843e01d9a2
	golang.org/x/oauth2 v0.0.0-20190115181402-5dab4167f31c
	golang.org/x/sync v0.0.0-20190227155943-e225da77a7e6
	golang.org/x/sys v0.0.0-20190222072716-a9d3bda3a223
	golang.org/x/text v0.3.2
	golang.org/x/time v0.0.0-20161028155119-f51c12702a4d // indirect
	gopkg.in/airbrake/gobrake.v2 v2.0.9 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/cheggaaa/pb.v1 v1.0.6 // indirect
	gopkg.in/gemnasium/logrus-airbrake-hook.v2 v2.1.2 // indirect
	gopkg.in/inf.v0 v0.9.0 // indirect
	gopkg.in/yaml.v1 v1.0.0-20140924161607-9f9df34309c0 // indirect
	k8s.io/api v0.0.0-20180712090710-2d6f90ab1293
	k8s.io/apimachinery v0.0.0-20180621070125-103fd098999d
	k8s.io/apiserver v0.0.0-20180914001516-67c892841170 // indirect
	k8s.io/client-go v0.0.0-20180806134042-1f13a808da65
	k8s.io/kube-openapi v0.0.0-20180216212618-50ae88d24ede // indirect
	k8s.io/kubernetes v1.11.3
)
