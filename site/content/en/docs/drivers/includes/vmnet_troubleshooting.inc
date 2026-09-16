### Cannot connect to the VM on macOS

On macOS 15 or later, `minikube start` can fail while waiting for the SSH
server, even though the VM boots and gets an IP address. For example, with the
`vfkit` driver:

```shell
minikube start --driver vfkit
```

```
😄  minikube v1.39.0 on Darwin 26.6.2 (arm64)
✨  Using the vfkit driver based on user configuration
🌐  Automatically selected the vmnet-shared network
👍  Starting "minikube" primary control-plane node in "minikube" cluster
🔥  Creating vfkit VM (CPUs=2, Memory=6144MB, Disk=20000MB) ...
🔥  Deleting "minikube" in vfkit ...
🤦  StartHost failed, but will try again: creating host: create: creating: timeout waiting for SSH server "192.168.64.3:22"
🔥  Creating vfkit VM (CPUs=2, Memory=6144MB, Disk=20000MB) ...
😿  Failed to start vfkit VM. Running "minikube delete" may fix it: creating host: create: creating: timeout waiting for SSH server "192.168.64.3:22"

❌  Exiting due to GUEST_PROVISION: error provisioning guest: Failed to start host: creating host: create: creating: timeout waiting for SSH server "192.168.64.3:22"
```

Starting with `--alsologtostderr -v=7` shows that the VM gets an IP address from
DHCP, but every connection to it fails immediately:

```
libmachine: IP: 192.168.64.3
libmachine: Waiting until SSH server "192.168.64.3:22" is accessible
libmachine: Dialing to SSH server "192.168.64.3:22"
libmachine: Failed to dial: dial tcp 192.168.64.3:22: connect: no route to host
```

macOS 15 introduced a privacy control for the local network. minikube connects
directly to the IP address of the VM, which is on a local network, so you must
allow the application you run minikube from to access the local network.

Commands you run in Terminal.app always have local network access, and the
**Local Network** entry for Terminal.app has no effect. Every other terminal
needs the permission, including third-party terminals such as iTerm2 or Ghostty
and the terminals embedded in editors and IDEs such as Visual Studio Code.

If you clicked **Don't Allow** when macOS asked for this permission, macOS does
not ask again. The permission can be off even though you do not remember denying
it, possibly from long before you installed minikube.

To fix this:

1. Choose Apple menu > **System Settings**, then click **Privacy & Security**
   in the sidebar.
2. Click **Local Network**.
3. Find the terminal application or IDE you run minikube from in the list, and
   turn it on.

Then start the cluster again.

For more information, see [Control access to your local network on
Mac](https://support.apple.com/guide/mac-help/mchla4f49138/mac).
