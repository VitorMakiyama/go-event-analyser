package handler

import (
	"encoding/json"
	"fmt"
	"go-event-analyser/repository"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type EventRequest struct {
	SubjectID  int64     `json:"subject_id"`
	Ocurrences int       `json:"ocurrences"`
	InsertTS   time.Time `json:"inserted_ts"`
}

type EventResponse struct {
	ID         int64     `json:"id"`
	SubjectID  int64     `json:"subject_id"`
	Ocurrences int       `json:"ocurrences"`
	InsertTS   time.Time `json:"inserted_ts"`
	LastUpdate time.Time `json:"last_update"`
}

func CreateEventResponse(e repository.Event) EventResponse {
	return EventResponse{
		ID:         e.ID,
		SubjectID:  e.SubjectID,
		Ocurrences: e.Ocurrences,
		InsertTS:   e.InsertTS,
		LastUpdate: e.LastUpdate,
	}
}

func CreateEvent(w http.ResponseWriter, r *http.Request) {
	request := EventRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		fmt.Println("error decoding request body: ", err, " Body:", r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	newEvent := repository.Event{
		SubjectID:  request.SubjectID,
		Ocurrences: request.Ocurrences,
		InsertTS:   time.Now(),
		LastUpdate: time.Now(),
	}
	newEvent.ID, err = repo.InsertEvent(newEvent)
	if err != nil {
		fmt.Println("error inserting event: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := CreateEventResponse(newEvent)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func GetEvent(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		fmt.Println("error getting query params: ", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	event, err := repo.GetEvent(id)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		fmt.Println("error getting event: ", err)
		return
	}

	response := CreateEventResponse(event)

	json.NewEncoder(w).Encode(response)
}

func UpdateEvent(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		fmt.Println("error getting query params: ", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	request := EventRequest{}
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		fmt.Println("error decoding request body: ", err, " Body:", r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	updatedEvent := repository.Event{
		ID:         id,
		SubjectID:  request.SubjectID,
		Ocurrences: request.Ocurrences,
		InsertTS:   request.InsertTS,
		LastUpdate: time.Now(),
	}
	updatedEvent, err = repo.UpdateEvent(updatedEvent)
	if err != nil {
		fmt.Println("error inserting event: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := CreateEventResponse(updatedEvent)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}
