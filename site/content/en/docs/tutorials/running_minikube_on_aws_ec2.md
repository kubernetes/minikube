---
title: "Running minikube on AWS EC2"
linkTitle: "Minikube on AWS EC2"
weight: 1
date: 2026-05-23
description: >
  How to run minikube on an AWS EC2 instance as a remote development environment
---

![minikube running on an iPad via SSH to EC2](/images/minikube-on-ipad-via-ssh.jpeg)

*Above: minikube running on an EC2 t3.medium, accessed from an iPad over SSH. The iPad itself can't run minikube — but it doesn't need to.*

This tutorial covers running minikube on an AWS EC2 instance, useful when you want to run your Kubernetes development cluster on a remote machine instead of locally.

## Why run minikube on EC2

Most people should run minikube locally — it's simpler. But a few cases where running on EC2 makes sense:

- **Your laptop can't handle it.** 4–8 GB RAM machines running Docker, an IDE, and browser tabs often can't fit a minikube cluster on top.
- **Corporate-managed laptops.** IT-locked machines where you can't install Docker, KVM, or other drivers locally.
- **You don't want minikube draining battery and heating your laptop.**
- **You need a cluster that outlives your laptop session** — for testing webhooks, long-running jobs, or async systems.
- **You already develop on a remote VM** (cloud workstation, SSH'd EC2, VS Code Remote).
- **You need a stable endpoint for teammates or external tools** to hit your dev cluster.

If you're already running minikube fine on your laptop, you don't need this guide.

## Prerequisites

- An AWS account with EC2 access
- Basic familiarity with SSH and EC2
- A key pair for SSH access

## Choosing an EC2 instance type

minikube requires at least **2 CPUs and 2 GB RAM** to start (this is a Kubernetes requirement, not a minikube limitation). On AWS EC2:

| Instance type | vCPUs | RAM | Works? | Notes |
|---------------|-------|-----|--------|-------|
| `t2.micro` | 1 | 1 GB | ❌ | Free tier, but minikube refuses to start: "Requested cpu count 1 is less than the minimum allowed of 2" |
| `t3.small`  | 2 | 2 GB | ⚠️ | Meets minimums but tight; expect OOM under load |
| `t3.medium` | 2 | 4 GB | ✅ | Recommended minimum for comfortable use |
| `t3.large`  | 2 | 8 GB | ✅ | Better for running multiple workloads |

For a development/learning setup, **`t3.medium` is the practical minimum**. Pricing varies by region and over time — as a rough guide it costs a few cents per hour in most regions (check the [current EC2 pricing](https://aws.amazon.com/ec2/pricing/) for your region). Always **stop the instance when you're not using it** to avoid charges.

## Launch the EC2 instance

1. In the AWS console, launch an EC2 instance with:
   - **AMI:** Ubuntu Server 24.04 LTS (or newer, e.g. 26.04)
   - **Instance type:** `t3.medium`
   - **Key pair:** your existing SSH key
   - **Security group:** inbound SSH (22) from your IP
   - **Storage:** at least 20 GB (the default 8 GB fills up fast with Docker images)
2. SSH in:
   ```bash
   ssh -i /path/to/your-key.pem ubuntu@<EC2_PUBLIC_IP>
   ```

## Install Docker

minikube needs a container engine on the host to run the cluster. Install Docker Engine by following the [official Docker Engine install guide](https://docs.docker.com/engine/install/ubuntu/), then:

```bash
sudo systemctl enable --now docker

# Add your user to the docker group so you can run docker without sudo
sudo usermod -aG docker $USER
newgrp docker

# Verify
docker run --rm hello-world
```

## Install minikube

See the [official install guide](https://minikube.sigs.k8s.io/docs/start/) for full options. The quick Linux path:

```bash
curl -fSLO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
sudo install minikube-linux-amd64 /usr/local/bin/minikube
rm minikube-linux-amd64

# Verify
minikube version
```

## Install kubectl

See the [kubectl handbook](https://minikube.sigs.k8s.io/docs/handbook/kubectl/) for usage tips with minikube. The quick install:

```bash
curl -fSLO "https://dl.k8s.io/release/$(curl -fSL https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install kubectl /usr/local/bin/kubectl
rm kubectl

# Verify
kubectl version --client
```

## Start your cluster

```bash
minikube start
```

You'll see something like this:

![minikube start output on EC2](/images/minikube-start-on-ec2.png)

Verify the cluster is running:

```bash
kubectl get nodes
```

![kubectl get nodes output](/images/kubectl-get-nodes-on-ec2.png)

## Stopping your cluster (and the EC2 instance)

When you're done for the day:

```bash
# Stop minikube (keeps the cluster state on disk)
minikube stop
```

Then **stop the EC2 instance from the AWS Console** — `Instance State → Stop`. A stopped EC2 instance costs near zero (you only pay for the attached EBS volume). When you're back, **Start** the instance and `minikube start` picks up where you left off.

> **Note:** a stopped instance gets a new public IP when you start it again. If you need it to stay fixed, attach an Elastic IP.

If you're done with the cluster entirely, **Terminate** the instance instead — that stops the EBS charge too. Stopping only makes sense if you plan to come back to the same cluster.

{{% pageinfo color="warning" %}}
Don't forget this step. A running `t3.medium` costs roughly a dollar a day depending on region; a stopped one costs only a few cents per month for storage.
{{% /pageinfo %}}

## Deploying to a remote cluster

With local minikube you'd write manifests in your editor and apply them directly. On a remote host you author them on the instance itself over SSH. `nano` and `vim` are both available if you'd rather edit interactively, but a heredoc is the quickest way for a small manifest:

```bash
cat > my-pod.yaml <<'EOF'
apiVersion: v1
kind: Pod
metadata:
  name: my-app
spec:
  containers:
  - name: my-app
    image: nginx
EOF

kubectl apply -f my-pod.yaml
```

For anything bigger than a quick test, edit locally and copy it over with `scp -i /path/to/your-key.pem my-pod.yaml ubuntu@<EC2_IP>:~/`, or use VS Code's Remote-SSH extension to edit files on the instance as if they were local.


## Accessing services from outside the EC2 instance

> **Note:** running minikube remotely works differently from the local drivers. Normally your editor, `docker`, and `kubectl` all run on the same machine as the cluster — here they run on the remote EC2 host. You'll edit files over SSH, and reach `kubectl` and services through an SSH session or port-forwarding rather than directly. Steps that "just work" locally usually need a tunnel here.

When minikube exposes a service, it binds to the EC2 instance's internal network — not the public internet. To reach it from your laptop, you need an SSH tunnel.

### Get the service URL inside the EC2 instance

```bash
minikube service your-service --url
# http://192.168.49.2:30000
```

That URL only works inside the EC2 instance. Note the IP and port — you'll tunnel to that next.

### Forward it to your laptop via SSH

From your **local laptop**:

```bash
ssh -i /path/to/your-key.pem -L 8080:192.168.49.2:30000 ubuntu@<EC2_PUBLIC_IP>
```

Replace `192.168.49.2:30000` with the URL `minikube service` gave you. Now `http://localhost:8080` on your laptop hits the service inside minikube on EC2.

This works for any minikube service. If you'd rather not expose your service via NodePort, `kubectl port-forward svc/your-service 8080:80` (run inside the EC2 instance, then tunnel to it the same way) is an alternative.

## Troubleshooting

### "Requested cpu count 1 is less than the minimum allowed of 2"

You're on `t2.micro` (1 vCPU). Kubernetes requires at least 2 CPUs to run a control plane, which is why minikube enforces this minimum. Resize to `t3.medium` or larger — see the [instance sizing section](#choosing-an-ec2-instance-type) above.

### Cluster is slow or pods get OOM-killed

Your instance is RAM-starved. Check usage:

```bash
free -h
```

If you're consistently above 80% memory use, bump to `t3.large` (8 GB RAM). Workloads with heavy controllers (Prometheus, Argo, etc.) typically need 8+ GB.

### Disk fills up over time

Docker images and minikube's storage build up. Free space:

```bash
df -h /
```

If `/` is above 80%, clean up:

```bash
docker system prune -a
minikube delete   # destroys the cluster — you can recreate it in seconds with minikube start
```

You can also resize the EBS volume in AWS Console without losing the instance.


## Next steps

- Deploy your first app: [Hello minikube]({{< ref "/docs/start" >}})
- Try multi-node clusters: [multi_node tutorial]({{< ref "/docs/tutorials/multi_node" >}})
- Explore add-ons: `minikube addons list`

## Notes

{{% pageinfo %}}
minikube is **not intended for production Kubernetes hosting** — for that, use a managed or self-managed Kubernetes cluster (see the Kubernetes [production environment](https://kubernetes.io/docs/setup/production-environment/) docs). This guide treats EC2 as an extension of your local development environment, like a remote dev box.
{{% /pageinfo %}}

{{% pageinfo %}}
This page refers to third-party products and services (Amazon EC2, Ubuntu, Docker). The minikube project authors aren't responsible for those third-party products or services. See the [CNCF website guidelines](https://github.com/cncf/foundation/blob/main/policies-guidance/website-guidelines.md) for more details.
{{% /pageinfo %}}
