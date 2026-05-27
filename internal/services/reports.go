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
	Sigma            string `json:"sigma"`
	StartDate        string `json:"start_date"`
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
		return ReportData{}, ErrorReportTypeNotFound{
			ReportType: reportType,
		}
	}

	return reportData, nil
}


func generateBasicReport(events []repository.Event) BasicReport {
	yearWeekMap := make(map[int]map[int]int) // map[{Year}]map[{Week}]{Occurrences}
	yearMonthMap := make(map[string]int)     // map[{"YearMonth"}]{{Occurrences}}
	var startDate time.Time = time.Now()
	var totalOccurrences int
	
	var occurrences []float64 // For calculating sigma (Variance^1/2)
	
	for _, e := range events {
		if e.InsertTS.Before(startDate) {
			// Gets the first date (oldest) from all events
			startDate = e.InsertTS
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
	weeklyFrequency, monthlyFrequency := calculateFrequencies(yearWeekMap, yearMonthMap, float64(totalOccurrences))
	
	var sigma = 0.0
	if len(occurrences) > 1 {
		var variance float64 = stat.Variance(occurrences, nil)
		sigma = math.Sqrt(variance)
	}
	
	return BasicReport{
		Weekly:           strconv.FormatFloat(weeklyFrequency, 'f', 2, 64),
		Monthly:          strconv.FormatFloat(monthlyFrequency, 'f', 2, 64),
		Sigma:            strconv.FormatFloat(sigma, 'f', 2, 64),
		StartDate:        startDate.Format(time.DateOnly),
		TotalOccurrences: strconv.Itoa(totalOccurrences),
	}
}

func calculateFrequencies(yearWeekMap map[int]map[int]int, yearMonthMap map[string]int, totalOccurrences float64) (weekly float64, monthly float64) {
	// nYears := len(yearWeekMap) // If need to calculate yearly frequency
	nWeeks := 0
	for _, year := range yearWeekMap {
		nWeeks += len(year)
	}
	
	nMonths := len(yearMonthMap)
	
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
	
	return weekly, monthly
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
