package handler

import (
	"encoding/json"
	"errors"
	"go-event-analyser/repository"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type EventRequest struct {
	SubjectID   int64  `json:"subject_id"`
	Occurrences int    `json:"occurrences"`
	InsertUTC   string `json:"insert_utc"`
}

type EventResponse struct {
	ID          int64     `json:"id"`
	SubjectID   int64     `json:"subject_id"`
	Occurrences int       `json:"occurrences"`
	InsertUTC   time.Time `json:"insert_utc"`
	LastUpdate  time.Time `json:"last_update"`
}

func CreateEventResponse(e repository.Event) EventResponse {
	return EventResponse{
		ID:          e.ID,
		SubjectID:   e.SubjectID,
		Occurrences: e.Occurrences,
		InsertUTC:   e.InsertUTC,
		LastUpdate:  e.LastUpdate,
	}
}

func CreateEventFromRequest(request EventRequest, parsedInsertUTC time.Time) repository.Event {
	newEvent := repository.Event{
		SubjectID:   request.SubjectID,
		Occurrences: request.Occurrences,
		InsertTS:    time.Now(),
		InsertUTC:   time.Now().UTC(),
		LastUpdate:  time.Now().UTC(),
	}

	if !parsedInsertUTC.IsZero() { // if it is not zero, insert_ts was sent
		newEvent.InsertTS = parsedInsertUTC.Local()
		newEvent.InsertUTC = parsedInsertUTC
	}

	return newEvent
}

func CreateEvent(w http.ResponseWriter, r *http.Request) {
	request := EventRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Println("error decoding request body: ", err, " Body:", r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	insertUTC, err := parseTimeFromVariousString(request.InsertUTC)
	if err != nil {
		log.Printf("error parsing time from timestring (%s): %v", request.InsertUTC, err)
	}

	newEvent := CreateEventFromRequest(request, insertUTC)

	// Verify if there is already a inserted_ts with the same date, return early with status 409 - Conflict
	foundE, err := repo.CheckEventExistenceByDate(newEvent.InsertUTC)
	if err == nil {
		response := CreateEventResponse(foundE)
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(response)
		return
	} else if !strings.Contains(err.Error(), "no rows in result") {
		// if is another error, other than "no rows in result..." something is really wrong!
		log.Printf("error checking event with date %s existence in db: %s\n", newEvent.InsertTS.Format(time.DateOnly), err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	newEvent.ID, err = repo.InsertEvent(newEvent)
	if err != nil {
		if errors.As(err, &repository.ErrorSubjectIDNotFound{}) {
			log.Println("subject_id not found in DB: ", err)
			w.WriteHeader(http.StatusNotFound)
			return
		} else {
			// Unknonw error
			log.Println("error inserting event: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	response := CreateEventResponse(newEvent)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func parseTimeFromVariousString(timeString string) (time.Time, error) {
	// Tries to parse using RFC3339
	parsedTime, err := time.Parse(time.RFC3339, timeString)
	if err == nil {
		log.Printf("Parsed %v, with RFC3339 (%s) the received time string: %s\n", parsedTime, time.RFC3339, timeString)
		return parsedTime, nil
	}

	// Then tries with RFC3339Nano
	parsedTime, err = time.Parse(time.RFC3339Nano, timeString)
	if err == nil {
		log.Printf("Parsed %v, with RFC3339Nano (%s) the received time string: %s\n", parsedTime, time.RFC3339Nano, timeString)
		return parsedTime, nil
	}

	// Then tries with DateTime
	parsedTime, err = time.Parse(time.DateTime, timeString)
	if err == nil {
		log.Printf("Parsed %v, with DateTime (%s) the received time string: %s\n", parsedTime, time.DateTime, timeString)
		return parsedTime, nil
	}

	// Then tries with a custom layout
	customLayout := "2006-01-02T15:04:05.999999"
	parsedTime, err = time.Parse(customLayout, timeString)
	if err == nil {
		log.Printf("Parsed %v, with custom layout (%s) the received time string: %s\n", parsedTime, customLayout, timeString)
		return parsedTime, nil
	}

	return time.Time{}, err
}

func GetEvent(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		log.Println("error getting query params: ", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	event, err := repo.GetEvent(id)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result") {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		log.Println("error getting event: ", err)
		return
	}

	response := CreateEventResponse(event)

	json.NewEncoder(w).Encode(response)
}

func UpdateEvent(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		log.Println("error getting query params: ", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	oldEvent, err := repo.GetEvent(id)
	if err != nil && errors.As(err, &repository.ErrorEventIDNotFound{}) {
		log.Println("event_id not found in DB: ", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	request := EventRequest{}
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Println("error decoding request body: ", err, " Body:", r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	updatedEvent := repository.Event{
		ID:          id,
		SubjectID:   request.SubjectID,
		Occurrences: request.Occurrences,
		LastUpdate:  time.Now().UTC(),
	}
	updatedEvent, err = repo.UpdateEvent(updatedEvent)
	if err != nil {
		log.Println("error inserting event: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := CreateEventResponse(updatedEvent)
	log.Printf("updated event %v to %v", oldEvent, updatedEvent)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
