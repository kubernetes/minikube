TODO: This document was manually maintained so might be incomplete. The
automation effort is tracked in
https://github.com/kubernetes/client-go/issues/234.

# HEAD (changes that will go into v5)

** Breaking changes**

pkg/api and pkg/apis are moved to
[k8s.io/api](https://github.com/kubernetes/api). Other kubernetes repositories
also import types from there, so they are composable with client-go.

Helper functions in pkg/api and pkg/apis are also removed. They are planned to
be exported in other repos. The issue is tracked
[here](https://github.com/kubernetes/kubernetes/issues/48209#issuecomment-314537745).
During the transition, you'll have to copy the helper functions to your projects. 

# v4.0.0

No significant changes since v4.0.0-beta.0.

# v4.0.0-beta.0

**New features:**

* Added OpenAPISchema support in the discovery client

    * [https://github.com/kubernetes/kubernetes/pull/44531](https://github.com/kubernetes/kubernetes/pull/44531)

* Added mutation cache filter: MutationCache is able to take the result of update operations and stores them in an LRU that can be used to provide a more current view of a requested object.

    * [https://github.com/kubernetes/kubernetes/pull/45838](https://github.com/kubernetes/kubernetes/pull/45838/commits/f88c7725b4f9446c652d160bdcfab7c6201bddea)

* Moved the remotecommand package (used by `kubectl exec/attach`) to client-go

    * [https://github.com/kubernetes/kubernetes/pull/41331](https://github.com/kubernetes/kubernetes/pull/41331)

* Added support for following redirects to the SpdyRoundTripper

    * [https://github.com/kubernetes/kubernetes/pull/44451](https://github.com/kubernetes/kubernetes/pull/44451)

* Added Azure Active Directory plugin

    * [https://github.com/kubernetes/kubernetes/pull/43987](https://github.com/kubernetes/kubernetes/pull/43987)

**Usability improvements:**

* Added several new examples and reorganized client-go/examples

    * [Related PRs](https://github.com/kubernetes/kubernetes/commits/release-1.7/staging/src/k8s.io/client-go/examples)

**API changes:**

* Added networking.k8s.io/v1 API

    * [https://github.com/kubernetes/kubernetes/pull/39164](https://github.com/kubernetes/kubernetes/pull/39164)

* ControllerRevision type added for StatefulSet and DaemonSet history.

    * [https://github.com/kubernetes/kubernetes/pull/45867](https://github.com/kubernetes/kubernetes/pull/45867)

* Added support for initializers

    * [https://github.com/kubernetes/kubernetes/pull/38058](https://github.com/kubernetes/kubernetes/pull/38058)

* Added admissionregistration.k8s.io/v1alpha1 API

    * [https://github.com/kubernetes/kubernetes/pull/46294](https://github.com/kubernetes/kubernetes/pull/46294)

**Breaking changes:**

* Moved client-go/util/clock to apimachinery/pkg/util/clock 

    * [https://github.com/kubernetes/kubernetes/pull/45933](https://github.com/kubernetes/kubernetes/pull/45933/commits/8013212db54e95050c622675c6706cce5de42b45)

* Some [API helpers](https://github.com/kubernetes/client-go/blob/release-3.0/pkg/api/helpers.go) were removed. 

* Dynamic client takes GetOptions as an input parameter

    * [https://github.com/kubernetes/kubernetes/pull/47251](https://github.com/kubernetes/kubernetes/pull/47251)

**Bug fixes:**

* PortForwarder: don't log an error if net.Listen fails. [https://github.com/kubernetes/kubernetes/pull/44636](https://github.com/kubernetes/kubernetes/pull/44636)

* oidc auth plugin not to override the Auth header if it's already exits. [https://github.com/kubernetes/kubernetes/pull/45529](https://github.com/kubernetes/kubernetes/pull/45529)

* The --namespace flag is now honored for in-cluster clients that have an empty configuration. [https://github.com/kubernetes/kubernetes/pull/46299](https://github.com/kubernetes/kubernetes/pull/46299)

* GCP auth plugin no longer overwrites existing Authorization headers. [https://github.com/kubernetes/kubernetes/pull/45575](https://github.com/kubernetes/kubernetes/pull/45575)

# v3.0.0

Bug fixes:
* Use OS-specific libs when computing client User-Agent in kubectl, etc. (https://github.com/kubernetes/kubernetes/pull/44423)
* kubectl commands run inside a pod using a kubeconfig file now use the namespace specified in the kubeconfig file, instead of using the pod namespace. If no kubeconfig file is used, or the kubeconfig does not specify a namespace, the pod namespace is still used as a fallback. (https://github.com/kubernetes/kubernetes/pull/44570)
* Restored the ability of kubectl running inside a pod to consume resource files specifying a different namespace than the one the pod is running in. (https://github.com/kubernetes/kubernetes/pull/44862)

# v3.0.0-beta.0

* Added dependency on k8s.io/apimachinery. The impacts include changing import path of API objects like `ListOptions` from `k8s.io/client-go/pkg/api/v1` to `k8s.io/apimachinery/pkg/apis/meta/v1`.
* Added generated listers (listers/) and informers (informers/)
* Kubernetes API changes:
  * Added client support for:
    * authentication/v1
    * authorization/v1
    * autoscaling/v2alpha1
    * rbac/v1beta1
    * settings/v1alpha1
    * storage/v1
  * Changed client support for:
    * certificates from v1alpha1 to v1beta1
    * policy from v1alpha1 to v1beta1
  * Deleted client support for:
    * extensions/v1beta1#Job
* CHANGED: pass typed options to dynamic client (https://github.com/kubernetes/kubernetes/pull/41887)

# v2.0.0

* Included bug fixes in k8s.io/kuberentes release-1.5 branch, up to commit 
  bde8578d9675129b7a2aa08f1b825ec6cc0f3420

# v2.0.0-alpha.1

* Removed top-level version folder (e.g., 1.4 and 1.5), switching to maintaining separate versions
  in separate branches.
* Clientset supported multiple versions per API group
* Added ThirdPartyResources example
* Kubernetes API changes
  * Apps API group graduated to v1beta1 
  * Policy API group graduated to v1beta1
  * Added support for batch/v2alpha1/cronjob
  * Renamed PetSet to StatefulSet
  

# v1.5.0

* Included the auth plugin (https://github.com/kubernetes/kubernetes/pull/33334)
* Added timeout field to RESTClient config (https://github.com/kubernetes/kubernetes/pull/33958)
