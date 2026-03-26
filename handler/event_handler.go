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

func CreateEvent(w http.ResponseWriter, r *http.Request) {
	request := EventRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Println("error decoding request body: ", err, " Body:", r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Validate subjectID
	_, err = repo.GetSubject(request.SubjectID)
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

	insertTS, err := parseTimeFromStringRFC3339(request.InsertTS)
	if err != nil {
		log.Printf("error parsing time from timestring (%s): %v", request.InsertTS, err)
	}
	log.Printf("CreateEvent - Received request body: %v\n", request)

	newEvent := CreateEventFromRequest(request, insertTS)

	// Verify if there is already a inserted_ts with the same date (comparing both on Local time), return early with status 409 - Conflict
	foundE, err := repo.CheckEventExistenceByDate(newEvent.InsertTS)
	if err == nil {
		// Found an Event with same inserted_ts, conflict!
		response := CreateEventResponse(foundE)
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(response)
		return
	} else if !strings.Contains(err.Error(), "no rows in result") {
		// if is another error, other than "no rows in result..." something is really wrong!
		log.Printf("error checking event with date %s existence in db: %s\n", newEvent.InsertTS.Format(time.RFC3339), err.Error())
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

func GetEvent(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		log.Println("error getting query params: ", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Println("GetEvent - Received request for id: ", id)

	event, err := repo.GetEvent(id)
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

func UpdateEvent(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		log.Println("error getting query params: ", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	oldEvent, err := repo.GetEvent(id)
	if err != nil {
		if errors.As(err, &repository.ErrorEventIDNotFound{}) {
			log.Println("event_id not found in DB: ", err)
			w.WriteHeader(http.StatusNotFound)
			return
		} else {
			// Unknwon error
			log.Println("error getting event: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
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

	updatedEvent := repository.Event{
		ID:          id,
		SubjectID:   request.SubjectID,
		Occurrences: request.Occurrences,
		LastUpdate:  time.Now(),
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
