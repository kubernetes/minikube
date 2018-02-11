#!/usr/bin/env bash
set -e

cd "$(dirname $0)" || exit

source demo-magic.sh -n

DEMO_PROMPT="${CYAN}$ "
MINIKUBE_CMD="../out/minikube"
TYPE_SPEED=15

p "# In this demo I will start up a 4-node minikube cluster."
p "# Three workers and one server"
p ""

p "# Start minikube master ..."
pe "$MINIKUBE_CMD start"

p ""
pe "$MINIKUBE_CMD node list"

p ""
p "# Add some nodes ..."
pe "$MINIKUBE_CMD node add"
pe "$MINIKUBE_CMD node add"
pe "$MINIKUBE_CMD node add"

p ""
pe "$MINIKUBE_CMD node list"

p ""
p "# Start nodes ..."
pe "$MINIKUBE_CMD node start"

p ""
pe "$MINIKUBE_CMD node list"

sleep 10

p ""
pe "kubectl get nodes"

p ""
p "# Installing Pod network ..."
pe "kubectl apply -f kube-flannel.yml"

p ""
p "# Wait for flannel to start ..."
pe "kubectl --namespace=kube-system rollout status ds/kube-flannel-ds"

p ""
p "# Deploy our pods ..."
pe "cat hello-deployment.yml"
pe "kubectl apply -f hello-deployment.yml"

p ""
pe "kubectl rollout status deployment/hello"

p ""
p "# Deploy our service ..."
pe "cat hello-svc.yml"
pe "kubectl apply -f hello-svc.yml"

p ""
pe "# Note Pod IPs ..."
pe "kubectl get pod -o wide"

p ""
pe "$MINIKUBE_CMD service list"

p ""
ip=$($MINIKUBE_CMD ip)
pe "for i in \$(seq 1 10); do curl http://$ip:31000; echo; sleep 0.4; done"

p "# Yay"
