/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Note: the example only works with the code within the same release/branch.
package main

import (
	"flag"
	"fmt"
	"log"
	"main/Entities"
	"main/Infrastructure/Providers"
	"main/Interfaces"
	"main/Shared"
	"math"
	"os"
	"strings"
	"time"
	//
	// Uncomment to load all auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth"
	//
	// Or uncomment to load specific auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

/*
func applyCpuAndMemoryConfiguration(pods []string, namespace string, client *kubernetes.Clientset, memory string, cpu string) error {

		for i := 0; i < len(pods); i = i + 1 {
			data := []byte(`{"spec":{"containers":[{"name":"server", "resources":{"requests":{"cpu":"` + cpu + `m","memory":"` + memory + `Mi"}, "limits":{"cpu":"` + cpu + `m","memory":"` + memory + `Mi"}}}]}}`)
			_, err := client.CoreV1().Pods(namespace).Patch(context.TODO(), pods[i], types.PatchType("application/strategic-merge-patch+json"), data, metav1.PatchOptions{})
			fmt.Println(i)
			if err != nil {
				return err
			}

		}

		return nil
	}

	func getAllPodsInNamespace(namespace string, client *kubernetes.Clientset, podsToExclude []string) []string {
		var response []string
		var skip bool = false
		pods, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
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
*/
func setupNewController(k8s *Providers.Kubernetes, servicesToExclude []string, controllerType string, minCPU float64, maxCPU float64, kp float64, ki float64, kd float64, adaptionInterval int) []Entities.Service {

	var svc []Entities.Service
	var next = false
	inCluster := Shared.ValidateEnvironmentVariable("IN_CLUSTER")
	k8s.CreateClient(inCluster == "true")
	svcNames := k8s.GetServiceNames("default")
	for i := 0; i < len(svcNames); i++ {
		next = false
		for _, j := range servicesToExclude {
			if svcNames[i] == j {
				next = true
				break
			}

		}
		if next {
			continue
		}
		labels := k8s.GetServiceSelector("default", svcNames[i])

		//k8s.GetDeploymentFromService(svcNames[i], "default") //fix it!

		var controller Interfaces.IController
		switch controllerType {
		case "ratioController":
			controller = &Entities.RatioController{}
			controller.Initialise(minCPU, maxCPU)

		case "AStar":
			controller = &Entities.AStarController{Info: Shared.ControllerParams{}}
			controller.Initialise(minCPU, maxCPU, 0.2)
		case "PID":
			controller = &Entities.PIDController{}
			controller.Initialise(minCPU, maxCPU, kp, ki, kd, float64(adaptionInterval)) //min, max, kp, ki, kd, delta
		case "MRAC":
			//quarkus test
			Ts := float64(adaptionInterval) //to-do: generalization
			tau := 519.8                    //to-do: generalization
			A_d := math.Exp(-Ts / tau)      //estimado
			B_d := -2.234 * (1 - A_d)       //estimado
			a_m := math.Exp(-Ts / 100.0)    //desejado
			b_m := 1 - a_m                  //desejado
			controller = &Entities.MRACAdaptativeController{}
			controller.Initialise(minCPU, maxCPU, 0.01, a_m, b_m, A_d, B_d)

		}

		//contMemory := Entities.Controller{Info: Shared.ControllerParams{}}
		//contMemory.Initialise(128, 128, 0.5)

		svc = append(svc, Entities.Service{Deployment: "quarkus", Namespace: "default", Name: svcNames[i], Controller: controller, Labels: labels, Protocol: ""})
	}
	return svc

}

func InitialiseMetricForwarder() []Entities.Service {
	var svc []Entities.Service
	servicesToExclude := strings.Split(os.Getenv("SERVICES_TO_EXCLUDE"), ",")
	inCluster := Shared.ValidateEnvironmentVariable("IN_CLUSTER")
	var k Providers.Kubernetes
	k.CreateClient(inCluster == "true")
	svcNames := k.GetServiceNames("default")
	var next = false

	for i := 0; i < len(svcNames); i++ {
		next = false
		for _, j := range servicesToExclude {
			if svcNames[i] == j {
				next = true
				break
			}

		}
		if next {
			continue
		}

		//contMemory := Entities.Controller{Info: Shared.ControllerParams{}}
		//contMemory.Initialise(128, 128, 0.5)
		svc = append(svc, Entities.Service{Namespace: "default", Name: svcNames[i], Protocol: ""})
	}
	return svc
}

func metricForwarderMode(poolingInterval int) {

	var customMetric Entities.CustomMetric
	customMetric.Initialise(true)
	services := InitialiseMetricForwarder()
	monitor := Entities.PrometheusMonitor{Prom: Providers.Prometheus{Address: Shared.ValidateEnvironmentVariable("PROMETHEUS_HOST"), Port: Shared.ValidateEnvironmentVariable("PROMETHEUS_PORT"), Protocol: Shared.ValidateEnvironmentVariable("PROMETHEUS_PROTOCOL")}}
	for {

		for _, service := range services {
			responseTime, err := monitor.GetResponseTimeFromService(service, 0.9, "MetricForwarder")
			if err != nil {
				log.Println(err)
			}
			fmt.Println("Forwarding metric to service:", service.Name, "response time:", responseTime)
			go customMetric.SetResponseTimeValueToService(responseTime, service) //response time in miliseconds

		}

		time.Sleep(time.Second * time.Duration(poolingInterval))

	}

}

func controllerMode(settings Shared.OperationSettings) {

	servicesToExclude := strings.Split(os.Getenv("SERVICES_TO_EXCLUDE"), ",")
	iteration := 0
	u1 := 0.0

	var k Providers.Kubernetes
	monitor := Entities.PrometheusMonitor{Prom: Providers.Prometheus{Address: Shared.ValidateEnvironmentVariable("PROMETHEUS_HOST"), Port: Shared.ValidateEnvironmentVariable("PROMETHEUS_PORT"), Protocol: Shared.ValidateEnvironmentVariable("PROMETHEUS_PROTOCOL")}}
	planner := Entities.Planner{K: &k, SkipList: servicesToExclude}
	services := setupNewController(&k, servicesToExclude, *settings.ControllerType, *settings.MinCPU, *settings.MaxCPU, *settings.Kp, *settings.Ki, *settings.Kd, settings.AdaptionInterval)
	exec := Entities.Executor{K8s: &k}
	var propertyToAdapt []string
	configMapName := ""
	previousValue := 0
	fmt.Println("PID params", *settings.Kp, *settings.Ki, *settings.Kd)
	if settings.AdaptionType == "ApplicationPropertyAdaption" {
		propertyToAdapt = strings.Split(Shared.ValidateEnvironmentVariable("PROPERTY_ADAPT_NAME"), `,`)
		configMapName = Shared.ValidateEnvironmentVariable("CONFIGMAP_NAME")
	}

	for true {
		fmt.Println("loop")
		for i := 0; i < len(services); i++ {
			Rt, errRT := monitor.GetResponseTimeFromService(services[i], 0.9, "ControllerMode")
			Er, errTX := monitor.GetErrorRateFromService(services[i])
			if errRT != nil {
				panic(errRT)
			}
			if errTX != nil {
				panic(errTX)
			}
			fmt.Println(services[i].Name, "RT:", Rt, "Err:", Er)
			//plan
			planner.GetPodsFromService(&services[i]) //it's important because pods can scale up or down.
			planner.GetContainerFromPod(&services[i])

			//Monitor and Analyze
			//RT expression: RT*(1+(Er)*(*settings.SetPointInSeconds*10))

			if !settings.PerformanceMeasurement {

				u1 = services[i].Controller.Update(*settings.SetPointInSeconds, Rt)

			} else {
				if len(settings.SetpointList)+1 == iteration {
					fmt.Println("Performance Measurement Terminated")
					os.Exit(0)
				}
				u1 = services[i].Controller.Update(settings.SetpointList[iteration], Rt)
				fmt.Println("[Operating on Performance Measurement Mode]")
				now := time.Now()
				fmt.Println("Timestamp", now, "setpoint", settings.SetpointList[iteration])

			}

			if Rt < 0.01 && Er == 0 { // condition to avoid high peak in experiment begining
				u1 = *settings.InitialValue
				services[i].Controller.SetPreviousValue(*settings.InitialValue)
			}

			//u2 := services[i].MemoryController.Update(5, monitor.GetResponseTimeFromService(services[i], 0.9)*(1+monitor.GetErrorRateFromService(services[i])))

			//fmt.Println("New value from Memory", int(u2))
			//Execute
			switch settings.AdaptionType {
			case "CPU":
				fmt.Println("New value from CPU", int(u1))
				go exec.SetCPUToPods(services[i].Namespace, services[i].Pods, int(u1), services[i].Container)
			case "ApplicationPropertyAdaption":
				fmt.Println("New value from "+propertyToAdapt[0]+","+propertyToAdapt[1]+" in configmap: "+configMapName, int(u1))
				if int(u1) == previousValue {
					fmt.Println("Skipping change, previous value is the same")
					break
				}

				go exec.SetEnvironmentPropretyValueToCM(services[i].Pods, "quarkus", services[i].Namespace, configMapName, propertyToAdapt, []int{int(u1), int(u1)})
				previousValue = int(u1)

			}

			//exec.SetMemoryToPods(services[i].Namespace, services[i].Pods, int(u2), services[i].Container)
			fmt.Println("Make change...", services[i].Container)

		}
		fmt.Println("Waiting", settings.AdaptionInterval, "seconds")
		time.Sleep(time.Second * time.Duration(settings.AdaptionInterval))
		iteration = iteration + 1
	}

}

func controllerModeMM(settings Shared.OperationSettings) {

	servicesToExclude := strings.Split(os.Getenv("SERVICES_TO_EXCLUDE"), ",")
	iteration := 0
	first := true

	u1 := 0.0

	var k Providers.Kubernetes
	monitor := Entities.PrometheusMonitor{Prom: Providers.Prometheus{Address: Shared.ValidateEnvironmentVariable("PROMETHEUS_HOST"), Port: Shared.ValidateEnvironmentVariable("PROMETHEUS_PORT"), Protocol: Shared.ValidateEnvironmentVariable("PROMETHEUS_PROTOCOL")}}
	planner := Entities.Planner{K: &k, SkipList: servicesToExclude}
	services := setupNewController(&k, servicesToExclude, *settings.ControllerType, *settings.MinCPU, *settings.MaxCPU, *settings.Kp, *settings.Ki, *settings.Kd, settings.AdaptionInterval)
	exec := Entities.Executor{K8s: &k}
	var propertyToAdapt []string
	configMapName := ""
	previousValue := 0
	if settings.AdaptionType == "ApplicationPropertyAdaption" {
		propertyToAdapt = strings.Split(Shared.ValidateEnvironmentVariable("PROPERTY_ADAPT_NAME"), `,`)
		configMapName = Shared.ValidateEnvironmentVariable("CONFIGMAP_NAME")
	}

	for true {
		fmt.Println("loop")
		for i := 0; i < (settings.AdaptionInterval)/(settings.AdaptionInterval/3); i++ {
			if first {
				break //in first iteration, don't capture smoothy
			}
			//getting smooth RT
			time.Sleep(time.Second * time.Duration(settings.AdaptionInterval/3))
			for i := 0; i < len(services); i++ {
				fmt.Println("Getting partial response time")
				RT, errRT := monitor.GetResponseTimeFromService(services[i], 0.9, "MultipleSetpoints")

				services[i].RT += RT
				fmt.Println("service", services[i].Name, "RT partial", services[i].RT)
				if errRT != nil {
					panic(errRT)
				}

			}

		}

		for i := 0; i < len(services); i++ {

			Er, errTX := monitor.GetErrorRateFromService(services[i])
			if first {
				services[i].RT, _ = monitor.GetResponseTimeFromService(services[i], 0.9, "controllerModeMM")
				first = false
			} else {
				services[i].RT = services[i].RT / float64((settings.AdaptionInterval)/(settings.AdaptionInterval/3))
			}

			if errTX != nil {
				panic(errTX)
			}
			fmt.Println(services[i].Name, "RT:", services[i].RT, "Err:", Er)
			//plan
			planner.GetPodsFromService(&services[i]) //it's important because pods can scale up or down.
			planner.GetContainerFromPod(&services[i])

			//Monitor and Analyze
			//RT expression: RT*(1+(Er)*(*settings.SetPointInSeconds*10))

			if !settings.PerformanceMeasurement {

				u1 = services[i].Controller.Update(*settings.SetPointInSeconds, services[i].RT)

			} else {
				if len(settings.SetpointList)+1 == iteration {
					fmt.Println("Performance Measurement Terminated")
					os.Exit(0)
				}
				u1 = services[i].Controller.Update(settings.SetpointList[iteration], services[i].RT)
				fmt.Println("[Operating on Performance Measurement Mode]")
				now := time.Now()
				fmt.Println("Timestamp", now, "setpoint", settings.SetpointList[iteration])

			}
			services[i].RT = 0 //set value to zero to restart counter

			//u2 := services[i].MemoryController.Update(5, monitor.GetResponseTimeFromService(services[i], 0.9)*(1+monitor.GetErrorRateFromService(services[i])))

			//fmt.Println("New value from Memory", int(u2))
			//Execute
			switch settings.AdaptionType {
			case "CPU":
				fmt.Println("New value from CPU", int(u1))
				go exec.SetCPUToPods(services[i].Namespace, services[i].Pods, int(u1), services[i].Container)
			case "ApplicationPropertyAdaption":
				fmt.Println("New value from "+propertyToAdapt[0]+","+propertyToAdapt[1]+" in configmap: "+configMapName, int(u1))
				if int(u1) == previousValue {
					fmt.Println("Skipping change, previous value is the same")
					break
				}

				go exec.SetEnvironmentPropretyValueToCM(services[i].Pods, "quarkus", services[i].Namespace, configMapName, propertyToAdapt, []int{int(u1), int(u1)})
				previousValue = int(u1)

			}

			//exec.SetMemoryToPods(services[i].Namespace, services[i].Pods, int(u2), services[i].Container)
			fmt.Println("Make change...", services[i].Container)

		}
		fmt.Println("Waiting", settings.AdaptionInterval, "seconds")

		iteration = iteration + 1
	}

}

func SystemModelingMode(settings Shared.OperationSettings) {

	fmt.Println("Operation Mode: System Modeling")
	samples := Shared.ConvertStringToInt(Shared.ValidateEnvironmentVariable("SAMPLES")) //Define how many samples will be collected for each input plant value
	fmt.Println("Samples per level:", samples)
	path := Shared.ValidateEnvironmentVariable("PATH_TO_SAVE_MODELING") //should terminate with .csv
	fmt.Println("Path to save data", path)
	data := make(map[string]Entities.InputOutput)
	minInput := *settings.MinCPU
	maxInput := *settings.MaxCPU
	sysm := Entities.CreateSystemModeling(samples, "CPU", "ResponseTime", path, minInput, maxInput)

	servicesToExclude := strings.Split(os.Getenv("SERVICES_TO_EXCLUDE"), ",")

	var k Providers.Kubernetes
	monitor := Entities.PrometheusMonitor{Prom: Providers.Prometheus{Address: Shared.ValidateEnvironmentVariable("PROMETHEUS_HOST"), Port: Shared.ValidateEnvironmentVariable("PROMETHEUS_PORT"), Protocol: Shared.ValidateEnvironmentVariable("PROMETHEUS_PROTOCOL")}}
	planner := Entities.Planner{K: &k, SkipList: servicesToExclude}

	services := setupNewController(&k, servicesToExclude, *settings.ControllerType, *settings.MinCPU, *settings.MaxCPU, *settings.Kp, *settings.Ki, *settings.Kd, settings.AdaptionInterval)
	exec := Entities.Executor{K8s: &k}
	//skipChange := true
	//outputValueAcc := 0
	outputAcc := 0.0
	minInputValue := sysm.GetMinInputValue()
	inputValue := sysm.GetMinInputValue()
	maxInputValue := sysm.GetMaxInputValue()
	fmt.Println("Minimum Input Value is:", inputValue)
	fmt.Println("Maximum Input Value is:", maxInput)
	skipCount := 0
	for {

		if skipCount >= samples {
			fmt.Println("Send data to csv")
			sysm.Capture(data) // send data to write in file

			if (inputValue + minInputValue*settings.ModelingPercentageStep) <= maxInputValue {
				inputValue = inputValue + minInputValue*settings.ModelingPercentageStep

				skipCount = 0

			} else {
				fmt.Println("Tests Ended")
				sysm.FinalizeExperiment()
				os.Exit(0)
			}

			for _, service := range services {

				planner.GetPodsFromService(&service) //it's important because pods can scale up or down.
				planner.GetContainerFromPod(&service)
				exec.SetCPUToPods(service.Namespace, service.Pods, int(inputValue), service.Container)
				fmt.Println("CPU updated to", inputValue, "in service", service.Name)
				fmt.Println("Reseting data to service", service.Name)
				data[service.Name] = Entities.InputOutput{Input: int(math.Round(inputValue)), Output: 0}

			}

		} else {

			for _, service := range services {
				planner.GetPodsFromService(&service) //it's important because pods can scale up or down.
				planner.GetContainerFromPod(&service)
				rt, _ := monitor.GetResponseTimeFromService(service, 0.9, "SystemModeling")
				outputAcc = data[service.Name].Output + rt
				fmt.Println("Saving response time:", rt, "seconds")
				data[service.Name] = Entities.InputOutput{Input: int(inputValue), Output: outputAcc}
				fmt.Println(data)
			}
			skipCount = skipCount + 1
			time.Sleep(time.Second * time.Duration(settings.AdaptionInterval))
			fmt.Println("Waiting", settings.AdaptionInterval, "seconds...")
		}

	}
}
func main() {

	var settings Shared.OperationSettings
	opMode := flag.String("opMode", "", "Available Types: MetricForwarder or Controller")
	controllerType := flag.String("controllerType", "", "Available Types: AStar, PID, ratioController")
	minCPU := flag.Float64("minCPU", 200, "Set Minimum CPU value in milicores")
	maxCPU := flag.Float64("maxCPU", 1024, "Set Maximum CPU value in milicores")
	adaptionInterval := flag.Float64("adaptionInterval", 30, "Set interval between adaptions")
	setPointInSeconds := flag.Float64("setpoint", 5, "Target controller value. It is the plant output")
	kp := flag.Float64("kp", 1, "Proporcional constant in PID controller")
	initialValue := flag.Float64("initialValue", 600, "Initial Value for Controled Variable")
	ki := flag.Float64("ki", 1, "Integral constant in PID controller")
	kd := flag.Float64("kd", 1, "Differential constant in PID controller")
	flag.Parse()
	fmt.Println("Params:", "minCPU:", *minCPU, "maxCPU:", *maxCPU, "controllerType:", *controllerType, "setpoint:", *setPointInSeconds)

	if !(*opMode == "MetricForwarder" || *opMode == "Controller" || *opMode == "SystemModeling") {

		*opMode = Shared.ValidateEnvironmentVariable("OP_MODE")
	}

	if !(*controllerType == "AStar" || *controllerType == "PID" || *controllerType == "MRAC" || *controllerType == "ratioController" || *opMode == "MetricForwarder") {

		*controllerType = Shared.ValidateEnvironmentVariable("CONTROLLER_TYPE")

	}

	switch *opMode {
	case "PerformanceMeasurement":
		simulationDuration := 24 * 60 //seconds -- simulation duration
		settings.Kp = kp
		settings.Kd = kd
		settings.Ki = ki
		settings.MaxCPU = maxCPU
		settings.InitialValue = initialValue
		settings.MinCPU = minCPU
		setPoints := []float64{1, 2, 1, 1.5, 1, 2}
		settings.ControllerType = controllerType
		settings.AdaptionInterval = int(*adaptionInterval)
		settings.AdaptionType = Shared.ValidateEnvironmentVariable("ADAPTION_TYPE")
		settings.PerformanceMeasurement = true
		slots := (simulationDuration / settings.AdaptionInterval) / len(setPoints)
		settings.SetpointList = RepeatEach(setPoints, slots)
		controllerMode(settings)
	case "SystemModeling":
		settings.ModelingPercentageStep = 0.1
		settings.AdaptionInterval = int(*adaptionInterval)
		settings.Kp = kp
		settings.Kd = kd
		settings.Ki = ki
		settings.MaxCPU = maxCPU
		settings.MinCPU = minCPU
		settings.SetPointInSeconds = setPointInSeconds
		settings.ControllerType = controllerType

		SystemModelingMode(settings)
	case "MetricForwarder":
		metricForwarderMode(30)
	case "Controller":
		settings.Kp = kp
		settings.InitialValue = initialValue
		settings.Kd = kd
		settings.Ki = ki
		settings.MaxCPU = maxCPU
		settings.MinCPU = minCPU
		settings.SetPointInSeconds = setPointInSeconds
		settings.ControllerType = controllerType
		settings.AdaptionInterval = int(*adaptionInterval)
		settings.AdaptionType = Shared.ValidateEnvironmentVariable("ADAPTION_TYPE")
		settings.PerformanceMeasurement = false
		controllerMode(settings)

	default:
		panic("Operation modes available are: MetricForwarder or Controller. Given: " + *controllerType)

	}
	//### End Setting Flag

	//[]string{"frontend-node-port", "kubernetes", "frontend-external", "prometheus-operated", "prometheus-operator", "prometheus", "redis-cart"}

}

func RepeatEach(values []float64, count int) []float64 {
	var result []float64
	for _, v := range values {
		for i := 0; i < count; i++ {
			result = append(result, v)
		}
	}
	return result
}

/*
var memory int = 128
	var cpu int = 200
	var podsToExclude []string
	podsToExclude = append(podsToExclude, "loadgenerator-c48b4f9c4-4sfdq")
	podsToExclude = append(podsToExclude, "redis-cart-7b76cf556-v6bnt")

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	for {
		fmt.Println("Making change...")
		fmt.Println("CPU:", cpu, "Memory:", memory)
		pods := getAllPodsInNamespace("default", clientset, podsToExclude)
		err := applyCpuAndMemoryConfiguration(pods, "default", clientset, strconv.Itoa(memory), strconv.Itoa(cpu))
		if err != nil {
			panic(err.Error())
		}
		/*fmt.Println("CPU")
		fmt.Scanf("CPU: %s", &inputCPU)
		fmt.Println("Memory")
		fmt.Scanf("Memory: %s", &inputMemory)*/

// Examples for error handling:
// - Use helper functions like e.g. errors.IsNotFound()
// - And/or cast to StatusError and use its properties like e.g. ErrStatus.Message
//namespace := "default"
//pod := "nginx"
//_, err = clientset.CoreV1().Pods(namespace).Get(context.TODO(), pod, metav1.GetOptions{})

/*} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
		fmt.Printf("Error getting pod %s in namespace %s: %v\n",
			pod, namespace, statusError.ErrStatus.Message)
	} else if err != nil {
		panic(err.Error())
	} else {
		fmt.Printf("Found pod %s in namespace %s\n", pod, namespace)
	}

	time.Sleep(60 * time.Second)
	cpu = cpu + 100
	memory = memory + 100
}
*/
