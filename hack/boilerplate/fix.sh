#!/usr/bin/env bash

# Copyright 2018 The Kubernetes Authors All rights reserved.
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

rootDir=$1
[[ "$rootDir" == "" ]] && export rootDir=`pwd`

function prepend() {
  ignore="vendor\|\_gopath\|assets.go"
  pattern=$1
  ref=$2
  headers=$3
  files=`hack/boilerplate/boilerplate.py --rootdir $rootDir | grep -v "$ignore" | grep "$pattern"`
  for f in $files; do
    echo $f;
    copyright="$(cat hack/boilerplate/boilerplate.$ref.txt | sed s/YEAR/2018/g)"

    if [ "$headers" != "" ]; then
      fileHeaders="$(cat $f | grep $headers)"
    fi

    if [ "$fileHeaders" != "" ]; then
        fileContent="$(cat $f | grep -v $headers)"
        printf '%s\n\n%s\n%s\n' "$fileHeaders" "$copyright" "$fileContent"  > $f ;
    else
      fileContent="$(cat $f)"
      printf '%s\n\n%s\n' "$copyright" "$fileContent"  > $f ;
    fi

    done
}

prepend "\.go" "go" "+build"
prepend "\.py" "py"
prepend "\.sh" "sh" "#!"
prepend Makefile Makefile
prepend Dockerfile Dockerfile
