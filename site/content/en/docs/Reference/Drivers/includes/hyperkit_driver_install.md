Download and install the latest minikube hyperkit driver:

```shell
curl -LO https://storage.googleapis.com/minikube/releases/latest/
docker-machine-driver-hyperkit \
&& sudo install -o root -m 4755 docker-machine-driver-hyperkit /usr/local/bin/
```