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
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 30
    scaleUp:
      stabilizationWindowSeconds: 30
  minReplicas: 1
  maxReplicas: 4
  metrics:
  - type: Object
    object:
      metric:
        name: response_time
      describedObject:
        apiVersion: v1
        kind: Service
        name: $item
      target:
        type: Value
        value: 0.5
---
        " >> manifests-response-time-stb-30.yaml
done



