package services

import (
	"fmt"
	"go-event-analyser/internal/repository"
	"log"
	"strings"
	"time"
)

type EventsService struct {
	repository repository.Repository
}

type EventsServiceBase interface {
	Create(e repository.Event) (repository.Event, error)
	Get(id int64) (repository.Event, error)
	Update(e repository.Event) (repository.Event, error)
	Delete(id int64) (int64, error)
}

func NewEventsService(repository repository.Repository) EventsServiceBase {
	return &EventsService{
		repository: repository,
	}
}

func (es *EventsService) Create(e repository.Event) (repository.Event, error) {
	// Validate subjectID
	_, err := es.repository.GetSubject(e.SubjectID)
	if err != nil {
		return repository.Event{}, err
	}

	// Verify if there is already a inserted_ts with the same date (comparing both on Local time), return early with status 409 - Conflict
	foundE, err := es.repository.CheckEventExistenceByDate(e.InsertTS)
	if err == nil {
		// Found an Event with same inserted_ts, conflict!
		log.Printf("EventsService - found an event with date %s in db\n", foundE.InsertTS.Format(time.DateOnly))
		return foundE, ErrorEventDateConflict{ date: foundE.InsertTS }
	} else if !strings.Contains(err.Error(), "no rows in result") {
		// if is another error, other than "no rows in result..." something is really wrong!
		log.Printf("EventsService - error checking event with date %s existence in db: %s\n", e.InsertTS.Format(time.RFC3339), err.Error())
		return foundE, err
	}

	e.ID, err = es.repository.InsertEvent(e)
	if err != nil {
		log.Println("EventsService - error inserting event: ", err)
	}
	return e, err
}

func (es *EventsService) Get(id int64) (repository.Event, error) {
	event, err := es.repository.GetEvent(id)
	if err != nil {
		log.Println("error getting event: ", err)
		return repository.Event{}, err
	}
	return event, nil
}

func (es *EventsService) Update(newEvent repository.Event) (repository.Event, error) {
	oldEvent, err := es.repository.GetEvent(newEvent.ID)
	if err != nil {
		// Unknwon error
		log.Println("error getting event: ", err)
		return repository.Event{}, err
	}

	updatedEvent := repository.Event{
		ID:          oldEvent.ID,
		SubjectID:   newEvent.SubjectID,
		Occurrences: newEvent.Occurrences,
		LastUpdate:  time.Now(),
	}
	updatedEvent, err = es.repository.UpdateEvent(updatedEvent)
	if err != nil {
		log.Println("error inserting event: ", err)
		return repository.Event{}, err
	}
	return updatedEvent, nil
}

func (es *EventsService) Delete(id int64) (int64, error) {
	panic("not implemented")
}

// Custom Errors for this postgres driver
type ErrorEventDateConflict struct {
	date time.Time
}

func (e ErrorEventDateConflict) Error() string {
	return fmt.Sprintf("event with date %s already exists", e.date.Format(time.DateOnly))
}
