---
title: "Module 2 - Deploy an app"                  
weight: 2
--- 

**Difficulty**: Beginner
**Estimated Time:** 10 minutes

The goal of this scenario is to help you deploy your first app on Kubernetes using kubectl. You will learn the basics about kubectl cli and how to interact with your application.

## Step 1 - kubectl basics

Type kubectl in the terminal to see its usage. The common format of a kubectl command is: kubectl action resource. This performs the specified action (like create, describe) on the specified resource (like node, container). You can use `--help` after the command to get additional info about possible parameters (`kubectl get nodes --help`).

Check the kubectl is configured to talk to your cluster, by running the `kubectl version` command:

```shell
kubectl version
```

OK, kubectl is installed and you can see both the client and the server versions.

To view the nodes in the cluster, run the `kubectl get nodes` command:

```shell
kubectl get nodes
```

Here we see the available nodes (1 in our case). Kubernetes will choose where to deploy our application based on Node available resources.

## Step 2 - Deploy our app

Let's deploy our first app on Kubernetes with the `kubectl create deployment` command. We need to provide the deployment name and app image location (include the full repository url for images hosted outside Docker Hub).

```shell
kubectl create deployment kubernetes-bootcamp --image=gcr.io/k8s-minikube/kubernetes-bootcamp:v1
```

Great! You just deployed your first application by creating a deployment. This performed a few things for you:

- searched for a suitable node where an instance of the application could be run (we have only 1 available node)
- scheduled the application to run on that Node
- configured the cluster to reschedule the instance on a new Node when needed

To list your deployments use the `get deployments` command:

```shell
kubectl get deployments
```

We see that there is 1 deployment running a single instance of your app. The instance is running inside a Docker container on your node.

## Step 3 - View our app

Pods that are running inside Kubernetes are running on a private, isolated network. By default they are visible from other pods and services within the same Kubernetes cluster, but not outside that network. When we use `kubectl`, we're interacting through an API endpoint to communicate with our application.

We will cover other options on how to expose your application outside the Kubernetes cluster in Module 4.

The `kubectl` command can create a proxy that will forward communications into the cluster-wide, private network. The proxy can be terminiated by pressing control-C and won't show any output while its running.

We will open a second terminal window to run the proxy.

```shell
echo -e "Starting Proxy. After starting it will not output a response. Please return to your original terminal window\n"; kubectl proxy
```

We now have a connection between our host (the online terminal) and the Kubernetes cluster. The proxy enabled direct access to the API from these terminals.

You can see all those APIs hosted through the proxy endpoint. For example, we can query the version directly through the API using the `curl` command:
```shell
curl http://localhost:8001/version
```

*Note: The proxy was run in a new tab, and the recent commands were executed in the original tab. The proxy still runs in the second tab, and this allowed our curl command to work using `localhost:8001`.*

**If Port 8001 is not accessible, ensure that the `kubectl proxy` started above is running.**

The API server will automatically create an endpoint for each pod, based on the pod name, that is also accessible through the proxy.

First we need to get the Pod name, and we'll store in the environment variable POD_NAME:
```shell
export POD_NAME=$(kubectl get pods -o go-template --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}')
echo Name of the Pod: $POD_NAME
```

You can access the Pod through the API by running:
```shell
curl http://localhost:8001/api/v1/namespaces/default/pods/$POD_NAME
```

In order for the new deployment to be accessible without using the Proxy, a Service is required which will be explained in the next modules.

{{% button link="/docs/tutorials/kubernetes_101/module3" %}}
