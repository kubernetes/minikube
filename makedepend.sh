#!/bin/sh

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

# Generate go dependencies, for make. Uses `go list".
# Usage: makedepend.sh [-t] output package path [extra]

PATH_FORMAT='{{ .ImportPath }}{{"\n"}}{{join .Deps "\n"}}'
FILE_FORMAT='{{ range $file := .GoFiles }} {{$.Dir}}/{{$file}}{{"\n"}}{{end}}'

if [ "$1" = "-t" ]
then
  PATH_FORMAT='{{ if .TestGoFiles }} {{.ImportPath}} {{end}}'
  shift
fi

out=$1
pkg=$2
path=$3
extra=$4

# check for mandatory parameters
test -n "$out$pkg$path" || exit 1

echo "$out: $extra\\"
go list -f "$PATH_FORMAT" $path |
  grep "$pkg" |
  xargs go list -f "$FILE_FORMAT" |
  sed -e "s|^ ${GOPATH}| \$(GOPATH)|;s/$/ \\\\/"
echo " #"
