package Entities

import (
	"fmt"
	"main/Infrastructure/Providers"
	"strconv"
	"time"
)

type Executor struct {
	K8s *Providers.Kubernetes
}

func (e *Executor) SetCPUToPods(namespace string, pods []string, cpu int, container string) error {
	e.K8s.SetCPUToPods(pods, namespace, cpu, container)
	return nil
}

func (e *Executor) SetMemoryToPods(namespace string, pods []string, memory int, container string) error {
	e.K8s.SetMemoryToPods(pods, namespace, memory, container)
	return nil
}

func (e *Executor) SetEnvironmentPropretyValueToCM(pods []string, deploymentName, namespace, configmapName string, propertyName []string, propertyValue []int) error {

	var strNumber []string
	for _, item := range propertyValue {
		strNumber = append(strNumber, strconv.Itoa(item))
	}

	fmt.Println("DEBUG EXEC", deploymentName)
	e.K8s.SetValueToConfigMap(namespace, configmapName, propertyName, strNumber)

	time.Sleep(time.Second * time.Duration(2))
	e.K8s.RestartDeployment(namespace, deploymentName) //the rollout will be controlled by deployment resource.
	/*e.K8s.ScaleKubernetesDeployment(namespace, deploymentName, 2)
	time.Sleep(time.Second * time.Duration(10))
	e.K8s.DeletePods(namespace, pods)
	time.Sleep(time.Second * time.Duration(2))
	e.K8s.ScaleKubernetesDeployment(namespace, deploymentName, 1)*/
	return nil
}
