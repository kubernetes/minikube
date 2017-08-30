# libvirt-go [![Build Status](https://travis-ci.org/libvirt/libvirt-go.svg?branch=master)](https://travis-ci.org/libvirt/libvirt-go) [![GoDoc](https://godoc.org/github.com/libvirt/libvirt-go?status.svg)](https://godoc.org/github.com/libvirt/libvirt-go)

Go bindings for libvirt.

Make sure to have `libvirt-dev` package (or the development files otherwise somewhere in your include path)

## Version Support

The libvirt go package provides API coverage for libvirt versions
from 1.2.0 onwards, through conditional compilation of newer APIs.

By default the binding will support APIs in libvirt.so, libvirt-qemu.so
and libvirt-lxc.so. Coverage for the latter two libraries can be dropped
from the build using build tags 'without_qemu' or 'without_lxc'
respectively.

## Development status

The Go API is considered to be production ready and aims to be kept
stable across future versions. Note, however, that the following
changes may apply to future versions:

* Existing structs can be augmented with new fields, but no existing
  fields will be changed / removed. New fields are needed when libvirt
  defines new typed parameters for various methods

* Any method with an 'flags uint32' parameter will have its parameter
  type changed to a specific typedef, if & when the libvirt API defines
  constants for the flags. To avoid breakage, always pass a literal
  '0' to any 'flags uint32' parameter, since this will auto-cast to
  any future typedef that is introduced.

## Documentation

* [api documentation for the bindings](https://godoc.org/github.com/libvirt/libvirt-go)
* [api documentation for libvirt](http://libvirt.org/html/libvirt-libvirt.html)

## Contributing

The libvirt project aims to add support for new APIs to libvirt-go
as soon as they are added to the main libvirt C library. If you
are submitting changes to the libvirt C library API, please submit
a libvirt-go change at the same time.

Bug fixes and other improvements to the libvirt-go library are
welcome at any time. The preferred submission method is to use
git send-email to submit patches to the libvir-list@redhat.com
mailing list. eg. to send a single patch

   git send-email --to libvir-list@redhat.com --subject-prefix "PATCH go" \
       --smtp-server=$HOSTNAME -1

Or to send all patches on the current branch, against master

   git send-email --to libvir-list@redhat.com --subject-prefix "PATCH go" \
       --smtp-server=$HOSTNAME --no-chain-reply-to --cover-letter --annotate \
       master..

Note the master GIT repository is at

* http://libvirt.org/git/?p=libvirt-go.git;a=summary

The following automatic read-only mirrors are available as a
convenience to allow contributors to "fork" the repository:

* https://gitlab.com/libvirt/libvirt-go
* https://github.com/libvirt/libvirt-go

While you can send pull-requests to these mirrors, they will be
re-submitted via emai to the mailing list for review before
being merged, unless they are trivial/obvious bug fixes.

## Testing

The core API unit tests are all written to use the built-in
test driver (test:///default), so they have no interaction
with the host OS environment.

Coverage of libvirt C library APIs / constants is verified
using automated tests. These can be run by passing the 'api'
build tag. eg  go test -tags api

For areas where the test driver lacks functionality, it is
possible to use the QEMU or LXC drivers to exercise code.
Such tests must be part of the 'integration_test.go' file
though, which is only run when passing the 'integration'
build tag. eg  go test -tags integration

In order to run the unit tests, libvirtd should be configured
to allow your user account read-write access with no passwords.
This can be easily done using polkit config files

```
# cat > /etc/polkit-1/localauthority/50-local.d/50-libvirt.pkla  <<EOF
[Passwordless libvirt access]
Identity=unix-group:berrange
Action=org.libvirt.unix.manage
ResultAny=yes
ResultInactive=yes
ResultActive=yes
EOF
```

(Replace 'berrange' with your UNIX user name).

One of the integration tests also requires that libvirtd is
listening for TCP connections on localhost, with sasl auth
This can be setup by editing /etc/libvirt/libvirtd.conf to
set

```
  listen_tls=0
  listen_tcp=1
  auth_tcp=sasl
  listen_addr="127.0.0.1"
```

and then start libvirtd with the --listen flag (this can
be set in /etc/sysconfig/libvirtd to make it persistent).

Then create a sasl user

```
   saslpasswd2 -a libvirt user
```

and enter "pass" as the password.

Alternatively a `Vagrantfile`, requiring use of virtualbox,
is included to run the integration tests:

* `cd ./vagrant`
* `vagrant up` to provision the virtual machine
* `vagrant ssh` to login to the virtual machine

Once inside, `sudo su -` and `go test -tags integration libvirt`.
