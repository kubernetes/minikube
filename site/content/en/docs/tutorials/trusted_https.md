---
title: "Zero-Configuration Trusted HTTPS with minikube"
linkTitle: "Trusted HTTPS Setup"
weight: 1
---

## Overview

minikube provides zero-configuration trusted HTTPS access for Kubernetes applications running in your local cluster.

When starting minikube with the `--https` flag:
1. **Root CA Generation**: An in-memory Certificate Authority (CA) is generated (`minikube-<profile>`) and signs a server TLS certificate for the cluster's IP address and Subject Alternative Names (SANs).
2. **Host OS Trust Store Integration**: The Root CA is automatically installed into your operating system's trust store (Linux NSS database `~/.pki/nssdb` with `TCu,cu,cu`, macOS Keychain, Windows Certificate Store) so web browsers and tools trust the cluster certificates without security warnings.
3. **Cluster TLS Secret Provisioning**: A standard Kubernetes TLS Secret named `minikube-tls` (containing `tls.crt` and `tls.key`) is created in `kube-system` and automatically synced into target namespaces.
4. **Ingress Host Discovery**: minikube automatically discovers hostnames in cluster `Ingress` resources, re-issues server certificates with all disovered SANs, updates `/etc/hosts`, and configures Traefik's default `TLSStore`.

---

## Complete Tutorial & Workflow Guide

### Step 1: Start minikube with `--https`

```bash
minikube start --https
```

Output:
```text
✅ Installed trusted CA minikube-minikube in the user trust store
✅ Installed TLS certificate for 192.168.49.2 in the cluster
```

> **Note for Web Browsers**: Browsers (Chrome / Edge) cache NSS trust databases in memory while running. Restart your browser once after cluster startup to reload the updated trust store.

---

### Step 2: Enable the Traefik Ingress Addon

Enable Traefik as the cluster Ingress Controller:

```bash
minikube addons enable traefik
```

---

### Step 3: Deploy an Application and Ingress

Deploy an example application and expose it with a standard Kubernetes `Ingress` resource:

```bash
cat << 'EOF' | kubectl apply -f -
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: myapp-ingress
  namespace: default
spec:
  ingressClassName: traefik
  rules:
  - host: traefik.example
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: podinfo-ui-svc
            port:
              number: 9898
EOF
```

> **Note**: You do not even need to specify a `tls:` section in your Ingress manifest! minikube configures Traefik's default `TLSStore` to automatically serve the trusted `minikube-tls` certificate for all HTTPS traffic.

---

### Step 4: Verify Trusted HTTPS Access

Minikube automatically syncs discovered Ingress hostnames to `/etc/hosts`. You can test access via `curl` or in your web browser:

```bash
curl -v https://traefik.example
```

Example Output:
```text
* Hostname traefik.example was found in DNS cache
* SSL connection using TLSv1.3 / TLS_AES_128_GCM_SHA256
* Server certificate:
*   subject: O=minikube; CN=minikube-minikube-server
*   issuer: O=minikube; CN=minikube-minikube
*   subjectAltName: "traefik.example" matches cert's "traefik.example"
* SSL certificate verified via OpenSSL.
* Established connection to traefik.example (192.168.49.2 port 443) using HTTP/2

< HTTP/2 200 OK
```

Open [**https://traefik.example**](https://traefik.example) in your browser — it connects securely with a valid green lock and zero security warnings.

---

### Step 5: Profile Deletion & Cleanup

When deleting the cluster profile, minikube automatically removes the CA certificate from your host operating system's trust store to leave your host clean:

```bash
minikube delete
```
