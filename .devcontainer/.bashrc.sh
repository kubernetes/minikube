alias k=kubectl
alias kgp='kubectl get pods -A'
alias kl='kubectl logs'
alias m=minikube
alias ml=minikube profile list
alias mld=minikube delete --all
alias mk=/workspaces/minikube/out/minikube
source <(kubectl completion bash)
complete -o default -F __start_kubectl k
source <(minikube completion bash)
