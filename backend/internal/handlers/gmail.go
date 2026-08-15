package handlers

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"slices"
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
// SyncGmail — main HTTP handler
// ---------------------------------------------------------------------------

// SyncGmail handles the Gmail sync request.
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
		investments, transactions, err := h.parseEmailContent(c.Request.Context(), email)
		if err != nil {
			log.Printf("Failed to parse email %s: %v", email.Id, err)
			continue
		}

		if len(investments) == 0 {
			log.Printf("No investments extracted from email %s", email.Id)
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

// stripHTML converts HTML content to plain text by walking the parse tree.
func (h *GmailHandler) stripHTML(body string) string {
	doc, err := html.Parse(strings.NewReader(body))
	if err != nil {
		// Fallback: crude tag removal
		return regexp.MustCompile(`<[^>]+>`).ReplaceAllString(body, " ")
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
	text = regexp.MustCompile(`[ \t]+`).ReplaceAllString(text, " ")
	text = regexp.MustCompile(`(\n\s*){3,}`).ReplaceAllString(text, "\n\n")
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

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to build Gemini HTTP request: %v", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		log.Printf("callGeminiForTransactions: Gemini HTTP call failed: %v", err)
		return nil, fmt.Errorf("Gemini HTTP call failed: %v", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("callGeminiForTransactions: failed to read Gemini response body: %v", err)
		return nil, fmt.Errorf("failed to read Gemini response body: %v", err)
	}

	log.Printf("callGeminiForTransactions: received response status=%d body_size=%d", resp.StatusCode, len(respBytes))

	if resp.StatusCode != http.StatusOK {
		log.Printf("callGeminiForTransactions: Gemini API returned error response: %s", string(respBytes))
		return nil, fmt.Errorf("Gemini API returned status %d: %s", resp.StatusCode, string(respBytes))
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
			strings.EqualFold(at.Category, "mutualFunds")

		// Clean name
		name := cleanInvestmentName(at.Name, isMF)
		if name == "" {
			name = at.Name
		}

		// Parse transaction date
		txDate := emailDate
		if at.TransactionDate != "" {
			if parsed, err := time.Parse("2006-01-02", at.TransactionDate); err == nil {
				txDate = parsed
			}
		}

		// Determine category and transaction type
		category := models.CategoryStocks
		if isMF {
			category = models.CategoryMutualFunds
		}

		txType := models.TransactionTypeBuy
		if strings.EqualFold(at.Type, "sell") {
			txType = models.TransactionTypeSell
		}

		txDateStr := txDate.Format("2006-01-02")
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
			Date:        txDate,
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

// fetchMarketPrice is the unified dispatcher for retrieving historical prices.
// Returns 0 on any error — callers must handle the 0 case gracefully.
func (h *GmailHandler) fetchMarketPrice(ctx context.Context, symbol, date string, isMF bool) float64 {
	if isMF {
		price := h.fetchMutualFundNAV(ctx, symbol, "", date)
		log.Printf("MF NAV for %s on %s: %.4f", symbol, date, price)
		return price
	}
	price := h.fetchStockPrice(ctx, symbol, date)
	log.Printf("Stock price for %s on %s: %.4f", symbol, date, price)
	return price
}

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
			return *closes[i]
		}
		diff := int(target.Sub(tDate).Hours() / 24)
		if diff >= 0 && diff < bestDiff {
			bestDiff = diff
			bestPrice = *closes[i]
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
	name = regexp.MustCompile(`[^A-Z0-9]+`).ReplaceAllString(name, "_")
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
		mfPrefixRegex := regexp.MustCompile(`(?i)^[^-]+(?:MF|MUTUAL\s+FUND)\s*-\s*`)
		name = mfPrefixRegex.ReplaceAllString(name, "")

		// Clean common suffixes: Direct, Regular, Plan, Growth, Option
		planRegex := regexp.MustCompile(`(?i)(?:\s*[-/]\s*|\s+)\b(?:DIRECT|REGULAR|PLAN|PL|OPT|OPTION)\b.*$`)
		name = planRegex.ReplaceAllString(name, "")
		growthRegex := regexp.MustCompile(`(?i)\s*[-/]\s*GROWTH\b.*$`)
		name = growthRegex.ReplaceAllString(name, "")
	} else {
		// Remove " - EQUITY SHARES" / " - NEW EQUITY SHARES" etc.
		stockRegex := regexp.MustCompile(`(?i)\s*-\s*(?:NEW\s+)?EQUITY(?:\s+SHARES)?.*$`)
		name = stockRegex.ReplaceAllString(name, "")
	}

	return strings.TrimSpace(name)
}
