"""
Copyright 2016 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
"""

"This package gets the LD flags used to set the version of kubernetes."

import json
import subprocess

K8S_PACKAGE = 'k8s.io/kubernetes/'
X_ARG_BASE = '-X k8s.io/minikube/vendor/k8s.io/kubernetes/pkg/version.'

def get_rev():
  with open('./Godeps/Godeps.json') as f:
    contents = json.load(f)
    for dep in contents['Deps']:
      if dep['ImportPath'].startswith(K8S_PACKAGE):
        return 'gitCommit=%s' % dep['Rev']

def get_version():
  # Update when vendor/k8s.io/kubernetes is updated.
  return 'gitVersion=v1.3.0-alpha.3-838+ba170aa191f8c7'

def get_tree_state():
  git_status = subprocess.check_output(['git', 'status', '--porcelain'])
  if git_status:
    result = 'dirty'
  else :
    result = 'clean'
  return 'gitTreeState=%s' % result

def main():
  args = [get_rev(), get_version(), get_tree_state()]
  return ' '.join([X_ARG_BASE + arg for arg in args])

if __name__ == '__main__':
  print main()
