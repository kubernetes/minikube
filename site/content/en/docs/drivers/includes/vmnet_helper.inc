### Install vmnet-helper

On macOS 26 or later, install vmnet-helper using [Homebrew](https://brew.sh/):

```shell
brew tap nirs/vmnet-helper
brew trust nirs/vmnet-helper
brew install vmnet-helper
```

On macOS 15 or earlier, install the latest version using the install script:

```shell
curl -fsSL https://github.com/minikube-machine/vmnet-helper/releases/latest/download/install.sh | bash
```

The command downloads the latest release from GitHub and installs it to
`/opt/vmnet-helper`, and configures a sudoers rule to allow running
vmnet-helper without a password.

### Grant permission to run vmnet-helper manually on macOS 15 or earlier

On macOS 15 or earlier, vmnet-helper must run as root to create a vmnet
interface. To let users in the staff group run it without a password, install the
default sudoers rule. The install script offers to add this automatically; if you
declined, run the command below manually:

```shell
sudo install -m 0640 /opt/vmnet-helper/share/doc/vmnet-helper/sudoers.d/vmnet-helper /etc/sudoers.d/
```

You can change the sudoers configuration to allow access to specific
users or other groups.

**IMPORTANT**: On macOS 15 or earlier, the vmnet-helper executable and the
directory where it is installed must be owned by root and may not be modifiable
by unprivileged users.
