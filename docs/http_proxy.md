## Using Minikube with an HTTP Proxy

Minikube creates a Virtual Machine that includes Kubernetes and a Docker daemon.
When Kubernetes attempts to schedule containers using Docker, the Docker daemon may require external network access to pull containers.

If you are behind an HTTP proxy, you may need to supply Docker with the proxy settings.
To do this, pass the required environment variables as flags during `minikube start`.

For example:

```shell
$ minikube start --docker-env HTTP_PROXY=http://$YOURPROXY:PORT \
                 --docker-env HTTPS_PROXY=https://$YOURPROXY:PORT
```

If your Virtual Machine address is 192.168.99.100, then chances are your proxy settings will prevent kubectl from directly reaching it.
To by-pass proxy configuration for this IP address, you should modify your no_proxy settings. You can do so with:

```shell
$ export no_proxy=$no_proxy,$(minikube ip)
```
