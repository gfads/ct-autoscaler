#!/bin/bash
for item in $(cat services.txt)
do
  echo "apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: vpa-$item
spec:
  targetRef:
    apiVersion: "apps/v1"
    kind: Deployment
    name: $item
  updatePolicy:
    updateMode: "Auto"
---
        " >> manifests-vpa.yaml
done