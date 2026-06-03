package services

import (
	"fmt"
	"go-event-analyser/internal/repository"
	"log"
	"math"
	"strconv"
	"time"

	"gonum.org/v1/gonum/stat"
)

var reportTypes = []string{
	"BASIC",
	"CHART",
}

type ReportData struct {
	Type    string `json:"type"`
	Details any    `json:"details"`
}

type BasicReport struct {
	Weekly           string `json:"weekly"`
	Monthly          string `json:"monthly"`
	Yearly           string `json:"yearly"`
	Sigma            string `json:"sigma"`
	StartDate        string `json:"start_date"`
	EndDate          string `json:"end_date"`
	TotalOccurrences string `json:"total_occurrences"`
}

type ReportsServiceBase interface {
	GetReportTypes() []string
	GetReport(reportType string, subject_id int64) (ReportData, error)
}

type ReportsService struct {
	repository repository.Repository
}

func NewReportsService(repository repository.Repository) ReportsServiceBase {
	return &ReportsService{
		repository: repository,
	}
}

const logTag = "ReportsServices - "

func (rs *ReportsService) GetReportTypes() []string {
	return reportTypes
}

func (rs *ReportsService) GetReport(reportType string, subject_id int64) (ReportData, error) {
	e, err := rs.repository.GetAllEventsFromSubject(subject_id)
	if err != nil {
		log.Printf("%serror getting all events from subject_id %d, err: %v\n", logTag, subject_id, err)
		return ReportData{}, err
	}

	var reportData = ReportData{
		Type: reportType,
	}
	switch reportType {
	case reportTypes[0]:
		// BASIC report
		reportData.Details = generateBasicReport(e)

	case reportTypes[1]:
		// CHART report
		reportData.Details = generateChartReport(e)
	default:
		// Unknown reportType
		return ReportData{Details: ""}, ErrorReportTypeNotFound{
			ReportType: reportType,
		}
	}

	return reportData, nil
}

func generateBasicReport(events []repository.Event) BasicReport {
	yearWeekMap := make(map[int]map[int]int) // map[{Year}]map[{Week}]{Occurrences}
	yearMonthMap := make(map[string]int)     // map[{"YearMonth"}]{{Occurrences}}
	var startDate time.Time = time.Now()
	var endDate time.Time = time.Now()
	if len(events) != 0 {
		endDate = events[len(events)-1].InsertTS
	}
	var totalOccurrences int

	var occurrences []float64 // For calculating sigma (Variance^1/2)

	for _, e := range events {
		if e.InsertTS.Before(startDate) {
			// Gets the first date (oldest) from all events
			startDate = e.InsertTS
		}
		if e.InsertTS.After(endDate) {
			// Gets the lasat date (newest) from all events
			endDate = e.InsertTS
		}
		year, week := e.InsertTS.ISOWeek()
		if week, _ := yearWeekMap[year]; week == nil {
			// Instantiate this year map of weeks, if i was not already instantiate
			yearWeekMap[year] = make(map[int]int)
		}
		yearMonthMap[fmt.Sprintf("%d%d", year, e.InsertTS.Month())] += e.Occurrences

		yearWeekMap[year][week] += e.Occurrences

		occurrences = append(occurrences, float64(e.Occurrences))
		totalOccurrences += e.Occurrences
	}
	weeklyFrequency, monthlyFrequency, yearlyFrequency := calculateFrequencies(startDate, endDate, float64(totalOccurrences))

	var sigma = 0.0
	if len(occurrences) > 1 {
		var variance float64 = stat.Variance(occurrences, nil)
		sigma = math.Sqrt(variance)
	}

	return BasicReport{
		Weekly:           strconv.FormatFloat(weeklyFrequency, 'f', 2, 64),
		Monthly:          strconv.FormatFloat(monthlyFrequency, 'f', 2, 64),
		Yearly:           strconv.FormatFloat(yearlyFrequency, 'f', 2, 64),
		Sigma:            strconv.FormatFloat(sigma, 'f', 2, 64),
		StartDate:        startDate.Format(time.DateOnly),
		EndDate:          endDate.Format(time.DateOnly),
		TotalOccurrences: strconv.Itoa(totalOccurrences),
	}
}

func calculateFrequencies(startDate time.Time, endDate time.Time, totalOccurrences float64) (weekly float64, monthly float64, yearly float64) {
	startYear, startWeek := startDate.ISOWeek()
	endYear, endWeek := endDate.ISOWeek()
	nYears := endYear - startYear // represents the number of full years, if nYears == 0 means the start and end Year are the same
	nWeeks:= (nYears * 52) + (endWeek - startWeek + 1) // the real average number of weeks is 52.1429
	
	nMonths := (nYears * 12) + (int(endDate.Month()) - int(startDate.Month()) + 1)

	weekly = totalOccurrences / float64(nWeeks)
	if math.IsNaN(weekly) {
		// If division by 0
		weekly = 0
	}

	monthly = totalOccurrences / float64(nMonths)
	if math.IsNaN(monthly) {
		// If division by 0
		monthly = 0
	}

	yearly = totalOccurrences / float64(nYears + 1) // when start and end Year are the same, add 1 because the same year is 1 year for this calculation
	if math.IsNaN(yearly) {
		yearly = 0
	}

	return weekly, monthly, yearly
}

func generateChartReport(e []repository.Event) string {
	//TODO: Implement it
	return "unimplemented"
}

// Custom Errors for this reports service
type ErrorReportTypeNotFound struct {
	ReportType string
}

func (e ErrorReportTypeNotFound) Error() string {
	return fmt.Sprintf("report type '%s' not found", e.ReportType)
}
