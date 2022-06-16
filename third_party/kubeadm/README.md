maintaining a fork of kubernetes constants to avoid depending on k8s.io/kubernetes/ as a lib
how to update this fork


# clone latest stable version

```
git clone --depth 1 --branch v1.22.4 git@github.com:kubernetes/kubernetes.git ./out/kubernetes
rm -rf ./third_party/kubeadm || true
mkdir -p ./third_party/kubeadm/app/
cp -r ./out/kubernetes/cmd/kubeadm/app/features ./third_party/kubeadm/app/
cp -r ./out/kubernetes/cmd/kubeadm/app/constants ./third_party/kubeadm/app/
rm ./third_party/kubeadm/app/features/*_test.go || true
rm ./third_party/kubeadm/app/constants/*_test.go || true
```