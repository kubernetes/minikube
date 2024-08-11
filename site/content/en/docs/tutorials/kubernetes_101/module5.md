---
title: "Module 5 - Scale up your app"
weight: 5
---

**Difficulty**: Beginner
**Estimated Time:** 10 minutes

The goal of this scenario is to scale a deployment with kubectl scale and to see the load balancing in action

## Step 1 - Scaling a deployment

First, let's list the deployments using the `get deployment` command:

```shell
kubectl get deployments
```

The output should be similar to:

```shell
NAME                  READY   UP-TO_DATE   AVAILABLE   AGE
kubernetes-bootcamp   1/1     1            1           11m
```

We should have 1 Pod. If not, run the command again. This shows:

- *NAME* lists the names of the Deployments in the cluster.
- *READY* shows the ratio of CURRENT/DESIRED replicas
- *UP-TO-DATE* displays the number of replicas that have been updated to achieve the desired state.
- *AVAILABLE* displays how many replicas of the application are available to your users.
- *AGE* displays the amount of time that the application has been running.

To see the ReplicaSet created by the Deployment, run:

```shell
kubectl get rs
```

Notice that the name of the ReplicaSet is always formatted as `[DEPLOYMENT-NAME]-[RANDOM-STRING]`. The random string is randomly generated and uses the pod-template-hash as a seed.

Two important columns of this command are:

- *DESIRED* displays the desired number of replicas of the application, which you define when you create the Deployment. This is the desired state.
- *CURRENT* displays how many replicas are currently running.

Next, let's scale the Deployment to 4 replicas. We'll use the `kubectl scale` command, following by the deployment type, name and desired number of instances:

```shell
kubectl scale deployments/kubernetes-bootcamp --replicas=4
```

To list your Deployments once again, use `get deployments`:

```shell
kubectl get deployments
```

The change was applied, and we have 4 instances of the application available. Next, let's check if the number of Pods changed:

```shell
kubectl get pods -o wide
```

There are 4 Pods now, with different IP addresses. The change was registered in the Deployment events log. The check that, use the describe command:

```shell
kubectl describe deployments/kubernetes-bootcamp
```

You can also view in the output of this command that there are 4 replicas now.

## Step 2 - Load Balancing

Let's check that the Service is load-balancing the traffic. To find out the exposed IP and Port we can use the describe service as we learned in the previous Module:

```shell
kubectl describe services/kubernetes-bootcamp
```

**Note for Docker Desktop users:** Due to Docker Desktop networking limitations, by default you're unable to access pods directly from the host. Run `minikube service kubernetes-bootcamp`, this will create a SSH tunnel from the pod to your host and open a window in your default browser that's connected to the service. Refresh the browser page to see the load-balancing working. The tunnel can be terminated by pressing control-C, then continue on to Step 3.

Create an environment variable called NODE_PORT that has a value as the Node port:

```shell
export NODE_PORT=$(kubectl get services/kubernetes-bootcamp -o go-template='{{(index .spec.ports 0).nodePort}}')
echo NODE_PORT=$NODE_PORT
```

Next, we'll do a `curl` to the exposed IP and port. Execute the command multiple times:

```shell
curl $(minikube ip):$NODE_PORT
```

We hit a different Pod with every request. This demonstrates that the load-balancing is working.

## Step 3 - Scale Down

To scale down the Service to 2 replicas, run again the `scale` command:

```shell
kubectl scale deployments/kubernetes-bootcamp --replicas=2
```

List the Deployments to check if the change was applied with the `get deployments` command:

```shell
kubectl get deployments
```

The number of replicas decreased to 2. List the number of Pods, with `get pods`:

```shell
kubectl get pods -o wide
```

This confirms that 2 Pods were terminated.

{{% button link="/docs/tutorials/kubernetes_101/module6" %}}
