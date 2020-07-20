#!/bin/bash

set -x

echo "automount ...";

if [ -d /var ]; then
    mkdir -p /var/data
    mkdir /data
    mount --bind /var/data /data

    mkdir -p /var/hostpath_pv
    mkdir /tmp/hostpath_pv
    mount --bind /var/hostpath_pv /tmp/hostpath_pv

    mkdir -p /var/hostpath-provisioner
    mkdir /tmp/hostpath-provisioner
    mount --bind /var/hostpath-provisioner /tmp/hostpath-provisioner
fi
