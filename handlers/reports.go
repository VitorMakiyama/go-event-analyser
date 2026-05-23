package handlers

import (
	"encoding/json"
	"go-event-analyser/internal/services"
	"log"
	"net/http"
)

type ReportsHandler struct {
	service services.ReportsServices
}

func NewReportsHandler(service services.ReportsServices) ReportsHandler{
	return ReportsHandler{}
}

func (rh *ReportsHandler) GetReportTypes(w http.ResponseWriter, r *http.Request) {
	log.Println("ReportsHandler - Getting all report types available...")
	types := rh.service.GetReportTypes()

	json.NewEncoder(w).Encode(types)
}
