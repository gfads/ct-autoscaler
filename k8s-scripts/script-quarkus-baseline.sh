#!/bin/bash

kind delete clusters kind
kind create cluster --config ./kind/cluster-with-node-port.yaml

export PATH_QUARKUS_LOAD=$REPOSITORY_BASE_PATH/loadtest/quarkus/baseline

cd $PATH_QUARKUS_LOAD

docker build --no-cache -t baselinetestk6 .
kind load docker-image baselinetestk6

export PATH_K8S_SCRIPTS=$REPOSITORY_BASE_PATH/k8s-scripts
cd $PATH_K8S_SCRIPTS/k6

kubectl create ns k6
sleep 3
kubectl apply -f ./quarkus/k6-baseline.yaml -n k6

echo "Loadtest have been started..."

sleep 10
kubectl logs $(kubectl get pods -n k6 --no-headers -o custom-columns=":metadata.name") -n k6 -f
