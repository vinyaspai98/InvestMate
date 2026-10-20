package handlers

import (
	_ "embed"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
)

// GmailHandler handles Gmail sync and investment extraction.
type GmailHandler struct {
	firestoreClient    *firestore.Client
	geminiAPIKey       string
	alphaVantageAPIKey string
}

// NewGmailHandler creates a new GmailHandler with the required API keys.
func NewGmailHandler(firestoreClient *firestore.Client, geminiAPIKey, alphaVantageAPIKey string) *GmailHandler {
	return &GmailHandler{
		firestoreClient:    firestoreClient,
		geminiAPIKey:       geminiAPIKey,
		alphaVantageAPIKey: alphaVantageAPIKey,
	}
}

// SyncGmail handles the Gmail sync request.
func (h *GmailHandler) SyncGmail(c *gin.Context) {

}
