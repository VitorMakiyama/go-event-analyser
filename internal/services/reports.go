package services

var reportTypes = []string{
	"BASIC",
	"CHART",
}

type ReportsServices struct {}

func NewReportsServices() ReportsServices {
	return ReportsServices{}
}

func (rs *ReportsServices) GetReportTypes() []string {
	return reportTypes
}
