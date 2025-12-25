# Copyright 2025 The Kubernetes Authors All rights reserved.
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

alias k=kubectl
alias kgp='kubectl get pods -A'
alias kl='kubectl logs'
alias m="minikube"
alias ml="minikube profile list"
alias mld="minikube delete --all"
alias mk="/workspaces/minikube/out/minikube"
source <(kubectl completion bash)
complete -o default -F __start_kubectl k
source <(minikube completion bash)
