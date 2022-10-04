---
title: "Registry Integration"
linkTitle: "Registry Integration"
weight: 6
description: >
  Cloud provider registry integration (Google Artifact Registry + Service Account + Permission)
aliases:
 - /docs/tasks/registry_integration
---

## Push Ngnix image to private Google Artifact Registry

 - Push a Nginx image to a private Google Artifact Registry using GCP-Auth addon.

## Deploy a pod (yaml) to minikube using that image

Create a new pod with your Nginx image:
```
kubectl run nginx --image=nginx --restart=Never --command -- sleep infinity
```

## Ensure that the Pod is working (running)

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
