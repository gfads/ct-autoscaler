package Entities

import (
	"main/Infrastructure/Providers"
	"main/Interfaces"
)

type Service struct {
	Name      string
	Namespace string
	Labels    map[string]string
	//	MaxCPU           int
	//	MinCPU           int
	Controller Interfaces.IController
	Deployment string
	//MemoryController   Interfaces.IController
	PropertyController Interfaces.IController
	Pods               []string
	Container          string
	Protocol           string
	RT                 float64
}

func (s *Service) setServiceLabels() {
	var k *Providers.Kubernetes
	s.Labels = k.GetServiceSelector(s.Name, s.Namespace)
}

func main() {

}
