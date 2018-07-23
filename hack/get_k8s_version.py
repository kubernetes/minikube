#!/usr/bin/env python

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

"This package gets the LD flags used to set the version of kubernetes."

import json
import os
import re
import subprocess
import sys
import time
from datetime import datetime

K8S_PACKAGE = 'k8s.io/kubernetes/'
X_ARGS = ['-X k8s.io/minikube/vendor/k8s.io/kubernetes/pkg/version.', '-X k8s.io/minikube/vendor/k8s.io/client-go/pkg/version.']

def get_commit():
  return 'gitCommit=%s' % get_from_godep('Rev')

def get_version():
    return 'gitVersion=%s' % get_from_godep('Comment')

def get_major_and_minor():
    major = ''
    minor = ''
    version = get_from_godep('Comment')
    # [kubernetes/hack/lib/version.sh]:
    # Try to match the "git describe" output to a regex to try to extract
    # the "major" and "minor" versions and whether this is the exact tagged
    # version or whether the tree is between two tagged versions.
    m = re.match('^v([0-9]+)\.([0-9]+)(\.[0-9]+)?([-].*)?$', version)
    if m:
        major = m.group(1)
        minor = m.group(2)
        if m.group(4):
            minor += "+"
    return ('gitMajor=%s' % major, 'gitMinor=%s' % minor)

def get_from_godep(key):
  with open('./Godeps/Godeps.json') as f:
    contents = json.load(f)
    for dep in contents['Deps']:
      if dep['ImportPath'].startswith(K8S_PACKAGE):
        return dep[key]

def get_tree_state():
  git_status = subprocess.check_output(['git', 'status', '--porcelain'])
  if git_status:
    result = 'dirty'
  else :
    result = 'clean'
  return 'gitTreeState=%s' % result

def get_build_date():
  build_date = datetime.utcfromtimestamp(int(os.environ.get('SOURCE_DATE_EPOCH', time.time())))
  return 'buildDate=%s' % build_date.strftime('%Y-%m-%dT%H:%M:%SZ')

def main():
  if len(sys.argv) > 1 and sys.argv[1] == "--k8s-version-only":
      return get_from_godep('Comment')
  major, minor = get_major_and_minor()
  args = [get_commit(), get_tree_state(), get_version(),
          major, minor, get_build_date()]
  ret = ''
  for xarg in X_ARGS:
    for arg in args:
      ret += xarg + arg + " "
  return ret

if __name__ == '__main__':
  sys.exit(main())
