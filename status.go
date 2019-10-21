package main

import (
	"encoding/json"
	"net/http"
	"sync"
)

type Status struct {
	Responses chan *StatusEntry         `json:"-"`
	Mutex     sync.Mutex                `json:"-"`
	Slowest   map[string][]*StatusEntry `json:"slowest"`
	Errors    map[string][]*StatusEntry `json:"errors"`
}

var _ http.Handler = &Status{}

func NewStatus() *Status {
	s := &Status{
		Responses: make(chan *StatusEntry, 100),
		Slowest:   make(map[string][]*StatusEntry),
		Errors:    make(map[string][]*StatusEntry),
	}

	go s.loop()

	return s
}

func (s *Status) Handler() http.Handler {
	return s
}

func (s *Status) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	data, err := json.Marshal(s)

	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)

		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.Write(data)
}

func (s *Status) Record(
	name string,
	latency float64,
	status int,
	response string,
	err error,
) {
	var encoded interface{}

	if json.Unmarshal([]byte(response), &encoded) != nil {
		encoded = response
	}

	errorString := ""

	if err != nil {
		errorString = err.Error()
	}

	s.Responses <- &StatusEntry{
		Name:     name,
		Latency:  latency,
		Status:   status,
		Response: encoded,
		Error:    errorString,
	}
}

func (s *Status) loop() {
	for entry := range s.Responses {
		s.Mutex.Lock()

		_, ok := s.Errors[entry.Name]

		if !ok {
			s.Errors[entry.Name] = make([]*StatusEntry, 0)
		}

		if len(entry.Error) > 0 {
			errors := append([]*StatusEntry{entry}, s.Errors[entry.Name]...)

			if len(errors) > 3 {
				s.Errors[entry.Name] = errors[:2]
			} else {
				s.Errors[entry.Name] = errors
			}
		} else {
			slowest, ok := s.Slowest[entry.Name]

			if !ok {
				s.Slowest[entry.Name] = []*StatusEntry{entry}
			} else {
				for i, e := range slowest {
					if e.Latency < entry.Latency {
						slowest = append(
							slowest[:i],
							append([]*StatusEntry{entry}, slowest[i:]...)...,
						)
						break
					}
				}

				if len(slowest) > 3 {
					s.Slowest[entry.Name] = slowest[:2]
				} else {
					s.Slowest[entry.Name] = slowest
				}
			}
		}

		s.Mutex.Unlock()
	}
}

type StatusEntry struct {
	Name     string      `json:"-"`
	Latency  float64     `json:"latency"`
	Status   int         `json:"status"`
	Response interface{} `json:"response"`
	Error    string      `json:"error,omitempty"`
}
