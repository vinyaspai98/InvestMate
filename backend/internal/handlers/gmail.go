package handlers

import (
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"sort"
	"strings"
	"time"

	"investmate-backend/internal/constants"
	"investmate-backend/internal/models"
	"investmate-backend/internal/services/caspdf"
	"investmate-backend/internal/services/gemini"
	"investmate-backend/internal/services/tradedate"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/html"
	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// ---------------------------------------------------------------------------
// AI extraction schema structs
// ---------------------------------------------------------------------------

// AIAssistantTransaction is the per-transaction shape that Gemini must return.
type AIAssistantTransaction struct {
	Category        string  `json:"category"`         // "stocks" | "mutualFunds"
	Name            string  `json:"name"`             // cleaned company / fund name
	ISIN            string  `json:"isin,omitempty"`   // 12-char Indian ISIN code
	Ticker          string  `json:"ticker"`           // Stock exchange ticker symbol (e.g. RELIANCE.BSE, TCS.BSE)
	Quantity        float64 `json:"quantity"`         // numeric, positive
	Type            string  `json:"type"`             // "buy" | "sell"
	TransactionDate string  `json:"transaction_date"` // YYYY-MM-DD
}

// AIExtractionResponse is the top-level JSON envelope Gemini returns.
type AIExtractionResponse struct {
	Transactions []AIAssistantTransaction `json:"transactions"`
}

// ---------------------------------------------------------------------------
// Gemini request/response wire types (minimal, REST-only)
// ---------------------------------------------------------------------------

type geminiPart struct {
	Text string `json:"text"`
}
type geminiContent struct {
	Role  string       `json:"role"`
	Parts []geminiPart `json:"parts"`
}
type geminiSchemaProperty struct {
	Type string `json:"type"`
}
type geminiSchemaObject struct {
	Type       string                          `json:"type"`
	Properties map[string]geminiSchemaProperty `json:"properties"`
	Required   []string                        `json:"required"`
}
type geminiResponseSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Required   []string               `json:"required"`
}
type geminiGenerationConfig struct {
	ResponseMIMEType string               `json:"responseMimeType"`
	ResponseSchema   geminiResponseSchema `json:"responseSchema"`
}
type geminiRequest struct {
	SystemInstruction geminiContent          `json:"system_instruction"`
	Contents          []geminiContent        `json:"contents"`
	GenerationConfig  geminiGenerationConfig `json:"generationConfig"`
}
type geminiCandidate struct {
	Content geminiContent `json:"content"`
}
type geminiResponse struct {
	Candidates []geminiCandidate `json:"candidates"`
}

// ---------------------------------------------------------------------------
// Yahoo Finance response types
// ---------------------------------------------------------------------------

type yfMeta struct {
	Currency           string  `json:"currency"`
	Symbol             string  `json:"symbol"`
	RegularMarketPrice float64 `json:"regularMarketPrice"`
}
type yfQuote struct {
	Close []*float64 `json:"close"`
	Open  []*float64 `json:"open"`
	High  []*float64 `json:"high"`
	Low   []*float64 `json:"low"`
}

// getDayPrice returns the average price for the day ((High+Low)/2) or falls back to Close.
func (q yfQuote) getDayPrice(index int) float64 {
	if index < 0 || index >= len(q.Close) || q.Close[index] == nil {
		return 0
	}
	closePrice := *q.Close[index]
	var highPrice, lowPrice float64
	if index < len(q.High) && q.High[index] != nil {
		highPrice = *q.High[index]
	}
	if index < len(q.Low) && q.Low[index] != nil {
		lowPrice = *q.Low[index]
	}
	if highPrice > 0 && lowPrice > 0 {
		return (highPrice + lowPrice) / 2.0
	}
	return closePrice
}
type yfIndicators struct {
	Quote []yfQuote `json:"quote"`
}
type yfResult struct {
	Meta       yfMeta       `json:"meta"`
	Timestamp  []int64      `json:"timestamp"`
	Indicators yfIndicators `json:"indicators"`
}
type yfChart struct {
	Result []yfResult `json:"result"`
}
type yfChartResponse struct {
	Chart yfChart `json:"chart"`
}

// ---------------------------------------------------------------------------
// Alpha Vantage response types
// ---------------------------------------------------------------------------

type avDailyMeta struct {
	LastRefreshed string `json:"3. Last Refreshed"`
}
type avDailyDay struct {
	Close string `json:"4. close"`
}
type avDailyResponse struct {
	MetaData   avDailyMeta           `json:"Meta Data"`
	TimeSeries map[string]avDailyDay `json:"Time Series (Daily)"`
}

// ---------------------------------------------------------------------------
// mfapi.in response types
// ---------------------------------------------------------------------------

type mfapiSearchResult struct {
	SchemeCode int    `json:"schemeCode"`
	SchemeName string `json:"schemeName"`
}
type mfapiNAVEntry struct {
	Date string `json:"date"`
	Nav  string `json:"nav"`
}
type mfapiSchemeResponse struct {
	Data []mfapiNAVEntry `json:"data"`
}

// ---------------------------------------------------------------------------
// GmailHandler
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// SyncGmailResponse represents the structured response of hybrid Gmail sync.
type SyncGmailResponse struct {
	Message                string     `json:"message"`
	SyncMode               string     `json:"syncMode"` // "cas_bootstrap", "incremental", "price_refresh", "full_reconcile"
	Synced                 int        `json:"synced"`
	Total                  int        `json:"total,omitempty"`
	CASProcessed           bool       `json:"casProcessed"`
	PortfolioEffectiveDate *time.Time `json:"portfolioEffectiveDate,omitempty"`
	LastGmailSync          time.Time  `json:"lastGmailSync"`
}

// SyncGmail handles the hybrid Gmail sync request:
// 1. Validates PAN card requirement.
// 2. If first sync or new CAS mail: bootstraps portfolio from CDSL CAS PDF and reconciles.
// 3. Incrementally syncs Demat transactions from CAS effective date / last sync.
// 4. If already synced and no new emails: executes fast price-only refresh directly from database.
func (h *GmailHandler) SyncGmail(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get user profile
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

	// Requirement 1 & 2: PAN card number is required for CAS PDF decryption
	if strings.TrimSpace(user.PANNumber) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "PAN card number is required to decrypt your CDSL CAS statement. Please update your PAN in your Profile.",
		})
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

	ctx := c.Request.Context()
	isFirstSync := user.LastGmailSync == nil || user.CASEffectiveDate == nil

	if isFirstSync {
		log.Printf("SyncGmail: First sync for user %s, initiating CAS bootstrap", userID)
		h.doCASBootstrapSync(c, ctx, gmailService, user, userID)
		return
	}

	// Requirement 5: Check if there are new CAS emails since last sync
	hasNewCAS := h.hasNewCASEmail(gmailService, user.LastGmailSync)
	if hasNewCAS {
		log.Printf("SyncGmail: New CAS email detected for user %s. Running full reconciliation.", userID)
		h.doCASBootstrapSync(c, ctx, gmailService, user, userID)
		return
	}

	// Check if there are new Demat transaction emails since last sync
	hasNewTx := h.hasNewDematTxEmail(gmailService, user.LastGmailSync)
	if hasNewTx {
		log.Printf("SyncGmail: New Demat transactions detected for user %s. Running incremental sync.", userID)
		h.doIncrementalSync(c, ctx, gmailService, user, userID)
		return
	}

	// If neither new CAS nor new Demat Tx emails exist: price-only refresh directly from DB
	log.Printf("SyncGmail: No new CAS or Demat emails for user %s. Running price-only refresh from database.", userID)
	h.doPriceRefresh(c, ctx, user, userID)
}

// hasNewCASEmail checks if any CDSL CAS emails arrived after the given timestamp.
func (h *GmailHandler) hasNewCASEmail(gmailService *gmail.Service, since *time.Time) bool {
	query := "subject:\"CDSL Consolidated Account Statement\""
	if since != nil && !since.IsZero() {
		query = fmt.Sprintf("%s after:%d", query, since.Unix())
	}
	resp, err := gmailService.Users.Messages.List("me").Q(query).MaxResults(1).Do()
	if err != nil {
		log.Printf("hasNewCASEmail error: %v", err)
		return false
	}
	return len(resp.Messages) > 0
}

// hasNewDematTxEmail checks if any Demat transaction emails arrived after the given timestamp.
func (h *GmailHandler) hasNewDematTxEmail(gmailService *gmail.Service, since *time.Time) bool {
	query := fmt.Sprintf("subject:\"%s\"", constants.CDSLEmailSubject)
	if since != nil && !since.IsZero() {
		query = fmt.Sprintf("%s after:%d", query, since.Unix())
	}
	resp, err := gmailService.Users.Messages.List("me").Q(query).MaxResults(1).Do()
	if err != nil {
		log.Printf("hasNewDematTxEmail error: %v", err)
		return false
	}
	return len(resp.Messages) > 0
}

// CASEmailItem represents a fetched CAS email with its decrypted PDF payload and metadata.
type CASEmailItem struct {
	MessageID string
	Subject   string
	EmailDate time.Time
	PDFBytes  []byte
}

// fetchAllCASEmails retrieves ALL CDSL CAS emails with PDF attachments (sorted chronologically from oldest to newest).
func (h *GmailHandler) fetchAllCASEmails(gmailService *gmail.Service, since *time.Time) ([]CASEmailItem, error) {
	query := "subject:\"CDSL Consolidated Account Statement\""
	if since != nil && !since.IsZero() {
		query = fmt.Sprintf("%s after:%d", query, since.Unix())
	}

	resp, err := gmailService.Users.Messages.List("me").Q(query).MaxResults(int64(constants.MaxEmailsToFetch)).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to query CAS emails: %w", err)
	}
	if len(resp.Messages) == 0 {
		return nil, fmt.Errorf("no CDSL Consolidated Account Statement emails found")
	}

	var casItems []CASEmailItem
	for _, msgMeta := range resp.Messages {
		fullMsg, err := gmailService.Users.Messages.Get("me", msgMeta.Id).Format("full").Do()
		if err != nil {
			continue
		}
		pdfBytes, err := h.extractPDFAttachment(gmailService, fullMsg)
		if err != nil || len(pdfBytes) == 0 {
			continue
		}

		emailDate := time.Unix(fullMsg.InternalDate/1000, 0)
		subject := ""
		if fullMsg.Payload != nil {
			for _, hdr := range fullMsg.Payload.Headers {
				if strings.EqualFold(hdr.Name, "Subject") {
					subject = hdr.Value
				}
				if strings.EqualFold(hdr.Name, "Date") && emailDate.IsZero() {
					if parsed, err := time.Parse(time.RFC1123Z, hdr.Value); err == nil {
						emailDate = parsed
					} else if parsed, err := time.Parse(time.RFC1123, hdr.Value); err == nil {
						emailDate = parsed
					}
				}
			}
		}

		casItems = append(casItems, CASEmailItem{
			MessageID: fullMsg.Id,
			Subject:   subject,
			EmailDate: emailDate,
			PDFBytes:  pdfBytes,
		})
	}

	if len(casItems) == 0 {
		return nil, fmt.Errorf("no CAS emails with a PDF attachment could be found")
	}

	// Sort chronologically: OLDEST first -> NEWEST last
	// Earlier months establish real trades with their exact dates;
	// the latest statement gives the most recent portfolio balances.
	sort.Slice(casItems, func(i, j int) bool {
		return casItems[i].EmailDate.Before(casItems[j].EmailDate)
	})

	return casItems, nil
}

// extractPDFAttachment searches recursively for a PDF attachment inside a Gmail message.
func (h *GmailHandler) extractPDFAttachment(gmailService *gmail.Service, msg *gmail.Message) ([]byte, error) {
	var searchPart func(part *gmail.MessagePart) ([]byte, error)
	searchPart = func(part *gmail.MessagePart) ([]byte, error) {
		if part == nil {
			return nil, nil
		}
		isPDF := strings.EqualFold(part.MimeType, "application/pdf") ||
			strings.HasSuffix(strings.ToLower(part.Filename), ".pdf")

		if isPDF && part.Body != nil {
			if part.Body.AttachmentId != "" {
				att, err := gmailService.Users.Messages.Attachments.Get("me", msg.Id, part.Body.AttachmentId).Do()
				if err != nil {
					return nil, fmt.Errorf("failed to download attachment %s: %w", part.Body.AttachmentId, err)
				}
				data, err := base64.URLEncoding.DecodeString(att.Data)
				if err != nil {
					if data2, err2 := base64.RawURLEncoding.DecodeString(att.Data); err2 == nil {
						return data2, nil
					}
					return base64.StdEncoding.DecodeString(att.Data)
				}
				return data, nil
			} else if part.Body.Data != "" {
				data, err := base64.URLEncoding.DecodeString(part.Body.Data)
				if err != nil {
					return base64.RawURLEncoding.DecodeString(part.Body.Data)
				}
				return data, nil
			}
		}

		for _, subPart := range part.Parts {
			if data, err := searchPart(subPart); err == nil && len(data) > 0 {
				return data, nil
			}
		}
		return nil, nil
	}

	if msg.Payload != nil {
		data, err := searchPart(msg.Payload)
		if err != nil {
			return nil, err
		}
		if len(data) > 0 {
			return data, nil
		}
	}

	return nil, fmt.Errorf("no PDF attachment found in email %s", msg.Id)
}

// doCASBootstrapSync processes ALL CDSL CAS statements chronologically from oldest to newest:
// 1. Extracts monthly trades (Section 2) with exact execution dates from each monthly statement.
// 2. Reconciles current holdings (Section 1) from the latest statement without creating dummy transactions on statement date.
// 3. Incrementally syncs subsequent Demat transaction emails up to the present date.
func (h *GmailHandler) doCASBootstrapSync(c *gin.Context, ctx context.Context, gmailService *gmail.Service, user *models.User, userID string) {
	casItems, err := h.fetchAllCASEmails(gmailService, nil)
	if err != nil {
		log.Printf("doCASBootstrapSync: No CAS email found (%v). Falling back to Demat transaction emails.", err)
		h.doIncrementalSync(c, ctx, gmailService, user, userID)
		return
	}

	log.Printf("doCASBootstrapSync: Found %d CDSL CAS statement emails. Processing chronologically...", len(casItems))

	totalTradesExtracted := 0
	totalHoldingsExtracted := 0
	var latestStatementDate time.Time
	var latestHoldings []caspdf.CASHolding

	// Track seen transactions across monthly statements to prevent duplicate insertions
	seenTransactions := make(map[string]bool)

	// Step A: Iterate through all CAS statements from OLDEST to NEWEST
	for idx, item := range casItems {
		log.Printf("Processing CAS statement [%d/%d] from %s (Subject: %q)", idx+1, len(casItems), item.EmailDate.Format("2006-01-02"), item.Subject)

		casResult, err := caspdf.ExtractFromPDF(ctx, item.PDFBytes, user.PANNumber, h.geminiAPIKey)
		if err != nil {
			log.Printf("Failed to decrypt/extract CAS PDF from email %s: %v. Continuing with remaining...", item.MessageID, err)
			continue
		}

		stmtDate, err := caspdf.ParseStatementDate(casResult.StatementDate)
		if err != nil {
			stmtDate = item.EmailDate
		}

		if stmtDate.After(latestStatementDate) || latestStatementDate.IsZero() {
			latestStatementDate = stmtDate
			latestHoldings = casResult.Holdings
		}

		// Save the trades from Section 2 for this month with their exact execution dates
		if len(casResult.Transactions) > 0 {
			savedTrades, err := h.saveCASTransactions(ctx, userID, casResult.Transactions, seenTransactions)
			if err != nil {
				log.Printf("Warning: error saving some CAS transactions: %v", err)
			}
			totalTradesExtracted += savedTrades
			log.Printf("CAS [%d/%d]: saved %d trading transactions with exact execution dates", idx+1, len(casItems), savedTrades)
		}
	}

	if latestStatementDate.IsZero() {
		latestStatementDate = time.Now().AddDate(0, -1, 0)
	}

	// Step B: Reconcile current holdings from the latest CAS statement (Section 1)
	savedHoldings, err := h.reconcileCASHoldings(ctx, userID, latestHoldings, latestStatementDate, seenTransactions)
	if err != nil {
		log.Printf("Failed to reconcile CAS holdings: %v", err)
	}
	totalHoldingsExtracted = savedHoldings

	now := time.Now()
	_, err = h.firestoreClient.Collection("users").Doc(userID).Update(ctx, []firestore.Update{
		{Path: "lastCASProcessed", Value: now},
		{Path: "casEffectiveDate", Value: latestStatementDate},
		{Path: "lastGmailSync", Value: now},
		{Path: "updatedAt", Value: now},
	})
	if err != nil {
		log.Printf("doCASBootstrapSync: Failed to update user CAS metadata: %v", err)
	}

	// Step C: Incrementally scan Demat transaction emails strictly starting from latestStatementDate to current date
	incEmails, err := h.fetchCDSLEmails(gmailService, userID, &latestStatementDate)
	syncedTxCount := 0
	if err == nil && len(incEmails) > 0 {
		log.Printf("doCASBootstrapSync: Processing %d incremental Demat emails after latest CAS date %v", len(incEmails), latestStatementDate)
		for _, email := range incEmails {
			investments, transactions, err := h.parseEmailContent(ctx, email)
			if err != nil || len(investments) == 0 {
				continue
			}
			if err := h.saveInvestmentData(ctx, userID, investments, transactions); err == nil {
				syncedTxCount++
			}
		}
	}

	mode := "cas_bootstrap"
	if user.LastGmailSync != nil {
		mode = "full_reconcile"
	}

	totalSynced := totalTradesExtracted + totalHoldingsExtracted + syncedTxCount
	c.JSON(http.StatusOK, SyncGmailResponse{
		Message: fmt.Sprintf("Processed %d CDSL CAS statements: extracted %d monthly trades with exact execution dates, reconciled %d portfolio holdings, and synced %d subsequent Demat transactions",
			len(casItems), totalTradesExtracted, totalHoldingsExtracted, syncedTxCount),
		SyncMode:               mode,
		Synced:                 totalSynced,
		CASProcessed:           true,
		PortfolioEffectiveDate: &latestStatementDate,
		LastGmailSync:          now,
	})
}

// doIncrementalSync processes only new Demat transaction emails since last sync.
func (h *GmailHandler) doIncrementalSync(c *gin.Context, ctx context.Context, gmailService *gmail.Service, user *models.User, userID string) {
	syncAfter := user.LastGmailSync
	if user.CASEffectiveDate != nil && (syncAfter == nil || user.CASEffectiveDate.After(*syncAfter)) {
		syncAfter = user.CASEffectiveDate
	}

	emails, err := h.fetchCDSLEmails(gmailService, userID, syncAfter)
	if err != nil {
		log.Printf("doIncrementalSync: Failed to fetch emails: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch emails"})
		return
	}

	now := time.Now()
	syncedCount := 0
	for _, email := range emails {
		investments, transactions, err := h.parseEmailContent(ctx, email)
		if err != nil || len(investments) == 0 {
			continue
		}
		if err := h.saveInvestmentData(ctx, userID, investments, transactions); err == nil {
			syncedCount++
		}
	}

	_, _ = h.firestoreClient.Collection("users").Doc(userID).Update(ctx, []firestore.Update{
		{Path: "lastGmailSync", Value: now},
		{Path: "updatedAt", Value: now},
	})

	c.JSON(http.StatusOK, SyncGmailResponse{
		Message:       fmt.Sprintf("Incremental sync completed. Processed %d new Demat transactions.", syncedCount),
		SyncMode:      "incremental",
		Synced:        syncedCount,
		Total:         len(emails),
		LastGmailSync: now,
	})
}

// doPriceRefresh updates investment values and profit/loss directly from database and live prices.
func (h *GmailHandler) doPriceRefresh(c *gin.Context, ctx context.Context, user *models.User, userID string) {
	iter := h.firestoreClient.Collection("users").Doc(userID).Collection("investments").Where("isActive", "==", true).Documents(ctx)
	todayStr := time.Now().Format("2006-01-02")
	batch := h.firestoreClient.Batch()
	updatedCount := 0

	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}
		var inv models.Investment
		if err := doc.DataTo(&inv); err != nil {
			continue
		}

		var currentPrice float64
		if inv.Category == models.CategoryStocks {
			ticker := inv.InvestmentData.Ticker
			if ticker == "" {
				ticker = inv.Name
			}
			currentPrice = h.fetchStockPrice(ctx, ticker, todayStr)
			if currentPrice > 0 {
				inv.InvestmentData.CurrentPrice = currentPrice
				inv.CurrentValue = inv.InvestmentData.Quantity * currentPrice
				inv.UpdatedAt = time.Now()
				batch.Set(doc.Ref, inv)
				updatedCount++
			}
		} else if inv.Category == models.CategoryMutualFunds {
			ticker := inv.InvestmentData.Ticker
			currentPrice = h.fetchMutualFundNAV(ctx, inv.Name, ticker, todayStr)
			if currentPrice == 0 && ticker != "" {
				currentPrice = h.fetchStockPrice(ctx, ticker, todayStr)
			}
			if currentPrice > 0 {
				inv.InvestmentData.CurrentPrice = currentPrice
				inv.CurrentValue = inv.InvestmentData.Units * currentPrice
				inv.UpdatedAt = time.Now()
				batch.Set(doc.Ref, inv)
				updatedCount++
			}
		}
	}

	if updatedCount > 0 {
		if _, err := batch.Commit(ctx); err != nil {
			log.Printf("doPriceRefresh: Batch commit failed: %v", err)
		}
	}

	now := time.Now()
	_, _ = h.firestoreClient.Collection("users").Doc(userID).Update(ctx, []firestore.Update{
		{Path: "lastGmailSync", Value: now},
		{Path: "updatedAt", Value: now},
	})

	c.JSON(http.StatusOK, SyncGmailResponse{
		Message:       fmt.Sprintf("Portfolio refreshed directly from market data (%d holdings updated).", updatedCount),
		SyncMode:      "price_refresh",
		Synced:        updatedCount,
		LastGmailSync: now,
	})
}

// saveCASTransactions saves trades executed in that month (Section 2) with their EXACT execution dates.
func (h *GmailHandler) saveCASTransactions(ctx context.Context, userID string,
	trades []caspdf.CASTransaction, seenTx map[string]bool) (int, error) {

	batch := h.firestoreClient.Batch()
	savedCount := 0

	for _, trade := range trades {
		if trade.Quantity <= 0 {
			continue
		}

		isMF := strings.EqualFold(trade.Category, "mutualFunds") ||
			strings.HasPrefix(strings.ToUpper(trade.ISIN), "INF") ||
			strings.Contains(strings.ToUpper(trade.Name), "ETF")

		category := models.CategoryStocks
		if isMF {
			category = models.CategoryMutualFunds
		}

		name := cleanInvestmentName(trade.Name, isMF)
		if name == "" {
			name = trade.Name
		}

		txType := models.TransactionTypeBuy
		if strings.EqualFold(trade.Type, "sell") {
			txType = models.TransactionTypeSell
		}

		// Exact trade execution date from Section 2
		tradeDate := mustParseDate(trade.TransactionDate)
		if tradeDate.IsZero() {
			log.Printf("Warning: skipping CAS trade %q with missing/invalid date %q", name, trade.TransactionDate)
			continue
		}

		effectiveTradeDate := tradedate.ResolveTradeDate(tradeDate)
		tradeDateStr := effectiveTradeDate.Format("2006-01-02")

		var invTicker string
		var tradePrice float64

		if isMF {
			if trade.Ticker != "" && !strings.HasPrefix(strings.ToUpper(trade.Ticker), "INF") {
				invTicker = trade.Ticker
			} else {
				invTicker = generateMFTicker(name)
			}
			if trade.Price > 0 {
				tradePrice = trade.Price
			} else {
				tradePrice = h.fetchMutualFundNAV(ctx, name, trade.ISIN, tradeDateStr)
				if tradePrice == 0 && invTicker != "" {
					tradePrice = h.fetchStockPrice(ctx, invTicker, tradeDateStr)
				}
			}
		} else {
			stockTicker := trade.Ticker
			if stockTicker == "" {
				stockTicker = trade.ISIN
			}
			invTicker = stockTicker
			if trade.Price > 0 {
				tradePrice = trade.Price
			} else {
				tradePrice = h.fetchStockPrice(ctx, stockTicker, tradeDateStr)
			}
		}

		tradeAmount := trade.Quantity * tradePrice

		// Deduplicate transaction to avoid inserting the same trade twice across monthly statements
		dedupKey := fmt.Sprintf("%s_%s_%s_%.4f_%s", category, invTicker, tradeDateStr, trade.Quantity, txType)
		if seenTx[dedupKey] {
			continue
		}
		seenTx[dedupKey] = true

		// Find or create investment doc to obtain relatedID
		var q firestore.Query
		if invTicker != "" {
			q = h.firestoreClient.Collection("users").Doc(userID).Collection("investments").
				Where("category", "==", category).
				Where("investmentData.ticker", "==", invTicker).
				Limit(1)
		} else {
			q = h.firestoreClient.Collection("users").Doc(userID).Collection("investments").
				Where("category", "==", category).
				Where("name", "==", name).
				Limit(1)
		}

		relatedID := ""
		docs, err := q.Documents(ctx).GetAll()
		if err == nil && len(docs) > 0 {
			relatedID = docs[0].Ref.ID
		} else {
			newRef := h.firestoreClient.Collection("users").Doc(userID).Collection("investments").NewDoc()
			var invData models.InvestmentData
			if isMF {
				invData = models.InvestmentData{FundName: name, Ticker: invTicker, Units: trade.Quantity, CurrentPrice: tradePrice}
			} else {
				invData = models.InvestmentData{Ticker: invTicker, Quantity: trade.Quantity, CurrentPrice: tradePrice}
			}
			newInv := models.Investment{
				Category:       category,
				Name:           name,
				InvestedAmount: tradeAmount,
				CurrentValue:   tradeAmount,
				CreatedAt:      effectiveTradeDate,
				UpdatedAt:      time.Now(),
				IsActive:       true,
				InvestmentData: invData,
			}
			batch.Set(newRef, newInv)
			relatedID = newRef.ID
		}

		desc := trade.Description
		if desc == "" {
			desc = fmt.Sprintf("%s (%s) %s trade from CDSL CAS", name, invTicker, strings.ToUpper(trade.Type))
		}

		txRef := h.firestoreClient.Collection("users").Doc(userID).Collection("transactions").NewDoc()
		tx := models.Transaction{
			RelatedID:   relatedID,
			Category:    category,
			Type:        txType,
			Quantity:    trade.Quantity,
			Price:       tradePrice,
			Amount:      tradeAmount,
			Date:        effectiveTradeDate,
			Description: desc,
			CreatedAt:   time.Now(),
		}
		batch.Set(txRef, tx)
		savedCount++
	}

	if savedCount > 0 {
		_, err := batch.Commit(ctx)
		return savedCount, err
	}
	return 0, nil
}

// reconcileCASHoldings reconciles current holdings (Section 1) from the latest CAS statement.
// Crucially, it does NOT create fake transactions with statement dates for assets that already have trade history.
func (h *GmailHandler) reconcileCASHoldings(ctx context.Context, userID string,
	holdings []caspdf.CASHolding, statementDate time.Time, seenTx map[string]bool) (int, error) {

	if len(holdings) == 0 {
		return 0, nil
	}

	batch := h.firestoreClient.Batch()
	todayStr := time.Now().Format("2006-01-02")
	stmtDateStr := statementDate.Format("2006-01-02")
	reconciledCount := 0

	for _, holding := range holdings {
		isMF := strings.EqualFold(holding.Category, "mutualFunds") ||
			strings.HasPrefix(strings.ToUpper(holding.ISIN), "INF") ||
			strings.Contains(strings.ToUpper(holding.Name), "ETF")

		category := models.CategoryStocks
		if isMF {
			category = models.CategoryMutualFunds
		}

		name := cleanInvestmentName(holding.Name, isMF)
		if name == "" {
			name = holding.Name
		}

		var invTicker string
		var txPrice float64
		var currentPrice float64

		if isMF {
			if holding.Ticker != "" && !strings.HasPrefix(strings.ToUpper(holding.Ticker), "INF") {
				invTicker = holding.Ticker
			} else {
				invTicker = generateMFTicker(name)
			}
			if holding.Price > 0 {
				txPrice = holding.Price
			} else {
				txPrice = h.fetchMutualFundNAV(ctx, name, holding.ISIN, stmtDateStr)
				if txPrice == 0 && invTicker != "" {
					txPrice = h.fetchStockPrice(ctx, invTicker, stmtDateStr)
				}
			}
			currentPrice = h.fetchMutualFundNAV(ctx, name, holding.ISIN, todayStr)
			if currentPrice == 0 && invTicker != "" {
				currentPrice = h.fetchStockPrice(ctx, invTicker, todayStr)
			}
			if currentPrice == 0 {
				currentPrice = txPrice
			}
			if txPrice == 0 {
				txPrice = currentPrice
			}
		} else {
			stockTicker := holding.Ticker
			if stockTicker == "" {
				stockTicker = holding.ISIN
			}
			invTicker = stockTicker
			if holding.Price > 0 {
				txPrice = holding.Price
			} else {
				txPrice = h.fetchStockPrice(ctx, stockTicker, stmtDateStr)
			}
			currentPrice = h.fetchStockPrice(ctx, stockTicker, todayStr)
			if currentPrice == 0 {
				currentPrice = txPrice
			}
			if txPrice == 0 {
				txPrice = currentPrice
			}
		}

		investedAmount := holding.InvestedAmount
		if investedAmount <= 0 {
			investedAmount = holding.Quantity * txPrice
		}
		currentVal := holding.Quantity * currentPrice

		var invData models.InvestmentData
		if isMF {
			invData = models.InvestmentData{
				FundName:     name,
				Units:        holding.Quantity,
				Ticker:       invTicker,
				NAV:          txPrice,
				CurrentPrice: currentPrice,
			}
		} else {
			invData = models.InvestmentData{
				Ticker:       invTicker,
				Quantity:     holding.Quantity,
				AveragePrice: txPrice,
				CurrentPrice: currentPrice,
			}
		}

		var query firestore.Query
		if invTicker != "" {
			query = h.firestoreClient.Collection("users").Doc(userID).Collection("investments").
				Where("category", "==", category).
				Where("investmentData.ticker", "==", invTicker).
				Limit(1)
		} else {
			query = h.firestoreClient.Collection("users").Doc(userID).Collection("investments").
				Where("category", "==", category).
				Where("name", "==", name).
				Limit(1)
		}

		docs, err := query.Documents(ctx).GetAll()
		var targetRef *firestore.DocumentRef
		isNew := true

		if err == nil && len(docs) > 0 {
			targetRef = docs[0].Ref
			isNew = false
			var existing models.Investment
			_ = docs[0].DataTo(&existing)
			existing.InvestmentData = invData
			existing.InvestedAmount = investedAmount
			existing.CurrentValue = currentVal
			existing.UpdatedAt = time.Now()
			batch.Set(targetRef, existing)
		} else {
			targetRef = h.firestoreClient.Collection("users").Doc(userID).Collection("investments").NewDoc()
			newInv := models.Investment{
				Category:       category,
				Name:           name,
				InvestedAmount: investedAmount,
				CurrentValue:   currentVal,
				CreatedAt:      statementDate,
				UpdatedAt:      time.Now(),
				IsActive:       true,
				InvestmentData: invData,
			}
			batch.Set(targetRef, newInv)
		}
		reconciledCount++

		// Only if this investment has NO recorded transactions from Section 2 or Demat emails:
		// Create an initial opening balance transaction so the holding has cost basis.
		// If it already has transactions, DO NOT CREATE A DUMMY TRANSACTION with statement date!
		if isNew {
			txDocs, _ := h.firestoreClient.Collection("users").Doc(userID).Collection("transactions").
				Where("relatedId", "==", targetRef.ID).Limit(1).Documents(ctx).GetAll()
			if len(txDocs) == 0 {
				txRef := h.firestoreClient.Collection("users").Doc(userID).Collection("transactions").NewDoc()
				openingTx := models.Transaction{
					RelatedID:   targetRef.ID,
					Category:    category,
					Type:        models.TransactionTypeBuy,
					Quantity:    holding.Quantity,
					Price:       txPrice,
					Amount:      investedAmount,
					Date:        tradedate.ResolveTradeDate(statementDate),
					Description: fmt.Sprintf("%s (%s) Initial Holding Balance as of %s", name, invTicker, statementDate.Format("02 Jan 2006")),
					CreatedAt:   time.Now(),
				}
				batch.Set(txRef, openingTx)
			}
		}
	}

	if reconciledCount > 0 {
		_, err := batch.Commit(ctx)
		return reconciledCount, err
	}
	return 0, nil
}

// ---------------------------------------------------------------------------
// Gmail connection and fetching helpers
// ---------------------------------------------------------------------------

// connectToGmail establishes connection to Gmail API using OAuth token.
func (h *GmailHandler) connectToGmail(ctx context.Context, accessToken string) (*gmail.Service, error) {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})
	client := oauth2.NewClient(ctx, ts)
	gmailService, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to create Gmail service: %v", err)
	}
	return gmailService, nil
}

// fetchCDSLEmails fetches emails from CDSL India after a specific timestamp.
func (h *GmailHandler) fetchCDSLEmails(gmailService *gmail.Service, userID string, lastSync *time.Time) ([]*gmail.Message, error) {
	query := fmt.Sprintf("subject:\"%s\"", constants.CDSLEmailSubject)
	if lastSync != nil && !lastSync.IsZero() {
		query = fmt.Sprintf("%s after:%d", query, lastSync.Unix())
		log.Printf("Syncing emails after %v (Unix: %d)", lastSync, lastSync.Unix())
	}

	listCall := gmailService.Users.Messages.List("me").Q(query).MaxResults(int64(constants.MaxEmailsToFetch))
	response, err := listCall.Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve messages: %v", err)
	}

	if len(response.Messages) == 0 {
		return []*gmail.Message{}, nil
	}

	// Reverse to process oldest first
	slices.Reverse(response.Messages)

	var emails []*gmail.Message
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

// ---------------------------------------------------------------------------
// Email body extraction
// ---------------------------------------------------------------------------

// extractEmailBody decodes and returns the raw body of a Gmail message.
// Prefers text/html parts. Returns empty string if nothing is found.
func (h *GmailHandler) extractEmailBody(email *gmail.Message) string {
	if email.Payload.Body.Data != "" {
		data, err := base64.URLEncoding.DecodeString(email.Payload.Body.Data)
		if err == nil {
			return string(data)
		}
	}
	for _, part := range email.Payload.Parts {
		if part.MimeType == "text/html" && part.Body.Data != "" {
			data, err := base64.URLEncoding.DecodeString(part.Body.Data)
			if err == nil {
				return string(data)
			}
		}
	}
	// Try plain-text part as last resort
	for _, part := range email.Payload.Parts {
		if part.MimeType == "text/plain" && part.Body.Data != "" {
			data, err := base64.URLEncoding.DecodeString(part.Body.Data)
			if err == nil {
				return string(data)
			}
		}
	}
	return ""
}

var (
	whitespaceRegex        = regexp.MustCompile(`[ \t]+`)
	multiNewlineRegex      = regexp.MustCompile(`(\n\s*){3,}`)
	mfTickerCleanRegex     = regexp.MustCompile(`[^A-Z0-9]+`)
	mfPrefixRegex          = regexp.MustCompile(`(?i)^[^-]+(?:MF|MUTUAL\s+FUND)\s*-\s*`)
	mfPlanSuffixRegex      = regexp.MustCompile(`(?i)(?:\s*[-/]\s*|\s+)\b(?:DIRECT|REGULAR|PLAN|PL|OPT|OPTION)\b.*$`)
	mfGrowthSuffixRegex    = regexp.MustCompile(`(?i)\s*[-/]\s*GROWTH\b.*$`)
	stockEquitySuffixRegex = regexp.MustCompile(`(?i)\s*-\s*(?:NEW\s+)?EQUITY(?:\s+SHARES)?.*$`)
)

// stripHTML converts HTML content to plain text by walking the parse tree.
func (h *GmailHandler) stripHTML(body string) string {
	doc, err := html.Parse(strings.NewReader(body))
	if err != nil {
		return strings.TrimSpace(body)
	}
	var buf strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			buf.WriteString(n.Data)
			buf.WriteString(" ")
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	text := buf.String()
	// Collapse whitespace
	text = whitespaceRegex.ReplaceAllString(text, " ")
	text = multiNewlineRegex.ReplaceAllString(text, "\n\n")
	// Decode common HTML entities
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&#39;", "'")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	return strings.TrimSpace(text)
}

// ---------------------------------------------------------------------------
// Gemini Flash AI extraction
// ---------------------------------------------------------------------------

//go:embed prompts/gemini_cdsl_system_instruction.txt
var geminiSystemInstruction string

// callGeminiForTransactions sends the email text to Gemini Flash and returns
// the structured list of extracted transactions.
func (h *GmailHandler) callGeminiForTransactions(ctx context.Context, emailText string) ([]AIAssistantTransaction, error) {
	if h.geminiAPIKey == "" {
		log.Printf("callGeminiForTransactions: GEMINI_API_KEY is not configured")
		return nil, fmt.Errorf("GEMINI_API_KEY is not configured")
	}

	log.Printf("callGeminiForTransactions: starting extraction for email text length=%d", len(emailText))

	schema := geminiResponseSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"transactions": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"category":         map[string]string{"type": "string"},
						"name":             map[string]string{"type": "string"},
						"isin":             map[string]string{"type": "string"},
						"ticker":           map[string]string{"type": "string"},
						"quantity":         map[string]string{"type": "number"},
						"type":             map[string]string{"type": "string"},
						"transaction_date": map[string]string{"type": "string"},
					},
					"required": []string{"category", "name", "ticker", "quantity", "type", "transaction_date"},
				},
			},
		},
		Required: []string{"transactions"},
	}

	reqBody := geminiRequest{
		SystemInstruction: geminiContent{
			Parts: []geminiPart{{Text: geminiSystemInstruction}},
		},
		Contents: []geminiContent{
			{
				Role:  "user",
				Parts: []geminiPart{{Text: emailText}},
			},
		},
		GenerationConfig: geminiGenerationConfig{
			ResponseMIMEType: "application/json",
			ResponseSchema:   schema,
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("callGeminiForTransactions: failed to marshal Gemini request: %v", err)
		return nil, fmt.Errorf("failed to marshal Gemini request: %v", err)
	}

	log.Printf("callGeminiForTransactions: sending request to Gemini endpoint with payload size=%d", len(bodyBytes))

	endpoint := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/gemini-3.5-flash-lite:generateContent?key=%s",
		h.geminiAPIKey,
	)

	respBytes, err := gemini.PostWithRetry(ctx, endpoint, bodyBytes)
	if err != nil {
		log.Printf("callGeminiForTransactions: Gemini call failed after retries: %v", err)
		return nil, fmt.Errorf("Gemini call failed: %w", err)
	}

	var gemResp geminiResponse
	if err := json.Unmarshal(respBytes, &gemResp); err != nil {
		log.Printf("callGeminiForTransactions: failed to unmarshal Gemini response envelope: %v", err)
		return nil, fmt.Errorf("failed to unmarshal Gemini response envelope: %v", err)
	}

	if len(gemResp.Candidates) == 0 || len(gemResp.Candidates[0].Content.Parts) == 0 {
		log.Printf("callGeminiForTransactions: Gemini returned no candidates")
		return nil, fmt.Errorf("Gemini returned no candidates")
	}

	rawJSON := gemResp.Candidates[0].Content.Parts[0].Text
	log.Printf("callGeminiForTransactions: parsed Gemini response text length=%d", len(rawJSON))

	var extraction AIExtractionResponse
	if err := json.Unmarshal([]byte(rawJSON), &extraction); err != nil {
		log.Printf("callGeminiForTransactions: failed to unmarshal AI extraction JSON: %v raw_json=%q", err, rawJSON)
		return nil, fmt.Errorf("failed to unmarshal AI extraction JSON (%q): %v", rawJSON, err)
	}

	log.Printf("callGeminiForTransactions: extracted %d transactions", len(extraction.Transactions))
	return extraction.Transactions, nil
}

// ---------------------------------------------------------------------------
// Email parsing (AI pipeline)
// ---------------------------------------------------------------------------

// parseEmailContent extracts investment and transaction records from a Gmail
// message using the Gemini Flash structured extraction pipeline.
func (h *GmailHandler) parseEmailContent(ctx context.Context, email *gmail.Message) ([]models.Investment, []models.Transaction, error) {
	body := h.extractEmailBody(email)
	if body == "" {
		return nil, nil, fmt.Errorf("empty email body for message %s", email.Id)
	}

	emailText := h.stripHTML(body)

	// Get email date for fallback
	emailDate := time.Now()
	if email.InternalDate != 0 {
		emailDate = time.Unix(email.InternalDate/1000, 0)
	}

	aiTxns, err := h.callGeminiForTransactions(ctx, emailText)
	if err != nil {
		return nil, nil, fmt.Errorf("Gemini extraction failed for email %s: %v", email.Id, err)
	}

	log.Printf("AI extracted %d transactions from email %s", len(aiTxns), email.Id)

	var investments []models.Investment
	var transactions []models.Transaction

	for _, at := range aiTxns {
		if (at.Ticker == "" && at.ISIN == "") || at.Quantity <= 0 {
			log.Printf("Skipping invalid AI transaction: %+v", at)
			continue
		}

		isMF := strings.HasPrefix(strings.ToUpper(at.ISIN), "INF") ||
			strings.HasPrefix(strings.ToUpper(at.Ticker), "INF") ||
			strings.Contains(strings.ToUpper(at.Name), "ETF") ||
			strings.Contains(strings.ToUpper(at.Ticker), "ETF") ||
			strings.EqualFold(at.Category, "mutualFunds")

		// Clean name
		name := cleanInvestmentName(at.Name, isMF)
		if name == "" {
			name = at.Name
		}

		// Parse transaction notification date
		notificationDate := emailDate
		if at.TransactionDate != "" {
			if parsed, err := time.Parse("2006-01-02", at.TransactionDate); err == nil {
				notificationDate = parsed
			}
		}

		// Requirement 6: Check for previous working day considering weekends & trade holidays
		effectiveTradeDate := tradedate.ResolveTradeDate(notificationDate)
		txDateStr := effectiveTradeDate.Format("2006-01-02")

		// Determine category and transaction type
		category := models.CategoryStocks
		if isMF {
			category = models.CategoryMutualFunds
		}

		txType := models.TransactionTypeBuy
		if strings.EqualFold(at.Type, "sell") {
			txType = models.TransactionTypeSell
		}

		todayStr := time.Now().Format("2006-01-02")
		var txPrice float64
		var currentPrice float64
		var invTicker string

		if isMF {
			// For Mutual Funds / ETFs, assign a real ticker symbol (e.g. PPFAS, MIRAE_ASSET_GOLD_ETF)
			if at.Ticker != "" && !strings.HasPrefix(strings.ToUpper(at.Ticker), "INF") {
				invTicker = at.Ticker
			} else {
				invTicker = generateMFTicker(name)
			}

			// Price (NAV) on transaction date
			txPrice = h.fetchMutualFundNAV(ctx, name, at.ISIN, txDateStr)
			if txPrice == 0 && invTicker != "" {
				txPrice = h.fetchStockPrice(ctx, invTicker, txDateStr)
			}

			// Price (NAV) on current date
			currentPrice = h.fetchMutualFundNAV(ctx, name, at.ISIN, todayStr)
			if currentPrice == 0 && invTicker != "" {
				currentPrice = h.fetchStockPrice(ctx, invTicker, todayStr)
			}

			if currentPrice == 0 {
				currentPrice = txPrice
			}
			if txPrice == 0 {
				txPrice = currentPrice
			}
			log.Printf("MF NAV for %s (%s) - TxDate (%s): %.4f, CurrentDate (%s): %.4f", name, invTicker, txDateStr, txPrice, todayStr, currentPrice)
		} else {
			// For Stocks, fetch stock price from Yahoo Finance / Alpha Vantage using Ticker symbol
			stockTicker := at.Ticker
			if stockTicker == "" {
				stockTicker = at.ISIN
			}
			invTicker = stockTicker
			// Price on transaction date
			txPrice = h.fetchStockPrice(ctx, stockTicker, txDateStr)
			// Price on current date
			currentPrice = h.fetchStockPrice(ctx, stockTicker, todayStr)
			if currentPrice == 0 {
				currentPrice = txPrice
			}
			if txPrice == 0 {
				txPrice = currentPrice
			}
			log.Printf("Stock price for %s - TxDate (%s): %.4f, CurrentDate (%s): %.4f", stockTicker, txDateStr, txPrice, todayStr, currentPrice)
		}

		// Calculate invested amount on transaction date & current holding value
		investedAmountOnTxDate := at.Quantity * txPrice
		currentVal := at.Quantity * currentPrice

		// Build investment record (financial math applied in saveInvestmentData)
		var invData models.InvestmentData
		if isMF {
			invData = models.InvestmentData{
				FundName:     name,
				Units:        at.Quantity,
				Ticker:       invTicker,
				NAV:          txPrice,
				CurrentPrice: currentPrice,
			}
		} else {
			invData = models.InvestmentData{
				Ticker:       invTicker,
				Quantity:     at.Quantity,
				AveragePrice: txPrice,
				CurrentPrice: currentPrice,
			}
		}

		investment := models.Investment{
			Category:       category,
			Name:           name,
			InvestedAmount: investedAmountOnTxDate,
			CurrentValue:   currentVal,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			IsActive:       true,
			InvestmentData: invData,
		}
		investments = append(investments, investment)

		// Build transaction record
		transaction := models.Transaction{
			Category:    category,
			Type:        txType,
			Quantity:    at.Quantity,
			Price:       txPrice,
			Amount:      investedAmountOnTxDate,
			Date:        effectiveTradeDate,
			Description: fmt.Sprintf("%s (%s) %s transaction from CDSL", name, invTicker, strings.ToUpper(at.Type)),
			CreatedAt:   time.Now(),
		}
		transactions = append(transactions, transaction)

		log.Printf("Parsed: name=%s isin=%s ticker=%s qty=%.4f type=%s txDate=%s txPrice=%.4f invested=%.2f currentPrice=%.4f currentValue=%.2f",
			name, at.ISIN, at.Ticker, at.Quantity, at.Type, txDateStr, txPrice, investedAmountOnTxDate, currentPrice, currentVal)
	}

	return investments, transactions, nil
}

// ---------------------------------------------------------------------------
// Market price fetchers
// ---------------------------------------------------------------------------

// fetchStockPrice fetches the closing price for the given ticker symbol on the given date (YYYY-MM-DD).
// It queries Yahoo Finance (.NS / .BO) first for free full historical data, with fallback to Alpha Vantage.
func (h *GmailHandler) fetchStockPrice(ctx context.Context, ticker, date string) float64 {
	base := strings.TrimSpace(ticker)
	base = strings.TrimSuffix(base, ".BSE")
	base = strings.TrimSuffix(base, ".NSE")
	base = strings.TrimSuffix(base, ".BO")
	base = strings.TrimSuffix(base, ".NS")

	// 1. Try Yahoo Finance with NSE (.NS)
	if price := h.fetchStockPriceYahoo(ctx, base+".NS", date); price > 0 {
		log.Printf("fetchStockPrice: found price %.4f for %s on %s via Yahoo Finance (.NS)", price, base, date)
		return price
	}

	// 2. Try Yahoo Finance with BSE (.BO)
	if price := h.fetchStockPriceYahoo(ctx, base+".BO", date); price > 0 {
		log.Printf("fetchStockPrice: found price %.4f for %s on %s via Yahoo Finance (.BO)", price, base, date)
		return price
	}

	// 3. Fallback to Alpha Vantage (if configured)
	if h.alphaVantageAPIKey != "" {
		if price := h.fetchStockPriceBySymbol(ctx, ticker, date); price > 0 {
			return price
		}
		if !strings.Contains(ticker, ".") {
			if fallbackPrice := h.fetchStockPriceBySymbol(ctx, ticker+".BSE", date); fallbackPrice > 0 {
				return fallbackPrice
			}
		}
	}

	return 0
}

// fetchStockPriceYahoo fetches historical or current stock prices from Yahoo Finance.
// It supports NSE (.NS) and BSE (.BO) tickers with full historical date range for free.
func (h *GmailHandler) fetchStockPriceYahoo(ctx context.Context, symbol, date string) float64 {
	target := mustParseDate(date)
	isRecent := time.Since(target).Hours() < 7*24

	p1 := target.AddDate(0, 0, -7).Unix()
	p2 := target.AddDate(0, 0, 2).Unix()

	endpoint := fmt.Sprintf(
		"https://query1.finance.yahoo.com/v8/finance/chart/%s?period1=%d&period2=%d&interval=1d",
		url.QueryEscape(symbol), p1, p2,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		log.Printf("fetchStockPriceYahoo: failed to build request for %s: %v", symbol, err)
		return 0
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("fetchStockPriceYahoo: HTTP error for %s: %v", symbol, err)
		return 0
	}
	defer resp.Body.Close()

	var yfResp yfChartResponse
	if err := json.NewDecoder(resp.Body).Decode(&yfResp); err != nil {
		log.Printf("fetchStockPriceYahoo: decode error for %s: %v", symbol, err)
		return 0
	}

	if len(yfResp.Chart.Result) == 0 {
		return 0
	}

	res := yfResp.Chart.Result[0]
	if len(res.Indicators.Quote) == 0 {
		return 0
	}

	closes := res.Indicators.Quote[0].Close
	timestamps := res.Timestamp

	// If looking for current/recent price and regularMarketPrice is present
	if isRecent && res.Meta.RegularMarketPrice > 0 {
		return res.Meta.RegularMarketPrice
	}

	// Match exact or closest prior business day on or before target date
	targetDateOnly := target.Format("2006-01-02")
	var bestPrice float64
	bestDiff := 999999

	for i, ts := range timestamps {
		if i >= len(closes) || closes[i] == nil {
			continue
		}
		tDate := time.Unix(ts, 0)
		tDateStr := tDate.Format("2006-01-02")
		if tDateStr == targetDateOnly {
			return res.Indicators.Quote[0].getDayPrice(i)
		}
		diff := int(target.Sub(tDate).Hours() / 24)
		if diff >= 0 && diff < bestDiff {
			bestDiff = diff
			bestPrice = res.Indicators.Quote[0].getDayPrice(i)
		}
	}

	if bestPrice > 0 {
		return bestPrice
	}

	// Fallback to regularMarketPrice if recent
	if isRecent && res.Meta.RegularMarketPrice > 0 {
		return res.Meta.RegularMarketPrice
	}

	return 0
}

func (h *GmailHandler) fetchStockPriceBySymbol(ctx context.Context, symbol, date string) float64 {
	target := mustParseDate(date)
	daysAgo := time.Since(target).Hours() / 24

	// If transaction date is older than 60 days, request full history immediately.
	outputSize := "compact"
	if daysAgo > 60 {
		outputSize = "full"
	}

	price := h.queryAlphaVantageDaily(ctx, symbol, date, outputSize)
	if price > 0 {
		return price
	}

	// If not found with compact, retry with full history
	if outputSize == "compact" {
		log.Printf("fetchStockPriceBySymbol: %s on %s not in compact window, retrying with outputsize=full", symbol, date)
		return h.queryAlphaVantageDaily(ctx, symbol, date, "full")
	}

	return 0
}

func (h *GmailHandler) queryAlphaVantageDaily(ctx context.Context, symbol, date, outputSize string) float64 {
	endpoint := fmt.Sprintf(
		"https://www.alphavantage.co/query?function=TIME_SERIES_DAILY&symbol=%s&outputsize=%s&apikey=%s",
		url.QueryEscape(symbol), outputSize, h.alphaVantageAPIKey,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		log.Printf("queryAlphaVantageDaily: failed to build request for %s: %v", symbol, err)
		return 0
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("queryAlphaVantageDaily: HTTP error for %s: %v", symbol, err)
		return 0
	}
	defer resp.Body.Close()

	var avResp avDailyResponse
	if err := json.NewDecoder(resp.Body).Decode(&avResp); err != nil {
		log.Printf("queryAlphaVantageDaily: decode error for %s: %v", symbol, err)
		return 0
	}

	target := mustParseDate(date)
	// Try exact date first, then walk back up to 10 trading days for weekends/holidays
	for i := 0; i < 10; i++ {
		d := target.AddDate(0, 0, -i).Format("2006-01-02")
		if day, ok := avResp.TimeSeries[d]; ok {
			var price float64
			fmt.Sscanf(day.Close, "%f", &price)
			if price > 0 {
				return price
			}
		}
	}

	// Fallback to LastRefreshed date ONLY if target date is recent (within last 7 days)
	// Do NOT use LastRefreshed for historical transactions from months or years ago!
	if time.Since(target).Hours() < 7*24 && avResp.MetaData.LastRefreshed != "" {
		if day, ok := avResp.TimeSeries[avResp.MetaData.LastRefreshed]; ok {
			var price float64
			fmt.Sscanf(day.Close, "%f", &price)
			if price > 0 {
				return price
			}
		}
	}

	log.Printf("queryAlphaVantageDaily: no AV data for %s near %s (outputsize=%s)", symbol, date, outputSize)
	return 0
}

// fetchMutualFundNAV fetches the historical or current NAV for an Indian mutual fund
// from mfapi.in by searching across scheme names, keywords, and ISIN.
func (h *GmailHandler) fetchMutualFundNAV(ctx context.Context, fundName, isin, date string) float64 {
	target := mustParseDate(date)
	isRecent := time.Since(target).Hours() < 7*24

	// Build search queries: cleaned name, shortened variations, and ISIN
	queries := []string{}
	cleaned := cleanInvestmentName(fundName, true)
	if cleaned != "" {
		queries = append(queries, cleaned)
		short := strings.TrimSpace(strings.ReplaceAll(cleaned, " FUND", ""))
		if short != "" && short != cleaned {
			queries = append(queries, short)
		}
		shortETF := strings.TrimSpace(strings.ReplaceAll(cleaned, " ETF", ""))
		if shortETF != "" && shortETF != cleaned && shortETF != short {
			queries = append(queries, shortETF)
		}
	}
	if fundName != "" && fundName != cleaned {
		queries = append(queries, fundName)
	}
	if isin != "" {
		queries = append(queries, isin)
	}

	// Try querying mfapi search with each candidate query
	var matchedSchemes []mfapiSearchResult
	seen := make(map[int]bool)

	for _, q := range queries {
		if q == "" {
			continue
		}
		searchURL := fmt.Sprintf("https://api.mfapi.in/mf/search?q=%s", url.QueryEscape(q))
		searchReq, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
		if err != nil {
			continue
		}
		searchReq.Header.Set("User-Agent", "Mozilla/5.0")

		searchResp, err := http.DefaultClient.Do(searchReq)
		if err != nil {
			continue
		}

		var results []mfapiSearchResult
		if err := json.NewDecoder(searchResp.Body).Decode(&results); err == nil {
			for _, r := range results {
				if !seen[r.SchemeCode] {
					seen[r.SchemeCode] = true
					matchedSchemes = append(matchedSchemes, r)
				}
			}
		}
		searchResp.Body.Close()

		if len(matchedSchemes) >= 5 {
			break
		}
	}

	if len(matchedSchemes) == 0 {
		log.Printf("fetchMutualFundNAV: no search results on mfapi for fund %q / isin %s", fundName, isin)
		return 0
	}

	// Check matched schemes for NAV data
	for _, schemeMeta := range matchedSchemes {
		navURL := fmt.Sprintf("https://api.mfapi.in/mf/%d", schemeMeta.SchemeCode)
		navReq, err := http.NewRequestWithContext(ctx, http.MethodGet, navURL, nil)
		if err != nil {
			continue
		}
		navReq.Header.Set("User-Agent", "Mozilla/5.0")

		navResp, err := http.DefaultClient.Do(navReq)
		if err != nil {
			continue
		}

		var scheme mfapiSchemeResponse
		err = json.NewDecoder(navResp.Body).Decode(&scheme)
		navResp.Body.Close()
		if err != nil || len(scheme.Data) == 0 {
			continue
		}

		// If looking for current/recent date and latest NAV is available
		if isRecent {
			var nav float64
			fmt.Sscanf(scheme.Data[0].Nav, "%f", &nav)
			if nav > 0 {
				log.Printf("fetchMutualFundNAV: found current NAV %.4f for %s (schemeCode=%d, %s)", nav, fundName, schemeMeta.SchemeCode, schemeMeta.SchemeName)
				return nav
			}
		}

		// For historical date, search through date entries within 10 days
		for i := 0; i < 10; i++ {
			candidate := target.AddDate(0, 0, -i).Format("02-01-2006")
			for _, entry := range scheme.Data {
				if entry.Date == candidate {
					var nav float64
					fmt.Sscanf(entry.Nav, "%f", &nav)
					if nav > 0 {
						log.Printf("fetchMutualFundNAV: found historical NAV %.4f on %s for %s (schemeCode=%d, %s)", nav, candidate, fundName, schemeMeta.SchemeCode, schemeMeta.SchemeName)
						return nav
					}
				}
			}
		}
	}

	log.Printf("fetchMutualFundNAV: no NAV data found for fund %q / isin %s near %s", fundName, isin, date)
	return 0
}

// generateMFTicker creates a clean, readable symbol from a mutual fund name.
func generateMFTicker(name string) string {
	name = strings.ToUpper(strings.TrimSpace(name))
	name = strings.ReplaceAll(name, " - DIRECT PLAN", "")
	name = strings.ReplaceAll(name, " - REGULAR PLAN", "")
	name = strings.ReplaceAll(name, " DIRECT PLAN", "")
	name = strings.ReplaceAll(name, " REGULAR PLAN", "")
	name = strings.ReplaceAll(name, "-GROWTH", "")
	name = strings.ReplaceAll(name, " GROWTH", "")
	name = strings.ReplaceAll(name, " FUND", "")
	name = mfTickerCleanRegex.ReplaceAllString(name, "_")
	name = strings.Trim(name, "_")
	if len(name) > 20 {
		name = name[:20]
	}
	return name
}

// mustParseDate parses a YYYY-MM-DD date and returns time.Now() on failure.
func mustParseDate(date string) time.Time {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return time.Now()
	}
	return t
}

// ---------------------------------------------------------------------------
// Firestore persistence — deterministic portfolio math
// ---------------------------------------------------------------------------

// saveInvestmentData persists the parsed investments and transactions to
// Firestore using a single atomic batch. For each investment it:
//   - Looks up an existing holding by ISIN (or fund name)
//   - BUY: adds quantity, updates weighted average cost basis
//   - SELL: reduces quantity, scales invested amount, deletes if qty <= 0
func (h *GmailHandler) saveInvestmentData(ctx context.Context, userID string, investments []models.Investment, transactions []models.Transaction) error {
	if len(investments) == 0 && len(transactions) == 0 {
		log.Println("No data to save to Firestore")
		return nil
	}

	batch := h.firestoreClient.Batch()

	// Track investment refs to link with transactions
	invRefs := make([]*firestore.DocumentRef, len(investments))

	for i, investment := range investments {
		if i >= len(transactions) {
			break
		}
		tx := transactions[i]

		// Look up existing investment by ISIN ticker or name
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
			log.Printf("Failed to query existing investments for %s: %v", investment.Name, err)
			continue
		}

		if len(docs) > 0 {
			// --- Update existing investment ---
			var existing models.Investment
			if err := docs[0].DataTo(&existing); err != nil {
				log.Printf("Failed to decode existing investment: %v", err)
				continue
			}
			docRef := docs[0].Ref
			invRefs[i] = docRef

			// Current price passed from parseEmailContent
			currentPrice := investment.InvestmentData.CurrentPrice
			if currentPrice == 0 {
				currentPrice = tx.Price
			}

			if tx.Type == models.TransactionTypeBuy {
				txQty := tx.Quantity
				txPrice := tx.Price // Price on transaction date

				if investment.Category == models.CategoryStocks {
					existingQty := existing.InvestmentData.Quantity
					existingInvested := existing.InvestedAmount

					newQty := existingQty + txQty
					newInvestedAmount := existingInvested + (txQty * txPrice)
					newAveragePrice := 0.0
					if newQty > 0 {
						newAveragePrice = newInvestedAmount / newQty
					}

					existing.InvestmentData.Quantity = newQty
					existing.InvestmentData.AveragePrice = newAveragePrice
					existing.InvestmentData.CurrentPrice = currentPrice
					existing.InvestedAmount = newInvestedAmount
					existing.CurrentValue = newQty * currentPrice // Total quantity * Current stock price

				} else if investment.Category == models.CategoryMutualFunds {
					existingUnits := existing.InvestmentData.Units
					existingInvested := existing.InvestedAmount

					newUnits := existingUnits + txQty
					newInvestedAmount := existingInvested + (txQty * txPrice)
					newAverageNAV := 0.0
					if newUnits > 0 {
						newAverageNAV = newInvestedAmount / newUnits
					}

					existing.InvestmentData.Units = newUnits
					existing.InvestmentData.NAV = newAverageNAV
					existing.InvestmentData.CurrentPrice = currentPrice
					existing.InvestedAmount = newInvestedAmount
					existing.CurrentValue = newUnits * currentPrice
				}

				existing.UpdatedAt = time.Now()
				batch.Set(docRef, existing)

			} else if tx.Type == models.TransactionTypeSell {
				txQty := tx.Quantity

				if investment.Category == models.CategoryStocks {
					existingQty := existing.InvestmentData.Quantity
					existingAvgPrice := existing.InvestmentData.AveragePrice

					newQty := existingQty - txQty
					if newQty <= 0 {
						log.Printf("Deleting stock investment %s (qty reached %.4f)", existing.Name, newQty)
						batch.Delete(docRef)
						continue
					}
					newInvestedAmount := newQty * existingAvgPrice

					existing.InvestmentData.Quantity = newQty
					existing.InvestmentData.AveragePrice = existingAvgPrice // unchanged cost basis
					existing.InvestmentData.CurrentPrice = currentPrice
					existing.InvestedAmount = newInvestedAmount
					existing.CurrentValue = newQty * currentPrice // Remaining quantity * Current stock price
					existing.UpdatedAt = time.Now()
					batch.Set(docRef, existing)

				} else if investment.Category == models.CategoryMutualFunds {
					existingUnits := existing.InvestmentData.Units
					existingNAV := existing.InvestmentData.NAV // average NAV (cost basis per unit)

					newUnits := existingUnits - txQty
					if newUnits <= 0 {
						log.Printf("Deleting MF investment %s (units reached %.4f)", existing.Name, newUnits)
						batch.Delete(docRef)
						continue
					}
					newInvestedAmount := newUnits * existingNAV

					existing.InvestmentData.Units = newUnits
					existing.InvestmentData.NAV = existingNAV // unchanged cost basis
					existing.InvestmentData.CurrentPrice = currentPrice
					existing.InvestedAmount = newInvestedAmount
					existing.CurrentValue = newUnits * currentPrice
					existing.UpdatedAt = time.Now()
					batch.Set(docRef, existing)
				}
			}

		} else {
			// --- Create new investment (only for BUY) ---
			if tx.Type == models.TransactionTypeBuy {
				txQty := tx.Quantity
				txPrice := tx.Price                  // Historical price on transaction date
				investedAmount := txQty * txPrice    // Invested amount on transaction date
				currentPrice := investment.InvestmentData.CurrentPrice
				if currentPrice == 0 {
					currentPrice = txPrice
				}
				currentValue := txQty * currentPrice // Current value on current date

				if investment.Category == models.CategoryStocks {
					investment.InvestmentData.AveragePrice = txPrice
					investment.InvestmentData.CurrentPrice = currentPrice
					investment.InvestmentData.Quantity = txQty
				} else if investment.Category == models.CategoryMutualFunds {
					investment.InvestmentData.NAV = txPrice
					investment.InvestmentData.CurrentPrice = currentPrice
					investment.InvestmentData.Units = txQty
				}

				investment.InvestedAmount = investedAmount
				investment.CurrentValue = currentValue

				ref := h.firestoreClient.Collection("users").Doc(userID).Collection("investments").NewDoc()
				batch.Set(ref, investment)
				invRefs[i] = ref
				log.Printf("Creating new investment: %s %s qty=%.4f txPrice=%.4f currentPrice=%.4f invested=%.2f currentValue=%.2f",
					investment.Category, investment.Name, txQty, txPrice, currentPrice, investedAmount, currentValue)
			} else {
				log.Printf("Skipping SELL for non-existent investment %s (%s)", investment.Name, investment.InvestmentData.Ticker)
			}
		}
	}

	// Save transaction records with relational links
	for i, transaction := range transactions {
		if i < len(invRefs) && invRefs[i] != nil {
			transaction.RelatedID = invRefs[i].ID
		}
		ref := h.firestoreClient.Collection("users").Doc(userID).Collection("transactions").NewDoc()
		batch.Set(ref, transaction)
	}

	// Commit atomic batch
	_, err := batch.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit Firestore batch: %v", err)
	}

	log.Printf("Firestore batch committed: %d investments, %d transactions", len(investments), len(transactions))
	return nil
}

// ---------------------------------------------------------------------------
// User profile helper
// ---------------------------------------------------------------------------

// getUserProfile retrieves user profile from Firestore.
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

// ---------------------------------------------------------------------------
// cleanInvestmentName — sanitises AI-returned names before storage
// ---------------------------------------------------------------------------

// cleanInvestmentName cleans up stock and mutual fund names.
// It strips plan/growth suffixes from MF names and equity-share suffixes from
// stock names. Retained unchanged from the original implementation.
func cleanInvestmentName(name string, isMutualFund bool) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}

	if isMutualFund {
		// Remove AMC / MF House prefix with hash (e.g. "MOTILAL OSWAL AMC LTD#MOTILAL OSWAL MF-...")
		if strings.Contains(name, "#") {
			parts := strings.SplitN(name, "#", 2)
			if len(parts) > 1 {
				name = parts[1]
			}
		}

		// Remove leading AMC prefix ending with MF or MUTUAL FUND before a dash
		// e.g. "MIRAE ASSET MF-MIRAE ASSET GOLD ETF" -> "MIRAE ASSET GOLD ETF"
		// e.g. "MOTILAL OSWAL MF-MOTILAL OSWAL MIDCAP 30 FUND" -> "MOTILAL OSWAL MIDCAP 30 FUND"
		name = mfPrefixRegex.ReplaceAllString(name, "")

		// Clean common suffixes: Direct, Regular, Plan, Growth, Option
		name = mfPlanSuffixRegex.ReplaceAllString(name, "")
		name = mfGrowthSuffixRegex.ReplaceAllString(name, "")
	} else {
		// Remove " - EQUITY SHARES" / " - NEW EQUITY SHARES" etc.
		name = stockEquitySuffixRegex.ReplaceAllString(name, "")
	}

	return strings.TrimSpace(name)
}
