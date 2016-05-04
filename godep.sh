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

echo "Updating Godeps"
echo
echo "Deleting Godeps and vendor dirs in 5 seconds if not cancelled..."
sleep 5

# delete Godeps and vendor directories
rm -rf Godeps vendor

# reproduce Godeps
GO15VENDOREXPERIMENT="1" godep save -v ./...

# apply patch to golang.org/x/net/trace/trace.go
patch vendor/golang.org/x/net/trace/trace.go - <<EOF
109a110
> 	/*
125a127
> 	*/
EOF

