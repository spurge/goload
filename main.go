package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/golang/glog"
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
	TargetsFileError          = ErrorCounter.WithLabelValues("targets_file")
	ParseUrlError             = ErrorCounter.WithLabelValues("url_parse")
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
	TargetsParamLength    = ParamGauge.WithLabelValues("targets_length")
	ConcurrencyParamValue = ParamGauge.WithLabelValues("concurrency")
	SleepParamValue       = ParamGauge.WithLabelValues("sleep")
	RequestLatencySummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "goload_request_latency",
			Help:       "Goload http request latency in milliseconds",
			Objectives: map[float64]float64{0.5: 0.05, 0.95: 0.005, 0.99: 0.001},
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
	var host string
	var port int
	var concurrency int
	var sleep int
	var targets string

	flag.StringVar(&host, "host", "0.0.0.0", "Hostname")
	flag.IntVar(&port, "port", 9115, "Port")
	flag.IntVar(&concurrency, "concurrency", 1, "Concurrency")
	flag.IntVar(&sleep, "sleep", 1, "Sleep")
	flag.StringVar(&targets, "targets", "", "Targets path")

	flag.Parse()

	go InitiateRequests(concurrency, time.Duration(sleep), targets)

	http.Handle("/metrics", promhttp.Handler())
	glog.Infof("Listens on %s:%d", host, port)
	glog.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), nil))
}

func InitiateRequests(concurrency int, sleep time.Duration, filename string) {
	glog.Infof("Loading targets from %s", filename)
	requests, err := LoadRequests(filename)

	if err != nil {
		TargetsFileError.Inc()
		glog.Errorf("Error reading targets file, %s: %s", filename, err)
	}

	ConcurrencyParamValue.Set(float64(concurrency))
	SleepParamValue.Set(float64(sleep))
	TargetsParamLength.Set(float64(len(requests)))

	for _, r := range requests {
		RequestStatusCounter.WithLabelValues(r.GetName(), "error")

		for s := 1; s < 6; s++ {
			RequestStatusCounter.WithLabelValues(r.GetName(), fmt.Sprintf("%dxx", s))
		}
	}

	glog.Infof("Starting %d request runners", concurrency)

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
		glog.Infof("%d targets requested, sleeping for %d seconds until next cycle", len(requests), sleep)
		time.Sleep(sleep * time.Second)
	}
}
