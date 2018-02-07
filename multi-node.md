
## Current setup and limitations

* tested only with virtualbox
* pods can communicate across nodes
* nodes get non-overlapping subnets for pod IPs
* CNI is deployed manually, out-of-band
* hard-coded pod subnets in CNI config (set up for `flannel`)
* some components are explicitly told to use `eth1`
* code is hacked-in, minimal refactoring done


## TODO:

* Parameterize machine name
* Start worker nodes
* Configure worker nodes

## Usage

## Add node

```
minikube node add [--cluster=default] --name=[<node_name>]
```

```
minikube node start <node_name>
```

### Cluster

* create
* delete
* add node

### Node

* create
* delete
* start
* stop

### Bootstrapper

* configure
* start
* restart
