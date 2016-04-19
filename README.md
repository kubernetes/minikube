# minikube
Run Kubernetes locally

## Build Instructions

    go build cli/main.go

## Run Instructions
Start the cluster with:

    ./main start
    Starting local Kubernetes cluster...
    2016/04/19 11:41:26 Machine exists!
    2016/04/19 11:41:27 Localkube is already running
    2016/04/19 11:41:27 Kubernetes is available at http://192.168.99.100:8080.
    2016/04/19 11:41:27 Run this command to use the cluster: 
    2016/04/19 11:41:27 kubectl config set-cluster localkube --insecure-skip-tls-verify=true --server=http://192.168.99.100:8080

Access the cluster with:

 First run the command from above:

    kubectl config set-cluster localkube --insecure-skip-tls-verify=true --server=http://192.168.99.100:8080

Then use kubectl normally:

    kubectl get pods


