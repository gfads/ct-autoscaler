# CT-Autoscaler



## Description

This repository contains source code and a guide to reproduce results using control-theory applied to Kubernetes Vertical Autoscaling.

## Repository Structure

```text
ct-autoscaler/
├── k8s-scripts/  -- folder with Kubernetes manifests and tools required by CT-Autoscaler
│   ├── controllerPod/ -- manisfests to create controller instance(s)
│   ├── quarkus/ -- quarkus external database manifests
│   ├── embedded-quarkus/ -- quarkus embedded database manifests 
├── loadtest/ -- folder with K6 loadtest scripts
├── src/ -- folder with CT-Autoscaler source code
└── README.md
```

## Requirements

- Golang 1.23.1+
- Kind 0.14.0+
- Linux Kernel 6.14+
- BPF, SYS_PTRACE, NET_RAW, DAC_READ_SEARCH and PERFMON kernel capabilities enabled
- Kubernetes 1.27.0+
- kubectl 1.31.2

## How to Run Test Experiments

export REPOSITORY_BASE_PATH="your basepath"

#### Setup plant and controller

In k8s-scripts folder, choose .sh file corresponding your wish: quarkus-embeeded, quarkus-external-db, plant-modeling, etc. For example:

```bash
./script-quarkus-embedded-plant.sh
```

This step will create Kubernetes cluster and start quarkus plant.

To start controller pod, go to folder controllerPod and choose an controller type and parametrization, for exemple:

```bash
kubectl apply -f /k8s-scripts/controllerPod/emb-quarkus-variable/AMIGO-PI.yaml
```

#### Setup Workload

The command bellow will create a new cluster to execute load generator. It is recommended execute this part in a separeted machine to avoid resource limitation problems.

Modify the file k8s-scripts/k6/quarkus/k6-quarkus.yaml to set your destination host IP ou name.

After that, execute the command bellow: 

```bash
./k8s-scripts/script-loadtest.sh
```