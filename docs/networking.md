## Networking

The minikube VM is exposed to the host system via a host-only IP address, that can be obtained with the `minikube ip` command.
Any services of type `NodePort` can be accessed over that IP address, on the NodePort.

To determine the NodePort for your service, you can use a `kubectl` command like this (note that `nodePort` begins with lowercase `n` in JSON output):

`kubectl get service $SERVICE --output='jsonpath="{.spec.ports[0].nodePort}"'`

We also have a shortcut for fetching the minikube IP and a service's `NodePort`:

`minikube service --url $SERVICE`

### LoadBalancer support

Services of type `LoadBalancer` can be exposed via the `minikube tunnel` command. 

````shell
$ sudo "PATH=$PATH" "HOME=$HOME" minikube tunnel
[sudo] password for balintp: 
INFO[0000] Creating docker machine client...            
INFO[0000] Creating k8s client...                       
INFO[0000] Setting up router...                         
INFO[0000] Setting up tunnel...                         
INFO[0000] Started minikube tunnel. Tunnel requires root access. Please wait for the password prompt! 
INFO[0005] Adding route for CIDR 10.96.0.0/12 to gateway 192.168.39.90 
INFO[0005] About to run command: [sudo ip route add 10.96.0.0/12 via 192.168.39.90] 
INFO[0005]                                              
TunnelState:
	minikube: Running
	route: 10.96.0.0/12 -> 192.168.39.90
	services: []
INFO[0055] Patched nginx with IP 10.101.207.169         
TunnelState:
	minikube: Running
	route: 10.96.0.0/12 -> 192.168.39.90
	services: [nginx]
```` 


`minikube tunnel` runs as a separate daemon, creates a network route on the host to the service CIDR of the cluster using the cluster's IP address as a gateway. 
Adding a route requires root privileges for the user, and thus there are differences in how to run this operating   

Recommended way to use on Linux with KVM2 driver and MacOSX with Hyperkit driver: 

`sudo "PATH=$PATH" "HOME=$HOME" minikube tunnel`

Using VirtualBox on Windows _both_ `minikube start` and `minikube tunnel` needs to be started from the same Administrator user session otherwise [VBoxManage can't recognize the created VM](https://forums.virtualbox.org/viewtopic.php?f=6&t=81551).     

