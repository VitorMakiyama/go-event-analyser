package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-event-analyser/internal/services"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReportsHandler_GetReportTypes(t *testing.T) {
	service := services.ReportsServiceFake{}
	handler := NewReportsHandler(service)

	// Creates a request
	req := httptest.NewRequest(http.MethodGet, "/reports", nil)
	req.Header.Set("Content-Type", "application/json")

	// Create the recoder (it captures the handler response)
	w := httptest.NewRecorder()

	// Calls my handler
	handler.GetReportTypes(w, req)

	// Analyse the result
	result := w.Result()

	assert.Equal(t, http.StatusOK, result.StatusCode)

	var reportTypes = []string{
		"BASIC",
		"CHART_DAILY",
		"CHART_WEEKLY",
		"CHART_MONTHLY",
		"CHART_YEARLY",
	}

	var res []string
	json.NewDecoder(w.Body).Decode(&res)
	for i, expectedType := range reportTypes {
		assert.Equal(t, expectedType, res[i])
	}
}

func TestReportsHandler_GetReport(t *testing.T) {
	tests := []struct {
		name           string
		reportType     string
		subjectID      string
		serviceFake    services.ReportsServiceFake
		wantStatus     int
		expectedResult services.ReportData
	}{
		{
			name:       "Ok - Get report data",
			reportType: "BASIC",
			subjectID:  "1",
			serviceFake: services.ReportsServiceFake{
				CallbackGetReport: func(reportType string, subject_id int64) (services.ReportData, error) {
					return services.ReportData{
						Type: reportType,
						Details: services.BasicReport{
							Weekly:           "1.00",
							Monthly:          "1.00",
							Sigma:            "0.00",
							StartDate:        "2026-05-01",
							TotalOccurrences: "1",
						},
					}, nil
				},
			},
			wantStatus: 200,
			expectedResult: services.ReportData{
				Type: "BASIC",
				Details: services.BasicReport{
					Weekly:           "1.00",
					Monthly:          "1.00",
					Sigma:            "0.00",
					StartDate:        "2026-05-01",
					TotalOccurrences: "1",
				},
			},
		},
		{
			name:       "Error - parsing subject_id",
			reportType: "TEST",
			subjectID:  "invalid",
			serviceFake: services.ReportsServiceFake{
				CallbackGetReport: func(reportType string, subject_id int64) (services.ReportData, error) {
					return services.ReportData{}, services.ErrorReportTypeNotFound{}
				},
			},
			wantStatus: 400,
			expectedResult: services.ReportData{},
		},
		{
			name:       "Error - report type not found",
			reportType: "TEST",
			subjectID:  "1",
			serviceFake: services.ReportsServiceFake{
				CallbackGetReport: func(reportType string, subject_id int64) (services.ReportData, error) {
					return services.ReportData{}, services.ErrorReportTypeNotFound{}
				},
			},
			wantStatus: 404,
			expectedResult: services.ReportData{},
		},
		{
			name:       "Error - unknown error - internal server error",
			reportType: "TEST",
			subjectID:  "1",
			serviceFake: services.ReportsServiceFake{
				CallbackGetReport: func(reportType string, subject_id int64) (services.ReportData, error) {
					return services.ReportData{}, errors.New("TEST")
				},
			},
			wantStatus: 500,
			expectedResult: services.ReportData{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewReportsHandler(tt.serviceFake)

			targetPath := fmt.Sprintf("/reports?type=%s&subject_id=%s", tt.reportType, tt.subjectID)
			// Creates a request
			req := httptest.NewRequest(http.MethodGet, targetPath, nil)
			req.Header.Set("Content-Type", "application/json")

			// Create the recoder (it captures the handler response)
			w := httptest.NewRecorder()

			// Calls my handler
			handler.GetReport(w, req)

			// Analyse the result
			result := w.Result()
			assert.Equal(t, tt.wantStatus, result.StatusCode)

			var res services.ReportData
			json.NewDecoder(w.Body).Decode(&res)

			// Convert the 'any' map back to the expected concrete struct
			switch tt.reportType {
			case "BASIC":
				var actualDetails services.BasicReport
				mapBytes, _ := json.Marshal(res.Details) // res.Details is a map[string]any
				json.Unmarshal(mapBytes, &actualDetails)
				res.Details = actualDetails
			case "CHART":
				// var actualDetails services.ChartReport
				// mapBytes, _ := json.Marshal(res.Details) // res.Details is a map[string]any
				// json.Unmarshal(mapBytes, &actualDetails)
				// res.Details = actualDetails
			}

			assert.Equal(t, tt.expectedResult.Type, res.Type)
			assert.Equal(t, tt.expectedResult.Details, res.Details)
		})

	}
}
