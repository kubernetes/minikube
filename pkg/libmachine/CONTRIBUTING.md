# Contributing to machine

[![GoDoc](https://godoc.org/github.com/docker/machine?status.png)](https://godoc.org/github.com/docker/machine)
[![Build Status](https://travis-ci.org/docker/machine.svg?branch=master)](https://travis-ci.org/docker/machine)
[![Windows Build Status](https://ci.appveyor.com/api/projects/status/github/docker/machine?svg=true)](https://ci.appveyor.com/project/dmp42/machine-fp5u5)
[![Coverage Status](https://coveralls.io/repos/docker/machine/badge.svg?branch=master&service=github)](https://coveralls.io/github/docker/machine?branch=master)

Want to hack on Machine? Awesome! Here are instructions to get you
started.

Machine is a part of the [Docker](https://www.docker.com) project, and follows
the same rules and principles. If you're already familiar with the way
Docker does things, you'll feel right at home.

Otherwise, please read [Docker's contributions
guidelines](https://github.com/docker/docker/blob/master/CONTRIBUTING.md).

# Building

The requirements to build Machine are:

1.  A running instance of Docker or a Golang 1.10 development environment
2.  The `bash` shell
3.  [Make](https://www.gnu.org/software/make/)

## Build using Docker containers

To build the `docker-machine` binary using containers, simply run:

    $ export USE_CONTAINER=true
    $ make build

## Local Go development environment

Make sure the source code directory is under a correct directory structure;
Example of cloning and preparing the correct environment `GOPATH`:

    $ mkdir docker-machine
    $ cd docker-machine
    $ export GOPATH="$PWD"
    $ go get github.com/docker/machine
    $ cd src/github.com/docker/machine

If you want to use your existing workspace, make sure your `GOPATH` is set to
the directory that contains your `src` directory, e.g.:

    $ export GOPATH=/home/yourname/work
    $ mkdir -p $GOPATH/src/github.com/docker
    $ cd $GOPATH/src/github.com/docker && git clone git@github.com:docker/machine.git
    $ cd machine

At this point, simply run:

    $ make build

## Built binary

After the build is complete a `bin/docker-machine` binary will be created.

You may call:

    $ make clean

to clean-up build results.

## Tests and validation

We use the usual `go` tools for this, to run those commands you need at least the linter which you can
install with `go get -u golang.org/x/lint/golint`

To run basic validation (dco, fmt), and the project unit tests, call:

    $ make test

If you want more indepth validation (vet, lint), and all tests with race detection, call:

    $ make validate

If you make a pull request, it is highly encouraged that you submit tests for
the code that you have added or modified in the same pull request.

## Code Coverage

To generate an html code coverage report of the Machine codebase, run:

    make coverage-serve

And navigate to <http://localhost:8000> (hit `CTRL+C` to stop the server).

### Native build

Alternatively, if you are building natively, you can simply run:

    make coverage-html


## List of all targets

### High-level targets

    make clean
    make build
    make test
    make validate

### Advanced build targets

Build for all supported OSes and architectures (binaries will be in the `bin` project subfolder):

    make build-x

Build for a specific list of OSes and architectures:

    TARGET_OS=linux TARGET_ARCH="amd64 arm" make build-x

You can further control build options through the following environment variables:

    DEBUG=true # enable debug build
    STATIC=true # build static (note: when cross-compiling, the build is always static)
    VERBOSE=true # verbose output
    PREFIX=folder # put binaries in another folder (not the default `./bin`)

Scrub build results:

    make build-clean

### Coverage targets

    make coverage-html
    make coverage-serve
    make coverage-send
    make coverage-generate
    make coverage-clean

### Tests targets

    make test-short
    make test-long
    make test-integration

### Validation targets

    make fmt
    make vet
    make lint
    make dco

### Managing dependencies

When you make a fresh copy of the repo, all the dependencies are in `vendor/` directory for the build to work.
This project uses [golang/dep](https://github.com/golang/dep) as vendor management tool. Please refer to `dep` documentation
for further details.

4. Verify the changes in your repo, commit and submit a pull request

## Integration Tests

### Setup

We use [BATS](https://github.com/sstephenson/bats) for integration testing, so,
first make sure to [install it](https://github.com/sstephenson/bats#installing-bats-from-source).

### Basic Usage

You first need to build, calling `make build`.

You can then invoke integration tests calling `DRIVER=foo make test-integration TESTSUITE`, where `TESTSUITE` is
one of the `test/integration` subfolder, and `foo` is the specific driver you want to test.

Examples:

```console
$ DRIVER=virtualbox make test-integration test/integration/core/core-commands.bats
 ✓ virtualbox: machine should not exist
 ✓ virtualbox: create
 ✓ virtualbox: ls
 ✓ virtualbox: run busybox container
 ✓ virtualbox: url
 ✓ virtualbox: ip
 ✓ virtualbox: ssh
 ✓ virtualbox: docker commands with the socket should work
 ✓ virtualbox: stop
 ✓ virtualbox: machine should show stopped after stop
 ✓ virtualbox: machine should now allow upgrade when stopped
 ✓ virtualbox: start
 ✓ virtualbox: machine should show running after start
 ✓ virtualbox: kill
 ✓ virtualbox: machine should show stopped after kill
 ✓ virtualbox: restart
 ✓ virtualbox: machine should show running after restart

17 tests, 0 failures
Cleaning up machines...
Successfully removed bats-virtualbox-test
```

To invoke a directory of tests recursively:

```console
$ DRIVER=virtualbox make test-integration test/integration/core/
...
```

### Extra Create Arguments

In some cases, for instance to test the creation of a specific base OS (e.g.
RHEL) as opposed to the default with the common tests, you may want to run
common tests with different create arguments than you get out of the box.

Keep in mind that Machine supports environment variables for many of these
flags.  So, for instance, you could run the command (substituting, of course,
the proper secrets):

    $ DRIVER=amazonec2 \
      AWS_VPC_ID=vpc-xxxxxxx \
      AWS_SECRET_ACCESS_KEY=yyyyyyyyyyyyy \
      AWS_ACCESS_KEY_ID=zzzzzzzzzzzzzzzz \
      AWS_AMI=ami-12663b7a \
      AWS_SSH_USER=ec2-user \
      make test-integration test/integration/core

in order to run the core tests on Red Hat Enterprise Linux on Amazon.

### Layout

The `test/integration` directory is laid out to divide up tests based on the
areas which the test.  If you are uncertain where to put yours, we are happy to
guide you.

At the time of writing, there is:

1.  A `core` directory which contains tests that are applicable to all drivers.
2.  A `drivers` directory which contains tests that are applicable only to
    specific drivers with sub-directories for each provider.
3.  A `cli` directory which is meant for testing functionality of the command
    line interface, without much regard for driver-specific details.

### Guidelines

The best practices for writing integration tests on Docker Machine are still a
work in progress, but here are some general guidelines from the maintainers:

1.  Ideally, each test file should have only one concern.
2.  Tests generally should not spin up more than one machine unless the test is
    deliberately testing something which involves multiple machines, such as an `ls`
    test which involves several machines, or a test intended to create and check
    some property of a Swarm cluster.
3.  BATS will print the output of commands executed during a test if the test
    fails.  This can be useful, for instance to dump the magic `$output` variable
    that BATS provides and/or to get debugging information.
4.  It is not strictly needed to clean up the machines as part of the test.  The
    BATS wrapper script has a hook to take care of cleaning up all created machines
    after each test.

# Drivers

Docker Machine has several included drivers that supports provisioning hosts
in various providers.  If you wish to contribute a driver, we ask the following
to ensure we keep the driver in a consistent and stable state:

-   Address issues filed against this driver in a timely manner
-   Review PRs for the driver
-   Be responsible for maintaining the infrastructure to run unit tests
    and integration tests on the new supported environment
-   Participate in a weekly driver maintainer meeting


Note: even if those are met does not guarantee a driver will be accepted.
If you have questions, please do not hesitate to contact us on IRC.
