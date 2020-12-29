package main

import (
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

var (
	ErrorCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "goload_errors_total",
			Help: "Goload total number of errors",
		},
		[]string{"error"},
	)
	TargetsFileError          = ErrorCounter.WithLabelValues("targets_file")
	ParseURLError             = ErrorCounter.WithLabelValues("url_parse")
	ParseTemplateError        = ErrorCounter.WithLabelValues("template_parse")
	ExecuteTemplateError      = ErrorCounter.WithLabelValues("template_execute")
	MissingTemplateEntryError = ErrorCounter.WithLabelValues("template_missing_entry")
	ExpectReCompileError      = ErrorCounter.WithLabelValues("expect_re_compile")
	RuntimeGauge              = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "goload_runtime",
			Help: "Goload runtime with parameters",
		},
		[]string{"targets_length", "concurrency", "sleep", "repeat"},
	)
	RequestLatencySummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "goload_request_latency_seconds",
			Help:       "Goload http request latency in seconds",
			Objectives: map[float64]float64{0.5: 0.05, 0.95: 0.005, 0.99: 0.001},
		},
		[]string{"name", "status"},
	)
	RequestStatusCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "goload_request_status_total",
			Help: "Goload total requests by status code",
		},
		[]string{"name", "status"},
	)
	ExpectedResponseCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "goload_expected_response_total",
			Help: "Goload total expected responses",
		},
		[]string{"name", "part"},
	)
)

func init() {
	prometheus.MustRegister(ErrorCounter)
	prometheus.MustRegister(RuntimeGauge)
	prometheus.MustRegister(RequestLatencySummary)
	prometheus.MustRegister(RequestStatusCounter)
	prometheus.MustRegister(ExpectedResponseCounter)

	logrus.SetLevel(logrus.FatalLevel)
	logrus.SetFormatter(&logrus.TextFormatter{})
}

func main() {
	var host string
	var port int
	var concurrency int
	var sleep int
	var repeat int
	var targets string
	var logLevel string
	var logFormat string

	flag.StringVar(&host, "host", "0.0.0.0", "Hostname")
	flag.IntVar(&port, "port", 9115, "Port")
	flag.IntVar(&concurrency, "concurrency", 1, "Concurrency")
	flag.IntVar(&sleep, "sleep", 1, "Sleep")
	flag.IntVar(&repeat, "repeat", -1, "Repeat, -1 <= infinite")
	flag.StringVar(&targets, "targets", "", "Targets path")
	flag.StringVar(&logLevel, "loglevel", "warn", "Log level")
	flag.StringVar(&logFormat, "logformat", "text", "Log format - text or json")

	flag.Parse()

	parsedLogLevel, err := logrus.ParseLevel(logLevel)

	if err != nil {
		logrus.
			WithError(err).
			WithField("loglevel", logLevel).
			Fatalf("Could not parse log level %s", logLevel)
	}

	logrus.SetLevel(parsedLogLevel)

	switch logFormat {
	case "text":
		logrus.SetFormatter(&logrus.TextFormatter{})
		break
	case "json":
		logrus.SetFormatter(&logrus.JSONFormatter{})
		break
	default:
		logrus.
			WithField("logformat", logFormat).
			Panicf("Unsupported log format %s", logFormat)
	}

	logrus.
		WithField("host", host).
		WithField("port", port).
		WithField("concurrency", concurrency).
		WithField("sleep", sleep).
		WithField("repeat", repeat).
		WithField("targets", targets).
		WithField("loglevel", logLevel).
		WithField("logformat", logFormat).
		Debug("Started Goload")

	closer := make(chan bool)

	status := NewStatus()

	go InitiateRequests(concurrency, time.Duration(sleep), repeat, targets, status, closer)
	go InitiateServer(host, port, status)

	<-closer
}

func InitiateServer(host string, port int, status *Status) {
	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/status", status.Handler())

	httpLogger := logrus.
		WithField("host", host).
		WithField("port", port)

	httpLogger.Info("Started HTTP server")
	httpLogger.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), nil))
}

func InitiateRequests(
	concurrency int,
	sleep time.Duration,
	repeat int,
	filename string,
	status *Status,
	closer chan bool,
) {
	reqLogger := logrus.
		WithField("concurrency", concurrency).
		WithField("sleep", sleep.String()).
		WithField("repeat", repeat).
		WithField("targets", filename)

	reqLogger.Info("Started request loop")

	requests, err := LoadRequests(filename)

	if err != nil {
		TargetsFileError.Inc()
		reqLogger.
			WithError(err).
			Error("Error reading targets file")
	}

	RuntimeGauge.
		WithLabelValues(
			strconv.Itoa(len(requests)),
			strconv.Itoa(concurrency),
			sleep.String(),
			strconv.Itoa(repeat),
		).
		SetToCurrentTime()

	for _, r := range requests {
		RequestStatusCounter.WithLabelValues(r.GetName(), "error")

		for _, status := range []string{"2xx", "4xx", "5xx"} {
			RequestStatusCounter.WithLabelValues(r.GetName(), status)
			RequestLatencySummary.WithLabelValues(r.GetName(), status)
		}
	}

	for i := 0; i < concurrency; i++ {
		go RunRequests(requests, sleep, repeat, status, closer)
	}
}

func RunRequests(
	requests []*Request,
	sleep time.Duration,
	repeat int,
	status *Status,
	closer chan bool,
) {
	collection := RequestCollection{Requests: requests}
	runner := Runner{
		History:  NewHistory(),
		Requests: &collection,
		Status:   status,
	}
	repeated := 0

	runLogger := logrus.
		WithField("requests", len(requests)).
		WithField("sleep", sleep.String()).
		WithField("repeat", repeat).
		WithField("repeated", repeated)

	for {
		runLogger.Info("Initiated requests")
		runner.Run()

		if repeat > -1 && repeated >= repeat {
			runLogger.Info("Number of repeats reached. Closing down.")
			closer <- true
			break
		}

		repeated++
		runLogger.Info("Requests ended. Sleeping intil next run.")
		time.Sleep(sleep * time.Second)
	}
}
