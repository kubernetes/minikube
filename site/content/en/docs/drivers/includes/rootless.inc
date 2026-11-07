### Increase inotify limits

When using a rootless driver, increasing the host inotify limits may be required
for workloads running multiple containers:

```shell
sudo tee /etc/sysctl.d/99-minikube-rootless.conf <<EOF
fs.inotify.max_user_watches = 524288
fs.inotify.max_user_instances = 512
EOF

sudo sysctl --system
```

### Allow binding to privileged ports

When using a rootless driver, workloads that need to bind to privileged ports
such as ports 80 or 443 may require allowing unprivileged users to bind to
ports starting at 80:

```shell
sudo tee /etc/sysctl.d/99-minikube-rootless-ports.conf <<EOF
net.ipv4.ip_unprivileged_port_start=80
EOF

sudo sysctl --system
```
