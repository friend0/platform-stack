package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type Todo struct {
	Name      string    `json:"name"`
	Completed bool      `json:"completed"`
	Due       time.Time `json:"due"`
}


func (s *Server) routes() {
	todosRouter := s.Router.PathPrefix("/todos").Subrouter()
	todosRouter.HandleFunc("/", s.Discover()).Methods("GET")
}


func (s *Server) Discover() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Handling /todos/ request from %v\n", r.Host)
		todos := []Todo{
			{Name: "Write presentation"},
			{Name: "Host meetup"},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(todos); err != nil {
			panic(err)
		}
	}
}

