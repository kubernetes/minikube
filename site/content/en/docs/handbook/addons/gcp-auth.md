---
title: "Automated Google Cloud Platform Authentication"
linkTitle: "GCP Auth"
weight: 1
date: 2020-07-15
---

If you have a containerized GCP app with a Kubernetes yaml, you can automatically add your credentials to all your deployed pods dynamically with this minikube addon. You just need to have a credentials file, which can be generated with `gcloud auth application-default login`. If you already have a json credentials file you want specify, use the GOOGLE_APPLICATION_CREDENTIALS environment variable.

- Start a cluster:
```
minikube start
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
```
minikube addons enable gcp-auth
🔎  Verifying gcp-auth addon...
📌  Your GCP credentials will now be mounted into every pod created in the minikube cluster.
📌  If you don't want credential mounted into a specific pod, add a label with the `gcp-auth-skip-secret` key to your pod configuration.
🌟  The 'gcp-auth' addon is enabled
```

- For credentials in an arbitrary path:
```
export GOOGLE_APPLICATION_CREDENTIALS=<creds-path>.json 
minikube addons enable gcp-auth
```

- Deploy your GCP app as normal:
```
kubectl apply -f test.yaml
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
