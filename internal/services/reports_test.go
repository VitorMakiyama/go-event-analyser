package services

import (
	"errors"
	"go-event-analyser/internal/repository"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReportsService_GetReportTypes(t *testing.T) {
	repository := repository.NewRepositoryFake()
	service := NewReportsService(repository)

	types := service.GetReportTypes()

	for i, rT := range reportTypes {
		assert.Equal(t, rT, types[i])
	}
}

func TestReportsService_GetReport(t *testing.T) {
	tests := []struct {
		name           string
		reportType     string
		subjectID      int64
		wantError      error
		expectedResult ReportData
	}{
		{
			name:       "OK - Basic report 1 event",
			reportType: reportTypes[0], // BASIC
			subjectID:  1,
			expectedResult: ReportData{
				Type: "BASIC",
				Details: BasicReport{
					Weekly:           "1.00",
					Monthly:          "1.00",
					Sigma:            "0.00",
					StartDate:        "2026-01-01",
					TotalOccurrences: "1",
				},
			},
		},
		{
			name:       "OK - Basic report multiple events",
			reportType: reportTypes[0], // BASIC
			subjectID:  2,
			expectedResult: ReportData{
				Type: "BASIC",
				Details: BasicReport{
					Weekly:           "2.00",
					Monthly:          "3.00",
					Sigma:            "1.00",
					StartDate:        "2026-01-01",
					TotalOccurrences: "6",
				},
			},
		},
		{
			name:       "OK - Basic report with no events",
			reportType: reportTypes[0], // BASIC
			subjectID:  0,
			expectedResult: ReportData{
				Type: "BASIC",
				Details: BasicReport{
					Weekly:           "0.00",
					Monthly:          "0.00",
					Sigma:            "0.00",
					StartDate:        time.Now().Format(time.DateOnly),
					TotalOccurrences: "0",
				},
			},
		},
		{
			name:           "Error - subject_id not found",
			reportType:     reportTypes[0], // BASIC
			subjectID:      -1,
			wantError:      errors.New(""),
			expectedResult: ReportData{},
		},
		{
			name:           "Error - report not found",
			reportType:     "TEST", // BASIC
			subjectID:      1,
			wantError:      ErrorReportTypeNotFound{},
			expectedResult: ReportData{},
		},
	}

	repository := repository.NewRepositoryFake()
	service := NewReportsService(repository)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calls my service
			result, err := service.GetReport(tt.reportType, tt.subjectID)

			// Analyse the result
			if tt.wantError != nil {
				assert.ErrorAs(t, err, &tt.wantError)
			} else {
				assert.NoError(t, err)
			}

			// Analyse the result
			assert.Equal(t, tt.expectedResult.Type, result.Type)
			assert.Equal(t, tt.expectedResult.Details, result.Details)
		})
	}
}
