package caspdf

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"investmate-backend/internal/services/gemini"
)

//go:embed prompts/gemini_cas_system_instruction.txt
var casSystemInstruction string

// CASHolding represents a portfolio holding lot on the statement date (Section 1).
type CASHolding struct {
	Category        string  `json:"category"`         // "stocks" | "mutualFunds"
	Name            string  `json:"name"`
	ISIN            string  `json:"isin,omitempty"`
	Ticker          string  `json:"ticker,omitempty"`
	Quantity        float64 `json:"quantity"`
	Price           float64 `json:"price,omitempty"`
	InvestedAmount  float64 `json:"invested_amount,omitempty"`
	Type            string  `json:"type"`             // "buy"
	TransactionDate string  `json:"transaction_date"` // YYYY-MM-DD
}

// CASTransaction represents an individual trade executed during the statement month (Section 2).
type CASTransaction struct {
	Category        string  `json:"category"`         // "stocks" | "mutualFunds"
	Name            string  `json:"name"`
	ISIN            string  `json:"isin,omitempty"`
	Ticker          string  `json:"ticker,omitempty"`
	Quantity        float64 `json:"quantity"`
	Price           float64 `json:"price,omitempty"`
	Type            string  `json:"type"`             // "buy" | "sell"
	TransactionDate string  `json:"transaction_date"` // YYYY-MM-DD
	Description     string  `json:"description,omitempty"`
}

// CASExtractionResult holds the extracted portfolio holdings (Section 1) and transactions (Section 2).
type CASExtractionResult struct {
	StatementDate string            `json:"statement_date"`
	Holdings      []CASHolding      `json:"holdings"`     // Section 1: All holdings on a given date
	Transactions  []CASTransaction  `json:"transactions"` // Section 2: All tradings performed in that month
}

// DecryptAndExtractText decrypts a CDSL CAS PDF with the user's PAN card number
// using pdftotext and returns the full layout text.
func DecryptAndExtractText(ctx context.Context, pdfBytes []byte, panNumber string) (string, error) {
	if len(pdfBytes) == 0 {
		return "", fmt.Errorf("empty PDF bytes provided")
	}

	tempFile, err := os.CreateTemp("", "cdsl_cas_*.pdf")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file for CAS PDF: %w", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.Write(pdfBytes); err != nil {
		tempFile.Close()
		return "", fmt.Errorf("failed to write CAS PDF to temporary file: %w", err)
	}
	tempFile.Close()

	panClean := strings.ToUpper(strings.TrimSpace(panNumber))

	// 1. Try with user-provided PAN as password
	if panClean != "" {
		cmd := exec.CommandContext(ctx, "pdftotext", "-layout", "-upw", panClean, tempFile.Name(), "-")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err == nil && stdout.Len() > 0 {
			log.Printf("Successfully decrypted CAS PDF with PAN, extracted %d bytes", stdout.Len())
			return stdout.String(), nil
		} else {
			log.Printf("pdftotext with PAN password failed: %v, stderr: %s", err, stderr.String())
		}
	}

	// 2. Fallback: Try without password in case the PDF is unencrypted
	cmd := exec.CommandContext(ctx, "pdftotext", "-layout", tempFile.Name(), "-")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err == nil && stdout.Len() > 0 {
		log.Printf("Successfully extracted text from unencrypted CAS PDF, %d bytes", stdout.Len())
		return stdout.String(), nil
	}

	return "", fmt.Errorf("failed to decrypt or read CAS PDF. Please verify your PAN card number: %s", stderr.String())
}

var (
	statementDateRegexes = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(?:statement|valuation|portfolio|holding|holdings)\s+(?:as\s+on|as\s+at|dated)\s*[:\-]?\s*([0-9]{1,2}[-\/\s][A-Za-z]{3,9}[-\/\s][0-9]{2,4}|[0-9]{1,2}[-\/\s][0-9]{1,2}[-\/\s][0-9]{2,4})`),
		regexp.MustCompile(`(?i)(?:for\s+the\s+period\s+ended|as\s+on\s+date)\s*[:\-]?\s*([0-9]{1,2}[-\/\s][A-Za-z]{3,9}[-\/\s][0-9]{2,4}|[0-9]{1,2}[-\/\s][0-9]{1,2}[-\/\s][0-9]{2,4})`),
		regexp.MustCompile(`(?i)as\s+on\s+([0-9]{1,2}-[A-Za-z]{3}-[0-9]{4})`),
	}

	dateFormats = []string{
		"02-Jan-2006",
		"02-January-2006",
		"02 Jan 2006",
		"02 January 2006",
		"02/01/2006",
		"02-01-2006",
		"2006-01-02",
		"2/1/2006",
		"2-1-2006",
		"02-Jan-06",
		"02/01/06",
	}
)

// ExtractStatementDateFromText attempts to discover the effective statement date using regex patterns.
func ExtractStatementDateFromText(text string) (time.Time, bool) {
	for _, re := range statementDateRegexes {
		matches := re.FindAllStringSubmatch(text, -1)
		for _, match := range matches {
			if len(match) > 1 {
				candidate := strings.TrimSpace(match[1])
				for _, layout := range dateFormats {
					if t, err := time.Parse(layout, candidate); err == nil {
						return t, true
					}
				}
			}
		}
	}
	return time.Time{}, false
}

// ExtractFromPDF orchestrates PDF decryption, text extraction, and Gemini AI parsing.
func ExtractFromPDF(ctx context.Context, pdfBytes []byte, panNumber, geminiAPIKey string) (*CASExtractionResult, error) {
	text, err := DecryptAndExtractText(ctx, pdfBytes, panNumber)
	if err != nil {
		return nil, err
	}

	return ExtractHoldingsFromText(ctx, text, geminiAPIKey)
}

// ExtractHoldingsFromText uses Gemini Flash structured output to extract all holdings.
func ExtractHoldingsFromText(ctx context.Context, text, geminiAPIKey string) (*CASExtractionResult, error) {
	regexDate, hasRegexDate := ExtractStatementDateFromText(text)
	defaultDateStr := time.Now().Format("2006-01-02")
	if hasRegexDate {
		defaultDateStr = regexDate.Format("2006-01-02")
	}

	if geminiAPIKey == "" {
		log.Println("ExtractHoldingsFromText: GEMINI_API_KEY is not configured, running heuristic parser fallback")
		return heuristicParser(text, defaultDateStr), nil
	}

	// Limit text length to avoid token limit if CAS statement is exceptionally large (e.g. > 100 pages)
	maxChars := 120000
	cleanText := text
	if len(cleanText) > maxChars {
		cleanText = cleanText[:maxChars]
	}

	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"statement_date": map[string]string{"type": "string"},
			"holdings": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"category":         map[string]string{"type": "string"},
						"name":             map[string]string{"type": "string"},
						"isin":             map[string]string{"type": "string"},
						"ticker":           map[string]string{"type": "string"},
						"quantity":         map[string]string{"type": "number"},
						"price":            map[string]string{"type": "number"},
						"invested_amount":  map[string]string{"type": "number"},
						"type":             map[string]string{"type": "string"},
						"transaction_date": map[string]string{"type": "string"},
					},
					"required": []string{"category", "name", "quantity", "type"},
				},
			},
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
						"price":            map[string]string{"type": "number"},
						"type":             map[string]string{"type": "string"},
						"transaction_date": map[string]string{"type": "string"},
						"description":      map[string]string{"type": "string"},
					},
					"required": []string{"category", "name", "quantity", "type", "transaction_date"},
				},
			},
		},
		"required": []string{"statement_date", "holdings"},
	}

	reqBody := map[string]interface{}{
		"system_instruction": map[string]interface{}{
			"parts": []map[string]string{{"text": casSystemInstruction}},
		},
		"contents": []map[string]interface{}{
			{
				"role":  "user",
				"parts": []map[string]string{{"text": cleanText}},
			},
		},
		"generationConfig": map[string]interface{}{
			"responseMimeType": "application/json",
			"responseSchema":   schema,
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Gemini CAS request: %w", err)
	}

	endpoint := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/gemini-3.5-flash-lite:generateContent?key=%s",
		geminiAPIKey,
	)

	respBytes, err := gemini.PostWithRetry(ctx, endpoint, bodyBytes)
	if err != nil {
		log.Printf("ExtractHoldingsFromText: Gemini call failed after retries: %v. Falling back to heuristic parser.", err)
		return heuristicParser(text, defaultDateStr), nil
	}

	var geminiEnvelope struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(respBytes, &geminiEnvelope); err != nil || len(geminiEnvelope.Candidates) == 0 || len(geminiEnvelope.Candidates[0].Content.Parts) == 0 {
		log.Printf("ExtractHoldingsFromText: failed to parse candidates, falling back: %v", err)
		return heuristicParser(text, defaultDateStr), nil
	}

	rawJSON := geminiEnvelope.Candidates[0].Content.Parts[0].Text
	var result CASExtractionResult
	if err := json.Unmarshal([]byte(rawJSON), &result); err != nil {
		log.Printf("ExtractHoldingsFromText: failed to parse JSON: %v raw: %s", err, rawJSON)
		return heuristicParser(text, defaultDateStr), nil
	}

	// Validate / fix statement date
	if result.StatementDate == "" || result.StatementDate == "YYYY-MM-DD" {
		result.StatementDate = defaultDateStr
	}

	// Ensure all holdings (Section 1) have proper transaction date and category
	for i := range result.Holdings {
		if result.Holdings[i].TransactionDate == "" {
			result.Holdings[i].TransactionDate = result.StatementDate
		}
		if result.Holdings[i].Type == "" {
			result.Holdings[i].Type = "buy"
		}
		// Respect user decision: ETFs are classified under mutualFunds
		if strings.Contains(strings.ToUpper(result.Holdings[i].Name), "ETF") ||
			strings.HasPrefix(strings.ToUpper(result.Holdings[i].ISIN), "INF") {
			result.Holdings[i].Category = "mutualFunds"
		}
	}

	// Ensure all transactions (Section 2) have proper date and category
	for i := range result.Transactions {
		if result.Transactions[i].TransactionDate == "" {
			result.Transactions[i].TransactionDate = result.StatementDate
		}
		if result.Transactions[i].Type == "" {
			result.Transactions[i].Type = "buy"
		}
		if strings.Contains(strings.ToUpper(result.Transactions[i].Name), "ETF") ||
			strings.HasPrefix(strings.ToUpper(result.Transactions[i].ISIN), "INF") {
			result.Transactions[i].Category = "mutualFunds"
		}
	}

	log.Printf("Gemini extracted %d CAS holdings (Sec 1) and %d trades (Sec 2), statement date: %s",
		len(result.Holdings), len(result.Transactions), result.StatementDate)
	return &result, nil
}

// heuristicParser extracts lines containing ISIN codes (INE... or INF...) as fallback.
func heuristicParser(text, defaultDate string) *CASExtractionResult {
	lines := strings.Split(text, "\n")
	var holdings []CASHolding
	seen := make(map[string]bool)

	isinRegex := regexp.MustCompile(`\b(IN[A-Z0-9]{10})\b`)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		match := isinRegex.FindString(trimmed)
		if match != "" && !seen[match] {
			seen[match] = true
			isMF := strings.HasPrefix(match, "INF") || strings.Contains(strings.ToUpper(trimmed), "MUTUAL FUND") || strings.Contains(strings.ToUpper(trimmed), "ETF")
			cat := "stocks"
			if isMF {
				cat = "mutualFunds"
			}

			// Extract words around ISIN
			words := strings.Fields(trimmed)
			name := trimmed
			if len(words) > 3 {
				name = strings.Join(words[:3], " ")
			}

			holdings = append(holdings, CASHolding{
				Category:        cat,
				Name:            name,
				ISIN:            match,
				Ticker:          match,
				Quantity:        1.0,
				Type:            "buy",
				TransactionDate: defaultDate,
			})
		}
	}

	return &CASExtractionResult{
		StatementDate: defaultDate,
		Holdings:      holdings,
		Transactions:  []CASTransaction{},
	}
}

// ParseStatementDate parses date strings into time.Time.
func ParseStatementDate(dateStr string) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)
	for _, layout := range dateFormats {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse date %q", dateStr)
}
