package Entities

import (
	"main/Infrastructure/Providers"
	"main/Interfaces"
)

type CustomMetric struct {
	customMetricProvider Interfaces.ICustomMetric
}

func (c *CustomMetric) Initialise(isExternal bool) {
	if isExternal {
		var CustomMetric Providers.CustomMetricIntegration
		CustomMetric.InitialiseCustomMetricIntegration()
		c.customMetricProvider = &CustomMetric
	}
}
func (c *CustomMetric) SetResponseTimeValueToService(value float64, service Service) {
	c.customMetricProvider.SetMetricValueToService(value, service.Name, service.Namespace, "response_time")
}
