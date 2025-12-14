---
title: "Minikube AI playground on Apple silicon"
linkTitle: "Minikube AI playground on Apple silicon"
weight: 1
date: 2024-10-04
---

This tutorial shows how to create an AI playground with minikube on Apple
silicon devices such as a MacBook Pro. We'll create a cluster that shares your
Mac's GPU using the krunkit driver, deploy two large language models, and
interact with the models using Open WebUI.

![Open WebUI Chat](/images/open-webui-chat.png)

## Prerequisites

- Apple silicon Mac
- [krunkit](https://github.com/containers/krunkit) v1.0.0 or later
- [vmnet-helper](https://github.com/nirs/vmnet-helper) v0.6.0 or later
- [generic-device-plugin](https://github.com/squat/generic-device-plugin)
- minikube v1.37.0 or later (krunkit driver only)

## Installing krunkit and vmnet-helper

Install latest krunkit:

```shell
brew tap slp/krunkit
brew install krunkit
krunkit --version
```

Install latest vmnet-helper:

```shell
curl -fsSL https://github.com/minikube-machine/vmnet-helper/releases/latest/download/install.sh | bash
/opt/vmnet-helper/bin/vmnet-helper --version
```

For more information, see the [krunkit driver](https://minikube.sigs.k8s.io/docs/drivers/krunkit/)
documentation.

## Download models

Download some models to the local disk. By keeping the models outside of
minikube, you can create and delete clusters quickly without downloading the
models again.

```shell
mkdir ~/models
cd ~/models
curl -LO 'https://huggingface.co/instructlab/granite-7b-lab-GGUF/resolve/main/granite-7b-lab-Q4_K_M.gguf?download=true'
curl -LO 'https://huggingface.co/TheBloke/TinyLlama-1.1B-Chat-v1.0-GGUF/resolve/main/tinyllama-1.1b-chat-v1.0.Q8_0.gguf?download=true'
```

**Important**: The model must be in *GGUF* format.

## Start minikube

Start a minikube cluster with the krunkit driver, mounting the `~/models`
directory at `/mnt/models`:

```shell
minikube start --driver krunkit --mount-string ~/models:/mnt/models
```

Output:
```
😄  minikube v1.37.0 on Darwin 15.6.1 (arm64)
✨  Using the krunkit (experimental) driver based on user configuration
👍  Starting "minikube" primary control-plane node in "minikube" cluster
🔥  Creating krunkit VM (CPUs=2, Memory=6144MB, Disk=20000MB) ...
🐳  Preparing Kubernetes v1.34.0 on Docker 28.4.0 ...
🔗  Configuring bridge CNI (Container Networking Interface) ...
🔎  Verifying Kubernetes components...
    ▪ Using image gcr.io/k8s-minikube/storage-provisioner:v5
🌟  Enabled addons: storage-provisioner, default-storageclass
🏄  Done! kubectl is now configured to use "minikube" cluster and "default" namespace by default
```

### Verifying that the GPU is available

The krunkit driver exposes your host GPU as a virtio-gpu device:

```
% minikube ssh -- tree /dev/dri
/dev/dri
|-- by-path
|   |-- platform-a007000.virtio_mmio-card -> ../card0
|   `-- platform-a007000.virtio_mmio-render -> ../renderD128
|-- card0
`-- renderD128
```

## Deploying the generic-device-plugin

To use the GPU in pods, we need the generic-device-plugin. Deploy it with:

```shell
cat <<'EOF' | kubectl apply -f -
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: generic-device-plugin
  namespace: kube-system
  labels:
    app.kubernetes.io/name: generic-device-plugin
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: generic-device-plugin
  template:
    metadata:
      labels:
        app.kubernetes.io/name: generic-device-plugin
    spec:
      priorityClassName: system-node-critical
      tolerations:
      - operator: "Exists"
        effect: "NoExecute"
      - operator: "Exists"
        effect: "NoSchedule"
      containers:
      - image: squat/generic-device-plugin
        args:
        - --device
        - |
          name: dri
          groups:
          - count: 4
            paths:
            - path: /dev/dri
        name: generic-device-plugin
        resources:
          requests:
            cpu: 50m
            memory: 10Mi
          limits:
            cpu: 50m
            memory: 20Mi
        ports:
        - containerPort: 8080
          name: http
        securityContext:
          privileged: true
        volumeMounts:
        - name: device-plugin
          mountPath: /var/lib/kubelet/device-plugins
        - name: dev
          mountPath: /dev
      volumes:
      - name: device-plugin
        hostPath:
          path: /var/lib/kubelet/device-plugins
      - name: dev
        hostPath:
          path: /dev
  updateStrategy:
    type: RollingUpdate
EOF
```

**Note**: This configuration allows up to 4 pods to use `/dev/dri`. You can
increase `count` to run more pods using the GPU.

Wait until the generic-device-plugin DaemonSet is available:

```shell
% kubectl get daemonset generic-device-plugin -n kube-system -w
NAME                    DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE   NODE SELECTOR   AGE
generic-device-plugin   1         1         1       1            1           <none>          45s
```

## Deploying the granite model

To play with the granite model you downloaded, start a llama-server pod serving
the model and a service to make the pod available to other pods.

```shell
cat <<'EOF' | kubectl apply -f -
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: granite
spec:
  replicas: 1
  selector:
    matchLabels:
      app: granite
  template:
    metadata:
      labels:
        app: granite
      name: granite
    spec:
      containers:
      - name: llama-server
        image: quay.io/ramalama/ramalama:latest
        command: [
          llama-server,
          --host, "0.0.0.0",
          --port, "8080",
          --model, /mnt/models/granite-7b-lab-Q4_K_M.gguf,
          --alias, "ibm/granite:7b",
          --ctx-size, "2048",
          --temp, "0.8",
          --cache-reuse, "256",
          -ngl, "999",
          --threads, "6",
          --no-warmup,
          --log-colors, auto,
        ]
        resources:
          limits:
            squat.ai/dri: 1
        volumeMounts:
        - name: models
          mountPath: /mnt/models
      volumes:
      - name: models
        hostPath:
          path: /mnt/models
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: granite
  name: granite
spec:
  ports:
  - protocol: TCP
    port: 8080
  selector:
    app: granite
EOF
```

Wait until the deployment is available:

```shell
% kubectl get deploy granite
NAME      READY   UP-TO-DATE   AVAILABLE   AGE
granite   1/1     1            1           8m17s
```

Check the granite service:

```shell
% kubectl get service granite
NAME      TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
granite   ClusterIP   10.105.145.9   <none>        8080/TCP   28m
```

## Deploying the tinyllama model

To play with the tinyllama model you downloaded, start a llama-server pod
serving the model and a service to make the pod available to other pods.

```shell
cat <<'EOF' | kubectl apply -f -
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: tinyllama
spec:
  replicas: 1
  selector:
    matchLabels:
      app: tinyllama
  template:
    metadata:
      labels:
        app: tinyllama
      name: tinyllama
    spec:
      containers:
      - name: llama-server
        image: quay.io/ramalama/ramalama:latest
        command: [
          llama-server,
          --host, "0.0.0.0",
          --port, "8080",
          --model, /mnt/models/tinyllama-1.1b-chat-v1.0.Q8_0.gguf,
          --alias, tinyllama,
          --ctx-size, "2048",
          --temp, "0.8",
          --cache-reuse, "256",
          -ngl, "999",
          --threads, "6",
          --no-warmup,
          --log-colors, auto,
        ]
        resources:
          limits:
            squat.ai/dri: 3
        volumeMounts:
        - name: models
          mountPath: /mnt/models
      volumes:
      - name: models
        hostPath:
          path: /mnt/models
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: tinyllama
  name: tinyllama
spec:
  ports:
  - protocol: TCP
    port: 8080
  selector:
    app: tinyllama
EOF
```

Wait until the deployment is available:

```
% kubectl get deploy tinyllama
NAME        READY   UP-TO-DATE   AVAILABLE   AGE
tinyllama   1/1     1            1           9m14s
```

Check the tinyllama service:

```shell
% kubectl get service tinyllama
NAME        TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
tinyllama   ClusterIP   10.98.219.117   <none>        8080/TCP   23m
```

## Deploying Open WebUI

The [Open WebUI](https://docs.openwebui.com) project provides an easy-to-use web
interface for interacting with OpenAI-compatible APIs such as our llama-server
pods.

To deploy Open WebUI, run:

```shell
---
cat <<'EOF' | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: open-webui
spec:
  replicas: 1
  selector:
    matchLabels:
      app: open-webui
  template:
    metadata:
      labels:
        app: open-webui
    spec:
      containers:
        - name: open-webui
          image: ghcr.io/open-webui/open-webui:dev-slim
          ports:
            - containerPort: 8080
          env:
            # Preconfigure OpenAI-compatible endpoints
            - name: OPENAI_API_BASE_URLS
              value: "http://granite:8080/v1;http://tinyllama:8080/v1"
          volumeMounts:
            - name: open-webui-data
              mountPath: /app/backend/data
      volumes:
        - name: open-webui-data
          persistentVolumeClaim:
            claimName: open-webui-data
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: open-webui-data
spec:
  storageClassName: standard
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
---
apiVersion: v1
kind: Service
metadata:
  name: open-webui
spec:
  ports:
  - protocol: TCP
    port: 8080
    nodePort: 30080
  selector:
    app: open-webui
  type: NodePort
EOF
```

We configured our llama-server OpenAI compatible API endpoints using the
`OPENAI_API_BASE_URLS` environment variable.
Check [Open WebUI documentation](https://docs.openwebui.com) to learn how to
configure it using the Admin Panel.

Wait until the deployment is available:

```shell
% kubectl get deploy open-webui
NAME         READY   UP-TO-DATE   AVAILABLE   AGE
open-webui   1/1     1            1           69s
```

## Interacting with the models

Open a browser with Open WebUI console:

```shell
open $(minikube service open-webui --url)
```

Create an Admin account to start using Open WebUI.

### Chatting with the granite model

You can start chatting with the "ibm/granite:7b" model.

Type a prompt:

```
> Write a very technical haiku about playing with large language models with Minikube on Apple silicon

Mighty model, Minikube,
Silicon-powered speed,
Learning's dance, ever-changing.
Through data streams it weaves,
Inference's wisdom, vast and deep,
Apple's heartbeat, in code, resounds.
Exploring AI's vast frontier,
Minikube, language model's playground,
Innovation's rhythm, forever.
```

### Chatting with the tinyllama model

Click the "New Chat" button on the left and select the "tinyllama" model from
the model menu in the top left.

Type a prompt:

```
> How do you feel inside this fancy Minikube cluster?

I do not have a physical body. However, based on the given text material, the
author is describing feeling inside a cluster of Minikube, a type of jellyfish.
The use of the word "fancy" suggests that the author is impressed or appreciates
the intricate design of the cluster, while the adjective "minikube" connotes its
smooth texture, delicate shape, and iridescent colors. The word "cluster"
suggests a group of these jellyfish, while "inside" implies being in the
vicinity or enclosed within.
```
