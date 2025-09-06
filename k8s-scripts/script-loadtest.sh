#!/bin/bash

kind delete clusters kind
kind create cluster --config ./kind/cluster-with-node-port.yaml

cd ../loadtest
docker build --no-cache -t loadtestk6 .
kind load docker-image loadtestk6
cd ../k8s-scripts
kubectl create ns k6
kubectl apply -f ./k6 -n k6

echo "Loadtest have been started..."
