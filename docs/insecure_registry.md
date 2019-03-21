# Enabling Docker Insecure Registry

Minikube allows users to configure the docker engine's `--insecure-registry` flag. You can use the `--insecure-registry` flag on the
`minikube start` command to enable insecure communication between the docker engine and registries listening to requests from the CIDR range.

One nifty hack is to allow the kubelet running in minikube to talk to registries deployed inside a pod in the cluster without backing them
with TLS certificates. Because the default service cluster IP is known to be available at 10.0.0.1, users can pull images from registries
deployed inside the cluster by creating the cluster with `minikube start --insecure-registry "10.0.0.0/24"`.

## Private Container Registries

**GCR/ECR/Docker**: Minikube has an addon, `registry-creds` which maps credentials into Minikube to support pulling from Google Container Registry (GCR), Amazon's EC2 Container Registry (ECR), and Private Docker registries.  You will need to run `minikube addons configure registry-creds` and `minikube addons enable registry-creds` to get up and running.  An example of this is below:

```shell
$ minikube addons configure registry-creds
Do you want to enable AWS Elastic Container Registry? [y/n]: n

Do you want to enable Google Container Registry? [y/n]: y
-- Enter path to credentials (e.g. /home/user/.config/gcloud/application_default_credentials.json):/home/user/.config/gcloud/application_default_credentials.json

Do you want to enable Docker Registry? [y/n]: n
registry-creds was successfully configured
$ minikube addons enable registry-creds
```

For additional information on private container registries, see [this page](https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/).

We recommend you use _ImagePullSecrets_, but if you would like to configure access on the minikube VM you can place the `.dockercfg` in the `/home/docker` directory or the `config.json` in the `/var/lib/kubelet` directory. Make sure to restart your kubelet (for kubeadm) process with `sudo systemctl restart kubelet`.
