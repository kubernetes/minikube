---
title: "Accessing apps"
weight: 3
description: >
  How to access applications running within minikube
aliases:
 - /docs/tasks/loadbalancer
 - /Handbook/loadbalancer/
 - /docs/tasks/nodeport
---

There are two major categories of services in Kubernetes:

* NodePort
* LoadBalancer

minikube supports either. Read on!

## NodePort access

A NodePort service is the most basic way to get external traffic directly to your service. NodePort, as the name implies, opens a specific port, and any traffic that is sent to this port is forwarded to the service.

### Getting the NodePort using the service command

We also have a shortcut for fetching the minikube IP and a service's `NodePort`:

```shell
minikube service <service-name> --url
```

### Using `minikube service` with tunnel

The network is limited if using the Docker driver on Darwin, Windows, or WSL, and the Node IP is not reachable directly.

Running minikube on Linux with the Docker driver will result in no tunnel being created.

Services of type `NodePort` can be exposed via the `minikube service <service-name> --url` command. It must be run in a separate terminal window to keep the [tunnel](https://en.wikipedia.org/wiki/Port_forwarding#Local_port_forwarding) open. Ctrl-C in the terminal can be used to terminate the process at which time the network routes will be cleaned up.

### Example of NodePort

1. Create a Kubernetes deployment

    ```shell
    kubectl create deployment hello-minikube1 --image=kicbase/echo-server:1.0
    ```

2. Create a Kubernetes service type NodePort

    ```shell
    kubectl expose deployment hello-minikube1 --type=NodePort --port=8080
    ```

3. Check Node Port

    ```shell
    $ kubectl get svc
    NAME              TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)          AGE
    hello-minikube1   NodePort    10.100.238.34   <none>        8080:31389/TCP   3s
    ```

4. Run service tunnel

    ```shell
    minikube service hello-minikube1 --url
    ```

    `minikube service hello-minikube1 --url` runs as a process, creating a [tunnel](https://en.wikipedia.org/wiki/Port_forwarding#Local_port_forwarding) to the cluster. The command exposes the service directly to any program running on the host operating system.

    <details>
    <summary>
    service output example
    </summary>
    <pre>
    $ minikube service hello-minikube1 --url
    http://127.0.0.1:57123
    ❗  Because you are using a Docker driver on darwin, the terminal needs to be open to run it.
    </pre>
    </details>

    Check ssh tunnel in another terminal

    ```shell
    $ ps -ef | grep docker@127.0.0.1
    ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -N docker@127.0.0.1 -p 55972 -i /Users/FOO/.minikube/machines/minikube/id_rsa -L TUNNEL_PORT:CLUSTER_IP:TARGET_PORT
    ```

5. Try in your browser

    Open in your browser (ensure there is no proxy set)

    ```shell
    http://127.0.0.1:TUNNEL_PORT
    ```

### Getting the NodePort using kubectl

The minikube VM is exposed to the host system via a host-only IP address, that can be obtained with the `minikube ip` command. Any services of type `NodePort` can be accessed over that IP address, on the NodePort.

To determine the NodePort for your service, you can use a `kubectl` command like this (note that `nodePort` begins with lowercase `n` in JSON output):

```shell
kubectl get service <service-name> --output='jsonpath="{.spec.ports[0].nodePort}"'
```

### Increasing the NodePort range

By default, minikube only exposes ports 30000-32767. If this does not work for you, you can adjust the range by using:

```shell
minikube start --extra-config=apiserver.service-node-port-range=1-65535
```

This flag also accepts a comma separated list of ports and port ranges.

----

## LoadBalancer access

A LoadBalancer service is the standard way to expose a service to the internet. With this method, each service gets its own IP address.

### Using `minikube tunnel`

Services of type `LoadBalancer` can be exposed via the `minikube tunnel` command. It must be run in a separate terminal window to keep the `LoadBalancer` running.  Ctrl-C in the terminal can be used to terminate the process at which time the network routes will be cleaned up.

### Example of LoadBalancer

1. Run the tunnel in a separate terminal

    It will ask for a password.

    ```shell
    minikube tunnel
    ```

    `minikube tunnel` runs as a process, creating a network route on the host to the service CIDR of the cluster using the cluster's IP address as a gateway.  The tunnel command exposes the external IP directly to any program running on the host operating system.

    <details>
    <summary>
    tunnel output example
    </summary>
    <pre>
    Password:
    Status:
    machine: minikube
    pid: 39087
    route: 10.96.0.0/12 -> 192.168.64.194
    minikube: Running
    services: [hello-minikube]
        errors:
      minikube: no errors
      router: no errors
      loadbalancer emulator: no errors
    ...
    ...
    ...
    </pre>
    </details>

2. Create a Kubernetes deployment

    ```shell
    kubectl create deployment hello-minikube1 --image=kicbase/echo-server:1.0
    ```

3. Create a Kubernetes service with type LoadBalancer

    ```shell
    kubectl expose deployment hello-minikube1 --type=LoadBalancer --port=8080
    ```

4. Check the external IP

    ```shell
    kubectl get svc
    ```
    <pre>
    $ kc get svc
    NAME              TYPE           CLUSTER-IP      EXTERNAL-IP     PORT(S)          AGE
    hello-minikube1   LoadBalancer   10.96.184.178   10.96.184.178   8080:30791/TCP   40s
    </pre>

    Note that without minikube tunnel, Kubernetes will show the external IP as "pending".

5. Try in your browser

    Open in your browser (ensure there is no proxy set)

    ```shell
    http://REPLACE_WITH_EXTERNAL_IP:8080
    ```

    Each service will get its own external IP.

----

### DNS resolution (experimental)

If you are on macOS, the tunnel command also allows DNS resolution for Kubernetes services from the host.

NOTE: docker driver doesn't support DNS resolution

### Cleaning up orphaned routes

If the `minikube tunnel` shuts down in an abrupt manner, it may leave orphaned network routes on your system. If this happens, the ~/.minikube/tunnels.json file will contain an entry for that tunnel. To remove orphaned routes, run:

````shell
minikube tunnel --cleanup
````

NOTE: `--cleanup` flag's default value is `true`.

### Avoiding password prompts

Adding a route requires root privileges for the user, and thus there are differences in how to run `minikube tunnel` depending on the OS. If you want to avoid entering the root password, consider setting NOPASSWD for "ip" and "route" commands:

<https://superuser.com/questions/1328452/sudoers-nopasswd-for-single-executable-but-allowing-others>

### Access to ports <1024 on Windows requires root permission

If you are using Docker driver on Windows, there is a chance that you have an old version of SSH client you might get an error like - `Privileged ports can only be forwarded by root.` or you might not be able to access the service even after `minikube tunnel` if the access port is less than 1024 but for ports greater than 1024 works fine.

In order to resolve this, ensure that you are running the latest version of SSH client. You can install the latest version of the SSH client on Windows by running the following in a Command Prompt with an Administrator Privileges (Requires [chocolatey package manager](https://chocolatey.org/install))
```cmd
choco install openssh
```
The latest version (`OpenSSH_for_Windows_7.7p1, LibreSSL 2.6.5`) which is available on Windows 10 by default doesn't work. You can track the issue with this over here - https://github.com/PowerShell/Win32-OpenSSH/issues/1693
