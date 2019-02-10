package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	ErrorCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "goload_errors",
			Help: "Goload error counter",
		},
		[]string{"error"},
	)
	ConcurrencyParamError     = ErrorCounter.WithLabelValues("concurrency_param")
	SleepParamError           = ErrorCounter.WithLabelValues("sleep_param")
	RequestFileError          = ErrorCounter.WithLabelValues("request_file")
	ParseTemplateError        = ErrorCounter.WithLabelValues("template_parse")
	ExecuteTemplateError      = ErrorCounter.WithLabelValues("template_execute")
	MissingTemplateEntryError = ErrorCounter.WithLabelValues("template_missing_entry")
	ParamGauge                = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "goload_params",
			Help: "Goload parameters",
		},
		[]string{"param"},
	)
	RequestParamLength    = ParamGauge.WithLabelValues("request_length")
	ConcurrencyParamValue = ParamGauge.WithLabelValues("concurrency")
	SleepParamValue       = ParamGauge.WithLabelValues("sleep")
	RequestLatencySummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "goload_request_latency",
			Help:       "Goload http request latency",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"name"},
	)
	RequestStatusCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "goload_request_status",
			Help: "Gload request status code",
		},
		[]string{"name", "status"},
	)
)

func init() {
	prometheus.MustRegister(ErrorCounter)
	prometheus.MustRegister(ParamGauge)
	prometheus.MustRegister(RequestLatencySummary)
	prometheus.MustRegister(RequestStatusCounter)
}

func main() {
	host := os.Getenv("HOST")
	port := os.Getenv("PORT")

	if host == "" {
		host = "localhost"
	}

	if port == "" {
		port = "8100"
	}

	var err error
	var concurrency int64
	var sleep int64

	concurrency, err = strconv.ParseInt(os.Getenv("CONCURRENCY"), 10, 16)

	if err != nil {
		concurrency = 1

		ConcurrencyParamError.Inc()
		log.Printf("CONCURRENCY environmental variable not set: %s", err)
	}

	sleep, err = strconv.ParseInt(os.Getenv("SLEEP"), 10, 16)

	if err != nil {
		sleep = 1

		SleepParamError.Inc()
		log.Printf("SLEEP environmental variable not set: %s", err)
	}

	filename := os.Getenv("REQUESTS")

	go InitiateRequests(int(concurrency), time.Duration(sleep), filename)

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%s", host, port), nil))
}

func InitiateRequests(concurrency int, sleep time.Duration, filename string) {
	requests, err := LoadRequests(filename)

	if err != nil {
		RequestFileError.Inc()
		log.Printf("Error reading request file: %s", err)
	}

	ConcurrencyParamValue.Set(float64(concurrency))
	SleepParamValue.Set(float64(sleep))
	RequestParamLength.Set(float64(len(requests)))

	for i := 0; i < concurrency; i++ {
		go RunRequests(requests, sleep)
	}
}

func RunRequests(requests []*Request, sleep time.Duration) {
	collection := RequestCollection{Requests: requests}
	runner := Runner{
		History:  NewHistory(),
		Requests: &collection,
	}

	for {
		runner.Run()
		time.Sleep(sleep * time.Second)
	}
}
