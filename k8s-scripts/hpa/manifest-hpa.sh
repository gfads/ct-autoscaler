#!/bin/bash
for item in $(cat services.txt)
do
  echo "apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: $item-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: $item
  minReplicas: 1
  maxReplicas: 4
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 75
---
        " >> manifests-75-4.yaml
done



