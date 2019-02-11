package main

type Runner struct {
	Requests RequestCollectionHandler
	History  HistoryHandler
}

func (r *Runner) Run() {
	for request := r.Requests.First(); request != nil; request = r.Requests.Next() {
		request.SetParser(r.History)

		response, err := request.Send()

		if err == nil {
			r.History.Record(request.GetName(), response.Body)
		}
	}
}
