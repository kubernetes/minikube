#!/bin/bash

# Copyright 2016 The Kubernetes Authors All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

PACKAGES="libmnl-1.0.3 libnfnetlink-1.0.1 libnetfilter_cttimeout-1.0.0 libnetfilter_cthelper-1.0.0 libnetfilter_queue-1.0.2 libnetfilter_conntrack-1.0.4"
CONNTRACK=conntrack-tools-1.4.2

fetch() {
    curl -s -S http://www.netfilter.org/projects/${1%-*}/files/$1.tar.bz2 | tar xj
}

for PACKAGE in $PACKAGES; do
    fetch $PACKAGE
    (cd $PACKAGE; ./configure && make LDFLAGS=-static install)
done

fetch $CONNTRACK
(cd $CONNTRACK; ./configure && make LDFLAGS=-static && rm -f src/conntrack && make LDFLAGS=-all-static)
cp $CONNTRACK/src/conntrack /go
