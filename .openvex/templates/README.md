# OpenVEX Templates Directory

This directory contains the OpenVEX data for this repository.
The files stored in this directory are used as templates by
`vexctl generate` when generating VEX data for a release or 
a specific artifact.

To add new statements to publish data about a vulnerability,
download [vexctl](https://github.com/openvex/vexctl)
and append new statements using `vexctl add`. For example:
```
vexctl add --in-place main.openvex.json pkg:oci/test CVE-2014-1234567 fixed
```
That will add a new VEX statement expressing that the impact of
CVE-2014-1234567 is under investigation in the test image. When
cutting a new release, for `pkg:oci/test` the new file will be
incorporated to the relase's VEX data.

## Read more about OpenVEX

To know more about generating, publishing and using VEX data
in your project, please check out the vexctl repository and
documentation: https://github.com/openvex/vexctl

OpenVEX also has an examples repository with samples and docs:
https://github.com/openvex/examples

