package kubeadm

import (
	"bytes"
	"crypto"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/docker/machine/libmachine"
	"github.com/golang/glog"
	download "github.com/jimmidyson/go-download"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"golang.org/x/sync/errgroup"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/sshutil"
	"k8s.io/minikube/pkg/util"
)

type KubeadmBootstrapper struct {
	c      *ssh.Client
	driver string // TODO(r2d4): get rid of this dependency
}

// TODO(r2d4): template this with bootstrapper.KubernetesConfig
const kubeletSystemdConf = `
[Service]
Environment="KUBELET_KUBECONFIG_ARGS=--kubeconfig=/etc/kubernetes/kubelet.conf --require-kubeconfig=true"
Environment="KUBELET_SYSTEM_PODS_ARGS=--pod-manifest-path=/etc/kubernetes/manifests --allow-privileged=true"
Environment="KUBELET_DNS_ARGS=--cluster-dns=10.0.0.10 --cluster-domain=cluster.local"
Environment="KUBELET_CADVISOR_ARGS=--cadvisor-port=0"
Environment="KUBELET_CGROUP_ARGS=--cgroup-driver=cgroupfs"
ExecStart=
ExecStart=/usr/bin/kubelet $KUBELET_KUBECONFIG_ARGS $KUBELET_SYSTEM_PODS_ARGS $KUBELET_DNS_ARGS $KUBELET_CADVISOR_ARGS $KUBELET_CGROUP_ARGS $KUBELET_EXTRA_ARGS
`

const kubeletService = `
[Unit]
Description=kubelet: The Kubernetes Node Agent
Documentation=http://kubernetes.io/docs/

[Service]
ExecStart=/usr/bin/kubelet
Restart=always
StartLimitInterval=0
RestartSec=10

[Install]
WantedBy=multi-user.target
`

func NewKubeadmBootstrapper(api libmachine.API) (*KubeadmBootstrapper, error) {
	h, err := api.Load(config.GetMachineName())
	if err != nil {
		return nil, errors.Wrap(err, "getting api client")
	}
	client, err := sshutil.NewSSHClient(h.Driver)
	if err != nil {
		return nil, errors.Wrap(err, "getting ssh client")
	}
	return &KubeadmBootstrapper{
		c:      client,
		driver: h.Driver.DriverName(),
	}, nil
}

func (k *KubeadmBootstrapper) GetClusterStatus() (string, error) {
	return "", nil
}

func (k *KubeadmBootstrapper) GetClusterLogs(follow bool) (string, error) {
	return "", nil
}

func (k *KubeadmBootstrapper) StartCluster(k8s bootstrapper.KubernetesConfig) error {
	kubeadmTmpl := "sudo /usr/bin/kubeadm init"
	kubeadmTmpl += " --cert-dir {{.CertDir}}"
	kubeadmTmpl += " --service-cidr {{.ServiceCIDR}}"
	kubeadmTmpl += " --apiserver-advertise-address {{.AdvertiseAddress}}"
	kubeadmTmpl += " --apiserver-bind-port {{.APIServerPort}}"
	t := template.Must(template.New("kubeadmTmpl").Parse(kubeadmTmpl))

	opts := struct {
		CertDir          string
		ServiceCIDR      string
		AdvertiseAddress string
		APIServerPort    int
	}{
		CertDir:          util.DefaultCertPath,
		ServiceCIDR:      util.DefaultServiceCIDR,
		AdvertiseAddress: k8s.NodeIP,
		APIServerPort:    util.APIServerPort,
	}

	b := bytes.Buffer{}
	if err := t.Execute(&b, opts); err != nil {
		return err
	}

	_, err := sshutil.RunCommandOutput(k.c, b.String())
	if err != nil {
		return err
	}

	return nil
}

func (k *KubeadmBootstrapper) RestartCluster(k8s bootstrapper.KubernetesConfig) error {
	tmpFile := "/tmp/cert.conf"

	restartTmpl := "sudo /usr/bin/kubeadm alpha phase kubeconfig client-certs"
	restartTmpl += " --cert-dir {{.CertDir}}"
	restartTmpl += " --server {{.IP}}"
	restartTmpl += " --client-name {{.MachineName}}"

	// Output to temp file, since we will have to write this file to a few places.
	restartTmpl += " > " + tmpFile
	t := template.Must(template.New("restartTmpl").Parse(restartTmpl))

	opts := struct {
		CertDir     string
		IP          string
		MachineName string
	}{
		CertDir:     util.DefaultCertPath,
		IP:          k8s.NodeIP,
		MachineName: k8s.NodeName,
	}

	b := bytes.Buffer{}
	if err := t.Execute(&b, opts); err != nil {
		return err
	}

	glog.Infoln("running cmd: %s", b.String())
	_, err := sshutil.RunCommandOutput(k.c, b.String())
	if err != nil {
		return errors.Wrapf(err, "regenerating kubeadm certs, running cmd: %s", b.String())
	}

	dsts := []string{"admin.conf", "controller-manager.conf", "kubelet.conf", "scheduler.conf"}
	for _, dst := range dsts {
		cmd := fmt.Sprintf("sudo cp %s %s", tmpFile, filepath.Join("/etc", "kubernetes", dst))
		_, err := sshutil.RunCommandOutput(k.c, cmd)
		if err != nil {
			return errors.Wrapf(err, "copying conf files, running cmd: %s", cmd)
		}
	}

	return nil
}

func (k *KubeadmBootstrapper) UpdateCluster(cfg bootstrapper.KubernetesConfig) error {
	files := []assets.CopyableFile{
		assets.NewMemoryAsset([]byte(kubeletService), "/lib/systemd/system", "kubelet.service", "0640"),
		assets.NewMemoryAsset([]byte(kubeletSystemdConf), "/etc/systemd/system/kubelet.service.d/", "10-kubeadm.conf", "0640"),
	}

	for _, f := range files {
		if err := sshutil.TransferFile(f, k.c); err != nil {
			return errors.Wrapf(err, "transferring kubeadm file: %+v", f)
		}
	}
	var g errgroup.Group
	for _, bin := range []string{"kubelet", "kubeadm"} {
		bin := bin
		g.Go(func() error {
			path, err := maybeDownloadAndCache(bin, cfg.KubernetesVersion)
			if err != nil {
				return errors.Wrapf(err, "downloading %s", bin)
			}
			f, err := assets.NewFileAsset(path, "/usr/bin", bin, "0641")
			if err != nil {
				return errors.Wrap(err, "making new file asset")
			}
			if err := sshutil.TransferFile(f, k.c); err != nil {
				return errors.Wrapf(err, "transferring kubeadm file: %+v", f)
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return errors.Wrap(err, "downloading binaries")
	}

	_, err := sshutil.RunCommandOutput(k.c, `
sudo systemctl daemon-reload &&
sudo systemctl enable kubelet &&
sudo systemctl start kubelet
`)
	if err != nil {
		return errors.Wrap(err, "starting kubelet")
	}

	return nil
}

func maybeDownloadAndCache(binary, version string) (string, error) {
	targetDir := constants.MakeMiniPath("cache", version)
	targetFilepath := filepath.Join(targetDir, binary)

	_, err := os.Stat(targetDir)
	// If it exists, do no verification and continue
	if err == nil {
		return targetFilepath, nil
	}
	if !os.IsNotExist(err) {
		return "", errors.Wrapf(err, "stat %s version %s at %s", binary, version, targetDir)
	}

	if err = os.MkdirAll(targetDir, 0777); err != nil {
		return "", errors.Wrapf(err, "mkdir %s", targetDir)
	}

	url := constants.GetKubernetesReleaseURL(binary, version)
	options := download.FileOptions{
		Mkdirs: download.MkdirAll,
	}

	options.Checksum = constants.GetKubernetesReleaseURLSha1(binary, version)
	options.ChecksumHash = crypto.SHA1

	fmt.Printf("Downloading %s %s\n", binary, version)
	if err := download.ToFile(url, targetFilepath, options); err != nil {
		return "", errors.Wrapf(err, "Error downloading %s %s", binary, version)
	}
	fmt.Printf("Finished Downloading %s %s\n", binary, version)

	return targetFilepath, nil
}
