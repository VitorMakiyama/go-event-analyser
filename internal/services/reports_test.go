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
					Yearly:           "1.00",
					Sigma:            "0.00",
					StartDate:        "2026-01-01",
					EndDate:          "2026-01-01",
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
					Weekly:           "1.22",
					Monthly:          "3.67",
					Yearly:           "5.50",
					Sigma:            "1.71",
					StartDate:        "2025-12-01",
					EndDate:          "2026-02-01",
					TotalOccurrences: "11",
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
					Yearly:           "0.00",
					Sigma:            "0.00",
					StartDate:        time.Now().Format(time.DateOnly),
					EndDate:          time.Now().Format(time.DateOnly),
					TotalOccurrences: "0",
				},
			},
		},
		{
			name:       "OK - Chart Daily report with 1 event",
			reportType: reportTypes[1], // CHART_DAILY
			subjectID:  1,
			wantError:  nil,
			expectedResult: ReportData{
				Type: "CHART_DAILY",
				Details: ChartReport{
					Data: []int{
						1,
					},
					XLabels: []string{
						"2026-01-01",
					},
				},
			},
		},
		{
			name:       "OK - Chart Daily report with multiple events",
			reportType: reportTypes[1], // CHART_DAILY
			subjectID:  2,
			wantError:  nil,
			expectedResult: ReportData{
				Type: "CHART_DAILY",
				Details: ChartReport{
					Data: []int{
						5, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
						1, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
						3,
					},
					XLabels: []string{
						"2025-12-01", "2025-12-02", "2025-12-03", "2025-12-04", "2025-12-05", "2025-12-06", "2025-12-07", "2025-12-08", "2025-12-09", "2025-12-10", "2025-12-11", "2025-12-12", "2025-12-13", "2025-12-14", "2025-12-15", "2025-12-16", "2025-12-17", "2025-12-18", "2025-12-19", "2025-12-20", "2025-12-21", "2025-12-22", "2025-12-23", "2025-12-24", "2025-12-25", "2025-12-26", "2025-12-27", "2025-12-28", "2025-12-29", "2025-12-30", "2025-12-31",
						"2026-01-01", "2026-01-02", "2026-01-03", "2026-01-04", "2026-01-05", "2026-01-06", "2026-01-07", "2026-01-08", "2026-01-09", "2026-01-10", "2026-01-11", "2026-01-12", "2026-01-13", "2026-01-14", "2026-01-15", "2026-01-16", "2026-01-17", "2026-01-18", "2026-01-19", "2026-01-20", "2026-01-21", "2026-01-22", "2026-01-23", "2026-01-24", "2026-01-25", "2026-01-26", "2026-01-27", "2026-01-28", "2026-01-29", "2026-01-30", "2026-01-31",
						"2026-02-01",
					},
				},
			},
		},
		{
			name:       "OK - Chart Weekly report with 1 event",
			reportType: reportTypes[2], // CHART_WEEKLY
			subjectID:  1,
			wantError:  nil,
			expectedResult: ReportData{
				Type: "CHART_WEEKLY",
				Details: ChartReport{
					Data: []int{
						1,
					},
					XLabels: []string{
						"2026-1",
					},
				},
			},
		},
		{
			name:       "OK - Chart Weekly report with multiple events",
			reportType: reportTypes[2], // CHART_WEEKLY
			subjectID:  2,
			wantError:  nil,
			expectedResult: ReportData{
				Type: "CHART_WEEKLY",
				Details: ChartReport{
					Data: []int{
						5, 0, 0, 0,
						1, 2, 0, 0, 3,
					},
					XLabels: []string{
						"2025-49", "2025-50", "2025-51", "2025-52",
						"2026-1", "2026-2", "2026-3", "2026-4", "2026-5",
					},
				},
			},
		},
		{
			name:       "OK - Chart Monthly report with 1 event",
			reportType: reportTypes[3], // CHART_MONTHLY
			subjectID:  1,
			wantError:  nil,
			expectedResult: ReportData{
				Type: "CHART_MONTHLY",
				Details: ChartReport{
					Data: []int{
						1,
					},
					XLabels: []string{
						"2026-Jan",
					},
				},
			},
		},
		{
			name:       "OK - Chart Monthly report with multiple events",
			reportType: reportTypes[3], // CHART_MONTHLY
			subjectID:  2,
			wantError:  nil,
			expectedResult: ReportData{
				Type: "CHART_MONTHLY",
				Details: ChartReport{
					Data: []int{
						5,
						3, 3,
					},
					XLabels: []string{
						"2025-Dez",
						"2026-Jan", "2026-Fev",
					},
				},
			},
		},
		{
			name:       "OK - Chart Yearly report with 1 event",
			reportType: reportTypes[4], // CHART_YEARLY
			subjectID:  1,
			wantError:  nil,
			expectedResult: ReportData{
				Type: "CHART_YEARLY",
				Details: ChartReport{
					Data: []int{
						1,
					},
					XLabels: []string{
						"2026",
					},
				},
			},
		},
		{
			name:       "OK - Chart Yearly report with multiple events",
			reportType: reportTypes[4], // CHART_YEARLY
			subjectID:  2,
			wantError:  nil,
			expectedResult: ReportData{
				Type: "CHART_YEARLY",
				Details: ChartReport{
					Data: []int{
						5,
						6,
					},
					XLabels: []string{
						"2025",
						"2026",
					},
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
			name:       "Error - report not found",
			reportType: "TEST", // BASIC
			subjectID:  1,
			wantError:  ErrorReportTypeNotFound{},
			expectedResult: ReportData{
				Type:    "",
				Details: "",
			},
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
