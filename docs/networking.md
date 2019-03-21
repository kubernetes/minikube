# Networking

The minikube VM is exposed to the host system via a host-only IP address, that can be obtained with the `minikube ip` command.
Any services of type `NodePort` can be accessed over that IP address, on the NodePort.

To determine the NodePort for your service, you can use a `kubectl` command like this (note that `nodePort` begins with lowercase `n` in JSON output):

`kubectl get service $SERVICE --output='jsonpath="{.spec.ports[0].nodePort}"'`

We also have a shortcut for fetching the minikube IP and a service's `NodePort`:

`minikube service --url $SERVICE`

## LoadBalancer emulation (`minikube tunnel`)

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

## (Advanced) Running tunnel as root to avoid entering password multiple times

`minikube tunnel` runs as a separate daemon, creates a network route on the host to the service CIDR of the cluster using the cluster's IP address as a gateway.
Adding a route requires root privileges for the user, and thus there are differences in how to run `minikube tunnel` depending on the OS.

Recommended way to use on Linux with KVM2 driver and MacOSX with Hyperkit driver:

`sudo -E minikube tunnel`

Using VirtualBox on Windows, Mac and Linux _both_ `minikube start` and `minikube tunnel` needs to be started from the same Administrator user session otherwise [VBoxManage can't recognize the created VM](https://forums.virtualbox.org/viewtopic.php?f=6&t=81551).
