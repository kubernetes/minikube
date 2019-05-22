# Networking

## Firewalls, VPN's, and proxies

minikube may require access from the host to the following IP ranges: 192.168.99.0/24, 192.168.39.0/24, and 10.96.0.0/12. These networks can be changed in minikube using `--host-only-cidr` and `--service-cluster-ip-range`.

* To use minikube with a proxy, see [Using HTTP/HTTPS proxies](http_proxy.md).

* If you are using minikube with a VPN, you may need to configure the VPN to allow local routing for  traffic to the afforementioned IP ranges.

* If you are using minikube with a local firewall, you will need to allow access from the host to the afforementioned IP ranges on TCP ports 22 and 8443. You will also need to add access from these IP's to TCP ports 443 and 53 externally to pull images.

## Access to NodePort services

The minikube VM is exposed to the host system via a host-only IP address, that can be obtained with the `minikube ip` command. Any services of type `NodePort` can be accessed over that IP address, on the NodePort.

To determine the NodePort for your service, you can use a `kubectl` command like this (note that `nodePort` begins with lowercase `n` in JSON output):

`kubectl get service $SERVICE --output='jsonpath="{.spec.ports[0].nodePort}"'`

We also have a shortcut for fetching the minikube IP and a service's `NodePort`:

`minikube service --url $SERVICE`

## Access to LoadBalancer services using `minikube tunnel`

Services of type `LoadBalancer` can be exposed via the `minikube tunnel` command.

````shell
minikube tunnel
````

Will output:

```text
out/minikube tunnel
Password: *****
Status:
        machine: minikube
        pid: 59088
        route: 10.96.0.0/12 -> 192.168.99.101
        minikube: Running
        services: []
    errors:
                minikube: no errors
                router: no errors
                loadbalancer emulator: no errors


````

Tunnel might ask you for password for creating and deleting network routes.

## Cleaning up orphaned routes

If the `minikube tunnel` shuts down in an unclean way, it might leave a network route around.
This case the ~/.minikube/tunnels.json file will contain an entry for that tunnel.
To cleanup orphaned routes, run:

````shell
minikube tunnel --cleanup
````

## Tunnel: Avoid entering password multiple times

`minikube tunnel` runs as a separate daemon, creates a network route on the host to the service CIDR of the cluster using the cluster's IP address as a gateway. Adding a route requires root privileges for the user, and thus there are differences in how to run `minikube tunnel` depending on the OS.

If you want to avoid entering the root password, consider setting NOPASSWD for "ip" and "route" commands:

https://superuser.com/questions/1328452/sudoers-nopasswd-for-single-executable-but-allowing-others

