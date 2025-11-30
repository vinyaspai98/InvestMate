package handlers

import (
	"context"
	"net/http"
	"time"

	"investmate-backend/internal/models"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	firebaseApp     *firebase.App
	firestoreClient *firestore.Client
	validator       *validator.Validate
}

func NewAuthHandler(firebaseApp *firebase.App, firestoreClient *firestore.Client) *AuthHandler {
	return &AuthHandler{
		firebaseApp:     firebaseApp,
		firestoreClient: firestoreClient,
		validator:       validator.New(),
	}
}

// VerifyToken verifies Firebase JWT token
func (h *AuthHandler) VerifyToken(c *gin.Context) {
	var req models.VerifyTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify Firebase ID token
	authClient, err := h.firebaseApp.Auth(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Auth client"})
		return
	}

	token, err := authClient.VerifyIDToken(c.Request.Context(), req.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// Get or create user profile
	user, err := h.getOrCreateUser(c.Request.Context(), token.UID, token.Claims)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user profile"})
		return
	}

	c.JSON(http.StatusOK, models.UserResponse{User: *user})
}

// GetProfile gets the authenticated user's profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	user, err := h.getUserProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, models.UserResponse{User: *user})
}

// UpdateProfile updates the authenticated user's profile
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build update data
	updateData := map[string]interface{}{
		"updatedAt": time.Now(),
	}

	if req.Name != nil {
		updateData["name"] = *req.Name
	}

	if req.PhoneNumber != nil {
		updateData["phoneNumber"] = *req.PhoneNumber
	}

	// Update user profile in Firestore
	userRef := h.firestoreClient.Collection("users").Doc(userID)
	_, err := userRef.Update(c.Request.Context(), []firestore.Update{
		{Path: "name", Value: updateData["name"]},
		{Path: "phoneNumber", Value: updateData["phoneNumber"]},
		{Path: "updatedAt", Value: updateData["updatedAt"]},
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	// Get updated user profile
	user, err := h.getUserProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated profile"})
		return
	}

	c.JSON(http.StatusOK, models.UserResponse{User: *user})
}

// UpdatePreferences updates the authenticated user's preferences
func (h *AuthHandler) UpdatePreferences(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.UpdatePreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build preference updates
	updates := []firestore.Update{
		{Path: "updatedAt", Value: time.Now()},
	}

	if req.Currency != nil {
		updates = append(updates, firestore.Update{Path: "preferences.currency", Value: *req.Currency})
	}

	if req.Theme != nil {
		updates = append(updates, firestore.Update{Path: "preferences.theme", Value: *req.Theme})
	}

	if req.Notifications != nil {
		updates = append(updates, firestore.Update{Path: "preferences.notifications", Value: *req.Notifications})
	}

	// Update user preferences in Firestore
	userRef := h.firestoreClient.Collection("users").Doc(userID)
	_, err := userRef.Update(c.Request.Context(), updates)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update preferences"})
		return
	}

	// Get updated user profile
	user, err := h.getUserProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated profile"})
		return
	}

	c.JSON(http.StatusOK, models.UserResponse{User: *user})
}

// Helper methods

func (h *AuthHandler) getOrCreateUser(ctx context.Context, userID string, claims map[string]interface{}) (*models.User, error) {
	userRef := h.firestoreClient.Collection("users").Doc(userID)
	doc, err := userRef.Get(ctx)

	if err != nil {
		// User doesn't exist, create new user
		email, _ := claims["email"].(string)
		name, _ := claims["name"].(string)
		if name == "" {
			name = email // fallback to email if name not available
		}

		user := models.User{
			ID:          userID,
			Email:       email,
			Name:        name,
			PhoneNumber: "",
			Preferences: models.UserPreferences{
				Currency: "INR",
				Theme:    "light",
				Notifications: models.NotificationPreferences{
					Email: true,
					Push:  false,
				},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			IsActive:  true,
		}

		_, err = userRef.Set(ctx, user)
		if err != nil {
			return nil, err
		}

		return &user, nil
	}

	// User exists, get user data
	var user models.User
	if err := doc.DataTo(&user); err != nil {
		return nil, err
	}
	user.ID = userID

	return &user, nil
}

func (h *AuthHandler) getUserProfile(ctx context.Context, userID string) (*models.User, error) {
	userRef := h.firestoreClient.Collection("users").Doc(userID)
	doc, err := userRef.Get(ctx)
	if err != nil {
		return nil, err
	}

	var user models.User
	if err := doc.DataTo(&user); err != nil {
		return nil, err
	}
	user.ID = userID

	return &user, nil
}
