package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"unifriend-api/models"
	"unifriend-api/tests/factory"

	"github.com/stretchr/testify/assert"
)

func TestGetMajors(t *testing.T) {
	SetupTestDB()
	defer models.TearDownTestDB()

	for i := 0; i < 5; i++ {
		major := factory.MajorFactory()
		models.DB.Create(&major)
	}
	var majors []models.Major
	models.DB.Find(&majors)

	req, _ := http.NewRequest("GET", "/api/majors", nil)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	for _, major := range majors {
		assert.Contains(t, rec.Body.String(), major.Name)
	}

	models.TearDownTestDB()
}
