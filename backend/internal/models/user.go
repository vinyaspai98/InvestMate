package models

import "time"

// User represents the user profile document
type User struct {
	ID                string          `json:"id" firestore:"-"`
	Email             string          `json:"email" firestore:"email"`
	Name              string          `json:"name" firestore:"name"`
	PhoneNumber       string          `json:"phoneNumber" firestore:"phoneNumber"`
	Preferences       UserPreferences `json:"preferences" firestore:"preferences"`
	CreatedAt         time.Time       `json:"createdAt" firestore:"createdAt"`
	UpdatedAt         time.Time       `json:"updatedAt" firestore:"updatedAt"`
	IsActive          bool            `json:"isActive" firestore:"isActive"`
	GmailAccessToken  string          `json:"-" firestore:"gmailAccessToken,omitempty"`                    // OAuth access token (not exposed in JSON)
	GmailRefreshToken string          `json:"-" firestore:"gmailRefreshToken,omitempty"`                   // OAuth refresh token (not exposed in JSON)
	GmailTokenExpiry  *time.Time      `json:"-" firestore:"gmailTokenExpiry,omitempty"`                    // Token expiration time
	LastGmailSync     *time.Time      `json:"lastGmailSync,omitempty" firestore:"lastGmailSync,omitempty"` // Last successful sync
}

// UserPreferences represents user settings and preferences
type UserPreferences struct {
	Currency      string                  `json:"currency" firestore:"currency"`
	Theme         string                  `json:"theme" firestore:"theme"`
	Notifications NotificationPreferences `json:"notifications" firestore:"notifications"`
}

// NotificationPreferences represents notification settings
type NotificationPreferences struct {
	Email bool `json:"email" firestore:"email"`
	Push  bool `json:"push" firestore:"push"`
}

// Request/Response models for API endpoints

// VerifyTokenRequest represents the token verification request
type VerifyTokenRequest struct {
	Token            string `json:"token" validate:"required"`
	GmailAccessToken string `json:"gmailAccessToken,omitempty"`
}

// UpdateProfileRequest represents the user profile update request
type UpdateProfileRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1"`
	PhoneNumber *string `json:"phoneNumber,omitempty" validate:"omitempty,min=10"`
}

// UpdatePreferencesRequest represents the user preferences update request
type UpdatePreferencesRequest struct {
	Currency      *string                  `json:"currency,omitempty" validate:"omitempty,oneof=INR USD EUR"`
	Theme         *string                  `json:"theme,omitempty" validate:"omitempty,oneof=light dark"`
	Notifications *NotificationPreferences `json:"notifications,omitempty"`
}

// UserResponse represents the user data response
type UserResponse struct {
	User User `json:"user"`
}

// CreateUserRequest represents the request to create a new user
type CreateUserRequest struct {
	Email       string `json:"email" validate:"required,email"`
	Name        string `json:"name" validate:"required,min=1"`
	PhoneNumber string `json:"phoneNumber,omitempty" validate:"omitempty,min=10"`
}
