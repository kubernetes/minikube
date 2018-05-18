# Docker Machine Parallels Driver

This is a plugin for [Docker Machine](https://docs.docker.com/machine/) allowing
to create Docker hosts locally on [Parallels Desktop for Mac](http://www.parallels.com/products/desktop/)

## Requirements
* OS X 10.9+
* [Docker Machine](https://docs.docker.com/machine/) 0.5.1+ (is bundled to
  [Docker Toolbox](https://www.docker.com/docker-toolbox) 1.9.1+)
* [Parallels Desktop](http://www.parallels.com/products/desktop/) 11.0.0+ **Pro** or
**Business** edition (_Standard edition is not supported!_)

## Installation
Install via Homebrew:

```console
$ brew install docker-machine-parallels
```

To install this plugin manually, download the binary `docker-machine-driver-parallels`
and  make it available by `$PATH`, for example by putting it to `/usr/local/bin/`:

```console
$ curl -L https://github.com/Parallels/docker-machine-parallels/releases/download/v1.3.0/docker-machine-driver-parallels > /usr/local/bin/docker-machine-driver-parallels

$ chmod +x /usr/local/bin/docker-machine-driver-parallels
```

The latest version of `docker-machine-driver-parallels` binary is available on
the ["Releases"](https://github.com/Parallels/docker-machine-parallels/releases) page.

## Usage
Official documentation for Docker Machine [is available here](https://docs.docker.com/machine/).

To create a Parallels Desktop virtual machine for Docker purposes just run this
command:

```
$ docker-machine create --driver=parallels prl-dev
```

Available options:

 - `--parallels-boot2docker-url`: The URL of the boot2docker image.
 - `--parallels-disk-size`: Size of disk for the host VM (in MB).
 - `--parallels-memory`: Size of memory for the host VM (in MB).
 - `--parallels-cpu-count`: Number of CPUs to use to create the VM (-1 to use the number of CPUs available).
 - `--parallels-video-size`: Size of video memory for host (in MB).
 - `--parallels-no-share`: Disable the sharing of `/Users` directory.
 - `--parallels-nested-virutalization`: Enable nested virtualization.

The `--parallels-boot2docker-url` flag takes a few different forms. By
default, if no value is specified for this flag, Machine will check locally for
a boot2docker ISO. If one is found, that will be used as the ISO for the
created machine. If one is not found, the latest ISO release available on
[boot2docker/boot2docker](https://github.com/boot2docker/boot2docker) will be
downloaded and stored locally for future use. Note that this means you must run
`docker-machine upgrade` deliberately on a machine if you wish to update the "cached"
boot2docker ISO.

This is the default behavior (when `--parallels-boot2docker-url=""`), but the
option also supports specifying ISOs by the `http://` and `file://` protocols.

Environment variables and default values:

| CLI option                          | Environment variable        | Default                  |
|-------------------------------------|-----------------------------|--------------------------|
| `--parallels-boot2docker-url`       | `PARALLELS_BOOT2DOCKER_URL` | *Latest boot2docker url* |
| `--parallels-cpu-count`             | `PARALLELS_CPU_COUNT`       | `1`                      |
| `--parallels-disk-size`             | `PARALLELS_DISK_SIZE`       | `20000`                  |
| `--parallels-memory`                | `PARALLELS_MEMORY_SIZE`     | `1024`                   |
| `--parallels-video-size`            | `PARALLELS_VIDEO_SIZE`      | `64`                     |
| `--parallels-no-share`              | -                           | `false`                  |
| `--parallels-nested-virtualization` | -                           | `false`                  |

## Development

### Build from Source
If you wish to work on Parallels Driver for Docker machine, you'll first need
[Go](http://www.golang.org) installed (version 1.7+ is required).
Make sure Go is properly installed, including setting up a [GOPATH](http://golang.org/doc/code.html#GOPATH).

Run these commands to build the plugin binary:

```bash
$ go get -d github.com/Parallels/docker-machine-parallels
$ cd $GOPATH/src/github.com/Parallels/docker-machine-parallels
$ make build
```

After the build is complete, `bin/docker-machine-driver-parallels` binary will
be created. If you want to copy it to the `${GOPATH}/bin/`, run `make install`.

### Acceptance Tests

We use [Bats](https://github.com/sstephenson/bats) for acceptance testing, so,
[install it](https://github.com/sstephenson/bats#installing-bats-from-source) first.

You also need to build the plugin binary by calling `make build`.

Then you can run acceptance tests using this command:

```bash
$ make test-acceptance
```

Acceptance tests will invoke the general `docker-machine` binary available by
`$PATH`. If you want to specify it explicitly, just set `MACHINE_BINARY` env variable:

```bash
$ MACHINE_BINARY=/path/to/docker-machine make test-acceptance
```

## Authors

* Mikhail Zholobov ([@legal90](https://github.com/legal90))
* Rickard von Essen ([@rickard-von-essen](https://github.com/rickard-von-essen))
