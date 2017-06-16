#### Updating Kubernetes

To update Kubernetes, follow these steps:

1. Make a clean GOPATH, with minikube in it.
This isn't strictly necessary, but it usually helps.

 ```shell
 mkdir -p $HOME/newgopath/src/k8s.io
 export GOPATH=$HOME/newgopath
 cd $HOME/newgopath/src/k8s.io
 git clone https://github.com/kubernetes/minikube.git
 ```

2. Copy your vendor directory back out to the new GOPATH.

 ```shell
 cd minikube
 ./hack/godeps/godep-restore.sh
 ```

3. Kubernetes should now be on your GOPATH. Check it out to the right version.
Make sure to also fetch tags, as Godep relies on these.

 ```shell
 cd $GOPATH/src/k8s.io/kubernetes
 git fetch --tags
 ```

 Then list all available Kubernetes tags:

 ```shell
 git tag
 ...
 v1.2.4
 v1.2.4-beta.0
 v1.3.0-alpha.3
 v1.3.0-alpha.4
 v1.3.0-alpha.5
 ...
```

 Then checkout the correct one and update its dependencies with:

 ```shell
 git checkout $DESIREDTAG
 ./hack/godeps/godep-restore.sh
 ```

4. Build and test minikube, making any manual changes necessary to build.

5. Update godeps

 ```shell
 cd $GOPATH/src/k8s.io/minikube
 ./hack/godeps/godep-save.sh
 ```

 6. Verify that the correct tag is marked in the Godeps.json file by running this script:

 ```shell
 python hack/get_k8s_version.py
 -X k8s.io/minikube/vendor/k8s.io/kubernetes/pkg/version.gitCommit=caf9a4d87700ba034a7b39cced19bd5628ca6aa3 -X k8s.io/minikube/vendor/k8s.io/kubernetes/pkg/version.gitVersion=v1.3.0-beta.2 -X k8s.io/minikube/vendor/k8s.io/kubernetes/pkg/version.gitTreeState=clean
```

The `-X k8s.io/minikube/vendor/k8s.io/kubernetes/pkg/version.gitVersion` flag should contain the right tag.

Once you've build and started minikube, you can also run:

```shell
kubectl version
Client Version: version.Info{Major:"1", Minor:"2", GitVersion:"v1.2.4", GitCommit:"3eed1e3be6848b877ff80a93da3785d9034d0a4f", GitTreeState:"clean"}
Server Version: version.Info{Major:"1", Minor:"3+", GitVersion:"v1.3.0-beta.2", GitCommit:"caf9a4d87700ba034a7b39cced19bd5628ca6aa3", GitTreeState:"clean"}
```

The Server Version should contain the right tag in `version.Info.GitVersion`.

If any manual changes were required, please commit the vendor changes separately.
This makes the change easier to view in GitHub.

```shell
git add vendor/
git commit -m "Updating Kubernetes to foo"
git add --all
git commit -m "Manual changes to update Kubernetes to foo"
```

As a final part of updating Kubernetes, a new version of localkube should be uploaded to GCS so that users can select this version of Kubernetes/localkube in later minikube/localkube builds. For instructions on how to do this, see [releasing_localkube.md](https://github.com/kubernetes/minikube/blob/master/docs/contributors/releasing_localkube.md)
