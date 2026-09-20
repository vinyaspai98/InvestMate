package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"investmate-backend/internal/models"

	"github.com/gin-gonic/gin"
)

func setupTestRouter(handler *AIHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Middleware injects mock userID
	router.Use(func(c *gin.Context) {
		c.Set("userID", "test-user-123")
		c.Next()
	})

	ai := router.Group("/api/v1/ai")
	{
		ai.POST("/scan-cas", handler.ScanCAS)
		ai.GET("/health-score", handler.GetHealthScore)
		ai.POST("/recommendations", handler.GetRecommendations)
		ai.POST("/forecast-profit", handler.ForecastProfit)
		ai.GET("/monte-carlo", handler.MonteCarlo)
	}

	return router
}

func TestAIHandler_ScanCAS(t *testing.T) {
	handler := NewAIHandler(nil, "")
	router := setupTestRouter(handler)

	reqPayload := models.ScanCASRequest{
		Text: "INF179K01BE2 HDFC Top 100 Fund Buy 100 800.00 2023-01-15\nINF846K01DP8 Parag Parikh Flexi Cap Buy 500 50.00 2023-05-10",
	}
	body, _ := json.Marshal(reqPayload)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/ai/scan-cas", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}

	var score models.PortfolioHealthScore
	if err := json.Unmarshal(w.Body.Bytes(), &score); err != nil {
		t.Fatalf("failed to unmarshal health score response: %v", err)
	}

	if score.OverallScore <= 0 || score.OverallScore > 100 {
		t.Errorf("expected overall score between 1 and 100, got %d", score.OverallScore)
	}
}

func TestAIHandler_ScanCAS_AutoExtractFromDematNoFileInput(t *testing.T) {
	handler := NewAIHandler(nil, "")
	router := setupTestRouter(handler)

	// Post empty JSON body - no file upload, no PDF password
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/ai/scan-cas", bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}

	var score models.PortfolioHealthScore
	if err := json.Unmarshal(w.Body.Bytes(), &score); err != nil {
		t.Fatalf("failed to unmarshal health score response: %v", err)
	}

	if score.OverallScore <= 0 || score.OverallScore > 100 {
		t.Errorf("expected overall score between 1 and 100, got %d", score.OverallScore)
	}
	if len(score.AuditDetails) == 0 {
		t.Errorf("expected audit details to be populated")
	}
	if score.Source == "" {
		t.Errorf("expected score.Source to be set")
	}
}

func TestAIHandler_GetHealthScore(t *testing.T) {
	handler := NewAIHandler(nil, "")
	router := setupTestRouter(handler)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/ai/health-score", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var score models.PortfolioHealthScore
	if err := json.Unmarshal(w.Body.Bytes(), &score); err != nil {
		t.Fatalf("failed to unmarshal health score: %v", err)
	}

	if score.OverallScore <= 0 {
		t.Errorf("expected positive overall score, got %d", score.OverallScore)
	}
}

func TestAIHandler_Recommendations(t *testing.T) {
	handler := NewAIHandler(nil, "")
	router := setupTestRouter(handler)

	reqPayload := models.RecommendationsRequest{
		TargetRisk: "moderate",
	}
	body, _ := json.Marshal(reqPayload)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/ai/recommendations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var actions []models.RecommendationAction
	if err := json.Unmarshal(w.Body.Bytes(), &actions); err != nil {
		t.Fatalf("failed to unmarshal recommendations: %v", err)
	}

	if len(actions) == 0 {
		t.Errorf("expected recommendations, got 0")
	}
}

func TestAIHandler_ForecastProfit(t *testing.T) {
	handler := NewAIHandler(nil, "")
	router := setupTestRouter(handler)

	reqPayload := models.ForecastProfitRequest{
		InitialBase: 1000000,
		Years:       20,
	}
	body, _ := json.Marshal(reqPayload)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/ai/forecast-profit", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var points []models.ProfitForecastPoint
	if err := json.Unmarshal(w.Body.Bytes(), &points); err != nil {
		t.Fatalf("failed to unmarshal forecast points: %v", err)
	}

	if len(points) != 20 {
		t.Errorf("expected 20 forecast points, got %d", len(points))
	}
}

func TestAIHandler_MonteCarlo(t *testing.T) {
	handler := NewAIHandler(nil, "")
	router := setupTestRouter(handler)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/ai/monte-carlo?simulations=500&years=20", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var mc models.MonteCarloResponse
	if err := json.Unmarshal(w.Body.Bytes(), &mc); err != nil {
		t.Fatalf("failed to unmarshal monte carlo response: %v", err)
	}

	if mc.SurvivalRate <= 0 {
		t.Errorf("expected positive survival rate, got %.2f", mc.SurvivalRate)
	}
	if len(mc.Trajectories) == 0 {
		t.Errorf("expected simulation trajectories")
	}
}
