package services

import (
	"fmt"
	"go-event-analyser/internal/repository"
	"log"
	"math"
	"slices"
	"strconv"
	"time"

	"gonum.org/v1/gonum/stat"
)

var reportTypes = []string{
	"BASIC",
	"CHART_DAILY",
	"CHART_WEEKLY",
	"CHART_MONTHLY",
	"CHART_YEARLY",
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

type ChartReport struct {
	Data    []int    `json:"data"`
	XLabels []string `json:"x_labels"`
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
		// CHART_DAILY report
		reportData.Details = generateDailyChartReport(e)
	case reportTypes[2]:
		// CHART_WEEKLY report
		reportData.Details = generateWeeklyChartReport(e)
	case reportTypes[3]:
		// CHART_MONTHLY report
		reportData.Details = generateMonthlyChartReport(e)
	case reportTypes[4]:
		// CHART_YEARLY report
		reportData.Details = generateYearlyChartReport(e)
	default:
		// Unknown reportType
		return ReportData{Details: ""}, ErrorReportTypeNotFound{
			ReportType: reportType,
		}
	}

	return reportData, nil
}

func generateBasicReport(events []repository.Event) BasicReport {
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
	//TODO: improve this frequencies calculation, especially for weekly
	startYear, startWeek := startDate.ISOWeek()
	endYear, endWeek := endDate.ISOWeek()
	nYears := endYear - startYear                       // represents the number of full years, if nYears == 0 means the start and end Year are the same
	nWeeks := (nYears * 52) + (endWeek - startWeek + 1) // the real average number of weeks is 52.1429

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

	yearly = totalOccurrences / float64(nYears+1) // when start and end Year are the same, add 1 because the same year is 1 year for this calculation
	if math.IsNaN(yearly) {
		yearly = 0
	}

	return weekly, monthly, yearly
}

// Returns slices ordered by InsertTS ascending
// years: 	slice with one of each year encountered
// months: 	slice of slices where the inner slice contains one of each month of a year, the outer slice represents the year, the same index of 'years' slices works here
// weeks: 	slice of slices where the inner slice contains one of each week of a year, the outer slice represents the year, the same index of 'years' slices works here
func getYearWeekMonthOrderedSlices(events []repository.Event) (years []int, months [][]int, weeks [][]int) {
	for _, e := range events {
		year, week := e.InsertTS.ISOWeek()
		month := int(e.InsertTS.Month())

		if !slices.Contains(years, year) {
			years = append(years, year)
			months = append(months, []int{})
			weeks = append(weeks, []int{})
		}

		yearIndex := slices.Index(years, year)
		if yearIndex != -1 && !slices.Contains(months[yearIndex], month) {
			months[yearIndex] = append(months[yearIndex], month)
		}

		if yearIndex != -1 && !slices.Contains(weeks[yearIndex], week) {
			weeks[yearIndex] = append(weeks[yearIndex], week)
		}
	}

	return
}

func generateDailyChartReport(events []repository.Event) ChartReport {
	var data []int
	var xLabels []string

	var lastDate time.Time

	for _, e := range events {
		d := e.InsertTS.Sub(lastDate)
		if d.Hours() > 24 && !lastDate.IsZero() {
			for !(lastDate.Year() == e.InsertTS.Year() && lastDate.Month() == e.InsertTS.Month() && lastDate.Day() == e.InsertTS.Day()) {
				data = append(data, 0)
				xLabels = append(xLabels, lastDate.Format(time.DateOnly))
				lastDate = lastDate.Add(24 * time.Hour)
			}
		}
		data = append(data, e.Occurrences)
		xLabels = append(xLabels, e.InsertTS.Format(time.DateOnly))
		lastDate = e.InsertTS.Add(24 * time.Hour)
	}

	return ChartReport{
		Data:    data,
		XLabels: xLabels,
	}
}

func generateWeeklyChartReport(events []repository.Event) ChartReport {
	var data []int
	var xLabels []string

	var lastDate time.Time

	for _, e := range events {
		lastYear, lastWeek := lastDate.ISOWeek()
		thisYear, thisWeek := e.InsertTS.ISOWeek()
		xLabel := fmt.Sprintf("%d-%d", thisYear, thisWeek)

		if !lastDate.IsZero() && lastWeek != thisWeek {
			for !(lastYear == thisYear && lastWeek == thisWeek) {
				// While thisYear and thisWeek are different from lastYear and lastWeek...
				lastDate = lastDate.Add(7 * 24 * time.Hour) // Add 7 days

				lastYear, lastWeek = lastDate.ISOWeek()
				zeroLabel := fmt.Sprintf("%d-%d", lastYear, lastWeek)

				data = append(data, 0)
				xLabels = append(xLabels, zeroLabel)

				if lastYear == thisYear && lastWeek == thisWeek {
					// Add this e occurrences
					data[len(data)-1] += e.Occurrences
				}
			}
		} else if !slices.Contains(xLabels, xLabel) {
			// If xLabels does NOT contains xLabel, add it
			data = append(data, e.Occurrences)
			xLabels = append(xLabels, xLabel)
			lastDate = e.InsertTS
		} else {
			// Else, lastWeek == thisWeek and the week already in the labels, so we add the occurrences!
			data[len(data)-1] += e.Occurrences
		}
	}

	return ChartReport{
		Data:    data,
		XLabels: xLabels,
	}
}

func generateMonthlyChartReport(events []repository.Event) ChartReport {
	//TODO: Implement it
	return ChartReport{}
}

func generateYearlyChartReport(events []repository.Event) ChartReport {
	var data []int
	var xLabels []string

	for _, e := range events {
		year := strconv.Itoa(e.InsertTS.Year())
		if slices.Contains(xLabels, year) {
			index := slices.Index(xLabels, year)

			data[index] += e.Occurrences
		} else {
			data = append(data, e.Occurrences)
			xLabels = append(xLabels, year)
		}
	}
	return ChartReport{
		Data:    data,
		XLabels: xLabels,
	}
}

// Custom Errors for this reports service
type ErrorReportTypeNotFound struct {
	ReportType string
}

func (e ErrorReportTypeNotFound) Error() string {
	return fmt.Sprintf("report type '%s' not found", e.ReportType)
}
