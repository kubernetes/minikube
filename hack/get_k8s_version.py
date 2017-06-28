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
import subprocess
import sys
from datetime import datetime

K8S_PACKAGE = 'k8s.io/kubernetes/'
X_ARGS = ['-X k8s.io/minikube/vendor/k8s.io/kubernetes/pkg/version.', '-X k8s.io/minikube/vendor/k8s.io/client-go/pkg/version.']

def get_rev():
  return 'gitCommit=%s' % get_from_godep('Rev')

def get_version():
    return 'gitVersion=%s' % get_from_godep('Comment')

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
  return 'buildDate=%s' % datetime.now().strftime('%Y-%m-%dT%H:%M:%SZ')

def main():
  if len(sys.argv) > 1 and sys.argv[1] == "--k8s-version-only":
      return get_from_godep('Comment')
  args = [get_rev(), get_version(), get_tree_state(), get_build_date()]
  ret = ''
  for xarg in X_ARGS:
    for arg in args:
      ret += xarg + arg + " "
  return ret

if __name__ == '__main__':
  sys.exit(main())
