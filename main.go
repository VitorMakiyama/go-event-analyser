package main

import (
	"go-event-analyser/handler"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Events
	eventsBase := "/events"
	r.Post(eventsBase, handler.CreateEvent)
	r.Get(eventsBase, handler.GetEvent)

	// Subjects
	subjectBase := "/subjects"
	r.Post(subjectBase, handler.CreateSubject)
	r.Get(subjectBase, handler.GetSubject)

	log.Println("Listening on port 3333...")
	http.ListenAndServe(":3333", r)
}
