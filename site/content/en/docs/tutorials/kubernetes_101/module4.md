---
title: "Module 4 - Expose your app publicly"                                
weight: 4
--- 

**Difficulty**: Beginner
**Estimated Time:** 10 minutes

In this scenario you will learn how to expose Kubernetes applications outside the cluster using the kubectl expose command. You will also learn how to view and apply labels to objects with the kubectl label command.

## Step 1 - Create a new service

Let's verify that our application is running. We'll use the `kubectl get` command and look for existing Pods:

```shell
kubectl get pods
```

Next, let's list the current Services from our cluster:

```shell
kubectl get services
```

We have a Service called Kubernetes that is created by default when minikube starts the cluster. To create a new service and expose it to external traffic we'll use the expose command with NodePort as parameter.

```shell
kubectl expose deployment/kubernetes-bootcamp --type="NodePort" --port 8080
```

Let's run again the `get services` command:

```shell
kubectl get services
```

**Note for Docker Desktop users:** Due to Docker Desktop networking limitations, by default you're unable to access pods directly from the host. Run `minikube service kubernetes-bootcamp`, this will create a SSH tunnel from the pod to your host and open a window in your default browser that's connected to the service. The tunnel can be terminated by pressing control-C, then continue on to Step 2.

We have now a running Service called kubernetes-bootcamp. Here we see that the Service received a unique cluster-IP, an internal port and an external-IP (the IP of the Node).

To find out what port was opened externally (by the NodePort option) we'll run the `describe service` command:

```shell
kubectl describe services/kubernetes-bootcamp
```

Create an environment variable called NODE_PORT that has the value of the Node port assigned:

```shell
export NODE_PORT=$(kubectl get services/kubernetes-bootcamp -o go-template='{{(index .spec.ports 0).nodePort}}')
echo NODE_PORT=$NODE_PORT
```

Now we can test that the app is exposed outside of the cluster using `curl`, the IP of the Node and the externally exposed port:

```shell
curl $(minikube ip):$NODE_PORT
```

And we get a response from the server. The Service is exposed.

## Step 2 - Using labels

The Deployment created automatically a label for our Pod. With `describe deployment` command you can see the name of the label:

```shell
kubectl describe deployment
```

Let's use this label to query our list of Pods. We'll use the `kubectl get pods` command with `-l` as a parameter, followed by the label values:

```shell
kubectl get pods -l app=kubernetes-bootcamp
```

You can do the same to list the existing services:

```shell
kubectl get services -l app=kubernetes-bootcamp
```

Get the name of the Pod and store it in the POD_NAME environment variable:

```shell
export POD_NAME=$(kubectl get pods -o go-template --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}')
echo Name of the Pod: $POD_NAME
```

To apply a new label we use the label command followed by the object type, object name and the new label:

```shell
kubectl label pods $POD_NAME version=v1
```

This will apply a new label to our Pod (we pinned the application version to the Pod), and we can check it with the describe pod command:

```shell
kubectl describe pods $POD_NAME
```

We see here that the label is attached new to our Pod. And we can query now the list of pods using the new label:

```shell
kubectl get pods -l version=v1
```

And we see the Pod.

## Step 3 - Deleting a service

To delete Services you can use the `delete service` command. Labels can be used also here:

```shell
kubectl delete service -l app=kubernetes-bootcamp
```

Confirm that the service is gone:

```shell
kubectl get services
```

This confirms that our Service was removed. To confirm that route is not exposed anymore you can `curl` the previously exposed IP and port:

```shell
curl $(minikube ip):$NODE_PORT
```

This proves that the app is not reachable anymore from outside of the cluster. You can confirm that the app is still running with a curl inside the pod:

```shell
kubectl exec -ti $POD_NAME -- curl localhost:8080
```

We see here that the application is up. This is because the Deployment is managing the application. To shut down the application, you would need to delete the Deployment as well.

{{% button link="/docs/tutorials/kubernetes_101/module5" %}}
