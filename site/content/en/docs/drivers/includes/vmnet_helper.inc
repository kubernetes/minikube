### Install vmnet-helper

```shell
curl -fsSL https://github.com/minikube-machine/vmnet-helper/releases/latest/download/install.sh | bash
```

The command downloads the latest release from github and installs it to
`/opt/vmnet-helper`.

### Grant permission to run vmnet-helper manually (if said no to script above)

vmnet-helper must run as root to create a vmnet interface. To let users in the
staff group run it without a password, install the default sudoers rule. The
install script offers to add this automatically; if you declined, run the command
below manually:

```shell
sudo install -m 0640 /opt/vmnet-helper/share/doc/vmnet-helper/sudoers.d/vmnet-helper /etc/sudoers.d/
```

You can change the sudoers configuration to allow access to specific
users or other groups.

**IMPORTANT**: The vmnet-helper executable and the directory where it is
installed must be owned by root and may not be modifiable by
unprivileged users.
