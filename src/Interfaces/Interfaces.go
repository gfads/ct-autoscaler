package Interfaces

type IMonitor interface {
	GetResponseTimeFromService() float64
	GetErrorRateFromService() float64
}

type IController interface {
	Initialise(p ...float64)
	Update(p ...float64) float64
	SetPreviousValue(value float64)
}

type ICustomMetric interface {
	SetMetricValueToService(value float64, serviceName string, serviceNamespace string, metricName string) error
}
