---
title: "Automated Google Cloud Platform Authentication"
linkTitle: "GCP Auth"
weight: 1
date: 2020-07-15
---


The `gcp-auth` addon automatically and dynamically configures pods to use your credentials, allowing applications to access Google Cloud services as if they were running within Google Cloud.  

The addon defaults to using your environment's [Application Default Credentials](https://google.aip.dev/auth/4110), which you can configure with `gcloud auth application-default login`. 
Alternatively, you can specify a JSON credentials file (e.g. service account key) by setting the `GOOGLE_APPLICATION_CREDENTIALS` environment variable to the location of that file.

The addon also defaults to using your local gcloud project, which you can configure with `gcloud config set project <project name>`. You can override this by setting the `GOOGLE_CLOUD_PROJECT` environment variable to the name of the desired project.

Once the addon is enabled, pods in your cluster will be configured with environment variables (e.g. `GOOGLE_APPLICATION_DEFAULTS`, `GOOGLE_CLOUD_PROJECT`) that are automatically used by GCP client libraries.  Additionally, the addon configures [registry pull secrets](https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/), allowing your cluster to access the container images hosted in [Artifact Registry](https://cloud.google.com/artifact-registry) and [Google Container Registry](https://cloud.google.com/container-registry).


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

Pods that were deployed to your minikube cluster before the `gcp-auth` addon was enabled will not be configured with GCP credentials. 
To resolve this problem, run:

`minikube addons enable gcp-auth --refresh`

## Adding new namespaces

### minikube v1.29.0+
Newly created namespaces automatically have the image pull secret configured, no action is required.

### minikube v1.28.0 and before
Namespaces that are added after enabling gcp-auth addon will not be configured with the image pull secret. 
To resolve this problem, run:

`minikube addons enable gcp-auth --refresh`
