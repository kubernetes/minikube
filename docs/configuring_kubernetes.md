## Configuring Kubernetes

Minikube has a "configurator" feature that allows users to configure the Kubernetes components with arbitrary values.
To use this feature, you can use the `--extra-config` flag on the `minikube start` command.

This flag is repeated, so you can pass it several times with several different values to set multiple options.

### Kubeadm bootstrapper

The kubeadm bootstrapper can be configured by the `--extra-config` flag on the `minikube start` command.  It takes a string of the form `component.key=value` where `component` is one of the strings

* kubelet
* apiserver
* controller-manager
* scheduler

and `key=value` is a flag=value pair for the component being configured.  For example,

```
minikube start --extra-config=apiserver.v=10 --extra-config=kubelet.max-pods=100
```

### Localkube

The configurator interpretes the `--extra-config` flags differently for localkube.

This flag takes a string of the form `component.key=value`, where `component` is one of the strings from the list below, `key` is a value on the
configuration struct and `value` is the value to set.

Valid `key`s can be found by examining the documentation for the Kubernetes `componentconfigs` for each component.
Here is the documentation for each supported configuration:

* [kubelet](https://godoc.org/k8s.io/kubernetes/pkg/apis/componentconfig#KubeletConfiguration)
* [apiserver](https://godoc.org/k8s.io/kubernetes/cmd/kube-apiserver/app/options#APIServer)
* [proxy](https://godoc.org/k8s.io/kubernetes/pkg/apis/componentconfig#KubeProxyConfiguration)
* [controller-manager](https://godoc.org/k8s.io/kubernetes/pkg/apis/componentconfig#KubeControllerManagerConfiguration)
* [etcd](https://godoc.org/github.com/coreos/etcd/etcdserver#ServerConfig)
* [scheduler](https://godoc.org/k8s.io/kubernetes/pkg/apis/componentconfig#KubeSchedulerConfiguration)

You can enable feature gates for alpha and experimental features with the `--feature-gates` flag on `minikube start`.  As of v1.5.1, the options are:

* AllAlpha=true|false (ALPHA - default=false)
* AllowExtTrafficLocalEndpoints=true|false (BETA - default=true)
* AppArmor=true|false (BETA - default=true)
* DynamicKubeletConfig=true|false (ALPHA - default=false)
* DynamicVolumeProvisioning=true|false (ALPHA - default=true)
* ExperimentalHostUserNamespaceDefaulting=true|false (ALPHA - default=false)
* StreamingProxyRedirects=true|false (ALPHA - default=false)

Note: All alpha and experimental features are not guaranteed to work with minikube.

#### Examples

To change the `MaxPods` setting to 5 on the Kubelet, pass this flag: `--extra-config=kubelet.MaxPods=5`.

This feature also supports nested structs. To change the `LeaderElection.LeaderElect` setting to `true` on the scheduler, pass this flag: `--extra-config=scheduler.LeaderElection.LeaderElect=true`.

To set the `AuthorizationMode` on the `apiserver` to `RBAC`, you can use: `--extra-config=apiserver.Authorization.Mode=RBAC`.

To enable all alpha feature gates, you can use: `--feature-gates=AllAlpha=true`
