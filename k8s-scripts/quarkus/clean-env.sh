kubectl delete -f quarkus-app-baseline.yaml
kubectl delete -f database-pg.yaml
sleep 2
kubectl apply -f database-pg.yaml
kubectl apply -f quarkus-app-baseline.yaml
kubectl get pods --watch
