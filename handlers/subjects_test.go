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

	"github.com/stretchr/testify/assert"
)

func TestSubjectsHandler_CreateSubject(t *testing.T) {
	tests := []struct {
		name           string
		body           interface{}
		serviceFake    services.SubjectsServiceBase
		wantStatus     int
		expectedResult SubjectResponse
	}{
		{
			name: "Ok - Create new Subject",
			body: CreateSubjectRequest{
				Name:        "TesteOK",
				Description: "Teste bem sucedido",
			},
			serviceFake: &services.SubjectsServiceFake{},
			wantStatus:  http.StatusCreated,
			expectedResult: SubjectResponse{
				ID:          0,
				Name:        "TesteOK",
				Description: "Teste bem sucedido",
			},
		},
		{
			name: "Error - Unknown service error - error inserting subject",
			body: CreateSubjectRequest{
				Name:        "TesteOK",
				Description: "Teste bem sucedido",
			},
			serviceFake: &services.SubjectsServiceFake{
				CallbackCreate: func(s repository.Subject) (repository.Subject, error) {
					return repository.Subject{}, errors.New("unknown error")
				},
			},
			wantStatus:     http.StatusInternalServerError,
			expectedResult: SubjectResponse{},
		},
		{
			name:       "Error - Decoding body - invalid JSON",
			body:       `{"name": 123, "description": false}`, // invalid JSON
			wantStatus: http.StatusBadRequest,
		},
		{
			name:        "Ok - using body as JSON string",
			body:        `{"name": "TesteJSONString", "description": "Teste"}`,
			serviceFake: &services.SubjectsServiceFake{},
			wantStatus:  http.StatusCreated,
			expectedResult: SubjectResponse{
				ID:          0,
				Name:        "TesteJSONString",
				Description: "Teste",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewSubjectsHandler(tt.serviceFake)

			jsonData, err := getJSONFromStringOrStruct(tt.body)
			if err != nil {
				t.Fatalf("could not serialize struct: %v", err)
			}

			// Creates a request
			req := httptest.NewRequest(http.MethodPost, "/subjects", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			// Create the recoder (it captures the handler response)
			w := httptest.NewRecorder()

			// Calls my handler
			handler.CreateSubject(w, req)

			// Analyse the result
			result := w.Result()
			assert.Equal(t, tt.wantStatus, result.StatusCode)

			var res SubjectResponse
			json.NewDecoder(w.Body).Decode(&res)
			assert.Equal(t, tt.expectedResult.ID, res.ID)
			assert.Equal(t, tt.expectedResult.Name, res.Name)
			assert.Equal(t, tt.expectedResult.Description, res.Description)
		})
	}
}

func TestSubjectsHandler_GetSubject(t *testing.T) {
	tests := []struct {
		name           string
		subjectID      string
		serviceFake    services.SubjectsServiceBase
		wantStatus     int
		expectedResult SubjectResponse
	}{
		{
			name:      "Ok - Get 1 Subject",
			subjectID: "1",
			serviceFake: &services.SubjectsServiceFake{
				CallbackGet: func(id int64) (repository.Subject, error) {
					return repository.Subject{
						ID:          id,
						Name:        "Test",
						Description: "Test",
					}, nil
				},
			},
			wantStatus: http.StatusOK,
			expectedResult: SubjectResponse{
				ID:          1,
				Name:        "Test",
				Description: "Test",
			},
		},
		{
			name:       "Error - error getting query params",
			subjectID:  "ERROR",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "Error - Subject id not found",
			subjectID: "404",
			serviceFake: &services.SubjectsServiceFake{
				CallbackGet: func(id int64) (repository.Subject, error) {
					return repository.Subject{}, repository.ErrorSubjectIDNotFound{}
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:      "Error - Internal server error - unknown error",
			subjectID: "1",
			serviceFake: &services.SubjectsServiceFake{
				CallbackGet: func(id int64) (repository.Subject, error) {
					return repository.Subject{}, errors.New("unknown error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewSubjectsHandler(tt.serviceFake)

			// Creates a request
			url := fmt.Sprintf("/subjects?id=%s", tt.subjectID)
			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.Header.Set("Content-Type", "application/json")

			// Create the recoder (it captures the handler response)
			w := httptest.NewRecorder()

			// Calls my handler
			handler.GetSubject(w, req)

			// Analyse the result
			result := w.Result()
			assert.Equal(t, tt.wantStatus, result.StatusCode)

			var res SubjectResponse
			json.NewDecoder(w.Body).Decode(&res)
			assert.Equal(t, tt.expectedResult.ID, res.ID)
			assert.Equal(t, tt.expectedResult.Name, res.Name)
			assert.Equal(t, tt.expectedResult.Description, res.Description)
		})
	}
}

func TestSubjectsHandler_GetSubject_ok_all_subjects(t *testing.T) {
	var serviceFake services.SubjectsServiceBase = &services.SubjectsServiceFake{
		CallbackGetAll: func() []repository.Subject {
			return []repository.Subject{
				{
					ID:          1,
					Name:        "Test 1",
					Description: "Test 1",
				},
				{
					ID:          2,
					Name:        "Test 2",
					Description: "Test 2",
				},
			}
		},
	}
	wantStatus := 200
	expectedResult := []SubjectResponse{
		{
			ID:          1,
			Name:        "Test 1",
			Description: "Test 1",
		},
		{
			ID:          2,
			Name:        "Test 2",
			Description: "Test 2",
		},
	}

	// Instantiate SubjectsHandler
	handler := NewSubjectsHandler(serviceFake)

	// Creates a request
	url := "/subjects"
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Content-Type", "application/json")

	// Create the recoder (it captures the handler response)
	w := httptest.NewRecorder()

	// Calls my handler
	handler.GetSubject(w, req)

	// Analyse the result
	result := w.Result()
	assert.Equal(t, wantStatus, result.StatusCode)

	var res []SubjectResponse
	json.NewDecoder(w.Body).Decode(&res)
	assert.Equal(t, len(expectedResult), len(res))
	for i, e := range expectedResult {
		assert.Equal(t, e.ID, res[i].ID)
		assert.Equal(t, e.Name, res[i].Name)
		assert.Equal(t, e.Description, res[i].Description)
	}
}
