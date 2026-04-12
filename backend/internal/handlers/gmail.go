package handlers

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"investmate-backend/internal/constants"
	"investmate-backend/internal/models"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/html"
	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type GmailHandler struct {
	firestoreClient *firestore.Client
}

func NewGmailHandler(firestoreClient *firestore.Client) *GmailHandler {
	return &GmailHandler{
		firestoreClient: firestoreClient,
	}
}

// SyncGmail handles the Gmail sync request
func (h *GmailHandler) SyncGmail(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get user profile to retrieve Gmail OAuth token
	user, err := h.getUserProfile(c.Request.Context(), userID)
	if err != nil {
		if strings.Contains(err.Error(), "NotFound") || strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "User profile not found. Please log in again to sync your profile."})
			return
		}
		log.Printf("Failed to get user profile for %s: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user profile"})
		return
	}

	if user.GmailAccessToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gmail access token not found. Please re-authenticate with Google."})
		return
	}

	// Connect to Gmail API
	gmailService, err := h.connectToGmail(c.Request.Context(), user.GmailAccessToken)
	if err != nil {
		log.Printf("Failed to connect to Gmail: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to Gmail"})
		return
	}

	// Fetch CDSL emails after last sync
	emails, err := h.fetchCDSLEmails(gmailService, userID, user.LastGmailSync)
	if err != nil {
		log.Printf("Failed to fetch emails: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch emails"})
		return
	}

	if len(emails) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "No new emails found",
			"synced":  0,
		})
		return
	}

	// Parse emails and extract investment data
	syncedCount := 0
	for _, email := range emails {
		investments, transactions, err := h.parseEmailContent(email)
		if err != nil {
			log.Printf("Failed to parse email %s: %v", email.Id, err)
			continue
		}

		// Save investment data to Firestore
		err = h.saveInvestmentData(c.Request.Context(), userID, investments, transactions)
		if err != nil {
			log.Printf("Failed to save investment data: %v", err)
			continue
		}

		syncedCount++
	}

	// Update last sync timestamp
	now := time.Now()
	_, err = h.firestoreClient.Collection("users").Doc(userID).Update(c.Request.Context(), []firestore.Update{
		{Path: "lastGmailSync", Value: now},
		{Path: "updatedAt", Value: now},
	})
	if err != nil {
		log.Printf("Failed to update last sync timestamp: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Gmail sync completed successfully",
		"synced":  syncedCount,
		"total":   len(emails),
	})
}

// connectToGmail establishes connection to Gmail API using OAuth token
func (h *GmailHandler) connectToGmail(ctx context.Context, accessToken string) (*gmail.Service, error) {
	// Create Gmail service with OAuth access token
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})
	client := oauth2.NewClient(ctx, ts)

	gmailService, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to create Gmail service: %v", err)
	}

	return gmailService, nil
}

// fetchCDSLEmails fetches emails from CDSL India after a specific timestamp
func (h *GmailHandler) fetchCDSLEmails(gmailService *gmail.Service, userID string, lastSync *time.Time) ([]*gmail.Message, error) {
	// Build query to filter emails
	query := fmt.Sprintf("subject:\"%s\"", constants.CDSLEmailSubject)

	// Only sync emails after the last successful sync
	if lastSync != nil && !lastSync.IsZero() {
		// Gmail after: operator supports Unix timestamps in seconds
		query = fmt.Sprintf("%s after:%d", query, lastSync.Unix())
		log.Printf("Syncing emails after %v (Unix: %d)", lastSync, lastSync.Unix())
	}

	// List messages
	listCall := gmailService.Users.Messages.List("me").Q(query).MaxResults(int64(constants.MaxEmailsToFetch))
	response, err := listCall.Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve messages: %v", err)
	}

	if len(response.Messages) == 0 {
		return []*gmail.Message{}, nil
	}

	// Fetch full message details
	var emails []*gmail.Message
	// Reverse messages to process from oldest to newest
	slices.Reverse(response.Messages)

	for _, msg := range response.Messages {
		fullMsg, err := gmailService.Users.Messages.Get("me", msg.Id).Format("full").Do()
		if err != nil {
			log.Printf("Failed to get message %s: %v", msg.Id, err)
			continue
		}
		emails = append(emails, fullMsg)
	}

	return emails, nil
}

// parseEmailContent extracts investment data from email
func (h *GmailHandler) parseEmailContent(email *gmail.Message) ([]models.Investment, []models.Transaction, error) {
	// Get email body
	var body string
	if email.Payload.Body.Data != "" {
		data, err := base64.URLEncoding.DecodeString(email.Payload.Body.Data)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to decode email body: %v", err)
		}
		body = string(data)
	} else if len(email.Payload.Parts) > 0 {
		// Multi-part email, find HTML part
		for _, part := range email.Payload.Parts {
			if part.MimeType == "text/html" && part.Body.Data != "" {
				data, err := base64.URLEncoding.DecodeString(part.Body.Data)
				if err != nil {
					continue
				}
				body = string(data)
				break
			}
		}
	}

	if body == "" {
		return nil, nil, fmt.Errorf("empty email body")
	}

	// Parse HTML to extract transaction data
	investments, transactions := h.extractTransactionsFromHTML(body, email)

	return investments, transactions, nil
}

// extractTransactionsFromHTML parses HTML and extracts transaction data using table parsing
func (h *GmailHandler) extractTransactionsFromHTML(htmlContent string, email *gmail.Message) ([]models.Investment, []models.Transaction) {
	var investments []models.Investment
	var transactions []models.Transaction

	// Parse HTML
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		log.Printf("Failed to parse HTML: %v", err)
		return investments, transactions
	}

	// Extract all table rows
	rows := h.extractTableData(doc)

	// Get email date as fallback
	emailDate := time.Now()
	if email.InternalDate != 0 {
		emailDate = time.Unix(email.InternalDate/1000, 0)
	}

	// Parse stock transactions from rows
	stockInvestments, stockTransactions := h.parseStockRows(rows, emailDate)
	investments = append(investments, stockInvestments...)
	transactions = append(transactions, stockTransactions...)

	// Fallback to legacy text-based parsing if no investments found (optional, but good for compatibility)
	if len(investments) == 0 {
		text := h.extractTextFromNode(doc)
		mfInvestments, mfTransactions := h.parseMutualFundTransactions(text, email)
		investments = append(investments, mfInvestments...)
		transactions = append(transactions, mfTransactions...)
	}

	return investments, transactions
}

// extractTableData recursively finds all tables and returns all rows with their cell contents
func (h *GmailHandler) extractTableData(n *html.Node) [][]string {
	var rows [][]string
	var findRows func(*html.Node)
	findRows = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "tr" {
			var row []string
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.ElementNode && (c.Data == "td" || c.Data == "th") {
					cellText := strings.TrimSpace(h.extractTextFromNode(c))
					// Clean up whitespace and newlines within the cell
					cellText = regexp.MustCompile(`\s+`).ReplaceAllString(cellText, " ")
					row = append(row, cellText)
				}
			}
			if len(row) > 0 {
				rows = append(rows, row)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findRows(c)
		}
	}
	findRows(n)
	return rows
}

// parseStockRows extracts transaction data from structured table rows
func (h *GmailHandler) parseStockRows(rows [][]string, emailDate time.Time) ([]models.Investment, []models.Transaction) {
	var investments []models.Investment
	var transactions []models.Transaction

	// Robust regex to extract fields from a potentially single-string row
	rowPattern := regexp.MustCompile(`(?i)(.*?)\s+(IN[EF][0-9A-Z]{9})\s+([\d,.]+)\s+(Credit|Debit|Buy|Sell)\s+(\d{2}[/-]\d{2}[/-]\d{4}.*)`)

	for _, row := range rows {
		rowText := strings.TrimSpace(strings.Join(row, " "))
		if rowText == "" || strings.Contains(strings.ToLower(rowText), "isin") {
			continue
		}

		matches := rowPattern.FindStringSubmatch(rowText)
		if len(matches) < 6 {
			continue
		}

		company := strings.TrimSpace(matches[1])
		isin := matches[2]
		quantityStr := matches[3]
		transType := strings.ToUpper(matches[4])
		dateStr := matches[5]

		company = regexp.MustCompile(`^\d+\s+`).ReplaceAllString(company, "")
		quantityStr = strings.ReplaceAll(quantityStr, ",", "")
		quantity, _ := strconv.ParseFloat(quantityStr, 64)
		if quantity == 0 {
			continue
		}

		txType := models.TransactionTypeBuy
		if strings.Contains(transType, "CREDIT") {
			txType = models.TransactionTypeBuy
		} else if strings.Contains(transType, "DEBIT") {
			txType = models.TransactionTypeSell
		}

		txDate := emailDate
		if dateStr != "" {
			parsedDate, err := h.parseDate(dateStr)
			if err == nil {
				txDate = parsedDate
			}
		}

		investment := models.Investment{
			Category:  models.CategoryStocks,
			Name:      company,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			IsActive:  true,
			InvestmentData: models.InvestmentData{
				Ticker:   isin,
				Quantity: quantity,
			},
		}
		investments = append(investments, investment)

		transaction := models.Transaction{
			Category:    models.CategoryStocks,
			Type:        txType,
			Quantity:    quantity,
			Date:        txDate,
			Description: fmt.Sprintf("%s (%s) %s transaction from CDSL", company, isin, transType),
			CreatedAt:   time.Now(),
		}
		transactions = append(transactions, transaction)

		log.Printf("Successfully parsed: Company: %s, ISIN: %s, Qty: %0.3f, Type: %s, Date: %v",
			company, isin, quantity, transType, txDate)
	}

	return investments, transactions
}

// parseDate is a helper to parse various date formats typically found in Indian statements
func (h *GmailHandler) parseDate(dateStr string) (time.Time, error) {
	formats := []string{
		"02/01/2006",
		"02-01-2006",
		"02-Jan-2006",
		"02 Jan 2006",
		"2006-01-02",
	}
	for _, format := range formats {
		t, err := time.Parse(format, dateStr)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("could not parse date: %s", dateStr)
}

// extractTextFromNode recursively extracts text from HTML nodes
func (h *GmailHandler) extractTextFromNode(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}

	var text string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		text += h.extractTextFromNode(c)
	}
	return text
}

// parseStockTransactions extracts stock transaction data
func (h *GmailHandler) parseStockTransactions(text string, email *gmail.Message) ([]models.Investment, []models.Transaction) {
	var investments []models.Investment
	var transactions []models.Transaction

	// Regex patterns for stock transactions
	// Example: "RELIANCE | BUY | 10 | 2500.00 | 25000.00"
	stockPattern := regexp.MustCompile(`([A-Z]+)\s*\|\s*(Debit|Credit)\s*\|\s*(\d+\.?\d*)\s*\|\s*(\d+\.?\d*)\s*\|\s*(\d+\.?\d*)`)
	matches := stockPattern.FindAllStringSubmatch(text, -1)
	log.Printf("match HTML: %v", len(matches))

	for _, match := range matches {
		if len(match) < 6 {
			continue
		}

		ticker := strings.TrimSpace(match[1])
		transType := strings.TrimSpace(match[2])
		quantity, _ := strconv.ParseFloat(match[3], 64)
		price, _ := strconv.ParseFloat(match[4], 64)
		amount, _ := strconv.ParseFloat(match[5], 64)

		// Create investment
		investment := models.Investment{
			Category:       models.CategoryStocks,
			Name:           ticker,
			InvestedAmount: amount,
			CurrentValue:   amount, // Will be updated with real-time prices later
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			IsActive:       true,
			InvestmentData: models.InvestmentData{
				Ticker:       ticker,
				Quantity:     quantity,
				AveragePrice: price,
				CurrentPrice: price,
				Exchange:     "NSE", // Default to NSE
			},
		}
		investments = append(investments, investment)

		// Create transaction
		var txType models.TransactionType
		if transType == "BUY" {
			txType = models.TransactionTypeBuy
		} else {
			txType = models.TransactionTypeSell
		}

		transaction := models.Transaction{
			Category:    models.CategoryStocks,
			Type:        txType,
			Amount:      amount,
			Quantity:    quantity,
			Price:       price,
			Date:        time.Now(), // Should parse from email date
			Description: fmt.Sprintf("%s %s transaction from CDSL", ticker, transType),
			CreatedAt:   time.Now(),
		}
		transactions = append(transactions, transaction)
	}

	return investments, transactions
}

// parseMutualFundTransactions extracts mutual fund transaction data
func (h *GmailHandler) parseMutualFundTransactions(text string, email *gmail.Message) ([]models.Investment, []models.Transaction) {
	var investments []models.Investment
	var transactions []models.Transaction

	// Regex patterns for mutual fund transactions
	// Example: "HDFC Equity Fund | 100.50 | 150.00 | 15075.00"
	mfPattern := regexp.MustCompile(`([A-Za-z\s]+Fund)\s*\|\s*(\d+\.?\d*)\s*\|\s*(\d+\.?\d*)\s*\|\s*(\d+\.?\d*)`)
	matches := mfPattern.FindAllStringSubmatch(text, -1)

	for _, match := range matches {
		if len(match) < 5 {
			continue
		}

		fundName := strings.TrimSpace(match[1])
		units, _ := strconv.ParseFloat(match[2], 64)
		nav, _ := strconv.ParseFloat(match[3], 64)
		amount, _ := strconv.ParseFloat(match[4], 64)

		// Create investment
		investment := models.Investment{
			Category:       models.CategoryMutualFunds,
			Name:           fundName,
			InvestedAmount: amount,
			CurrentValue:   amount,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			IsActive:       true,
			InvestmentData: models.InvestmentData{
				FundName: fundName,
				Units:    units,
				NAV:      nav,
			},
		}
		investments = append(investments, investment)

		// Create transaction
		transaction := models.Transaction{
			Category:    models.CategoryMutualFunds,
			Type:        models.TransactionTypeBuy,
			Amount:      amount,
			Quantity:    units,
			Price:       nav,
			Date:        time.Now(),
			Description: fmt.Sprintf("%s purchase from CDSL", fundName),
			CreatedAt:   time.Now(),
		}
		transactions = append(transactions, transaction)
	}

	return investments, transactions
}

// saveInvestmentData saves parsed investment data to Firestore
func (h *GmailHandler) saveInvestmentData(ctx context.Context, userID string, investments []models.Investment, transactions []models.Transaction) error {
	if len(investments) == 0 && len(transactions) == 0 {
		log.Println("No data to save to Firestore")
		return nil
	}

	batch := h.firestoreClient.Batch()

	// Track investment refs to link with transactions
	invRefs := make([]*firestore.DocumentRef, len(investments))

	// Save investments
	for i, investment := range investments {
		tx := transactions[i]

		// Find existing investment by Category and ISIN (Ticker)
		// If ISIN is missing, fallback to Name
		var query firestore.Query
		if investment.InvestmentData.Ticker != "" {
			query = h.firestoreClient.Collection("users").Doc(userID).Collection("investments").
				Where("category", "==", investment.Category).
				Where("investmentData.ticker", "==", investment.InvestmentData.Ticker).
				Limit(1)
		} else {
			query = h.firestoreClient.Collection("users").Doc(userID).Collection("investments").
				Where("category", "==", investment.Category).
				Where("name", "==", investment.Name).
				Limit(1)
		}

		docs, err := query.Documents(ctx).GetAll()
		if err != nil {
			log.Printf("Failed to query existing investments: %v", err)
			continue
		}

		if len(docs) > 0 {
			// Update existing investment
			var existing models.Investment
			docs[0].DataTo(&existing)
			docRef := docs[0].Ref
			invRefs[i] = docRef

			if tx.Type == models.TransactionTypeBuy {
				// Add quantities/units
				if investment.Category == models.CategoryStocks {
					existing.InvestmentData.Quantity += investment.InvestmentData.Quantity
				} else if investment.Category == models.CategoryMutualFunds {
					existing.InvestmentData.Units += investment.InvestmentData.Units
				}
				existing.UpdatedAt = time.Now()
				batch.Set(docRef, existing)
			} else if tx.Type == models.TransactionTypeSell {
				// Subtract quantities/units
				if investment.Category == models.CategoryStocks {
					existing.InvestmentData.Quantity -= investment.InvestmentData.Quantity
				} else if investment.Category == models.CategoryMutualFunds {
					existing.InvestmentData.Units -= investment.InvestmentData.Units
				}

				// Check if quantity became zero or negative
				qty := 0.0
				if investment.Category == models.CategoryStocks {
					qty = existing.InvestmentData.Quantity
				} else {
					qty = existing.InvestmentData.Units
				}

				if qty <= 0 {
					log.Printf("Deleting investment %s as quantity reached %0.3f", existing.Name, qty)
					batch.Delete(docRef)
				} else {
					existing.UpdatedAt = time.Now()
					batch.Set(docRef, existing)
				}
			}
		} else {
			// Create new investment only for Buy transactions
			if tx.Type == models.TransactionTypeBuy {
				ref := h.firestoreClient.Collection("users").Doc(userID).Collection("investments").NewDoc()
				batch.Set(ref, investment)
				invRefs[i] = ref
			}
		}
	}

	// Save transactions
	for i, transaction := range transactions {
		// Link transaction to investment if possible
		if invRefs[i] != nil {
			transaction.RelatedID = invRefs[i].ID
		}

		ref := h.firestoreClient.Collection("users").Doc(userID).Collection("transactions").NewDoc()
		batch.Set(ref, transaction)
	}

	// Commit batch
	_, err := batch.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit batch: %v", err)
	}

	return nil
}

// getUserProfile retrieves user profile from Firestore
func (h *GmailHandler) getUserProfile(ctx context.Context, userID string) (*models.User, error) {
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
