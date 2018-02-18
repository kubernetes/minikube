#!/bin/sh

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
