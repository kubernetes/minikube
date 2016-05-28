### Run localkube in a docker container (experimental)

**Warning:** This is very experimental code at the moment.

#### How to build

```console
$ make VERSION=vX.Y.Z
```

#### How to run

```console
$ docker run -d \
    --volume=/:/rootfs:ro \
    --volume=/sys:/sys:rw \
    --volume=/var/lib/docker:/var/lib/docker:rw \
    --volume=/var/lib/kubelet:/var/lib/kubelet:rw \
    --volume=/var/run:/var/run:rw \
    --net=host \
    --pid=host \
    --privileged \
    gcr.io/google_containers/localkube-amd64:vX.Y.Z \
    /localkube start \
    --containerized
```