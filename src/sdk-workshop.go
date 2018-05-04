package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

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
	//endpointA = "http://localhost:8081/health"
	//endpointB = "http://localhost:8082/health"
	endpointA = "http://www.mocky.io/v2/5aec67303200006700fa48fb"
	endpointB = "http://www.mocky.io/v2/5aec674b3200004a00fa48fc"
)

type serverStatus struct {
	StatusCode int    `json:"status"`
	ApiVersion string `json:"version"`
	LatencyMs  int64  `json:"-"`
	Error      string `json:"error"`
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
	s, err := queryServer(endpoint)
	if err != nil {
		return err
	}

	// Add Inventory
	e.SetInventoryItem("api", "version", s.ApiVersion)

	// Add Metrics
	set, err := e.NewMetricSet("status")
	if err != nil {
		return err
	}

	set.SetMetric("status", s.StatusCode, metric.GAUGE)
	set.SetMetric("latency", s.LatencyMs, metric.GAUGE)

	if s.StatusCode >= 500 {
		// Add Event
		e.AddEvent(event.New(s.Error, "error"))
	}

	return nil
}

func queryServer(endpoint string) (s serverStatus, err error) {
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return
	}

	start := time.Now()
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	latency := time.Now().Sub(start)

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &s)

	s.LatencyMs = latency.Nanoseconds() / 1000000

	return
}

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}
