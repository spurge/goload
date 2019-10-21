package main

type Runner struct {
	Requests RequestCollectionHandler
	History  HistoryHandler
	Status   *Status
}

func (r *Runner) Run() {
	for request := r.Requests.First(); request != nil; request = r.Requests.Next() {
		request.SetParser(r.History)

		response, err := request.Send()

		r.Status.Record(
			request.GetName(),
			response.Latency,
			response.RealStatusCode,
			response.Body,
			err,
		)

		if err == nil {
			r.History.Record(request.GetName(), response.Body)
		}
	}
}
