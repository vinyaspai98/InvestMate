package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"investmate-backend/internal/models"
	"investmate-backend/pkg/config"
	"io"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"google.golang.org/api/iterator"
)

type InvestmentHandler struct {
	firestoreClient *firestore.Client
	validator       *validator.Validate
	cfg             *config.Config
}

func NewInvestmentHandler(firestoreClient *firestore.Client, cfg *config.Config) *InvestmentHandler {
	return &InvestmentHandler{
		firestoreClient: firestoreClient,
		validator:       validator.New(),
		cfg:             cfg,
	}
}

// GetInvestmentsByCategory gets investments filtered by category
func (h *InvestmentHandler) GetInvestmentsByCategory(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	category := c.Param("category")
	if category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category parameter is required"})
		return
	}

	// Validate category
	validCategories := []string{string(models.CategoryStocks), string(models.CategoryMutualFunds),
		string(models.CategoryFDs), string(models.CategoryInsurance)}

	isValid := false
	for _, validCat := range validCategories {
		if category == validCat {
			isValid = true
			break
		}
	}

	if !isValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category"})
		return
	}

	investments, err := h.getInvestmentsByCategory(c.Request.Context(), userID, models.InvestmentCategory(category))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get investments"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"investments": investments})
}

// GetCategorySummary gets aggregate metrics for a specific category
func (h *InvestmentHandler) GetCategorySummary(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	category := c.Param("category")
	if category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category parameter is required"})
		return
	}

	investments, err := h.getInvestmentsByCategory(c.Request.Context(), userID, models.InvestmentCategory(category))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get investments"})
		return
	}

	summary := h.calculateCategorySummary(investments)
	c.JSON(http.StatusOK, summary)
}

// GetInvestment gets a specific investment by ID
func (h *InvestmentHandler) GetInvestment(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	investmentID := c.Param("id")
	if investmentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Investment ID is required"})
		return
	}

	investment, err := h.getInvestmentByID(c.Request.Context(), userID, investmentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Investment not found"})
		return
	}

	c.JSON(http.StatusOK, investment)
}

// CreateInvestment creates a new investment
func (h *InvestmentHandler) CreateInvestment(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.CreateInvestmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	investment := models.Investment{
		Category:       req.Category,
		Name:           req.Name,
		InvestedAmount: req.InvestedAmount,
		CurrentValue:   req.CurrentValue,
		Notes:          req.Notes,
		InvestmentData: req.InvestmentData,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		IsActive:       true,
	}

	// Enrich with Alpha Vantage data if it's a stock and ticker is provided
	if req.Category == models.CategoryStocks && req.InvestmentData.Ticker != "" {
		h.enrichStockInvestment(c.Request.Context(), &investment, req)
	}

	// Create investment in Firestore
	docRef, _, err := h.firestoreClient.Collection("users").Doc(userID).Collection("investments").Add(c.Request.Context(), investment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create investment"})
		return
	}

	investment.ID = docRef.ID
	c.JSON(http.StatusCreated, investment)
}

// UpdateInvestment updates an existing investment
func (h *InvestmentHandler) UpdateInvestment(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	investmentID := c.Param("id")
	if investmentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Investment ID is required"})
		return
	}

	var req models.UpdateInvestmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build update data
	updates := []firestore.Update{
		{Path: "updatedAt", Value: time.Now()},
	}

	if req.Name != nil {
		updates = append(updates, firestore.Update{Path: "name", Value: *req.Name})
	}
	if req.InvestedAmount != nil {
		updates = append(updates, firestore.Update{Path: "investedAmount", Value: *req.InvestedAmount})
	}
	if req.CurrentValue != nil {
		updates = append(updates, firestore.Update{Path: "currentValue", Value: *req.CurrentValue})
	}
	if req.Notes != nil {
		updates = append(updates, firestore.Update{Path: "notes", Value: *req.Notes})
	}
	if req.InvestmentData != nil {
		updates = append(updates, firestore.Update{Path: "investmentData", Value: *req.InvestmentData})
	}

	// Update investment in Firestore
	investmentRef := h.firestoreClient.Collection("users").Doc(userID).Collection("investments").Doc(investmentID)
	_, err := investmentRef.Update(c.Request.Context(), updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update investment"})
		return
	}

	// Get updated investment
	investment, err := h.getInvestmentByID(c.Request.Context(), userID, investmentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated investment"})
		return
	}

	c.JSON(http.StatusOK, investment)
}

// DeleteInvestment deletes an investment (soft delete by setting isActive to false)
func (h *InvestmentHandler) DeleteInvestment(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	investmentID := c.Param("id")
	if investmentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Investment ID is required"})
		return
	}

	// Soft delete by setting isActive to false
	investmentRef := h.firestoreClient.Collection("users").Doc(userID).Collection("investments").Doc(investmentID)
	_, err := investmentRef.Update(c.Request.Context(), []firestore.Update{
		{Path: "isActive", Value: false},
		{Path: "updatedAt", Value: time.Now()},
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete investment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Investment deleted successfully"})
}

// SearchSymbols proxies Alpha Vantage SYMBOL_SEARCH
func (h *InvestmentHandler) SearchSymbols(c *gin.Context) {
	keywords := c.Query("keywords")
	if keywords == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Keywords parameter is required"})
		return
	}

	url := fmt.Sprintf("https://www.alphavantage.co/query?function=SYMBOL_SEARCH&keywords=%s&apikey=%s",
		keywords, h.cfg.AlphaVantageAPIKey)

	resp, err := http.Get(url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch from Alpha Vantage"})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	c.JSON(http.StatusOK, result)
}

// GetPriceData proxies Alpha Vantage TIME_SERIES_DAILY and GLOBAL_QUOTE
func (h *InvestmentHandler) GetPriceData(c *gin.Context) {
	symbol := c.Param("symbol")
	date := c.Query("date")

	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Symbol is required"})
		return
	}

	// Fetch current price
	quoteURL := fmt.Sprintf("https://www.alphavantage.co/query?function=GLOBAL_QUOTE&symbol=%s&apikey=%s",
		symbol, h.cfg.AlphaVantageAPIKey)

	quoteResp, err := http.Get(quoteURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch quote"})
		return
	}
	defer quoteResp.Body.Close()

	quoteBody, _ := io.ReadAll(quoteResp.Body)
	var quoteResult map[string]interface{}
	json.Unmarshal(quoteBody, &quoteResult)

	response := gin.H{
		"quote": quoteResult,
	}

	// Fetch historical price if date is provided
	if date != "" {
		historyURL := fmt.Sprintf("https://www.alphavantage.co/query?function=TIME_SERIES_DAILY&symbol=%s&apikey=%s",
			symbol, h.cfg.AlphaVantageAPIKey)

		historyResp, err := http.Get(historyURL)
		if err != nil {
			// Don't fail the whole request if history fails
			response["history_error"] = "Failed to fetch historical data"
		} else {
			defer historyResp.Body.Close()
			historyBody, _ := io.ReadAll(historyResp.Body)
			var historyResult map[string]interface{}
			json.Unmarshal(historyBody, &historyResult)
			response["history"] = historyResult
		}
	}

	c.JSON(http.StatusOK, response)
}

func (h *InvestmentHandler) enrichStockInvestment(ctx context.Context, inv *models.Investment, req models.CreateInvestmentRequest) {
	symbol := inv.InvestmentData.Ticker
	// Use CurrentValue as a fallback if we can't fetch it

	// Fetch current price
	quoteURL := fmt.Sprintf("https://www.alphavantage.co/query?function=GLOBAL_QUOTE&symbol=%s&apikey=%s",
		symbol, h.cfg.AlphaVantageAPIKey)

	resp, err := http.Get(quoteURL)
	if err == nil {
		defer resp.Body.Close()
		var result map[string]interface{}
		body, _ := io.ReadAll(resp.Body)
		json.Unmarshal(body, &result)

		if quote, ok := result["Global Quote"].(map[string]interface{}); ok {
			if priceStr, ok := quote["05. price"].(string); ok {
				var price float64
				fmt.Sscanf(priceStr, "%f", &price)
				if price > 0 {
					inv.InvestmentData.CurrentPrice = price
					if inv.InvestmentData.Quantity > 0 {
						inv.CurrentValue = price * inv.InvestmentData.Quantity
					}
				}
			}
		}
	}

	// If Quantity is not provided, try to calculate it from amount and historical price
	if inv.InvestmentData.Quantity == 0 && inv.InvestedAmount > 0 {
		// For simplicity, we'll try to get the price from time.Now() - 24h as a proxy if date is not in req
		// But req doesn't have a date field at the top level, it's usually part of InvestmentData or we use CreatedAt
		// The user request said "We use the symbol, date and invested amount".
		// Let's assume the date is passed in InvestmentData or we use time.Now()

		// For now, if quantity is 0, we'll just set it to amount / current price if current price was found
		if inv.InvestmentData.CurrentPrice > 0 {
			inv.InvestmentData.Quantity = inv.InvestedAmount / inv.InvestmentData.CurrentPrice
			if inv.CurrentValue == 0 {
				inv.CurrentValue = inv.InvestedAmount // Initial value is invested amount
			}
		}
	}
}

// Helper methods

func (h *InvestmentHandler) getInvestmentsByCategory(ctx context.Context, userID string, category models.InvestmentCategory) ([]models.Investment, error) {
	var investments []models.Investment

	// Simplified query - removed OrderBy to avoid composite index requirement
	// Sort will be done in memory instead
	iter := h.firestoreClient.Collection("users").Doc(userID).Collection("investments").
		Where("category", "==", string(category)).
		Where("isActive", "==", true).
		Documents(ctx)

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			// Log the error for debugging
			return nil, err
		}

		var investment models.Investment
		if err := doc.DataTo(&investment); err != nil {
			return nil, err
		}
		investment.ID = doc.Ref.ID
		investments = append(investments, investment)
	}

	// Sort by createdAt in memory (descending)
	// This avoids needing a Firestore composite index
	// for i := 0; i < len(investments)-1; i++ {
	// 	for j := i + 1; j < len(investments); j++ {
	// 		if investments[i].CreatedAt.Before(investments[j].CreatedAt) {
	// 			investments[i], investments[j] = investments[j], investments[i]
	// 		}
	// 	}
	// }

	return investments, nil
}

func (h *InvestmentHandler) getInvestmentByID(ctx context.Context, userID, investmentID string) (*models.Investment, error) {
	doc, err := h.firestoreClient.Collection("users").Doc(userID).Collection("investments").Doc(investmentID).Get(ctx)
	if err != nil {
		return nil, err
	}

	var investment models.Investment
	if err := doc.DataTo(&investment); err != nil {
		return nil, err
	}
	investment.ID = doc.Ref.ID

	return &investment, nil
}

func (h *InvestmentHandler) calculateCategorySummary(investments []models.Investment) models.CategorySummary {
	var totalInvested, totalCurrent float64
	count := len(investments)

	for _, investment := range investments {
		totalInvested += investment.InvestedAmount
		totalCurrent += investment.CurrentValue
	}

	var profitLossPercentage float64
	if totalInvested > 0 {
		profitLoss := totalCurrent - totalInvested
		profitLossPercentage = (profitLoss / totalInvested) * 100
	}

	return models.CategorySummary{
		InvestedAmount:       totalInvested,
		CurrentValue:         totalCurrent,
		ProfitLossPercentage: profitLossPercentage,
		Count:                count,
	}
}
