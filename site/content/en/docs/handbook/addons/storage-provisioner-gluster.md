---
title: "Using the Storage Provisioner Gluster Addon"
linkTitle: "Storage Provisioner Gluster"
weight: 1
date: 2019-01-09
---

## storage-provisioner-gluster addon
[Gluster](https://gluster.org/), a scalable network filesystem that provides dynamic provisioning of PersistentVolumeClaims.

### Starting Minikube
This addon works within Minikube, without any additional configuration.

```shell
$ minikube start
```

### Enabling storage-provisioner-gluster
To enable this addon, simply run:

```
$ minikube addons enable storage-provisioner-gluster
```

Within one minute, the addon manager should pick up the change and you should see several Pods in the `storage-gluster` namespace:

```
$ kubectl -n storage-gluster get pods
NAME                                      READY     STATUS              RESTARTS   AGE
glusterfile-provisioner-dbcbf54fc-726vv   1/1       Running             0          1m
glusterfs-rvdmz                           0/1       Running             0          40s
heketi-79997b9d85-42c49                   0/1       ContainerCreating   0          40s
```

Some of the Pods need a little more time to get up an running than others, but in a few minutes everything should have been deployed and all Pods should be `READY`:

```
$ kubectl -n storage-gluster get pods
NAME                                      READY     STATUS    RESTARTS   AGE
glusterfile-provisioner-dbcbf54fc-726vv   1/1       Running   0          5m
glusterfs-rvdmz                           1/1       Running   0          4m
heketi-79997b9d85-42c49                   1/1       Running   1          4m
```

Once the Pods have status `Running`, the `glusterfile` StorageClass should have been marked as `default`:

```
$ kubectl get sc
NAME                    PROVISIONER               AGE
glusterfile (default)   gluster.org/glusterfile   3m
```

### Creating PVCs
The storage in the Gluster environment is limited to 10 GiB. This is because the data is stored in the Minikube VM (a sparse file `/srv/fake-disk.img`).

The following `yaml` creates a PVC, starts a CentOS developer Pod that generates a website and deploys an NGINX webserver that provides access to the website:

```
---
#
# Minimal PVC where a developer can build a website.
#
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: website
spec:
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      storage: 2Mi
  storageClassName: glusterfile
---
#
# This pod will just download a fortune phrase and store it (as plain text) in
# index.html on the PVC. This is how we create websites?
#
# The root of the website stored on the above PVC is mounted on /mnt.
#
apiVersion: v1
kind: Pod
metadata:
  name: centos-webdev
spec:
  containers:
  - image: centos:latest
    name: centos
    args:
    - curl
    - -o/mnt/index.html
    - https://api.ef.gy/fortune
    volumeMounts:
    - mountPath: /mnt
      name: website
  # once the website is created, the pod will exit
  restartPolicy: Never
  volumes:
  - name: website
    persistentVolumeClaim:
      claimName: website
---
#
# Start a NGINX webserver with the website.
# We'll skip creating a service, to keep things minimal.
#
apiVersion: v1
kind: Pod
metadata:
  name: website-nginx
spec:
  containers:
  - image: gcr.io/google_containers/nginx-slim:0.8
    name: nginx
    ports:
    - containerPort: 80
      name: web
    volumeMounts:
    - mountPath: /usr/share/nginx/html
      name: website
  volumes:
  - name: website
    persistentVolumeClaim:
      claimName: website
```

Because the PVC has been created with the `ReadWriteMany` accessMode, both Pods can access the PVC at the same time. Other website developer Pods can use the same PVC to update the contents of the site.

The above configuration does not expose the website on the Minikube VM. One way to see the contents of the website is to SSH into the Minikube VM and fetch the website there:

```
$ kubectl get pods -o wide
NAME            READY     STATUS      RESTARTS   AGE       IP           NODE
centos-webdev   0/1       Completed   0          1m        172.17.0.9   minikube
website-nginx   1/1       Running     0          24s       172.17.0.9   minikube
$ minikube ssh
                         _             _            
            _         _ ( )           ( )           
  ___ ___  (_)  ___  (_)| |/')  _   _ | |_      __  
/' _ ` _ `\| |/' _ `\| || , <  ( ) ( )| '_`\  /'__`\
| ( ) ( ) || || ( ) || || |\`\ | (_) || |_) )(  ___/
(_) (_) (_)(_)(_) (_)(_)(_) (_)`\___/'(_,__/'`\____)

$ curl http://172.17.0.9
I came, I saw, I deleted all your files.
$ 
```
