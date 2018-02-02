## Networking

The minikube VM is exposed to the host system via a host-only IP address, that can be obtained with the `minikube ip` command.
Any services of type `NodePort` can be accessed over that IP address, on the NodePort.

To determine the NodePort for your service, you can use a `kubectl` command like this:

`kubectl get service $SERVICE --output='jsonpath="{.spec.ports[0].NodePort}"'`

We also have a shortcut for fetching the minikube IP and a service's `NodePort`:

`minikube service --url $SERVICE`

The same works for accessing addons:

`minikube addons open --url $ADDON`

As a shortcut, the dashboard addon also has its own command:

`minikube dashboard --url`

All three commands will open the URL automatically in the default web
browser when leaving out the `--url` parameter. For this to work, the
web browser must run on the same host as the minikube VM.

Minikube supports port forwarding to make the minikube VM ports
available outside of the local host. When requested via the
``--proxy`` parameter, minikube changes the URL so
that the local host is specified in the URL and acts as proxy:

`minikube dashboard --proxy --url`

Minikube keeps running and forwarding connections until
interupted. The default is to listen on all interfaces, pick a random
port and use the local host name in the modified URL. When a specific
port is desired and/or a specific interface should be used, then
``--proxyaddress`` can be used:

```
--proxyaddress :8080
--proxyaddress 192.168.1.1:0
--proxyaddress my-host:8080
```

Note that the host name in ``--proxyaddress`` will be resolved on the
local host before listening on the port, so the resulting URL will
have a fixed IP address. This is also useful when the default, the
local host name, cannot be resolved outside of the local host.

A fixed port in `--proxyaddress` only works when opening exactly one
URL. `minikube service` and `minikube addon open` can potentially open
multiple URLs, in which case the port has to be chosen automatically.
