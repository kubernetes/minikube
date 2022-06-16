---
title: "Automated Google Cloud Platform Authentication"
linkTitle: "GCP Auth"
weight: 1
date: 2020-07-15
---


The gcp-auth addon automatically and dynamically configures pods to use your credentials, allowing applications to access Google Cloud services as if they were running within Google Cloud.  

The addon normally uses the [Google Application Default Credentials](https://google.aip.dev/auth/4110) as configured with `gcloud auth application-default login`. If you already have a json credentials file you want specify, such as to use a service account, set the GOOGLE_APPLICATION_CREDENTIALS environment variable to point to that file.

The addon normally uses the default gcloud project as configured with `gcloud config set project <project name>`. If you want to use a different project, set the `GOOGLE_CLOUD_PROJECT` environment variable to the desired project.

The pods are configured with the `GOOGLE_APPLICATION_DEFAULTS` environment variable is set, which is automatically used by GCP client libraries, and the `GOOGLE_CLOUD_PROJECT` environment variable is set, as are several other historical environment variables.  The addon also configures  [registry pull secrets](https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/) to allow the cluster to access container images hosted in your project's [Artifact Registry](https://cloud.google.com/artifact-registry) and [Google Container Registry](https://cloud.google.com/container-registry).

## Tutorial

- Start a cluster:

```shell
minikube start
```

```
😄  minikube v1.12.0 on Darwin 10.15.5
✨  Automatically selected the docker driver. Other choices: hyperkit, virtualbox
👍  Starting control plane node minikube in cluster minikube
🔥  Creating docker container (CPUs=2, Memory=3892MB) ...
🐳  Preparing Kubernetes v1.18.3 on Docker 19.03.2 ...
🔎  Verifying Kubernetes components...
🌟  Enabled addons: default-storageclass, storage-provisioner
🏄  Done! kubectl is now configured to use "minikube"
```

- Enable the `gcp-auth` addon:

```shell
minikube addons enable gcp-auth
```

```
🔎  Verifying gcp-auth addon...
📌  Your GCP credentials will now be mounted into every pod created in the minikube cluster.
📌  If you don't want credential mounted into a specific pod, add a label with the `gcp-auth-skip-secret` key to your pod configuration.
🌟  The 'gcp-auth' addon is enabled
```

- For credentials in an arbitrary path:

```shell
export GOOGLE_APPLICATION_CREDENTIALS=<creds-path>.json
minikube addons enable gcp-auth
```

- Deploy your GCP app as normal:

```shell
kubectl apply -f test.yaml
```

```
deployment.apps/pytest created
```

Everything should work as expected. You can run `kubectl describe` on your pods to see the environment variables we inject.

As explained in the output above, if you have a pod you don't want to inject with your credentials, all you need to do is add the `gcp-auth-skip-secret` label:
<pre>
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pytest
spec:
  selector:
    matchLabels:
      app: pytest
  replicas: 2
  template:
    metadata:
      labels:
        app: pytest
        <b>gcp-auth-skip-secret: "true"</b>
    spec:
      containers:
      - name: py-test
        imagePullPolicy: Never
        image: local-pytest
        ports:
          - containerPort: 80
</pre>


## Refreshing existing pods

If you had already deployed pods to your minikube cluster before enabling the gcp-auth addon, then these pods will not have any GCP credentials. There are two ways to solve this issue.  

1. If you use a Deployment to deploy your pods, just delete the existing pods with `kubectl delete pod <pod_name>`. The deployment will then automatically recreate the pod and it will have the correct credentials.

2. minikube can delete and recreate your pods for you, by running `minikube addons enable gcp-auth --refresh`. It does not matter if you have already enabled the addon or not. 
