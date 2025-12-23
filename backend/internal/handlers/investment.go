package handlers

import (
	"context"
	"net/http"
	"time"

	"investmate-backend/internal/models"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"google.golang.org/api/iterator"
)

type InvestmentHandler struct {
	firestoreClient *firestore.Client
	validator       *validator.Validate
}

func NewInvestmentHandler(firestoreClient *firestore.Client) *InvestmentHandler {
	return &InvestmentHandler{
		firestoreClient: firestoreClient,
		validator:       validator.New(),
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
