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

type TransactionHandler struct {
	firestoreClient *firestore.Client
	validator       *validator.Validate
}

func NewTransactionHandler(firestoreClient *firestore.Client) *TransactionHandler {
	return &TransactionHandler{
		firestoreClient: firestoreClient,
		validator:       validator.New(),
	}
}

// GetTransactionsByCategory gets transactions filtered by category
func (h *TransactionHandler) GetTransactionsByCategory(c *gin.Context) {
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

	transactions, err := h.getTransactionsByCategory(c.Request.Context(), userID, models.InvestmentCategory(category))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get transactions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"transactions": transactions})
}

// GetTransactionsByRelatedID gets transactions for a specific investment or loan
func (h *TransactionHandler) GetTransactionsByRelatedID(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	relatedID := c.Param("relatedId")
	if relatedID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Related ID parameter is required"})
		return
	}

	transactions, err := h.getTransactionsByRelatedID(c.Request.Context(), userID, relatedID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get transactions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"transactions": transactions})
}

// CreateTransaction creates a new transaction
func (h *TransactionHandler) CreateTransaction(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify that the related investment/loan exists and belongs to the user
	exists, err := h.verifyRelatedEntity(c.Request.Context(), userID, req.RelatedID, req.Category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify related entity"})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Related investment/loan not found"})
		return
	}

	transaction := models.Transaction{
		RelatedID:   req.RelatedID,
		Category:    req.Category,
		Type:        req.Type,
		Amount:      req.Amount,
		Quantity:    req.Quantity,
		Price:       req.Price,
		Date:        req.Date,
		Description: req.Description,
		Fees:        req.Fees,
		Notes:       req.Notes,
		CreatedAt:   time.Now(),
	}

	// Create transaction in Firestore
	docRef, _, err := h.firestoreClient.Collection("users").Doc(userID).Collection("transactions").Add(c.Request.Context(), transaction)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transaction"})
		return
	}

	transaction.ID = docRef.ID
	c.JSON(http.StatusCreated, transaction)
}

// GetAllTransactions gets all transactions for the authenticated user
func (h *TransactionHandler) GetAllTransactions(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	transactions, err := h.getAllUserTransactions(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get transactions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"transactions": transactions})
}

// Helper methods

func (h *TransactionHandler) getTransactionsByCategory(ctx context.Context, userID string, category models.InvestmentCategory) ([]models.Transaction, error) {
	var transactions []models.Transaction

	iter := h.firestoreClient.Collection("users").Doc(userID).Collection("transactions").
		Where("category", "==", string(category)).
		OrderBy("date", firestore.Desc).
		Documents(ctx)

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var transaction models.Transaction
		if err := doc.DataTo(&transaction); err != nil {
			return nil, err
		}
		transaction.ID = doc.Ref.ID
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

func (h *TransactionHandler) getTransactionsByRelatedID(ctx context.Context, userID, relatedID string) ([]models.Transaction, error) {
	var transactions []models.Transaction

	iter := h.firestoreClient.Collection("users").Doc(userID).Collection("transactions").
		Where("relatedId", "==", relatedID).
		OrderBy("date", firestore.Desc).
		Documents(ctx)

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var transaction models.Transaction
		if err := doc.DataTo(&transaction); err != nil {
			return nil, err
		}
		transaction.ID = doc.Ref.ID
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

func (h *TransactionHandler) getAllUserTransactions(ctx context.Context, userID string) ([]models.Transaction, error) {
	var transactions []models.Transaction

	iter := h.firestoreClient.Collection("users").Doc(userID).Collection("transactions").
		OrderBy("date", firestore.Desc).
		Documents(ctx)

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var transaction models.Transaction
		if err := doc.DataTo(&transaction); err != nil {
			return nil, err
		}
		transaction.ID = doc.Ref.ID
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

func (h *TransactionHandler) verifyRelatedEntity(ctx context.Context, userID, relatedID string, category models.InvestmentCategory) (bool, error) {
	if category == models.CategoryLoans {
		// Check if loan exists
		_, err := h.firestoreClient.Collection("users").Doc(userID).Collection("loans").Doc(relatedID).Get(ctx)
		if err != nil {
			return false, nil // Document doesn't exist
		}
		return true, nil
	} else {
		// Check if investment exists
		_, err := h.firestoreClient.Collection("users").Doc(userID).Collection("investments").Doc(relatedID).Get(ctx)
		if err != nil {
			return false, nil // Document doesn't exist
		}
		return true, nil
	}
}
