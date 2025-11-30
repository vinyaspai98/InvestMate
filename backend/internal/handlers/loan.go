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

type LoanHandler struct {
	firestoreClient *firestore.Client
	validator       *validator.Validate
}

func NewLoanHandler(firestoreClient *firestore.Client) *LoanHandler {
	return &LoanHandler{
		firestoreClient: firestoreClient,
		validator:       validator.New(),
	}
}

// GetAllLoans gets all loans for the authenticated user
func (h *LoanHandler) GetAllLoans(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	loans, err := h.getAllUserLoans(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get loans"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"loans": loans})
}

// GetLoan gets a specific loan by ID
func (h *LoanHandler) GetLoan(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	loanID := c.Param("id")
	if loanID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Loan ID is required"})
		return
	}

	loan, err := h.getLoanByID(c.Request.Context(), userID, loanID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Loan not found"})
		return
	}

	c.JSON(http.StatusOK, loan)
}

// CreateLoan creates a new loan
func (h *LoanHandler) CreateLoan(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.CreateLoanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	loan := models.Loan{
		LoanType:           req.LoanType,
		LenderName:         req.LenderName,
		PrincipalAmount:    req.PrincipalAmount,
		OutstandingBalance: req.OutstandingBalance,
		InterestRate:       req.InterestRate,
		EMI:                req.EMI,
		Tenure:             req.Tenure,
		RemainingTenure:    req.RemainingTenure,
		StartDate:          req.StartDate,
		MaturityDate:       req.MaturityDate,
		Notes:              req.Notes,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
		IsActive:           true,
	}

	// Create loan in Firestore
	docRef, _, err := h.firestoreClient.Collection("users").Doc(userID).Collection("loans").Add(c.Request.Context(), loan)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create loan"})
		return
	}

	loan.ID = docRef.ID
	c.JSON(http.StatusCreated, loan)
}

// UpdateLoan updates an existing loan
func (h *LoanHandler) UpdateLoan(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	loanID := c.Param("id")
	if loanID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Loan ID is required"})
		return
	}

	var req models.UpdateLoanRequest
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

	if req.LenderName != nil {
		updates = append(updates, firestore.Update{Path: "lenderName", Value: *req.LenderName})
	}
	if req.OutstandingBalance != nil {
		updates = append(updates, firestore.Update{Path: "outstandingBalance", Value: *req.OutstandingBalance})
	}
	if req.InterestRate != nil {
		updates = append(updates, firestore.Update{Path: "interestRate", Value: *req.InterestRate})
	}
	if req.EMI != nil {
		updates = append(updates, firestore.Update{Path: "emi", Value: *req.EMI})
	}
	if req.RemainingTenure != nil {
		updates = append(updates, firestore.Update{Path: "remainingTenure", Value: *req.RemainingTenure})
	}
	if req.Notes != nil {
		updates = append(updates, firestore.Update{Path: "notes", Value: *req.Notes})
	}

	// Update loan in Firestore
	loanRef := h.firestoreClient.Collection("users").Doc(userID).Collection("loans").Doc(loanID)
	_, err := loanRef.Update(c.Request.Context(), updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update loan"})
		return
	}

	// Get updated loan
	loan, err := h.getLoanByID(c.Request.Context(), userID, loanID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated loan"})
		return
	}

	c.JSON(http.StatusOK, loan)
}

// DeleteLoan deletes a loan (soft delete by setting isActive to false)
func (h *LoanHandler) DeleteLoan(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	loanID := c.Param("id")
	if loanID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Loan ID is required"})
		return
	}

	// Soft delete by setting isActive to false
	loanRef := h.firestoreClient.Collection("users").Doc(userID).Collection("loans").Doc(loanID)
	_, err := loanRef.Update(c.Request.Context(), []firestore.Update{
		{Path: "isActive", Value: false},
		{Path: "updatedAt", Value: time.Now()},
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete loan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Loan deleted successfully"})
}

// GetLoanSummary gets aggregate metrics for all loans
func (h *LoanHandler) GetLoanSummary(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	loans, err := h.getAllUserLoans(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get loans"})
		return
	}

	summary := h.calculateLoanSummary(loans)
	c.JSON(http.StatusOK, summary)
}

// Helper methods

func (h *LoanHandler) getAllUserLoans(ctx context.Context, userID string) ([]models.Loan, error) {
	var loans []models.Loan

	iter := h.firestoreClient.Collection("users").Doc(userID).Collection("loans").
		Where("isActive", "==", true).
		OrderBy("createdAt", firestore.Desc).
		Documents(ctx)

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var loan models.Loan
		if err := doc.DataTo(&loan); err != nil {
			return nil, err
		}
		loan.ID = doc.Ref.ID
		loans = append(loans, loan)
	}

	return loans, nil
}

func (h *LoanHandler) getLoanByID(ctx context.Context, userID, loanID string) (*models.Loan, error) {
	doc, err := h.firestoreClient.Collection("users").Doc(userID).Collection("loans").Doc(loanID).Get(ctx)
	if err != nil {
		return nil, err
	}

	var loan models.Loan
	if err := doc.DataTo(&loan); err != nil {
		return nil, err
	}
	loan.ID = doc.Ref.ID

	return &loan, nil
}

func (h *LoanHandler) calculateLoanSummary(loans []models.Loan) models.LoanSummary {
	var totalPrincipal, totalOutstanding float64
	count := len(loans)

	for _, loan := range loans {
		totalPrincipal += loan.PrincipalAmount
		totalOutstanding += loan.OutstandingBalance
	}

	return models.LoanSummary{
		PrincipalAmount:    totalPrincipal,
		OutstandingBalance: totalOutstanding,
		Count:              count,
	}
}
