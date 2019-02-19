## Using Minikube with an HTTP Proxy

minikube requires access to the internet via HTTP, HTTPS, and DNS protocols. If a HTTP proxy is required to access the internet, you may need to pass the proxy connection information to both minikube and Docker using environment variables:

* HTTP_PROXY - The URL to your HTTP proxy
* HTTPS_PROXY - The URL to your HTTPS proxy
* NO_PROXY - A comma-separated list of hosts which should not go through the proxy.

The NO_PROXY variable here is important: Without setting it, minikube may not be able to access resources within the VM. minikube has two important ranges of internal IP's:

* 192.168.0.0/12: Used by the minikube VM. Configurable for some hypervisors via `--host-only-cidr`
* 10.96.0.0/12: Used by service cluster IP's. Configurable via  `--service-cluster-ip-range`

## Example Usage

### macOS and Linux

```
export HTTP_PROXY=http://<proxy hostname:port>
export HTTPS_PROXY=https://<proxy hostname:port>
export NO_PROXY=localhost,127.0.0.1,10.96.0.0/12,192.168.0.0/16

minikube start --docker-env=HTTP_PROXY=$HTTP_PROXY --docker-env HTTPS_PROXY=$HTTPS_PROXY
```

To make the exported variables permanent, consider adding the declarations to ~/.bashrc or wherever your user-set environment variables are stored.

### Windows

```
set HTTP_PROXY=http://<proxy hostname:port>
set HTTPS_PROXY=https://<proxy hostname:port>
set NO_PROXY=localhost,127.0.0.1,10.96.0.0/12,192.168.0.0/16

minikube start --docker-env=HTTP_PROXY=$HTTP_PROXY --docker-env HTTPS_PROXY=$HTTPS_PROXY --docker-env NO_PROXY=$NO_PROXY
```

To set these environment variables permanently, consider adding these to your [system settings](https://support.microsoft.com/en-au/help/310519/how-to-manage-environment-variables-in-windows-xp) or using [setx](https://stackoverflow.com/questions/5898131/set-a-persistent-environment-variable-from-cmd-exe)

## Additional Information

- [Configure Docker to use a proxy server](https://docs.docker.com/network/proxy/)
