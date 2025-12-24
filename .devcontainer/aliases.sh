alias k=kubectl
alias mk=minikube
source <(kubectl completion bash)
complete -o default -F __start_kubectl k
source <(minikube completion bash)
