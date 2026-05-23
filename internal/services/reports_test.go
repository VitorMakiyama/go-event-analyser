package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReportsServices_GetReportTypes(t *testing.T) {
	service := NewReportsServices()

	types := service.GetReportTypes()

	for i, rT := range reportTypes {
		assert.Equal(t, rT, types[i])
	}
}