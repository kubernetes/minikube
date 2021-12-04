---
title: "About the Benchmarking Process"
linkTitle: "About the Benchmarking Process"
weight: 1
--- 

## What's the difference between the four images?
In the benchmarking charts you'll see four images: Few Large Layers, Few Small Layers, Many Large Layers, and Many Small Layers

All the images use the same base image: `gcr.io/buildpacks/builder:v1`

#### Few vs Many
Few will copy two files while many will copy 20 files.

#### Small vs Large
Small will copy a 20MB file while large will copy a 123MB file.

Using this info you can see the following:
- Few Large Layers: copies two 123MB files
- Few Small Layers: copies two 20MB files
- Many Large Layers: copies 20 123MB files
- Many Small Layers: copies 20 20MB files

Finally, as the last layer, a simplistic 11 line Go app is copied in.

## Iterative vs Initial
There are two graphs for each benchmark, iterative and inital.

#### Inital
Initial simulates loading the image for the first time.

All existing images and cache is removed/cleared from minikube and Docker between runs to replicate what the user would experience when loading for the first time.

#### Iterative
Iterative simulates only the Go app (last layer of the image) changing.

This is the exact use case of [Skaffold](https://github.com/GoogleContainerTools/skaffold), where if the user changes a file the Go binary is rebuilt and the image is re-loaded.

Bewteen runs the cache and existing image is left alone, only the Go binary is changed.


## How are the benchmarks conducted?
```
// Pseudo code of running docker-env benchmark

startMininkube() // minikube start --container-runtime=docker

for image in [fewLargeLayers, fewSmallLayers, ...] {
	buildGoBinary()

	// inital simulation
	for i in runCount {
		startTimer()

		runDockerEnvImageLoad(image)

		stopTimer()

		verifyImageSuccessfullyLoaded()

		storeTimeTaken()

		removeImage()

		clearDockerCache()
	}

	// iterative simulation
	for i in runCount {
		updateGoBinary()

		startTimer()

		runDockerEnvImageLoad(image)

		stopTimer()

		verifyImageSuccessfullyLoaded()

		storeTimeTaken() // skip if first run
	}

	clearDockerCache()

	calculateAndRecordAverageTime()
}

deleteMinikube() // minkube delete --all
```
