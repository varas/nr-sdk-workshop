package main

import (
	sdkArgs "github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/data/event"
	"github.com/newrelic/infra-integrations-sdk/data/metric"
	"github.com/newrelic/infra-integrations-sdk/integration"
)

type argumentList struct {
	sdkArgs.DefaultArgumentList
}

const (
	integrationName    = "com.new-relic.sdk-workshop"
	integrationVersion = "0.1.0"
)

var (
	args argumentList
	endpointA = "http://localhost:8881"
	endpointB = "http://localhost:8882"
)

type serverStatus struct {
	statusCode int
	apiVersion string
	latencyMs int
	error string
}

func main() {
	// Initialize integration
	i, err := integration.New(integrationName, integrationVersion, integration.Args(&args))
	panicOnErr(err)

	// Create entities (name must be unique)
	a, err := i.Entity("instance-a", "web-server")
	panicOnErr(err)
	b, err := i.Entity("instance-b", "web-server")
	panicOnErr(err)

	// Fetch data (populate entities)
	err = monitorizeWebServer(a, endpointA)
	if err != nil {
		i.Logger().Errorf("cannot fetch data for endpoint: %s", endpointA)
	}
	err = monitorizeWebServer(b, endpointB)
	if err != nil {
		i.Logger().Errorf("cannot fetch data for endpoint: %s", endpointB)
	}

	// Push to New Relic
	panicOnErr(i.Publish())
}

func monitorizeWebServer(e *integration.Entity, endpoint string) error {
	s := queryServer(endpoint)

	// Add Inventory
	e.SetInventoryItem("api", "version", s.apiVersion)

	// Add Metrics
	set, err := e.NewMetricSet("status")
	if err != nil {
		return err
	}

	set.SetMetric("status", s.statusCode, metric.GAUGE)
	set.SetMetric("latency", s.latencyMs, metric.DELTA)

	if s.statusCode >= 500 {
		// Add Event
		e.AddEvent(event.New(s.error, "error"))
	}

	return nil
}

func queryServer(endpoint string) serverStatus {
	status := 200
	version := "1.0.0"
	errorMsg := ""
	latency := 150

	return serverStatus{
		statusCode: status,
		apiVersion: version,
		latencyMs: latency,
		error: errorMsg,
	}
}

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}
