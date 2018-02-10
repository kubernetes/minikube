#!/usr/bin/env bash
set -e

cd "$(dirname $0)" || exit

source demo-magic.sh -n

TYPE_SPEED=15
DEMO_PROMPT="${CYAN}$ "

p "# In this demo I will start up a 4-node minikube cluster."
p "# Three workers and one server"
p ""

p "# Start minikube master ..."
pe "./minikube start"

p ""
pe "./minikube node list"

p ""
p "# Add some nodes ..."
pe "./minikube node add"
pe "./minikube node add"
pe "./minikube node add"

p ""
pe "./minikube node list"

p ""
p "# Start nodes ..."
pe "./minikube node start"

p ""
pe "./minikube node list"

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
pe "./minikube service list"

p ""
ip=$(./minikube ip)
pe "for i in \$(seq 1 10); do curl http://$ip:31000; echo; sleep 0.4; done"

p "# Yay"
