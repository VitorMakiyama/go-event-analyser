package handlers

import (
	"encoding/json"
	"go-event-analyser/internal/services"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReportsHandler_GetReportTypes(t *testing.T) {
	service := services.NewReportsServices()
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
	"CHART",
	}

	var res []string
	json.NewDecoder(w.Body).Decode(&res)
	for i, expectedType := range reportTypes {
		assert.Equal(t, expectedType, res[i])
	}
}
