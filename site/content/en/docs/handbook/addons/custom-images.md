---
title: "Config the Addon to Use Custom Registries and Images"
linkTitle: "Custom Images"
weight: 2
date: 2021-01-13
---

If you have trouble to access default images, or want to use images from a private registry or local version, you could achieve
this by flag `--images` and `--registries`.

We defined and named all images used by an addon, you could view them by command `addons images`:

```shell
minikube addons images efk
```

```
    ▪ efk has following images:
|----------------------|------------------------------|-------------------|
|      IMAGE NAME      |        DEFAULT IMAGE         | DEFAULT REGISTRY  |
|----------------------|------------------------------|-------------------|
| Elasticsearch        | elasticsearch:v5.6.2         | k8s.gcr.io        |
| FluentdElasticsearch | fluentd-elasticsearch:v2.0.2 | k8s.gcr.io        |
| Alpine               | alpine:3.6                   |                   |
| Kibana               | kibana/kibana:5.6.2          | docker.elastic.co |
|----------------------|------------------------------|-------------------|
```

The `DEFAULT IMAGE` and `DEFAULT REGISTRY` columns indicate which images are used by default.
An empty registry means the image is stored locally or default registry `docker.io`.

The `IMAGE NAME` column is used to customize the corresponding image and registry.

Assume we have a private registry at `localhost:5555` to replace `k8s.gcr.io` and a locally built Kibana called `kibana/kibana:5.6.2-custom`.

We could load local images to minikube by:

```shell
minikube cache add kibana/kibana:5.6.2-custom
```

Then we can start `efk` addon with flags `--images` and `--registries`.
The format is `IMAGE_NAME=CUSTOM_VALUE`, separated by commas, where the `IMAGE_NAME` is the value of `IMAGE NAME` column in the table above.

```shell
minikube addons enable efk --images="Kibana=kibana/kibana:5.6.2-custom" --registries="Kibana=,Elasticsearch=localhost:5555,FluentdElasticsearch=localhost:5555"
```

```
    ▪ Using image localhost:5555/elasticsearch:v5.6.2
    ▪ Using image localhost:5555/fluentd-elasticsearch:v2.0.2
    ▪ Using image alpine:3.6
    ▪ Using image kibana/kibana:5.6.2-custom
🌟  The 'efk' addon is enabled
```

Now the `efk` addon is using the custom registry and images.