#!/bin/bash
./out/minikube delete
rm ./out/minikube
make && ./out/minikube start
