<!--[metadata]>
+++
title = "Driver options and operating system defaults"
description = "Identify active machines"
keywords = ["machine, driver, base, operating system"]
[menu.main]
parent="smn_machine_drivers"
weight=-1
+++
<![end-metadata]-->

# Driver options and operating system defaults

When Docker Machine provisions containers on local network provider or with a
remote, cloud provider such as Amazon Web Services, you must define both the
driver for your provider and a base operating system. There are over 10
supported drivers and a generic driver for adding machines for other providers.

Each driver has a set of options specific to that provider.  These options
provide information to machine such as connection credentials, ports, and so
forth.  For example, to create an Azure machine:

Grab your subscription ID from the portal, then run `docker-machine create` with
these details:

```bash
$ docker-machine create -d azure --azure-subscription-id="SUB_ID" --azure-subscription-cert="mycert.pem" A-VERY-UNIQUE-NAME
```

To see a list of providers and review the options available to a provider, see
the [Docker Machine driver reference](../index.md).

In addition to the provider, you have the option of identifying a base operating
system. It is an option because Docker Machine has defaults for both local and
remote providers. For local providers such as VirtualBox, Fusion, Hyper-V, and
so forth, the default base operating system is Boot2Docker. For cloud providers,
the base operating system is the latest Ubuntu LTS the provider supports.

| Operating System        | Version | Notes              |
| ----------------------- | ------- | ------------------ |
| Boot2Docker             | 1.5+    | default for local  |
| Ubuntu                  | 12.04+  | default for remote |
| RancherOS               | 0.3+    |                    |
| Debian                  | 8.0+    | experimental       |
| RedHat Enterprise Linux | 7.0+    | experimental       |
| CentOS                  | 7+      | experimental       |
| Fedora                  | 21+     | experimental       |

To use a different base operating system on a remote provider, specify the
provider's image flag and one of its available images. For example, to select a
`debian-8-x64` image on DigitalOcean you would supply the
`--digitalocean-image=debian-8-x64` flag.

If you change the base image for a provider, you may also need to change
the SSH user. For example, the default Red Hat AMI on EC2 expects the
SSH user to be `ec2-user`, so you would have to specify this with
`--amazonec2-ssh-user ec2-user`.
