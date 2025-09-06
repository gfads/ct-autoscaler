kubectl delete -f manifests
sleep 2
kubectl apply -f manifests
kubectl get pods --watch
