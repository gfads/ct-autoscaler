#!/bin/bash

#export PATH=$HOME/.istioctl/bin:$PATH

kind delete clusters kind
kind create cluster --config ./kind/cluster-with-node-port.yaml

#kubectl label ns --namespace default --all istio-injection=enabled

#istioctl install -y --set profile=demo
#kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.23/samples/addons/prometheus.yaml
#kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.23/samples/addons/grafana.yaml

helm install beyla -n beyla --create-namespace  grafana/beyla --values ./beyla/beyla-values.yaml --version 1.4.0
LATEST=v0.77.1
#Prometheus
kubectl create ns prom
curl -sL https://github.com/prometheus-operator/prometheus-operator/releases/download/${LATEST}/bundle.yaml | kubectl create -f -
kubectl wait --for=condition=Ready pods -l  app.kubernetes.io/name=prometheus-operator -n default
kubectl apply -f ./prom

#Grafana
helm install grafana grafana/grafana -n prom
kubectl apply -f ./grafana

echo "Grafana pass"
kubectl get secret --namespace prom grafana -o jsonpath="{.data.admin-password}" | base64 --decode ; echo
#Online Boutique
## ingress

#kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
#kubectl apply -f ./ingress
#echo "You need create a CNAME record to teste.aws.dev"
# Quarkus
kubectl create ns quarkus
sleep 3
kubectl apply -f ./quarkus/database-pg.yaml
kubectl apply -f ./quarkus/quarkus-app-baseline.yaml

# Output


#echo "Waiting for pending pods..."
#sleep 120

#kubectl port-forward -n locust --address 0.0.0.0 svc/locust-k8s 3003:80 &
#kubectl port-forward  --address 0.0.0.0 svc/prometheus 9090 &
#kubectl port-forward  --address 0.0.0.0 -n prom svc/grafana 3000:80 &

echo "Enabling metrics-server"

kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/download/v0.5.0/components.yaml
kubectl patch -n kube-system deployment metrics-server --type=json \
-p '[{"op":"add","path":"/spec/template/spec/containers/0/args/-","value":"--kubelet-insecure-tls"}]'

echo "Enabling metric adapter"

export API_SERVER_METRICS_PATH=/home/ubuntu/custom-metric-api-server/custom-metrics-apiserver/test-adapter-deploy

echo "api server container should be available"
echo "https://github.com/kubernetes-sigs/custom-metrics-apiserver/archive/refs/tags/v1.30.0.tar.gz"

kind load docker-image kubernetes-sigs/k8s-test-metrics-adapter-amd64
kubectl create ns controller
sleep 10

kubectl apply -f $API_SERVER_METRICS_PATH/testing-adapter.yaml

echo "Build controller/metric forwarder"

cd ..
docker build -t controller .
kind load docker-image controller

echo "Run metric forwarder"
cd ./k8s-scripts/controllerPod

kubectl apply -f auth.yaml
kubectl apply -f pod.yaml

