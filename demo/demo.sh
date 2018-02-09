#!/usr/bin/env bash

. demo-magic.sh -n

TYPE_SPEED=15
DEMO_PROMPT="${GREEN}➜ ${CYAN}\W $ "

p "# In this demo I will start up a 4-node minikube cluster."
p "# Three workers and one server"
p ""

p "# Start minikube master ..."
pe "out/minikube start"

pe "out/minikube nodes list"

p "# Add some nodes ..."
pe "out/minikube node add"
pe "out/minikube node add"
pe "out/minikube node add"

pe "out/minikube nodes list"


p "# Start nodes ..."
pe "out/minikube node start"

pe "out/minikube nodes list"

PROMPT_TIMEOUT=10
wait

pe "kubectl get nodes"

p "# Installing Pod network ..."
pe "kubectl apply -f kube-flannel.yml"

p "# Wait for flannel to start ..."
pe "kubectl --namespace=kube-system rollout status ds/kube-flannel-ds"

p "# Deploy our pods ..."
pe "cat hello-from.yml"
pe "kubectl apply -f hello-from.yml"

pe "kubectl rollout status deployment/hello-from"

pe "# Note Pod IPs ..."
pe "kubectl get pod -o wide"

pe "out/minikube service list"

ip=$(out/minikube ip)
pe "for i in \$(seq 1 10); do curl http://$ip:31000; echo; done"

p "# Yay"
