<!--[metadata]>
+++
title = "Install Machine"
description = "How to install Docker Machine"
keywords = ["machine, orchestration, install, installation, docker, documentation"]
[menu.main]
parent="workw_machine"
weight=-80
+++
<![end-metadata]-->

# Install Docker Machine

On OS X and Windows, Machine is installed along with other Docker products when
you install the Docker Toolbox. For details on installing Docker Toolbox, see
the <a href="https://docs.docker.com/installation/mac/" target="_blank">Mac OS X
installation</a> instructions or <a
href="https://docs.docker.com/installation/windows" target="_blank">Windows
installation</a> instructions.

If you want only Docker Machine, you can install the Machine binaries directly by following the instructions in the next section. You can find the latest versions of the binaries are on the <a href="https://github.com/docker/machine/releases/" target="_blank"> docker/machine release page</a> on GitHub.

## Installing Machine Directly

1.  Install <a href="https://docs.docker.com/installation/"
    target="_blank">the Docker binary</a>.

2.  Download the Docker Machine binary and extract it to your PATH.

    If you are running OS X or Linux:

        $ curl -L https://github.com/docker/machine/releases/download/v0.6.0/docker-machine-`uname -s`-`uname -m` > /usr/local/bin/docker-machine && \
        chmod +x /usr/local/bin/docker-machine

    If you are running Windows with git bash

        $ if [[ ! -d "$HOME/bin" ]]; then mkdir -p "$HOME/bin"; fi && \
        curl -L https://github.com/docker/machine/releases/download/v0.6.0/docker-machine-Windows-x86_64.exe > "$HOME/bin/docker-machine.exe" && \
        chmod +x "$HOME/bin/docker-machine.exe"

    Otherwise, download one of the releases from the <a href="https://github.com/docker/machine/releases/" target="_blank"> docker/machine release page</a> directly.

3.  Check the installation by displaying the Machine version:

        $ docker-machine version
        docker-machine version 0.6.0, build 61388e9

## Installing bash completion scripts

The Machine repository supplies several `bash` scripts that add features such
as:

-   command completion
-   a function that displays the active machine in your shell prompt
-   a function wrapper that adds a `docker-machine use` subcommand to switch the
    active machine

To install the scripts, copy or link them into your `/etc/bash_completion.d` or
`/usr/local/etc/bash_completion.d` directory. To enable the `docker-machine` shell
prompt, add `$(__docker_machine_ps1)` to your `PS1` setting in `~/.bashrc`.

    PS1='[\u@\h \W$(__docker_machine_ps1)]\$ '

You can find additional documentation in the comments at the <a href="https://github.com/docker/machine/tree/master/contrib/completion/bash" target="_blank">top of each script</a>.

## Where to go next

-   [Docker Machine overview](overview.md)
-   Create and run a Docker host on your [local system using VirtualBox](get-started.md)
-   Provision multiple Docker hosts [on your cloud provider](get-started-cloud.md)
-   <a href="https://docs.docker.com/machine/drivers/" target="_blank">Docker Machine driver reference</a>
-   <a href="https://docs.docker.com/machine/reference/" target="_blank">Docker Machine subcommand reference</a>
