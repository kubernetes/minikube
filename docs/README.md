# Advanced Topics and Tutorials

## Cluster Configuration

* **Alternative Runtimes** ([alternative_runtimes.md](alternative_runtimes.md)): How to run minikube without Docker as the container runtime

* **Environment Variables** ([env_vars.md](env_vars.md)): The different environment variables that minikube understands

* **Minikube Addons** ([addons.md](addons.md)): Information on configuring addons to be run on minikube

* **Configuring Kubernetes** ([configuring_kubernetes.md](configuring_kubernetes.md)): Configuring different Kubernetes components in minikube

* **Caching Images** ([cache.md](cache.md)): Caching non-minikube images in minikube

* **GPUs** ([gpu.md](gpu.md)): Using NVIDIA GPUs on minikube

* **OpenID Connect Authentication** ([openid_connect_auth.md](openid_connect_auth.md)): Using OIDC Authentication on minikube

### Installation and debugging

* **Driver installation** ([drivers.md](drivers.md)): In depth instructions for installing the various hypervisor drivers

* **Debugging minikube** ([debugging.md](debugging.md)): General practices for debugging the minikube binary itself

### Developing on the minikube cluster

* **Reusing the Docker Daemon** ([reusing_the_docker_daemon.md](reusing_the_docker_daemon.md)): How to point your docker CLI to the docker daemon running inside minikube

* **Building images within the VM** ([building_images_within_the_vm.md](building_images_within_the_vm.md)): How to build a container image within the minikube VM

#### Storage

* **Persistent Volumes** ([persistent_volumes.md](persistent_volumes.md)): Persistent Volumes in Minikube and persisted locations in the VM

* **Host Folder Mounting** ([host_folder_mount.md](host_folder_mount.md)): How to mount your files from your host into the minikube VM

* **Syncing files into the VM** ([syncing-files.md](syncing-files.md)): How to sync files from your host into the minikube VM

#### Networking

* **HTTP Proxy** ([http_proxy.md](http_proxy.md)): Instruction on how to run minikube behind a HTTP Proxy

* **Insecure or Private Registries** ([insecure_registry.md](insecure_registry.md)): How to use private or insecure registries with minikube

* **Accessing etcd from inside the cluster** ([accessing_etcd.md](accessing_etcd.md))

* **Networking** ([networking.md](networking.md)): FAQ about networking between the host and minikube VM

* **Offline** ([offline.md](offline.md)): Details about using minikube offline
