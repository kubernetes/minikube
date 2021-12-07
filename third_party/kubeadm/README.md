maintaining a fork of kubernetes constants to avoid depending on k8s.io/kubernetes/ as a lib
how to update this fork

# clone kuberentes

# pick latest stable version
$ git reset --hard v1.22.4

# cd into minikube

$ mkdir -p ./third_party/kubeadm/app/
$ cp -r ../kubernetes/cmd/kubeadm/app/features ./third_party/kubeadm/app/
$ cp -r ../kubernetes/cmd/kubeadm/app/constants ./third_party/kubeadm/app/