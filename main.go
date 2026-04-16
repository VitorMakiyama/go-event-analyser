package main

import (
	"encoding/json"
	"go-event-analyser/handlers"
	"go-event-analyser/internal/repository"
	"go-event-analyser/internal/services"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var LocalTZ time.Location

func main() {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Ping...
	r.Get("/ping", ping)

	var repository repository.Repository = repository.NewPostgreSQLRepository()
	
	// Events
	service := services.NewEventsService(repository)
	eventHandler := handlers.NewEventsHandler(service)
	eventsBase := "/events"
	r.Post(eventsBase, eventHandler.CreateEvent)
	r.Get(eventsBase, eventHandler.GetEvent)
	r.Put(eventsBase, eventHandler.UpdateEvent)

	// Subjects
	subjectHandler := handlers.NewSubjectHandler(repository)
	subjectBase := "/subjects"
	r.Post(subjectBase, subjectHandler.CreateSubject)
	r.Get(subjectBase, subjectHandler.GetSubject)

	log.Println("Listening on port 3333...")
	http.ListenAndServe(":3333", r)
}

// Function for the client to make sure the server is up and running!
func ping(w http.ResponseWriter, _ *http.Request) {
	json.NewEncoder(w).Encode("pong")
}
