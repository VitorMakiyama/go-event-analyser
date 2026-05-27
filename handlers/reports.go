package handlers

import (
	"encoding/json"
	"errors"
	"go-event-analyser/internal/services"
	"log"
	"net/http"
	"strconv"
)

type ReportsHandler struct {
	service services.ReportsServiceBase
}

func NewReportsHandler(service services.ReportsServiceBase) ReportsHandler {
	return ReportsHandler{
		service: service,
	}
}

const logTag = "ReportsHandler - "

// GetReportTypes returns a list of all available report categories supported by the system.
// It returns the report types as a slice of strings, which is useful for populating
// dropdown menus or validating incoming report requests in the android app.
func (rh *ReportsHandler) GetReportTypes(w http.ResponseWriter, r *http.Request) {
	log.Println(logTag, "Getting all report types available...")
	types := rh.service.GetReportTypes()

	json.NewEncoder(w).Encode(types)
}

// GetReport returns the corresponding report received from `type` query param for the
// corresponding `subject_id`.
func (rh *ReportsHandler) GetReport(w http.ResponseWriter, r *http.Request) {
	reportType := r.URL.Query().Get("type")
	id, err := strconv.ParseInt(r.URL.Query().Get("subject_id"), 10, 64)
	if err != nil {
		log.Println(logTag, "error getting query params: ", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("%sGetting report data for %s report and subject_id %d...", logTag, reportType, id)

	report, err := rh.service.GetReport(reportType, id)
	if err != nil {
		if errors.As(err, &services.ErrorReportTypeNotFound{}) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		log.Println(logTag, "error getting report: ", err)
		return
	}

	json.NewEncoder(w).Encode(report)
}
