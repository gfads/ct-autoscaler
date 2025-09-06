package Entities

import (
	//"main/Entities"
	"fmt"
	"main/Infrastructure/Providers"
)

type Planner struct {
	K        *Providers.Kubernetes
	SkipList []string
}

func (p *Planner) GetContainerFromPod(service *Service) {
	if p.skip(service.Name) {
		return
	}
	fmt.Println(service.Name)
	service.Container = p.K.GetContainerFromPod(service.Namespace, service.Pods[0])

}
func (p *Planner) GetPodsFromService(service *Service) {
	pods := p.K.GetPodsAndLabels(service.Namespace)
	var podList []string

	countLabels := len(service.Labels)
	//get keys from Map
	for k := range pods {
		count := 0
		for j := range service.Labels {
			if pods[k][j] == service.Labels[j] {
				count += 1

			}
			if count == countLabels {
				podList = append(podList, k)
				break
			}

		}

	}
	service.Pods = podList
	service.Deployment = ""

}

func (p *Planner) skip(serviceName string) bool {
	for _, item := range p.SkipList {
		if item == serviceName {
			return true
		}
	}
	return false
}
