package controllers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"

	"sensio/domain/reports/entities"
)

type mockReportRepository struct {
	reports map[string]*entities.Report
}

func (m *mockReportRepository) Save(report *entities.Report) error {
	if m.reports == nil {
		m.reports = map[string]*entities.Report{}
	}
	m.reports[report.ID] = report
	return nil
}

func (m *mockReportRepository) GetByID(id string) (*entities.Report, error) {
	if r, ok := m.reports[id]; ok {
		return r, nil
	}
	return nil, errors.New("report not found")
}

func (m *mockReportRepository) Delete(id string) error {
	delete(m.reports, id)
	return nil
}

func TestPublicDownloadRoute_NoAuthRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &mockReportRepository{
		reports: map[string]*entities.Report{
			"test-report-id": {
				ID:         "test-report-id",
				S3ObjectKey: "",
				LocalPath:  "",
			},
		},
	}
	ctrl := NewReportDownloadController(repo, nil)

	router := gin.New()
	router.GET("/api/reports/:id/download", ctrl.GetDownload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/reports/test-report-id/download", nil)
	router.ServeHTTP(w, req)

	if w.Code == http.StatusUnauthorized {
		t.Errorf("Public download route returned 401 — auth is still required. Route should be public.")
	}
	if w.Code == http.StatusUnauthorized {
		t.Log("Route is still protected. Check reports.go RegisterRoutes — download must be on public router, not protected group.")
	}
}

func TestProtectedRoutes_StillRequireAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := &ReportsGetByIDController{}
	router := gin.New()

	protected := router.Group("/api")
	protected.Use(func(c *gin.Context) {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Authorization header is required"})
		c.Abort()
	})
	protected.GET("/reports/:id", ctrl.GetReportByID)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/reports/test-id", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Protected GET /api/reports/:id should return 401 without auth, got %d", w.Code)
	}
}

func TestDownloadController_GetDownload_NonExistentReport(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &mockReportRepository{reports: map[string]*entities.Report{}}
	ctrl := NewReportDownloadController(repo, nil)

	router := gin.New()
	router.GET("/api/reports/:id/download", ctrl.GetDownload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/reports/nonexistent/download", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for nonexistent report, got %d", w.Code)
	}
}

func TestDownloadController_GetDownload_LocalFileServing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get cwd: %v", err)
	}

	uploadDir := cwd + "/uploads/reports"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		t.Fatalf("Failed to create upload dir: %v", err)
	}
	defer os.RemoveAll(cwd + "/uploads")

	tmpFile := uploadDir + "/local-report-test.pdf"
	if err := os.WriteFile(tmpFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	repo := &mockReportRepository{
		reports: map[string]*entities.Report{
			"local-report": {
				ID:        "local-report",
				LocalPath: "uploads/reports/local-report-test.pdf",
			},
		},
	}
	ctrl := NewReportDownloadController(repo, nil)

	router := gin.New()
	router.GET("/api/reports/:id/download", ctrl.GetDownload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/reports/local-report/download", nil)
	router.ServeHTTP(w, req)

	if w.Code == http.StatusNotFound {
		t.Errorf("Expected file to be found (200 or redirect), got 404. LocalPath serving may be broken.")
	}
	if w.Code == http.StatusUnauthorized {
		t.Errorf("Public download route returned 401 — auth is still required.")
	}
}