package services

import (
	"errors"
	"go-event-analyser/internal/repository"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubjectsService_Create(t *testing.T) {
	tests := []struct {
		name                string
		subjectToBeInserted repository.Subject
		wantError           error
		expectedResult      repository.Subject
	}{
		{
			name: "Ok - Create new Subject",
			subjectToBeInserted: repository.Subject{
				ID:          0, // DB creates it
				Name:        "TestOK",
				Description: "TestOK",
			},
			wantError: nil,
			expectedResult: repository.Subject{
				ID:          3,
				Name:        "TestOK",
				Description: "TestOK",
			},
		},
		{
			name: "Error - Insert repository error",
			subjectToBeInserted: repository.Subject{
				ID:          -1,
				Name:        "",
				Description: "",
			},
			wantError: errors.New(""),
			expectedResult: repository.Subject{
				ID:          -1,
				Name:        "",
				Description: "",
			},
		},
	}

	repository := repository.NewRepositoryFake()
	service := NewSubjectsService(repository)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calls my service
			subject, err := service.Create(tt.subjectToBeInserted)

			// Analyse the result
			if tt.wantError != nil {
				assert.ErrorAs(t, err, &tt.wantError)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult.ID, subject.ID)
			assert.Equal(t, tt.expectedResult.Name, subject.Name)
			assert.Equal(t, tt.expectedResult.Description, subject.Description)
		})
	}
}

func TestSubjectsService_Get(t *testing.T) {
	tests := []struct {
		name           string
		subjectID      int64
		wantError      error
		expectedResult repository.Subject
	}{
		{
			name:      "Ok - Get Subject",
			subjectID: 1,
			wantError: nil,
			expectedResult: repository.Subject{
				ID:          1,
				Name:        "T1",
				Description: "T1-Desc",
			},
		},
		{
			name:           "Error - Subject id not found",
			subjectID:      0,
			wantError:      repository.ErrorSubjectIDNotFound{},
			expectedResult: repository.Subject{},
		},
		{
			name:           "Error - Unknown repository error",
			subjectID:      -1,
			wantError:      errors.New(""),
			expectedResult: repository.Subject{},
		},
	}

	repository := repository.NewRepositoryFake()
	service := NewSubjectsService(repository)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calls my service
			subject, err := service.Get(tt.subjectID)

			// Analyse the result
			if tt.wantError != nil {
				assert.ErrorAs(t, err, &tt.wantError)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult.ID, subject.ID)
			assert.Equal(t, tt.expectedResult.Name, subject.Name)
			assert.Equal(t, tt.expectedResult.Description, subject.Description)
		})
	}
}
