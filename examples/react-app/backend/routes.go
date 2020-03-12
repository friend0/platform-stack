package main

import (
	"encoding/json"
	"net/http"
	"time"
)

type Todo struct {
	Name      string    `json:"name"`
	Completed bool      `json:"completed"`
	Due       time.Time `json:"due"`
}

type Todos []Todo

func (s *Server) routes() {
	todosRouter := s.Router.PathPrefix("/todos").Subrouter()
	todosRouter.HandleFunc("/", s.Discover()).Methods("GET")
}


func (s *Server) Discover() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		todos := Todos{
			Todo{Name: "Write presentation"},
			Todo{Name: "Host meetup"},
		}

		json.NewEncoder(w).Encode(todos)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(todos); err != nil {
			panic(err)
		}
	}
}

