package handlers

import (
	"encoding/json"
	"errors"
	"go-event-analyser/internal/repository"
	"go-event-analyser/internal/services"
	"log"
	"net/http"
	"strconv"
	"time"
)

type EventsHandler struct {
	service services.EventsServiceBase
}

func NewEventsHandler(service services.EventsServiceBase) EventsHandler {
	return EventsHandler{
		service: service,
	}
}

type EventRequest struct {
	SubjectID   int64  `json:"subject_id"`
	Occurrences int    `json:"occurrences"`
	InsertTS    string `json:"insert_ts"`
}

type EventResponse struct {
	ID          int64     `json:"id"`
	SubjectID   int64     `json:"subject_id"`
	Occurrences int       `json:"occurrences"`
	InsertTS    time.Time `json:"insert_ts"`
	LastUpdate  time.Time `json:"last_update"`
}

func CreateEventResponse(e repository.Event) EventResponse {
	return EventResponse{
		ID:          e.ID,
		SubjectID:   e.SubjectID,
		Occurrences: e.Occurrences,
		InsertTS:    e.InsertTS,
		LastUpdate:  e.LastUpdate,
	}
}

func CreateEventFromRequest(request EventRequest, parsedInsertTS time.Time) repository.Event {
	newEvent := repository.Event{
		SubjectID:   request.SubjectID,
		Occurrences: request.Occurrences,
		InsertTS:    time.Now(),
		LastUpdate:  time.Now(),
	}

	if !parsedInsertTS.IsZero() { // if it is not zero, insert_ts was sent
		newEvent.InsertTS = parsedInsertTS.Local()
	}

	return newEvent
}

func (e *EventsHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	request := EventRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Println("CreateEvent - error decoding request body: ", err, " Body:", r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	insertTS, err := parseTimeFromStringRFC3339(request.InsertTS)
	if err != nil {
		log.Printf("CreateEvent - error parsing time from timestring (%s): %v", request.InsertTS, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Printf("CreateEvent - Received request body: %v\n", request)

	newEvent := CreateEventFromRequest(request, insertTS)

	createdEvent, err := e.service.Create(newEvent)

	if err != nil {
		if errors.As(err, &repository.ErrorSubjectIDNotFound{}) {
			log.Println("CreateEvent - subject_id not found in DB: ", err)
			w.WriteHeader(http.StatusNotFound)
			return
		} else if errors.As(err, &services.ErrorEventDateConflict{}) {
			// Found an Event with same inserted_ts, conflict!
			log.Println("CreateEvent - conflict: ", err)
			response := CreateEventResponse(createdEvent)
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(response)
			return
		} else {
			// Unknown error
			log.Println("CreateEvent - error inserting event: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	response := CreateEventResponse(createdEvent)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

/*
* Function that parses a string date in RFC3339 format ("2006-01-02T15:04:05Z07:00")
 */
func parseTimeFromStringRFC3339(timeString string) (time.Time, error) {
	// Tries to parse using RFC3339 PREFERRED
	parsedTime, err := time.Parse(time.RFC3339, timeString)
	if err == nil {
		log.Printf("Parsed time string (%s), with RFC3339 (%s): %v\n", timeString, time.RFC3339, parsedTime)
		return parsedTime, nil
	}

	return time.Time{}, err
}

func (e *EventsHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		log.Println("error getting query params: ", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Println("GetEvent - Received request for id: ", id)

	event, err := e.service.Get(id)
	if err != nil {
		if errors.As(err, &repository.ErrorEventIDNotFound{}) {
			log.Println("event_id not found in DB: ", err)
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println("error getting event: ", err)
		}
		return
	}

	response := CreateEventResponse(event)

	json.NewEncoder(w).Encode(response)
}

func (e *EventsHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		log.Println("error getting query params: ", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Println("UpdateEvent - Received request for id: ", id)

	request := EventRequest{}
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Println("error decoding request body: ", err, " Body:", r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Printf("UpdateEvent - Received request body: %v\n", request)


	newEvent := repository.Event{
		ID:          id,
		SubjectID:   request.SubjectID,
		Occurrences: request.Occurrences,
	}
	updatedEvent, err := e.service.Update(newEvent)
	if err != nil {
		if errors.As(err, &repository.ErrorEventIDNotFound{}) {
			log.Println("event_id not found in DB: ", err)
			w.WriteHeader(http.StatusNotFound)
			return
		} else {
			// Unknwon error
			log.Println("error inserting event: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	response := CreateEventResponse(updatedEvent)
	log.Printf("updated event %v to %v", newEvent, updatedEvent)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
