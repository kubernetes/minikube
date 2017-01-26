# libvirt-go [![Build Status](https://travis-ci.org/rgbkrk/libvirt-go.svg?branch=master)](https://travis-ci.org/rgbkrk/libvirt-go) [![GoDoc](https://godoc.org/gopkg.in/alexzorin/libvirt-go.v2?status.svg)](http://godoc.org/gopkg.in/alexzorin/libvirt-go.v2)

Go bindings for libvirt.

Make sure to have `libvirt-dev` package (or the development files otherwise somewhere in your include path)

## Version Support

The minimum supported version of libvirt is **1.2.2**. Due to the
API/ABI compatibility promise of libvirt, more recent versions of
libvirt should work too.

Some features require a more recent version of libvirt. They are
disabled by default. If you want to enable them, build using one of
those additional tags (you need to use only the most recent one you
are interested in):

 - **1.2.14**

For example:

    go build -tags libvirt.1.2.14

### OS Compatibility Matrix

To quickly see what version of libvirt your OS can easily support (may be outdated). Obviously, nothing below 1.2.2 is usable with these bindings.

| OS Release   | libvirt Version                |
| ------------ | ------------------------------ |
| FC19         | 1.2.9 from libvirt.org/sources |
| Debian 8     | 1.2.9 from jessie              |
| Debian 7     | 1.2.9 from wheezy-backports    |
| Ubuntu 14.04 | 1.2.2 from trusty              |
| Ubuntu 16.04 | 1.3.1 from xenial              |
| RHEL 7       | 1.2.17                         |
| RHEL 6       | 0.10.x                         |
| RHEL 5       | 0.8.x                          |


### 0.9.x Support

Previously there was support for libvirt 0.9.8 and below, however this is no longer being updated. These releases were tagged `v1.x` at `gopkg.in/alexzorin/libvirt-go.v1` [(docs)](http://gopkg.in/alexzorin/libvirt-go.v1).

## Documentation

* [api documentation for the bindings](http://godoc.org/github.com/rgbkrk/libvirt-go)
* [api documentation for libvirt](http://libvirt.org/html/libvirt-libvirt.html)

## Contributing

Please fork and write tests.

Integration tests are available where functionality isn't provided by the test driver, see `integration_test.go`.

A `Vagrantfile` is included to run the integration tests:

* `cd ./vagrant`
* `vagrant up` to provision the virtual machine
* `vagrant ssh` to login to the virtual machine

Once inside, `sudo su -` and `go test -tags integration libvirt`.
