package Providers

import (
	"fmt"
	"log"
	"main/Shared"
	"net/http"
	"strings"
)

type CustomMetricIntegration struct {
	port     string
	host     string
	protocol string
}

func (c *CustomMetricIntegration) InitialiseCustomMetricIntegration() {
	c.host = Shared.ValidateEnvironmentVariable("CUSTOM_METRIC_PROVIDER_HOST")
	c.protocol = Shared.ValidateEnvironmentVariable("CUSTOM_METRIC_PROVIDER_PROTOCOL")
	c.port = Shared.ValidateEnvironmentVariable("CUSTOM_METRIC_PROVIDER_PORT")
}
func (c *CustomMetricIntegration) SetMetricValueToService(value float64, serviceName string, serviceNamespace string, metricName string) error {
	valueConverted := fmt.Sprintf("%f", value)
	path := "write-metrics/namespaces/" + serviceNamespace + "/service/" + serviceName + "/" + metricName
	url := c.protocol + "://" + c.host + ":" + c.port + "/" + path
	payload := strings.NewReader("\"" + valueConverted + "\"")
	req, _ := http.NewRequest("POST", url, payload)
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)

	if res.StatusCode != 200 {
		log.Println("Request to metric integration service has been failed with status code", res.StatusCode)
	}
	if err != nil {
		log.Println(err)
	}
	defer res.Body.Close()

	return err
}
