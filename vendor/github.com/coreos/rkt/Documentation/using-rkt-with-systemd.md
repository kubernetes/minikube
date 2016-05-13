# Using rkt with systemd

`rkt` is designed to cooperate with init systems, like [`systemd`][systemd]. rkt implements a simple CLI that directly executes processes, and does not interpose a long-running daemon, so the lifecycle of rkt pods can be directly managed by systemd. Standard systemd idioms like `systemctl start` and `systemctl stop` work out of the box.

In the shell excerpts below, a `#` prompt indicates commands that require root privileges, while the `$` prompt denotes commands issued as an unprivileged user.

## systemd-run

The [`systemd-run`][systemd-run] utility is a convenient shortcut for testing a service before making it permanent in a unit file. To start a "daemonized" container that forks the container processes into the background, wrap the invocation of `rkt` with `systemd-run`:

```
# systemd-run --slice=machine rkt run coreos.com/etcd:v2.0.10
Running as unit run-3247.service.
```

The `--slice=machine` option to `systemd-run` places the service in `machine.slice` rather than the host's `system.slice`, isolating containers in their own cgroup area.

Invoking a rkt container through systemd-run in this way creates a transient service unit that can be managed with the usual systemd tools:

```
$ systemctl status run-3247.service
● run-3247.service - /home/iaguis/work/coreos/go/src/github.com/coreos/rkt/build-rkt/bin/rkt run coreos.com/etcd:v2.0.10
   Loaded: loaded
  Drop-In: /run/systemd/system/run-3247.service.d
           └─50-Description.conf, 50-ExecStart.conf
   Active: active (running) since Mon 2015-10-26 17:38:06 CET; 41s ago
 Main PID: 3254 (ld-linux-x86-64)
   CGroup: /system.slice/run-3247.service
           ├─3254 stage1/rootfs/usr/lib/ld-linux-x86-64.so.2 stage1/rootfs/usr/bin/systemd-nspawn --boot --register=true --link-jou...
           ├─3321 /usr/lib/systemd/systemd --default-standard-output=tty --log-target=null --log-level=warning --show-status=0
           └─system.slice
             ├─etcd.service
             │ └─3326 /etcd
             └─systemd-journald.service
               └─3322 /usr/lib/systemd/systemd-journald
```

Since every pod is registered with `machined` with a machine name of the form `rkt-$UUID`, the systemd tools can inspect pod logs, or stop and restart pod "machines". Use the `machinectl` tool to print the list of rkt pods:

```
$ machinectl list
MACHINE                                  CLASS     SERVICE
rkt-f0261476-7044-4a84-b729-e0f7a47dcffe container nspawn

1 machines listed.
```

Given the name of this rkt machine, `journalctl` can inspect its logs, or `machinectl` can shut it down:

```
# journalctl -M rkt-f0261476-7044-4a84-b729-e0f7a47dcffe
Oct 26 17:38:11 locke rkt[3254]: [25966.375411] etcd[4]: 2015/10/26 16:38:11 raft: ce2a822cea30bfca became follower at term 0
Oct 26 17:38:11 locke rkt[3254]: [25966.375685] etcd[4]: 2015/10/26 16:38:11 raft: newRaft ce2a822cea30bfca [peers: [], ter...term: 0]
Oct 26 17:38:11 locke rkt[3254]: [25966.375942] etcd[4]: 2015/10/26 16:38:11
# machinectl poweroff rkt-f0261476-7044-4a84-b729-e0f7a47dcffe
$ machinectl list
MACHINE CLASS SERVICE

0 machines listed.
```

## Managing pods as systemd services

### Simple Unit File

The following is a simple example of a unit file using `rkt` to run an `etcd` instance under systemd service management:

```
[Unit]
Description=etcd

[Service]
ExecStart=/usr/bin/rkt run --mds-register=false coreos.com/etcd:v2.0.10
KillMode=mixed
Restart=always
```

This unit can now be managed using the standard `systemctl` commands:

```
# systemctl start etcd.service
# systemctl stop etcd.service
# systemctl restart etcd.service
# systemctl enable etcd.service
# systemctl disable etcd.service
```

Note that no `ExecStop` clause is required. Setting [`KillMode=mixed`][systemd-killmode-mixed] means that running `systemctl stop etcd.service` will send `SIGTERM` to `stage1`'s `systemd`, which in turn will initiate orderly shutdown inside the pod. Systemd is additionally able to send the cleanup `SIGKILL` to any lingering service processes, after a timeout. This comprises complete pod lifecycle management with familiar, well-known system init tools.

### Advanced Unit File

A more advanced unit example takes advantage of a few convenient `systemd` features:

1. Inheriting environment variables specified in the unit with `--inherit-env`. This feature helps keep units concise, instead of layering on many flags to `rkt run`.
2. Using the dependency graph to start our pod after networking has come online. This is helpful if your application requires outside connectivity to fetch remote configuration (for example, from `etcd`).
3. Set resource limits for this `rkt` pod. This can also be done in the unit file, rather than flagged to `rkt run`.

Here is what it looks like all together:

```
[Unit]
# Metadata
Description=MyApp
Documentation=https://myapp.com/docs/1.3.4
# Wait for networking
Requires=network-online.target
After=network-online.target

[Service]
Slice=machine.slice
# Resource limits
Delegate=true
CPUShares=512
MemoryLimit=1G
# Env vars
Environment=HTTP_PROXY=192.0.2.3:5000
Environment=STORAGE_PATH=/opt/myapp
Environment=TMPDIR=/var/tmp
# Fetch the app (not strictly required, `rkt run` will fetch the image if there is not one)
ExecStartPre=/usr/bin/rkt fetch myapp.com/myapp-1.3.4
# Start the app
ExecStart=/usr/bin/rkt run --inherit-env --port=http:8888 myapp.com/myapp-1.3.4
KillMode=mixed
Restart=always
```

rkt must be the main process of the service in order to support [isolators][systemd-isolators] correctly and to be well-integrated with [systemd-machined][systemd-machined]. To ensure that rkt is the main process of the service, the pattern `/bin/sh -c "foo ; rkt run ..."` should be avoided, because in that case the main process is `sh`.

In most cases, the parameters `Environment=` and `ExecStartPre=` can simply be used instead of starting a shell. If shell invocation is unavoidable, use `exec` to ensure rkt replaces the preceding shell process:

```
ExecStart=/bin/sh -c "foo ; exec rkt run ..."
```

### Socket-activated service

`rkt` supports [socket-activated services][systemd-socket-activated]. This means systemd will listen on a port on behalf of a container, and start the container when receiving a connection. An application needs to be able to accept sockets from systemd's native socket passing interface in order to handle socket activation.

To make socket activation work, add a [socket-activated port][aci-socketActivated] to the app container manifest:

```json
...
{
...
    "app": {
        ...
        "ports": [
            {
                "name": "8080-tcp",
                "protocol": "tcp",
                "port": 8080,
                "count": 1,
                "socketActivated": true
            }
        ]
    }
}
```

Then you will need a pair of `.service` and `.socket` unit files. The socket unit file will have the same port you've set in the image manifest of your app, and the service unit file will run `rkt`:

```
# my-socket-activated-app.socket
[Unit]
Description=My socket-activated app's socket

[Socket]
ListenStream=8080
```

```
# my-socket-activated-app.service
[Unit]
Description=My socket-activated app

[Service]
ExecStart=/usr/bin/rkt run myapp.com/my-socket-activated-app:v1.0
KillMode=mixed
```

Finally, start the socket unit:

```
# systemctl start my-socket-activated-app.socket
$ systemctl status my-socket-activated-app.socket
● my-socket-activated-app.socket - My socket-activated app's socket
   Loaded: loaded (/etc/systemd/system/my-socket-activated-app.socket; static; vendor preset: disabled)
   Active: active (listening) since Thu 2015-07-30 12:24:50 CEST; 2s ago
   Listen: [::]:8080 (Stream)

Jul 30 12:24:50 locke-work systemd[1]: Listening on My socket-activated app's socket.
Jul 30 12:24:50 locke-work systemd[1]: Starting My socket-activated app's socket.
```

Now a new connection to port 8080 will start your container to handle the request.

## Other tools for managing pods

Let us assume the service from the simple example unit file, above, is started on the host.

### ps auxf

The snippet below taken from output of `ps auxf` shows several things:

1. `rkt` `exec`s stage1's `systemd-nspawn` instead of using `fork-exec` technique. That is why rkt itself is not listed by `ps`.
2. `systemd-nspawn` runs a typical boot sequence - it spawns `systemd` inside the container, which in turn spawns our desired service(s).
3. There can be also other services running, which may be `systemd`-specific, like `systemd-journald`.

```
$ ps auxf
USER       PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND
root      7258  0.2  0.0  19680  2664 ?        Ss   12:38   0:02 stage1/rootfs/usr/lib/ld-linux-x86-64.so.2 stage1/rootfs/usr/bin/systemd-nspawn --boot --register=true --link-journal=try-guest --quiet --keep-unit --uuid=6d0d9608-a744-4333-be21-942145a97a5a --machine=rkt-6d0d9608-a744-4333-be21-942145a97a5a --directory=stage1/rootfs -- --default-standard-output=tty --log-target=null --log-level=warning --show-status=0
root      7275  0.0  0.0  27348  4316 ?        Ss   12:38   0:00  \_ /usr/lib/systemd/systemd --default-standard-output=tty --log-target=null --log-level=warning --show-status=0
root      7277  0.0  0.0  23832  6100 ?        Ss   12:38   0:00      \_ /usr/lib/systemd/systemd-journald
root      7343  0.3  0.0  10652  7332 ?        Ssl  12:38   0:04      \_ /etcd
```

### systemd-cgls

The `systemd-cgls` command prints the list of cgroups active on the system. The inner `system.slice` shown in the excerpt below is a cgroup in rkt's `stage1`, below which an in-container systemd has been started to shepherd pod apps with complete process lifecycle management:

```
$ systemd-cgls
├─1 /usr/lib/systemd/systemd --switched-root --system --deserialize 21
machine.slice
│ ├─etcd.service
│ │ ├─7258 stage1/rootfs/usr/lib/ld-linux-x86-64.so.2 stage1/rootfs/usr/bin/systemd-nspawn --boot --register=true --link-journal=try-guest --quiet --keep-unit --uuid=6d0d9608-a744-4333-be21-942145a97a5a --machine=rkt-6d0d9608-a744-4333-be21-942145a97a5a --directory=stage1/rootfs -- --default-standard-output=tty --log-target=null --log-level=warning --show-status=0
│ │ ├─7275 /usr/lib/systemd/systemd --default-standard-output=tty --log-target=null --log-level=warning --show-status=0
│ │ └─system.slice
│ │   ├─systemd-journald.service
│ │   │ └─7277 /usr/lib/systemd/systemd-journald
│ │   └─etcd.service
│ │     └─7343 /etcd
```

### systemd-cgls --all

To display all active cgroups, use the `--all` flag. This will show two cgroups for `mount` in the host's `system.slice`. One mount cgroup is for the `stage1` root filesystem, the other for the `stage2` root (the pod's filesystem). Inside the pod's `system.slice` there are more `mount` cgroups -- mostly for bind mounts of standard `/dev`-tree device files.

```
$ systemd-cgls --all
├─1 /usr/lib/systemd/systemd --switched-root --system --deserialize 21
├─system.slice
│ ├─var-lib-rkt-pods-run-6d0d9608\x2da744\x2d4333\x2dbe21\x2d942145a97a5a-stage1-rootfs.mount
│ ├─var-lib-rkt-pods-run-6d0d9608\x2da744\x2d4333\x2dbe21\x2d942145a97a5a-stage1-rootfs-opt-stage2-etcd-rootfs.mount
│ ├─etcd.service
│ │ ├─7258 stage1/rootfs/usr/lib/ld-linux-x86-64.so.2 stage1/rootfs/usr/bin/systemd-nspawn --boot --register=true --link-journal=try-guest --quiet --keep-unit --uuid=6d0d9608-a744-4333-be21-942145a97a5a --machine=rkt-6d0d9608-a744-4333-be21-942145a97a5a --directory=stage1/rootfs -- --default-standard-output=tty --log-target=null --log-level=warning --show-status=0
│ │ ├─7275 /usr/lib/systemd/systemd --default-standard-output=tty --log-target=null --log-level=warning --show-status=0
│ │ └─system.slice
│ │   ├─proc-sys-kernel-random-boot_id.mount
│ │   ├─opt-stage2-etcd-rootfs-dev-random.mount
│ │   ├─opt-stage2-etcd-rootfs-dev-net-tun.mount
│ │   ├─-.mount
│ │   ├─system-prepare\x2dapp.slice
│ │   ├─opt-stage2-etcd-rootfs-dev-pts.mount
│ │   ├─opt-stage2-etcd-rootfs-sys.mount
│ │   ├─tmp.mount
│ │   ├─opt-stage2-etcd-rootfs.mount
│ │   ├─systemd-journald.service
│ │   │ └─7277 /usr/lib/systemd/systemd-journald
│ │   ├─opt-stage2-etcd-rootfs-proc.mount
│ │   ├─opt-stage2-etcd-rootfs-dev-urandom.mount
│ │   ├─etcd.service
│ │   │ └─7343 /etcd
│ │   ├─opt-stage2-etcd-rootfs-dev-tty.mount
│ │   ├─opt-stage2-etcd-rootfs-dev-console.mount
│ │   ├─run-systemd-nspawn-incoming.mount
│ │   ├─opt-stage2-etcd-rootfs-dev-zero.mount
│ │   ├─exit-watcher.service
│ │   ├─opt-stage2-etcd-rootfs-dev-null.mount
│ │   ├─opt-stage2-etcd-rootfs-dev-full.mount
│ │   └─opt-stage2-etcd-rootfs-dev-shm.mount
```


[aci-socketActivated]: https://github.com/appc/spec/blob/master/spec/aci.md#image-manifest-schema
[systemd]: http://www.freedesktop.org/wiki/Software/systemd/
[systemd-isolators]: https://github.com/appc/spec/blob/master/spec/ace.md#isolators
[systemd-killmode-mixed]: http://www.freedesktop.org/software/systemd/man/systemd.kill.html#KillMode=
[systemd-machined]: http://www.freedesktop.org/software/systemd/man/systemd-machined.service.html
[systemd-run]: http://www.freedesktop.org/software/systemd/man/systemd-run.html
[systemd-socket-activated]: http://www.freedesktop.org/software/systemd/man/sd_listen_fds.html
