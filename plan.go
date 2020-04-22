package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
)

const (
	namespace = "browserstack_plan"
)

type PlanStatusResponse struct {
	AutomatePlan                   string `json:"automate_plan"`
	ParallelSessionsRunning        int    `json:"parallel_sessions_running"`
	TeamParallelSessionsMaxAllowed int    `json:"team_parallel_sessions_max_allowed"`
	ParallelSessionsMaxAllowed     int    `json:"parallel_sessions_max_allowed"`
	QueuedSessions                 int    `json:"queued_sessions"`
	QueuedSessionsMaxAllowed       int    `json:"queued_sessions_max_allowed"`
}

type planApiCollector struct{}

var (
	addr     = flag.String("listen-address", "127.0.0.1:5123", "The address to listen on for HTTP requests.")
	username = flag.String("username", "", "The username for authentication to the API endpoint.")
	password = flag.String("password", "", "The password for authentication to the API endpoint.")
)

const requestURL = "https://api.browserstack.com/automate/plan.json"

var (
	parallelSessionsRunningDesc        = prometheus.NewDesc(namespace+"_parallel_sessions_running", "", []string{"plan"}, prometheus.Labels{})
	teamParallelSessionsMaxAllowedDesc = prometheus.NewDesc(namespace+"_team_parallel_sessions_max_allowed", "", []string{"plan"}, prometheus.Labels{})
	parallelSessionsMaxAllowedDesc     = prometheus.NewDesc(namespace+"_parallel_sessions_max_allowed", "", []string{"plan"}, prometheus.Labels{})
	queuedSessionsDesc                 = prometheus.NewDesc(namespace+"_queued_sessions", "", []string{"plan"}, prometheus.Labels{})
	queuedSessionsMaxAllowedDesc       = prometheus.NewDesc(namespace+"_queued_sessions_max_allowed", "", []string{"plan"}, prometheus.Labels{})
)

func (c planApiCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- parallelSessionsRunningDesc
	ch <- teamParallelSessionsMaxAllowedDesc
	ch <- parallelSessionsMaxAllowedDesc
	ch <- queuedSessionsDesc
	ch <- queuedSessionsMaxAllowedDesc
}

func (c planApiCollector) Collect(ch chan<- prometheus.Metric) {
	client := &http.Client{Timeout: time.Second * 2}
	request, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		log.Errorf("Error creating a request to %s : %v\n", requestURL, err)
	}
	request.SetBasicAuth(*username, *password)
	response, err := client.Do(request)
	if err != nil {
		log.Errorln("Error issueing a request to %s : %v\n", requestURL, err)
	}

	var planStatusResponse PlanStatusResponse
	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Errorln("Error reading a request to %s : %v\n", requestURL, err)
	}
	json.Unmarshal(content, &planStatusResponse)

	ch <- prometheus.MustNewConstMetric(
		parallelSessionsRunningDesc,
		prometheus.GaugeValue,
		float64(planStatusResponse.ParallelSessionsRunning),
		*username,
	)
	ch <- prometheus.MustNewConstMetric(
		teamParallelSessionsMaxAllowedDesc,
		prometheus.GaugeValue,
		float64(planStatusResponse.TeamParallelSessionsMaxAllowed),
		*username,
	)
	ch <- prometheus.MustNewConstMetric(
		parallelSessionsMaxAllowedDesc,
		prometheus.GaugeValue,
		float64(planStatusResponse.ParallelSessionsMaxAllowed),
		*username,
	)
	ch <- prometheus.MustNewConstMetric(
		queuedSessionsDesc,
		prometheus.GaugeValue,
		float64(planStatusResponse.QueuedSessions),
		*username,
	)
	ch <- prometheus.MustNewConstMetric(
		queuedSessionsMaxAllowedDesc,
		prometheus.GaugeValue,
		float64(planStatusResponse.QueuedSessionsMaxAllowed),
		*username,
	)
}

func main() {
	flag.Parse()

	var c planApiCollector
	prometheus.MustRegister(c)

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*addr, nil))
}
