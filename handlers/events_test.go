package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go-event-analyser/internal/repository"
	"go-event-analyser/internal/services"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEventHandler_CreateEvent(t *testing.T) {
	timeString := "2026-03-24T22:16:20-03:00"
	parsedTime, _ := time.Parse(time.RFC3339, timeString)

	tests := []struct {
		name           string
		body           interface{}
		serviceFake    services.EventsServiceBase
		wantStatus     int
		expectedResult EventResponse
	}{
		{
			name: "Ok - Create new Event",
			body: EventRequest{
				SubjectID:   1,
				Occurrences: 1,
				InsertTS:    timeString,
			},
			serviceFake: &services.EventsServiceFake{},
			wantStatus:  http.StatusCreated,
			expectedResult: EventResponse{
				ID:          0,
				SubjectID:   1,
				Occurrences: 1,
				InsertTS:    parsedTime,
				LastUpdate:  time.Now(),
			},
		},
		{
			name: "Error - Subject id not found",
			body: EventRequest{
				SubjectID:   0,
				Occurrences: 0,
				InsertTS:    timeString,
			},
			serviceFake: &services.EventsServiceFake{
				CallbackCreate: func(e repository.Event) (repository.Event, error) {
					return repository.Event{}, repository.ErrorSubjectIDNotFound{}
				},
			},
			wantStatus:     http.StatusNotFound,
			expectedResult: EventResponse{},
		},
		{
			name: "Error - error parsing timestring",
			body: EventRequest{
				SubjectID:   1,
				Occurrences: 0,
				InsertTS:    "error",
			},
			wantStatus:     http.StatusBadRequest,
			expectedResult: EventResponse{},
		},
		{
			name: "Error - Conflict event already exists",
			body: EventRequest{
				SubjectID:   1,
				Occurrences: 1,
				InsertTS:    "2026-01-01T00:00:00-03:00",
			},
			serviceFake: &services.EventsServiceFake{
				CallbackCreate: func(e repository.Event) (repository.Event, error) {
					return repository.Event{
						ID:          1,
						SubjectID:   1,
						Occurrences: 1,
						InsertTS:    time.Date(2026, time.January, 01, 0, 0, 0, 0, time.Now().Location()),
						LastUpdate:  time.Time{},
					}, services.ErrorEventDateConflict{}
				},
			},
			wantStatus: http.StatusConflict,
			expectedResult: EventResponse{
				ID:          1,
				SubjectID:   1,
				Occurrences: 1,
				InsertTS:    time.Date(2026, time.January, 01, 0, 0, 0, 0, time.Now().Location()),
				LastUpdate:  time.Now(),
			},
		},
		{
			name: "Error - Unknown service error",
			body: EventRequest{
				SubjectID:   1,
				Occurrences: 1,
				InsertTS:    "2026-01-01T00:00:00-03:00",
			},
			serviceFake: &services.EventsServiceFake{
				CallbackCreate: func(e repository.Event) (repository.Event, error) {
					return repository.Event{}, errors.New("unknown error")
				},
			},
			wantStatus: http.StatusInternalServerError,
			expectedResult: EventResponse{},
		},
		{
			name: "Error - Decoding body - invalid JSON",
			body: `{"subject_id": 123, "occurrences": "i should be an int"}`, // invalid JSON
			wantStatus:     http.StatusBadRequest,
		},
		{
			name: "Ok - using body as JSON string",
			body: `{"subject_id": 123, "occurrences": 5, "insert_ts": "2026-03-24T22:16:20-03:00"}`,
			serviceFake: &services.EventsServiceFake{
				CallbackUpdate: func(e repository.Event) (repository.Event, error) {
					return repository.Event{
						ID:          e.ID,
						SubjectID:   e.SubjectID,
						Occurrences: e.Occurrences,
						InsertTS:    parsedTime,
						LastUpdate:  time.Time{},
					}, nil
				},
			},
			wantStatus:  http.StatusCreated,
			expectedResult: EventResponse{
				ID:          0,
				SubjectID:   123,
				Occurrences: 5,
				InsertTS:    parsedTime,
				LastUpdate:  time.Now(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewEventsHandler(tt.serviceFake)

			jsonData,  err := getJSONFromStringOrStruct(tt.body)
			if err != nil {
				t.Fatalf("could not serialize struct: %v", err)
			}

			// Creates a request
			req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			// create the recoder (it captures the handler response)
			w := httptest.NewRecorder()

			// Calls my handler
			handler.CreateEvent(w, req)

			// Analyse the result
			result := w.Result()
			assert.Equal(t, tt.wantStatus, result.StatusCode)

			var res EventResponse
			json.NewDecoder(w.Body).Decode(&res)
			assert.Equal(t, tt.expectedResult.ID, res.ID)
			assert.Equal(t, tt.expectedResult.SubjectID, res.SubjectID)
			assert.Equal(t, tt.expectedResult.Occurrences, res.Occurrences)
			assert.Equal(t, tt.expectedResult.InsertTS, res.InsertTS)
		})
	}
}

func getJSONFromStringOrStruct(body interface{}) (jsonData []byte, err error) {
	// Verify body type
	switch v := body.(type) {
	case string:
		// Se já é string, apenas transforma em bytes
		jsonData = []byte(v)
	case []byte:
		jsonData = v
	default:
		// Se for struct, faz o Marshal normalmente
		jsonData, err = json.Marshal(v)
	}
	return jsonData, err
}

func TestEventHandler_GetEvent(t *testing.T) {
	tests := []struct {
		name           string
		eventID        string
		serviceFake    services.EventsServiceBase
		wantStatus     int
		expectedResult EventResponse
	}{
		{
			name:    "Ok - Get Event",
			eventID: "1",
			serviceFake: &services.EventsServiceFake{
				CallbackGet: func(id int64) (repository.Event, error) {
					return repository.Event{
						ID:          id,
						SubjectID:   2,
						Occurrences: 3,
						InsertTS:    time.Date(2026, time.March, 24, 16, 20, 0, 0, time.Now().Location()),
						LastUpdate:  time.Time{},
					}, nil
				},
			},
			wantStatus: http.StatusOK,
			expectedResult: EventResponse{
				ID:          1,
				SubjectID:   2,
				Occurrences: 3,
				InsertTS:    time.Date(2026, time.March, 24, 16, 20, 0, 0, time.Now().Location()),
				LastUpdate:  time.Now(),
			},
		},
		{
			name:    "Error - Event id not found",
			eventID: "404",
			serviceFake: &services.EventsServiceFake{
				CallbackGet: func(id int64) (repository.Event, error) {
					return repository.Event{}, repository.ErrorEventIDNotFound{}
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "Error - error getting query params",
			eventID:    "ERROR",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "Error - Internal server error - unknown error",
			eventID: "1",
			serviceFake: &services.EventsServiceFake{
				CallbackGet: func(id int64) (repository.Event, error) {
					return repository.Event{}, errors.New("unknown error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewEventsHandler(tt.serviceFake)

			// Creates a request
			url := fmt.Sprintf("/events?id=%s", tt.eventID)
			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.Header.Set("Content-Type", "application/json")

			// create the recoder (it captures the handler response)
			w := httptest.NewRecorder()

			// Calls my handler
			handler.GetEvent(w, req)

			// Analyse the result
			result := w.Result()
			assert.Equal(t, tt.wantStatus, result.StatusCode)

			var res EventResponse
			json.NewDecoder(w.Body).Decode(&res)
			assert.Equal(t, tt.expectedResult.ID, res.ID)
			assert.Equal(t, tt.expectedResult.SubjectID, res.SubjectID)
			assert.Equal(t, tt.expectedResult.Occurrences, res.Occurrences)
			assert.Equal(t, tt.expectedResult.InsertTS, res.InsertTS)
		})
	}
}

func TestEventHandler_UpdateEvent(t *testing.T) {
	timeString := "2026-03-24T22:16:20-03:00"
	dateTime := time.Date(2026, time.March, 24, 16, 20, 0, 0, time.Now().Location())

	tests := []struct {
		name           string
		body           interface{} // can be EventRequest or string (for invalid JSON)
		eventID        string
		serviceFake    services.EventsServiceBase
		wantStatus     int
		expectedResult EventResponse
	}{
		{
			name: "Ok - Update Event",
			eventID: "1",
			body: EventRequest{
				SubjectID:   1,
				Occurrences: 11,
				InsertTS:    timeString,
			},
			serviceFake: &services.EventsServiceFake{
				CallbackUpdate: func(e repository.Event) (repository.Event, error) {
					return repository.Event{
						ID:          e.ID,
						SubjectID:   e.SubjectID,
						Occurrences: e.Occurrences,
						InsertTS:    dateTime,
						LastUpdate:  time.Time{},
					}, nil
				},
			},
			wantStatus:  http.StatusOK,
			expectedResult: EventResponse{
				ID:          1,
				SubjectID:   1,
				Occurrences: 11,
				InsertTS:    dateTime,
				LastUpdate:  time.Now(),
			},
		},
		{
			name:       "Error - error getting query params",
			eventID:    "ERROR",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "Error - Event id not found",
			eventID: "404",
			serviceFake: &services.EventsServiceFake{
				CallbackUpdate: func(e repository.Event) (repository.Event, error) {
					return repository.Event{}, repository.ErrorEventIDNotFound{}
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "Error - Internal server error - unknown error",
			eventID: "1",
			serviceFake: &services.EventsServiceFake{
				CallbackUpdate: func(e repository.Event) (repository.Event, error) {
					return repository.Event{}, errors.New("unknown error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "Error - Decoding body - invalid JSON",
			eventID: "1",
			body: `{"subject_id": 123, "occurrences": }`, // invalid JSON
			wantStatus:     http.StatusBadRequest,
		},
		{
			name: "Error - Decoding body - invalid data type",
			eventID: "1",
			body: `{"subject_id": 123, "occurrences": "i should be an int"}`, // invalid JSON
			wantStatus:     http.StatusBadRequest,
		},
		{
			name: "Ok - using body as JSON string",
			eventID: "123",
			body: `{"subject_id": 123, "occurrences": 5, "insert_ts": "2026-03-24T22:16:20-03:00"}`,
			serviceFake: &services.EventsServiceFake{
				CallbackUpdate: func(e repository.Event) (repository.Event, error) {
					return repository.Event{
						ID:          e.ID,
						SubjectID:   e.SubjectID,
						Occurrences: e.Occurrences,
						InsertTS:    dateTime,
						LastUpdate:  time.Time{},
					}, nil
				},
			},
			wantStatus:  http.StatusOK,
			expectedResult: EventResponse{
				ID:          123,
				SubjectID:   123,
				Occurrences: 5,
				InsertTS:    dateTime,
				LastUpdate:  time.Now(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewEventsHandler(tt.serviceFake)

			jsonData,  err := getJSONFromStringOrStruct(tt.body)
			if err != nil {
				t.Fatalf("could not serialize struct: %v", err)
			}

			// Creates a request
			url := fmt.Sprintf("/events?id=%s", tt.eventID)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			// create the recoder (it captures the handler response)
			w := httptest.NewRecorder()

			// Calls my handler
			handler.UpdateEvent(w, req)

			// Analyse the result
			result := w.Result()
			assert.Equal(t, tt.wantStatus, result.StatusCode)

			var res EventResponse
			json.NewDecoder(w.Body).Decode(&res)
			assert.Equal(t, tt.expectedResult.ID, res.ID)
			assert.Equal(t, tt.expectedResult.SubjectID, res.SubjectID)
			assert.Equal(t, tt.expectedResult.Occurrences, res.Occurrences)
			assert.Equal(t, tt.expectedResult.InsertTS, res.InsertTS)
		})
	}
}
