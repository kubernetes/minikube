## Installation

If you're a Mac or Windows user, the [Docker Toolbox](https://www.docker.com/docker-toolbox) will install Docker Machine {{VERSION}} for you, alongside the latest versions of the Docker Engine, Compose and Kitematic.

You can use the usual commands to install or upgrade:

On OS X
```console
$ curl -L https://github.com/docker/machine/releases/download/{{VERSION}}/docker-machine-`uname -s`-`uname -m` >/usr/local/bin/docker-machine && \
  chmod +x /usr/local/bin/docker-machine
```
On Linux
```console
$ curl -L https://github.com/docker/machine/releases/download/{{VERSION}}/docker-machine-`uname -s`-`uname -m` >/tmp/docker-machine &&
    chmod +x /tmp/docker-machine &&
    sudo cp /tmp/docker-machine /usr/local/bin/docker-machine
```
On Windows with git bash
```console
$ if [[ ! -d "$HOME/bin" ]]; then mkdir -p "$HOME/bin"; fi && \
curl -L https://github.com/docker/machine/releases/download/{{VERSION}}/docker-machine-Windows-x86_64.exe > "$HOME/bin/docker-machine.exe" && \
chmod +x "$HOME/bin/docker-machine.exe"
```

Otherwise, download one of the releases from the [release page](https://github.com/docker/machine/releases/) directly.

See the install [docs](https://docs.docker.com/machine/install-machine/) for more install options and instructions.

## Changelog

*Edit the changelog below by hand*

{{CHANGELOG}}

## Thank You

Thank you very much to our active users and contributors. If you have filed detailed bug reports, THANK YOU!
Please continue to do so if you encounter any issues. It's your hard work that makes Docker Machine better.

The following authors contributed changes to this release:

{{CONTRIBUTORS}}

Great thanks to all of the above! We appreciate it. Keep up the great work!

## Checksums

{{CHECKSUM}}

