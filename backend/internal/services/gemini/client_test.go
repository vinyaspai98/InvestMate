package gemini

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestExtractRetryDelay(t *testing.T) {
	// 1. Exact Gemini payload as reported by user
	userPayload := []byte(`{
  "error": {
    "code": 429,
    "message": "You exceeded your current quota, please check your plan and billing details. For more information on this error, head to: https://ai.google.dev/gemini-api/docs/rate-limits. To monitor your current usage, head to: https://ai.dev/rate-limit. \n* Quota exceeded for metric: generativelanguage.googleapis.com/generate_content_free_tier_requests, limit: 15, model: gemini-3.5-flash-lite\nPlease retry in 15.10362494s.",
    "status": "RESOURCE_EXHAUSTED",
    "details": [
      {
        "@type": "type.googleapis.com/google.rpc.Help",
        "links": [
          {
            "description": "Learn more about Gemini API quotas",
            "url": "https://ai.google.dev/gemini-api/docs/rate-limits"
          }
        ]
      },
      {
        "@type": "type.googleapis.com/google.rpc.QuotaFailure",
        "violations": [
          {
            "quotaMetric": "generativelanguage.googleapis.com/generate_content_free_tier_requests",
            "quotaId": "GenerateRequestsPerMinutePerProjectPerModel-FreeTier",
            "quotaDimensions": {
              "location": "global",
              "model": "gemini-3.5-flash-lite"
            },
            "quotaValue": "15"
          }
        ]
      },
      {
        "@type": "type.googleapis.com/google.rpc.RetryInfo",
        "retryDelay": "15s"
      }
    ]
  }
}`)

	delay := ExtractRetryDelay(userPayload)
	// 15s parsed from details + 1.5s buffer = 16.5s
	if delay < 16*time.Second || delay > 17*time.Second {
		t.Errorf("expected delay around 16.5s, got %v", delay)
	}

	// 2. Payload without details array, but with message regex
	msgOnlyPayload := []byte(`{
  "error": {
    "code": 429,
    "message": "Quota exceeded. Please retry in 10s.",
    "status": "RESOURCE_EXHAUSTED"
  }
}`)
	delayMsg := ExtractRetryDelay(msgOnlyPayload)
	// 10s + 1.5s = 11.5s
	if delayMsg < 11*time.Second || delayMsg > 12*time.Second {
		t.Errorf("expected delay around 11.5s, got %v", delayMsg)
	}

	// 3. Fallback when no delay info present
	genericPayload := []byte(`{"error": {"code": 429, "message": "Too many requests"}}`)
	delayGeneric := ExtractRetryDelay(genericPayload)
	if delayGeneric != DefaultRetryDelay {
		t.Errorf("expected DefaultRetryDelay (%v), got %v", DefaultRetryDelay, delayGeneric)
	}
}

func TestRateLimiterPacing(t *testing.T) {
	// Test with a short interval: 50ms
	interval := 50 * time.Millisecond
	limiter := NewRateLimiter(interval)

	ctx := context.Background()

	start := time.Now()
	// First call should not wait
	if err := limiter.Wait(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Second call should wait at least 50ms
	if err := limiter.Wait(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	elapsed := time.Since(start)
	if elapsed < 45*time.Millisecond {
		t.Errorf("expected at least ~50ms elapsed, got %v", elapsed)
	}
}

func TestPostWithRetry_SucceedsAfter429(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := atomic.AddInt32(&attempts, 1)
		if current == 1 {
			// Return 429 with short retry delay
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{
				"error": {
					"code": 429,
					"message": "Please retry in 0.1s.",
					"details": [{"@type": "type.googleapis.com/google.rpc.RetryInfo", "retryDelay": "0.1s"}]
				}
			}`))
			return
		}
		// Second attempt succeeds
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"candidates": [{"content": {"parts": [{"text": "success"}]}}]}`))
	}))
	defer server.Close()

	// Use client with zero interval for fast test, but allow retries
	client := NewClient(1*time.Millisecond, 2, server.Client())

	ctx := context.Background()
	body, err := client.PostWithRetry(ctx, server.URL, []byte(`{}`))
	if err != nil {
		t.Fatalf("expected success on retry, got err: %v", err)
	}

	if string(body) == "" || atomic.LoadInt32(&attempts) != 2 {
		t.Errorf("expected 2 attempts, got %d, body=%s", attempts, string(body))
	}
}

func TestPostWithRetry_NonRetryableError(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": {"code": 400, "message": "Invalid argument"}}`))
	}))
	defer server.Close()

	client := NewClient(1*time.Millisecond, 2, server.Client())
	ctx := context.Background()
	_, err := client.PostWithRetry(ctx, server.URL, []byte(`{}`))
	if err == nil {
		t.Fatal("expected error on 400 Bad Request, got nil")
	}

	if atomic.LoadInt32(&attempts) != 1 {
		t.Errorf("expected exactly 1 attempt for non-retryable error, got %d", attempts)
	}
}
