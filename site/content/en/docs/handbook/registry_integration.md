---
title: "Google Artifact Registry Integration"
linkTitle: "Google Artifact Registry Integration"
weight: 6
description: >
  Pull an image from Google's Artifact Registry into a minikube pod
aliases:
 - /docs/tasks/artifact_registry_integration
---

## Push Image to Google Artifact Registry

##### Authenticate with Artifact Registry:

`gcloud beta auth configure-docker us-central1-docker.pkg.dev`

##### Configure Docker with Artifact Registry Credentials:

Check Github for the latest [GoogleCloudPlatform/docker-credential-gcr](https://github.com/GoogleCloudPlatform/docker-credential-gcr/releases) release.

 - `VERSION=2.1.6`

For Linux, enter "linux", for Apple Mac OS, enter "darwin" and for Microsoft Windows, enter "windows".

 - `OS=linux`

For AMD64, enter "amd64", for ARM64, enter "arm64" and for 32-bit OS', enter "386".

 - `ARCH=amd64`

 Curl the appropriate version of `docker-credential-gcr`:

```
curl -fsSL "https://github.com/GoogleCloudPlatform/docker-credential-gcr/releases/download/v${VERSION}/docker-credential-gcr_${OS}_${ARCH}-${VERSION}.tar.gz" \
  | tar xz --to-stdout ./docker-credential-gcr \
  > /usr/bin/docker-credential-gcr
```

Make `docker-credential-gcr` executable:

`chmod +x /usr/bin/docker-credential-gcr`

##### Configure Docker Artifact Registry Interaction

`docker-credential-gcr configure-docker --registries=us-central1-docker.pkg.dev`

##### Run Docker Push

Run `docker push` to push an image from Docker to Artifact Registry.

See [Docker push documentation](https://docs.docker.com/engine/reference/commandline/push/) for more information.

## Deploy a pod to minikube with Artifact Registry Image

Create a new pod with your Image:
```
kubectl run nginx --image=nginx --restart=Never --command -- sleep infinity
```

## Ensure Pod is Running

Obtain a list of all pods:

```
kubectl get pods -A
```

```
NAMESPACE     NAME                               READY   STATUS    RESTARTS      AGE
default       nginx                              1/1     Running   0             117s
```

Describe the pod "nginx":
```
kubectl describe po nginx
```

```
Name:             nginx
Namespace:        default
Priority:         0
Service Account:  default
Node:             minikube/192.168.49.2
Start Time:       Tue, 04 Oct 2022 16:02:40 -0700
Labels:           run=nginx
Annotations:      <none>
Status:           Running
IP:               172.17.0.3
IPs:
  IP:  172.17.0.3
Containers:
  nginx:
    Container ID:  docker://fada0cc1f9d0c7bdb5f224fc6655875b6975e92b63663755a686c8a1a7f168e6
    Image:         nginx
    Image ID:      docker-pullable://nginx@sha256:0b970013351304af46f322da1263516b188318682b2ab1091862497591189ff1
    Port:          <none>
    Host Port:     <none>
    Command:
      sleep
      infinity
    State:          Running
      Started:      Tue, 04 Oct 2022 16:02:46 -0700
    Ready:          True
    Restart Count:  0
    Environment:    <none>
    Mounts:
      /var/run/secrets/kubernetes.io/serviceaccount from kube-api-access-xbnk6 (ro)
Conditions:
  Type              Status
  Initialized       True
  Ready             True
  ContainersReady   True
  PodScheduled      True
Volumes:
  kube-api-access-xbnk6:
    Type:                    Projected (a volume that contains injected data from multiple sources)
    TokenExpirationSeconds:  3607
    ConfigMapName:           kube-root-ca.crt
    ConfigMapOptional:       <nil>
    DownwardAPI:             true
QoS Class:                   BestEffort
Node-Selectors:              <none>
Tolerations:                 node.kubernetes.io/not-ready:NoExecute op=Exists for 300s
                             node.kubernetes.io/unreachable:NoExecute op=Exists for 300s
Events:
  Type    Reason     Age    From               Message
  ----    ------     ----   ----               -------
  Normal  Scheduled  3m29s  default-scheduler  Successfully assigned default/nginx to minikube
  Normal  Pulling    3m29s  kubelet            Pulling image "nginx"
  Normal  Pulled     3m23s  kubelet            Successfully pulled image "nginx" in 6.037499962s
  Normal  Created    3m23s  kubelet            Created container nginx
  Normal  Started    3m23s  kubelet            Started container nginx
```

If you have any questions or comments please do not hesitate to reach out to the [community](http://localhost:1313/community/).