package Entities

import (
	"errors"
	p "main/Infrastructure/Providers"
)

type PrometheusMonitor struct {
	Prom p.Prometheus
}

func (m *PrometheusMonitor) GetResponseTimeFromService(service Service, percentile float64, opMode string) (float64, error) {
	query := ""
	//percentileStr := strconv.FormatFloat(percentile, 'f', -1, 64)
	//query := `scalar(histogram_quantile(` + percentileStr + `, sum(irate(istio_request_duration_milliseconds_bucket{destination_service=~"` + service.Name + `.` + service.Namespace + `.svc.cluster.local"}[1m])) by (le)) / 1000)`
	//avg by(server_address) (topk by(server_address) (5, rate(http_server_request_duration_seconds_bucket{server_address="frontend"}[1m])))
	if service.Protocol == "" {
		m.ValidateCommunicationProtocol(&service)
	}

	switch service.Protocol {
	case "http":
		if opMode != "SystemModeling" {
			query = `2*scalar(sum(rate({__name__=~"http_server_request_duration_seconds_sum|http_server_request_duration_sum",server_address="` + service.Name + `"} [20s])) / sum(rate({__name__=~"http_server_request_duration_seconds_count|http_server_request_duration_count",server_address="` + service.Name + `"} [20s])))`
		} else {
			query = `2*scalar(sum(rate({__name__=~"http_server_request_duration_seconds_sum|http_server_request_duration_sum",server_address="` + service.Name + `"} [1m])) / sum(rate({__name__=~"http_server_request_duration_seconds_count|http_server_request_duration_count",server_address="` + service.Name + `"} [1m])))`
		}
	case "rpc":
		query = `3*scalar(quantile by(server_address) (1, histogram_quantile(0.99, rate(rpc_client_duration_seconds_bucket{server_address="` + service.Name + `"}[1m]))))`
	default:
		return 0, errors.New("Response Time -- Service protocol is nil " + service.Name)

	}

	res := m.Prom.MakeQuery(query)
	return res, nil
}

func (m *PrometheusMonitor) ValidateCommunicationProtocol(service *Service) {
	queryRPC := `scalar(count by(server_address) (rpc_server_duration_seconds_bucket{server_address="` + service.Name + `"}))`
	queryHTTP := `scalar(count by(server_address) (http_server_request_duration_seconds_bucket{server_address="` + service.Name + `"}))`
	resHTTP := m.Prom.MakeQuery(queryHTTP)
	if resHTTP > 0 {
		service.Protocol = "http"
		return
	}
	resRPC := m.Prom.MakeQuery(queryRPC)

	if resRPC > 0 {
		service.Protocol = "rpc"
		return
	}
}

func (m *PrometheusMonitor) GetErrorRateFromService(service Service) (float64, error) {
	query := ""
	if service.Protocol == "" {
		m.ValidateCommunicationProtocol(&service)
	}
	switch service.Protocol {
	case "http":
		// antes de 15/02/25 -- query = `100*scalar(sum by(server_address) (topk by(server_address) (5, irate(http_server_request_duration_seconds_bucket{server_address="frontend", service_name="server", http_response_status_code!~"302|200"}[1m]))) / sum by(server_address) (topk by(server_address) (5, irate(http_server_request_duration_seconds_bucket{server_address="` + service.Name + `"}[1m]))))`
		query = `scalar(sum(rate({__name__=~"http_server_request_duration_seconds_count|http_server_request_duration_count",server_address="` + service.Name + `",http_response_status_code=~"(4|5).*"}[1m])) /  sum(rate({__name__=~"http_server_request_duration_seconds_count|http_server_request_duration_count",server_address="` + service.Name + `"}[1m])))`
	case "rpc":
		query = `100*scalar(sum by(server_address) (rate(rpc_server_duration_seconds_bucket{server_address="$backend", rpc_grpc_status_code!="0"}[1m])) / sum by(server_address) (rate(rpc_server_duration_seconds_bucket{server_address="` + service.Name + `"}[1m])))`
	default:
		return 0, errors.New("Err TX -- Service protocol is nil")

	}

	res := m.Prom.MakeQuery(query)
	return res, nil
}
