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
alias reset-view='echo "To reset VS Code layout:" && echo "  1. Press Ctrl+Shift+P (or Cmd+Shift+P on Mac)" && echo "  2. Type: View: Reset View Locations" && echo "  3. Press Enter"'
source <(kubectl completion bash)
complete -o default -F __start_kubectl k
source <(minikube completion bash)

# Display welcome message on first terminal entry
if [ ! -f "$HOME/.devcontainer-welcome-shown" ]; then
    echo ""
    echo "╔══════════════════════════════════════════════════════════════════════╗"
    echo "║  Welcome to minikube in GitHub Codespaces                           ║"
    echo "╚══════════════════════════════════════════════════════════════════════╝"
    echo ""
    echo "To start minikube simply type:"
    echo ""
    echo "    minikube start"
    echo ""
    echo "Useful aliases:"
    echo "  m          - minikube"
    echo "  k          - kubectl"
    echo "  reset-view - Show how to reset VS Code layout"
    echo ""
    touch "$HOME/.devcontainer-welcome-shown"
fi
