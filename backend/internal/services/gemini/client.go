package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// DefaultMinInterval enforces a max of 12 queries per minute (60s / 12 = 5.0s).
	// We use 5100ms (5.1 seconds) to provide a safe margin so calls never exceed 11.76 RPM.
	DefaultMinInterval = 5100 * time.Millisecond

	// DefaultMaxRetries is the maximum number of retries upon encountering 429 or transient errors.
	DefaultMaxRetries = 3

	// DefaultRetryDelay is the fallback backoff duration if Gemini doesn't specify retryDelay.
	DefaultRetryDelay = 16 * time.Second
)

var retryRegex = regexp.MustCompile(`(?i)retry in ([0-9]+(?:\.[0-9]+)?)s`)

// RateLimiter paces requests to Gemini to guarantee that consecutive calls
// do not exceed the configured maximum rate (12 requests/minute by default).
type RateLimiter struct {
	mu          sync.Mutex
	lastRequest time.Time
	interval    time.Duration
}

// NewRateLimiter creates a RateLimiter with the specified minimum interval between calls.
func NewRateLimiter(interval time.Duration) *RateLimiter {
	return &RateLimiter{
		interval: interval,
	}
}

// SetInterval dynamically updates the minimum interval between calls.
func (rl *RateLimiter) SetInterval(interval time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.interval = interval
}

// Wait blocks until the required interval has passed since the last request,
// or returns an error if ctx is cancelled.
func (rl *RateLimiter) Wait(ctx context.Context) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	if !rl.lastRequest.IsZero() {
		elapsed := now.Sub(rl.lastRequest)
		if elapsed < rl.interval {
			sleepDuration := rl.interval - elapsed
			log.Printf("[Gemini RateLimiter] Pacing request (max 12 req/min): waiting %v before next query...", sleepDuration.Round(time.Millisecond))
			select {
			case <-time.After(sleepDuration):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	rl.lastRequest = time.Now()
	return nil
}

// Client wraps an HTTP client with rate limiting and automatic retry handling for Gemini API.
type Client struct {
	limiter    *RateLimiter
	httpClient *http.Client
	maxRetries int
}

// GlobalClient is the default shared client instance across all handlers and services.
var GlobalClient = NewClient(DefaultMinInterval, DefaultMaxRetries, http.DefaultClient)

// NewClient creates a new rate-limited Gemini Client.
func NewClient(interval time.Duration, maxRetries int, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		limiter:    NewRateLimiter(interval),
		httpClient: httpClient,
		maxRetries: maxRetries,
	}
}

// SetRateLimitInterval modifies the global rate limiter interval.
func SetRateLimitInterval(interval time.Duration) {
	GlobalClient.limiter.SetInterval(interval)
}

// ExtractRetryDelay inspects Gemini's 429 response body to find the server-recommended
// retry duration (from RetryInfo or error message) and adds a small safety margin.
func ExtractRetryDelay(respBytes []byte) time.Duration {
	var errPayload struct {
		Error struct {
			Code    int                      `json:"code"`
			Message string                   `json:"message"`
			Status  string                   `json:"status"`
			Details []map[string]interface{} `json:"details"`
		} `json:"error"`
	}

	if err := json.Unmarshal(respBytes, &errPayload); err == nil {
		// 1. Check details array for type.googleapis.com/google.rpc.RetryInfo
		for _, d := range errPayload.Error.Details {
			if typeVal, ok := d["@type"].(string); ok && strings.Contains(typeVal, "RetryInfo") {
				if rd, ok := d["retryDelay"].(string); ok && rd != "" {
					if dur, err := time.ParseDuration(rd); err == nil && dur > 0 {
						return clampDelay(dur + 1500*time.Millisecond)
					}
				}
			}
		}

		// 2. Check message string for "Please retry in <seconds>s"
		if match := retryRegex.FindStringSubmatch(errPayload.Error.Message); len(match) > 1 {
			if sec, err := strconv.ParseFloat(match[1], 64); err == nil && sec > 0 {
				dur := time.Duration(sec * float64(time.Second))
				return clampDelay(dur + 1500*time.Millisecond)
			}
		}
	}

	// 3. Fallback regex on raw response bytes
	if match := retryRegex.FindSubmatch(respBytes); len(match) > 1 {
		if sec, err := strconv.ParseFloat(string(match[1]), 64); err == nil && sec > 0 {
			dur := time.Duration(sec * float64(time.Second))
			return clampDelay(dur + 1500*time.Millisecond)
		}
	}

	return DefaultRetryDelay
}

func clampDelay(dur time.Duration) time.Duration {
	if dur < 5*time.Second {
		return 5 * time.Second
	}
	if dur > 60*time.Second {
		return 60 * time.Second
	}
	return dur
}

// PostWithRetry executes a JSON POST request to the Gemini API endpoint.
// It automatically:
// 1. Enforces rate limiting (max 12 queries per minute, >= 5.1s between calls).
// 2. Intercepts HTTP 429 (quota exceeded), extracts retryDelay, and retries up to maxRetries.
// 3. Intercepts transient 5xx server errors and retries.
// 4. Returns the response bytes on HTTP 200 OK.
func (c *Client) PostWithRetry(ctx context.Context, endpoint string, bodyBytes []byte) ([]byte, error) {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		// Enforce rate limiting slot before dispatching
		if err := c.limiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter wait cancelled: %w", err)
		}

		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to build Gemini HTTP request: %w", err)
		}
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(httpReq)
		if err != nil {
			lastErr = fmt.Errorf("Gemini HTTP call failed: %w", err)
			log.Printf("[Gemini Client] HTTP request error (attempt %d/%d): %v", attempt+1, c.maxRetries+1, err)
			if attempt < c.maxRetries {
				select {
				case <-time.After(2 * time.Second):
					continue
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			}
			continue
		}

		respBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to read Gemini response: %w", err)
			continue
		}

		if resp.StatusCode == http.StatusOK {
			return respBytes, nil
		}

		// Handle 429 Quota / Rate Limit
		if resp.StatusCode == http.StatusTooManyRequests {
			retryDelay := ExtractRetryDelay(respBytes)
			log.Printf("[Gemini Client] Rate limit exceeded (429 RESOURCE_EXHAUSTED). Waiting %v before retry (attempt %d/%d)...",
				retryDelay.Round(time.Millisecond), attempt+1, c.maxRetries+1)
			lastErr = fmt.Errorf("Gemini 429 quota exceeded: %s", string(respBytes))

			if attempt < c.maxRetries {
				select {
				case <-time.After(retryDelay):
					continue
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			}
			continue
		}

		// Handle transient server errors (500, 502, 503, 504)
		if resp.StatusCode == http.StatusBadGateway ||
			resp.StatusCode == http.StatusServiceUnavailable ||
			resp.StatusCode == http.StatusGatewayTimeout ||
			resp.StatusCode == http.StatusInternalServerError {
			log.Printf("[Gemini Client] Transient error (status %d). Backing off before retry (attempt %d/%d)...",
				resp.StatusCode, attempt+1, c.maxRetries+1)
			lastErr = fmt.Errorf("Gemini transient error status %d: %s", resp.StatusCode, string(respBytes))

			if attempt < c.maxRetries {
				select {
				case <-time.After(3 * time.Second):
					continue
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			}
			continue
		}

		// Non-retryable client error (e.g. 400 Bad Request, 401, 403)
		return nil, fmt.Errorf("Gemini returned non-retryable status %d: %s", resp.StatusCode, string(respBytes))
	}

	return nil, fmt.Errorf("Gemini request failed after %d attempts: %w", c.maxRetries+1, lastErr)
}

// PostWithRetry is a package-level helper that delegates to GlobalClient.
func PostWithRetry(ctx context.Context, endpoint string, bodyBytes []byte) ([]byte, error) {
	return GlobalClient.PostWithRetry(ctx, endpoint, bodyBytes)
}
