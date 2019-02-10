package main

import (
	"time"
)

type Runner struct {
	Requests RequestCollectionHandler
	History  HistoryHandler
}

func (r *Runner) Run() {
	for request := r.Requests.First(); request != nil; request = r.Requests.Next() {
		request.SetParser(r.History)
		requestLatency := RequestLatencySummary.WithLabelValues(request.GetName())

		then := time.Now()
		response, err := request.Send()
		requestLatency.Observe(float64(time.Since(then).Nanoseconds()) / 1000)

		if err != nil {
			RequestStatusCounter.WithLabelValues(request.GetName(), "error").Inc()
		} else {
			RequestStatusCounter.WithLabelValues(request.GetName(), response.StatusCode).Inc()
			r.History.Record(request.GetName(), response.Body)
		}
	}
}
