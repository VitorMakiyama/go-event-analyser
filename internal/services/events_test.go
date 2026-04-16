package services

import (
	"errors"
	"fmt"
	"go-event-analyser/internal/repository"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEventService_Create(t *testing.T) {
	dateTimeWithConflict := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.Now().Location())
	dateTime := time.Date(2026, time.March, 24, 16, 20, 0, 0, time.Now().Location())

	tests := []struct {
		name              string
		eventToBeInserted repository.Event
		wantError         error
		expectedResult    repository.Event
	}{
		{
			name: "Ok - Create new Event",
			eventToBeInserted: repository.Event{
				ID:          0, // DB creates it
				SubjectID:   1,
				Occurrences: 1,
				InsertTS:    dateTime,
				LastUpdate:  time.Time{},
			},
			wantError: nil,
			expectedResult: repository.Event{
				ID:          2,
				SubjectID:   1,
				Occurrences: 1,
				InsertTS:    dateTime,
				LastUpdate:  time.Now(),
			},
		},
		{
			name: "Error - Subject id not found",
			eventToBeInserted: repository.Event{
				SubjectID:   -1,
				Occurrences: 0,
				InsertTS:    dateTime,
			},
			wantError:      repository.ErrorSubjectIDNotFound{},
			expectedResult: repository.Event{},
		},
		{
			name: "Error - Conflict event already exists",
			eventToBeInserted: repository.Event{
				SubjectID:   1,
				Occurrences: 1,
				InsertTS:    dateTimeWithConflict,
			},
			wantError: ErrorEventDateConflict{},
			expectedResult: repository.Event{
				ID:          1,
				SubjectID:   1,
				Occurrences: 1,
				InsertTS:    dateTimeWithConflict,
				LastUpdate:  time.Now(),
			},
		},
		{
			name: "Error - Unknown CheckEventExistenceByDate repository error",
			eventToBeInserted: repository.Event{
				SubjectID:   1,
				Occurrences: 1,
				InsertTS:    time.Time{},
			},
			wantError:      errors.New(""),
			expectedResult: repository.Event{},
		},
		{
			name: "Error - Insert repository error",
			eventToBeInserted: repository.Event{
				ID:          -1,
				SubjectID:   1,
				Occurrences: -1,
				InsertTS:    time.Date(2026, time.April, 16, 0, 0, 0, 0, time.Local),
			},
			wantError: errors.New(""),
			expectedResult: repository.Event{
				ID:          -1,
				SubjectID:   1,
				Occurrences: -1,
				InsertTS:    time.Date(2026, time.April, 16, 0, 0, 0, 0, time.Local),
				LastUpdate:  time.Time{},
			},
		},
	}

	repository := repository.NewRepositoryFake()
	service := NewEventsService(repository)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calls my service
			event, err := service.Create(tt.eventToBeInserted)

			// Analyse the result
			if tt.wantError != nil {
				assert.ErrorAs(t, err, &tt.wantError)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult.ID, event.ID)
			assert.Equal(t, tt.expectedResult.SubjectID, event.SubjectID)
			assert.Equal(t, tt.expectedResult.Occurrences, event.Occurrences)
			assert.Equal(t, tt.expectedResult.InsertTS, event.InsertTS)
		})
	}
}

func TestEventService_Get(t *testing.T) {
	tests := []struct {
		name           string
		eventID        int64
		wantError      error
		expectedResult repository.Event
	}{
		{
			name:      "Ok - Get Event",
			eventID:   1,
			wantError: nil,
			expectedResult: repository.Event{
				ID:          1,
				SubjectID:   1,
				Occurrences: 1,
				InsertTS:    time.Date(2026, time.January, 01, 0, 0, 0, 0, time.Now().Location()),
				LastUpdate:  time.Now(),
			},
		},
		{
			name:           "Error - Event id not found",
			eventID:        0,
			wantError:      repository.ErrorEventIDNotFound{},
			expectedResult: repository.Event{},
		},
		{
			name:           "Error - Unknown repository error",
			eventID:        -1,
			wantError:      errors.New(""),
			expectedResult: repository.Event{},
		},
	}

	repository := repository.NewRepositoryFake()
	service := NewEventsService(repository)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calls my service
			event, err := service.Get(tt.eventID)

			// Analyse the result
			if tt.wantError != nil {
				assert.ErrorAs(t, err, &tt.wantError)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult.ID, event.ID)
			assert.Equal(t, tt.expectedResult.SubjectID, event.SubjectID)
			assert.Equal(t, tt.expectedResult.Occurrences, event.Occurrences)
			assert.Equal(t, tt.expectedResult.InsertTS, event.InsertTS)
		})
	}
}

func TestEventService_Update(t *testing.T) {
	dateTimeWithConflict := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.Now().Location())

	tests := []struct {
		name             string
		eventToBeUpdated repository.Event
		wantError        error
		expectedResult   repository.Event
	}{
		{
			name: "Ok - Update Event's occurrence",
			eventToBeUpdated: repository.Event{
				ID:          1,
				SubjectID:   1,
				Occurrences: 100,
				InsertTS:    dateTimeWithConflict,
				LastUpdate:  time.Time{},
			},
			wantError: nil,
			expectedResult: repository.Event{
				ID:          1,
				SubjectID:   1,
				Occurrences: 100,
				InsertTS:    dateTimeWithConflict,
				LastUpdate:  time.Now(),
			},
		},
		{
			name: "Ok - Update Event's Subject ID",
			eventToBeUpdated: repository.Event{
				ID:          1,
				SubjectID:   2,
				Occurrences: 1,
				InsertTS:    dateTimeWithConflict,
				LastUpdate:  time.Time{},
			},
			wantError: nil,
			expectedResult: repository.Event{
				ID:          1,
				SubjectID:   2,
				Occurrences: 1,
				InsertTS:    dateTimeWithConflict,
				LastUpdate:  time.Now(),
			},
		},
		{
			name: "Error event not found",
			eventToBeUpdated: repository.Event{
				ID:          -1, // For returning a error in GetEvent
				SubjectID:   1,
				Occurrences: 1,
				LastUpdate:  time.Time{},
			},
			wantError:      repository.ErrorEventIDNotFound{},
			expectedResult: repository.Event{},
		},
		{
			name: "Error updating event",
			eventToBeUpdated: repository.Event{
				ID:          1,
				SubjectID:   -1, // For returning a error in UpdateEvent
				Occurrences: 1,
				LastUpdate:  time.Time{},
			},
			wantError:      errors.New(""),
			expectedResult: repository.Event{},
		},
	}

	repository := repository.NewRepositoryFake()
	service := NewEventsService(repository)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calls my service
			event, err := service.Update(tt.eventToBeUpdated)

			// Analyse the result
			if tt.wantError != nil {
				assert.ErrorAs(t, err, &tt.wantError)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult.ID, event.ID)
			assert.Equal(t, tt.expectedResult.SubjectID, event.SubjectID)
			assert.Equal(t, tt.expectedResult.Occurrences, event.Occurrences)
			assert.Equal(t, tt.expectedResult.InsertTS, event.InsertTS)
		})
	}
}

func TestEventService_Error(t *testing.T) {
	date :=  time.Date(2026, time.March, 24, 16, 20, 0, 0, time.Now().Location())
	e := ErrorEventDateConflict{
		date: date,
	}

	expectedString := fmt.Sprintf("event with date %s already exists", date.Format(time.DateOnly))
	assert.Equal(t, expectedString, e.Error())
}
