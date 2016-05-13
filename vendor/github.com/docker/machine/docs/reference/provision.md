<!--[metadata]>
+++
title = "provision"
description = "Re-run provisioning on a created machine."
keywords = ["machine, provision, subcommand"]
[menu.main]
parent="smn_machine_subcmds"
+++
<![end-metadata]-->

# provision

Re-run provisioning on a created machine.

Sometimes it may be helpful to re-run Machine's provisioning process on a
created machine.  Reasons for doing so may include a failure during the original
provisioning process, or a drift from the desired system state (including the
originally specified Swarm or Engine configuration).

Usage is `docker-machine provision [name]`.  Multiple names may be specified.

    $ docker-machine provision foo bar
    Copying certs to the local machine directory...
    Copying certs to the remote machine...
    Setting Docker configuration on the remote daemon...

The Machine provisioning process will:

1.  Set the hostname on the instance to the name Machine addresses it by (e.g.
    `default`).
2.  Install Docker if it is not present already.
3.  Generate a set of certificates (usually with the default, self-signed CA) and
    configure the daemon to accept connections over TLS.
4.  Copy the generated certificates to the server and local config directory.
5.  Configure the Docker Engine according to the options specified at create
    time.
6.  Configure and activate Swarm if applicable.
