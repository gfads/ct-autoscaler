package Providers

import (
	"context"
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	//
	// Uncomment to load all auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth"
	//
	// Or uncomment to load specific auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

type Kubernetes struct {
	PodsToExclude []string
	MaxCPU        int
	MinCPU        int
	MinMemory     int
	MaxMemory     int
	client        *kubernetes.Clientset
}

func (k *Kubernetes) GetServiceNames(namespace string) []string {
	services, err := k.client.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	var serviceNames []string

	for i := 0; i < len(services.Items); i++ {
		serviceNames = append(serviceNames, services.Items[i].Name)

	}
	return serviceNames
}

func (k *Kubernetes) GetServiceSelector(namespace string, serviceName string) map[string]string {
	services, err := k.client.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	for i := 0; i < len(services.Items); i++ {
		if services.Items[i].Name == serviceName {
			return services.Items[i].Spec.DeepCopy().Selector
		}

	}
	return nil

}

func (k *Kubernetes) ApplyCpuAndMemoryConfiguration(pods []string, namespace string, client *kubernetes.Clientset, memory string, cpu string) error {

	for i := 0; i < len(pods); i = i + 1 {
		// TO-DO: does it adapt only the first container in pod?
		data := []byte(`{"spec":{"containers":[{"name":"server", "resources":{"requests":{"cpu":"` + cpu + `m","memory":"` + memory + `Mi"}, "limits":{"cpu":"` + cpu + `m","memory":"` + memory + `Mi"}}}]}}`)
		_, err := client.CoreV1().Pods(namespace).Patch(context.TODO(), pods[i], types.PatchType("application/strategic-merge-patch+json"), data, metav1.PatchOptions{})
		fmt.Println(i)
		if err != nil {
			return err
		}

	}

	return nil
}

func (k *Kubernetes) SetCPUToPods(pods []string, namespace string, cpu int, container string) error {
	cpuStr := strconv.Itoa(cpu)
	for i := 0; i < len(pods); i = i + 1 {
		data := []byte(`{"spec":{"containers":[{"name":"` + container + `", "resources":{"requests":{"cpu":"` + cpuStr + `m"}, "limits":{"cpu":"` + cpuStr + `m"}}}]}}`)
		_, err := k.client.CoreV1().Pods(namespace).Patch(context.TODO(), pods[i], types.PatchType("application/strategic-merge-patch+json"), data, metav1.PatchOptions{})
		fmt.Println(i)
		if err != nil {
			return err
		}

	}

	return nil
}

func (k *Kubernetes) SetMemoryToPods(pods []string, namespace string, memory int, container string) error {
	memoryStr := strconv.Itoa(memory)
	for i := 0; i < len(pods); i = i + 1 {
		data := []byte(`{"spec":{"containers":[{"name":"` + container + `", "resources":{"requests":{"memory":"` + memoryStr + `Mi"}, "limits":{"memory":"` + memoryStr + `Mi"}}}]}}`)
		_, err := k.client.CoreV1().Pods(namespace).Patch(context.TODO(), pods[i], types.PatchType("application/strategic-merge-patch+json"), data, metav1.PatchOptions{})
		fmt.Println(i)
		if err != nil {
			return err
		}

	}

	return nil
}
func (k *Kubernetes) GetDeploymentFromService(serviceName string, namespace string) string {
	service, err := k.client.CoreV1().Services(namespace).Get(context.TODO(), serviceName, metav1.GetOptions{})
	if err != nil {
		log.Fatalf("Erro ao recuperar o Service: %v", err)
	}

	// O selector do Service, que é um conjunto de labels
	serviceSelector := service.Spec.Selector
	if len(serviceSelector) == 0 {
		log.Fatalf("O Service %s não tem um selector de labels", serviceName)
	}

	// Listar os Pods que correspondem ao selector do Service
	pods, err := k.client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(&metav1.LabelSelector{MatchLabels: serviceSelector}),
	})
	if err != nil {
		log.Fatalf("Erro ao listar os Pods: %v", err)
	}

	// Para cada Pod, verificar a associação com um Deployment
	for _, pod := range pods.Items {
		// Aqui, usamos o campo 'OwnerReferences' para identificar o controlador que gerencia o Pod
		for _, owner := range pod.OwnerReferences {
			if owner.Kind == "Deployment" {
				// O Deployment foi encontrado
				deploymentName := owner.Name

				return deploymentName
			}
		}
	}
	return ""
}
func (k *Kubernetes) GetPodsAndLabels(namespace string) map[string]map[string]string {

	podLabels := make(map[string]map[string]string)
	pods, err := k.client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for i := 0; i < len(pods.Items); i++ {
		podLabels[pods.Items[i].Name] = pods.Items[i].Labels

	}
	return podLabels
}

func (k *Kubernetes) GetPods(namespace string, podsToExclude []string) []string {
	var response []string
	var skip bool = false
	pods, err := k.client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	for i := 0; i < len(pods.Items); i++ {
		for j := 0; j < len(podsToExclude); j++ {
			if pods.Items[i].GetName() == podsToExclude[j] {
				fmt.Println("skip", podsToExclude[j])
				skip = true
			}
		}

		//fmt.Println(pods.Items[i].GetName())
		if skip {
			continue
		}
		response = append(response, pods.Items[i].GetName())

	}
	return response
}

func (k *Kubernetes) GetContainerFromPod(namespace string, podName string) string {
	pod, err := k.client.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		return err.Error()
	}
	return pod.Spec.Containers[0].Name // TO-DO: need improvement to generalize
}

func (k *Kubernetes) SetValueToConfigMap(namespace string, configmapName string, key []string, newValue []string) string {
	configMap, err := k.client.CoreV1().ConfigMaps(namespace).Get(context.TODO(), configmapName, metav1.GetOptions{})
	if err != nil {
		log.Fatalf("Failed to update ConfigMap: %v", err)
	}
	for i, _ := range key {
		configMap.Data[key[i]] = newValue[i]
	}

	updatedConfigMap, err := k.client.CoreV1().ConfigMaps(namespace).Update(context.TODO(), configMap, metav1.UpdateOptions{})
	if err != nil {
		log.Fatalf("Failed to update ConfigMap: %v", err)
	}
	return fmt.Sprintf("The ConfigMap '%s' was successfuly updated: Data %v\n", updatedConfigMap.Name, updatedConfigMap.Data)

}

func (k *Kubernetes) ScaleKubernetesDeployment(namespace string, deploymentName string, replicasNumber int32) string {
	deployment, err := k.client.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, metav1.GetOptions{})
	if err != nil {
		log.Fatalf("Error to get deployment: %v", err)
	}

	deployment.Spec.Replicas = &replicasNumber

	// Atualizar o Deployment com o novo número de réplicas
	updatedDeployment, err := k.client.AppsV1().Deployments(namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
	if err != nil {
		log.Fatalf("Error to scale Deployment: %v", err)
	}
	return fmt.Sprintf("Deployment '%s' updated successfuly. New number of replicas: %d\n", updatedDeployment.Name, *updatedDeployment.Spec.Replicas)

}

func (k *Kubernetes) DeletePods(namespace string, podNames []string) {
	deletePolicy := metav1.DeletePropagationForeground // Delete Policy, can be "Foreground", "Background" or "Orphan"

	for _, podName := range podNames {

		err := k.client.CoreV1().Pods(namespace).Delete(context.TODO(), podName, metav1.DeleteOptions{
			PropagationPolicy: &deletePolicy,
		})
		if err != nil {
			log.Fatalf("Error to delete pod: %v", err)
		}
	}

}

func (k *Kubernetes) CreateClient(clusterMode bool) {
	var kubeconfig *string
	var config *rest.Config
	var err error
	if !clusterMode {

		if home := homedir.HomeDir(); home != "" {
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		} else {
			kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
		}
		flag.Parse()

		// use the current context in kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			panic(err.Error())
		}
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
	}
	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	k.client = clientset
}

func (k *Kubernetes) RestartDeployment(namespace string, deploymentName string) error {
	deploy, err := k.client.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if deploy.Spec.Template.ObjectMeta.Annotations == nil {
		deploy.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}
	deploy.Spec.Template.ObjectMeta.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

	_, err = k.client.AppsV1().Deployments(namespace).Update(context.TODO(), deploy, metav1.UpdateOptions{})
	return err
}
