#!/bin/bash

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

# 1) "cat" together all summary files.
# 2) Process all summary files.
gsutil cat gs://minikube-builds/logs/master/*/*_summary.json \
| $DIR/process_data.sh
