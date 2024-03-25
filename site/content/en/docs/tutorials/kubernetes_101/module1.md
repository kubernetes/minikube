---
title: "Module 1 - Create a Kubernetes Cluster"
weight: 1
--- 

**Difficulty**: Beginner
**Estimated Time:** 10 minutes

The goal of this scenario is to deploy a local development Kubernetes cluster using minikube

## Step 1 - Cluster up and running

If you haven't already, first [install minikube]({{< ref "/docs/start" >}}). Check that it is properly installed, by running the *minikube version* command:

```shell
minikube version
```

Once minikube is installed, start the cluster, by running the *minikube start* command:

```shell
minikube start
```

Great! You now have a runnning Kubernetes cluster in your terminal. minikube started a virtual environment for you, and a Kubernetes cluster is now running in that environment.

## Step 2 - Cluster version

To interact with Kubernetes during this bootcamp we'll use the command line interface, kubectl. We'll explain kubectl in detail in the next modules, but for now, we're just going to look at some cluster information. To check if kubectl is installed you can run the *kubectl version* command:

```shell
kubectl version
```

OK, kubectl is configured and we can see both the version of the client and as well as the server. The client version is the kubectl version; the server version is the Kubernetes version installed on the master. You can also see details about the build.

## Step 3 - Cluster details

Let's view the cluster details. We'll do that by running *kubectl cluster-info*:

```shell
kubectl cluster-info
```

During this tutorial, we'll be focusing on the command line for deploying and exploring our application. To view the nodes in the cluster, run the *kubectl get nodes* command:

```shell
kubectl get nodes
```

This command shows all nodes that can be used to host our applications. Now we have only one node, and we can see that its status is ready (it is ready to accept applications for deployment).

{{% button link="/docs/tutorials/kubernetes_101/module2" %}}
