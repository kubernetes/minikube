#!/bin/bash -x

WEB_PORT=8080
SLEEP_SEC=6

## deployment "hello-minikube" created
# sudo kubectl run hello-minikube --image=k8s.gcr.io/echoserver:1.4 --port=8080
sudo kubectl run hello-minikube --image=k8s.gcr.io/echoserver:1.4 --port=${WEB_PORT}
sleep ${SLEEP_SEC}

## service "hello-minikube" exposed
sudo kubectl expose deployment hello-minikube --type=NodePort

sleep ${SLEEP_SEC}

# We have now launched an echoserver pod but we have to wait until the pod is up 
#   before curling/accessing it
# via the exposed service.
# To check whether the pod is up and running we can use the following:
#   NAME                              READY     STATUS              RESTARTS   AGE
#   hello-minikube-3383150820-vctvh   1/1       ContainerCreating   0          3s
sudo kubectl get pod

sleep ${SLEEP_SEC}

# We can see that the pod is still being created from the ContainerCreating status
#  NAME                              READY     STATUS    RESTARTS   AGE
# hello-minikube-3383150820-vctvh   1/1       Running   0          13s
sudo kubectl get pod

sleep ${SLEEP_SEC}

# We can see that the pod is now Running and we will now be able to curl it:
#  CLIENT VALUES:
#  client_address=192.168.99.1
#  command=GET
#  real path=/
#  ...
curl $(sudo minikube service hello-minikube --url)

sleep ${SLEEP_SEC}

# service "hello-minikube" deleted
sudo kubectl delete service hello-minikube

sleep ${SLEEP_SEC}

# deployment "hello-minikube" deleted
sudo kubectl delete deployment hello-minikube

sleep ${SLEEP_SEC}

# Stopping local Kubernetes cluster...
# Machine stopped.
sudo minikube stop

