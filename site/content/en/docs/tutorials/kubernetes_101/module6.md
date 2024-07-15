---
title: "Module 6 - Update your app"
weight: 6
---

**Difficulty:** Beginner
**Estimated Time:** 10 minutes

The goal of this scenario is to update a deployed application with kubectl set image and to rollback with the rollout undo command.

## Step 1 - Update the version of the app

To list your deployments, run the `get deployments` command:

```shell
kubectl get deployments
```

To list the running Pods, run the `get pods` command:

```shell
kubectl get pods
```

To view the current image version of the app, run the `describe pods` command and look for the `Image` field:

```shell
kubectl describe pods
```

To update the image of the application to version 2, use the `set image` command, followed by the deployment name and the new image version:

```shell
kubectl set image deployments/kubernetes-bootcamp kubernetes-bootcamp=gcr.io/k8s-minikube/kubernetes-bootcamp:v2
```

The command notified the Deployment to use a different image for your app and initiated a rolling update. Check the status of the new Pods, and view the old one terminating with the `get pods` command:

```shell
kubectl get pods
```

## Step 2 - Verify an update

First, check that the app is running. To find the exposed IP and Port, run the `describe service` command:

```shell
kubectl describe services/kubernetes-bootcamp
```

**Note for Docker Desktop users:** Due to Docker Desktop networking limitations, by default you're unable to access pods directly from the host. Run `minikube service kubernetes-bootcamp`, this will create a SSH tunnel from the pod to your host and open a window in your default browser that's connected to the service. The tunnel can be terminated by pressing control-C, then continue the tutorial after the `curl $(minikube ip):$NODE_PORT` command.

Create an environment variable called `NODE_PORT` that has the value of the Node port assigned:

```shell
export NODE_PORT=$(kubectl get services/kubernetes-bootcamp -o go-template='{{(index .spec.ports 0).nodePort}}')
echo NODE_PORT=$NODE_PORT
```

Next, do a `curl` to the exposed IP and port:

```shell
curl $(minikube ip):$NODE_PORT
```

Every time you run the `curl` command, you will hit a different Pod. Notice that all Pods are running the latest version (v2).

You can also confirm the update by running the `rollout status` command:

```shell
kubectl rollout status deployments/kubernetes-bootcamp
```

To view the current image version of the app, run the `describe pods` command:

```shell
kubectl describe pods
```

In the `Image` field of the output, verify that you are running the latest image version (v2).


## Step 3 - Rollback an update

Let's perform another update, and deploy an image tagged with `v10`:

```shell
kubectl set image deployments/kubernetes-bootcamp kubernetes-bootcamp=gcr.io/k8s-minikube/kubernetes-bootcamp:v10
```

Use `get deployments` to see the status of the deployment:

```shell
kubectl get deployments
```

Notice that the output doesn't list the desired number of available Pods. Run the `get pods` command to list all Pods:

```shell
kubectl get pods
```

Notice that some of the Pods have a status of `ImagePullBackOff`.

To get more insight into the problem, run the `describe pods` command:

```shell
kubectl describe pods
```

In the `Events` section of the output for the affected Pods, notice that the `v10` image version did not exist in the repository.

To roll back the deployment to your last working version, use the `rollout undo` command:

```shell
kubectl rollout undo deployments/kubernetes-bootcamp
```

The `rollout undo` command reverts the deployment to the previous known state (v2 of the image). Updates are versioned and you can revert to any previously known state of a deployment.

Use the `get pods` commands to list the Pods again:

```shell
kubectl get pods
```

Four Pods are running. To check the image deployed on these Pods, use the `describe pods` command:

```shell
kubectl describe pods
```

The deployment is once again using a stable version of the app (v2). The rollback was successful.
