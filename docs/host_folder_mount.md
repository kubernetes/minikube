# Mounting Host Folders

`minikube mount /path/to/dir/to/mount:/vm-mount-path` is the recommended way to mount directories into minikube so that they can be used in your local Kubernetes cluster. The command works on all supported platforms. Below is an example workflow for using `minikube mount`:

```shell
# terminal 1
$ mkdir ~/mount-dir
$ minikube mount ~/mount-dir:/mount-9p
Mounting /home/user/mount-dir/ into /mount-9p on the minikubeVM
This daemon process needs to stay alive for the mount to still be accessible...
ufs starting
# This process has to stay open, so in another terminal...
```

```shell
# terminal 2
$ echo "hello from host" > ~/mount-dir/hello-from-host
$ kubectl run -i --rm --tty ubuntu --overrides='
{
  "apiVersion": "v1",
  "kind": "Pod",
  "metadata": {
    "name": "ubuntu"
  },
  "spec": {
        "containers": [
          {
            "name": "ubuntu",
            "image": "ubuntu:14.04",
            "args": [
              "bash"
            ],
            "stdin": true,
            "stdinOnce": true,
            "tty": true,
            "workingDir": "/mount-9p",
            "volumeMounts": [{
              "mountPath": "/mount-9p",
              "name": "host-mount"
            }]
          }
        ],
    "volumes": [
      {
        "name": "host-mount",
        "hostPath": {
          "path": "/mount-9p"
        }
      }
    ]
  }
}
' --image=ubuntu:14.04 --restart=Never -- bash

Waiting for pod default/ubuntu to be running, status is Pending, pod ready: false
Waiting for pod default/ubuntu to be running, status is Running, pod ready: false
# ======================================================================================
# We are now in the pod
#=======================================================================================
root@ubuntu:/mount-9p# cat hello-from-host
hello from host
root@ubuntu:/mount-9p# echo "hello from pod" > /mount-9p/hello-from-pod
root@ubuntu:/mount-9p# ls
hello-from-host  hello-from-pod
root@ubuntu:/mount-9p# exit
exit
Waiting for pod default/ubuntu to terminate, status is Running
pod "ubuntu" deleted
# ======================================================================================
# We are back on the host
#=======================================================================================
$ cat ~/mount-dir/hello-from-pod
hello from pod
```

Some drivers themselves provide host-folder sharing options, but we plan to deprecate these in the future as they are all implemented differently and they are not configurable through minikube.
