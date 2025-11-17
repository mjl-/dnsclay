package njalla

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"time"
)

const apiEndpoint = "https://njal.la/api/1/"

// clientInterface defines the interface for API clients
type clientInterface interface {
	call(ctx context.Context, method string, params interface{}, result interface{}) error
}

// RetryConfig holds configuration for retry attempts
type RetryConfig struct {
	MaxRetries   int
	BaseDelay    time.Duration
	MaxDelay     time.Duration
	RandomFactor float64
}

// DefaultRetryConfig returns a reasonable default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:   3,
		BaseDelay:    100 * time.Millisecond,
		MaxDelay:     2 * time.Second,
		RandomFactor: 0.5,
	}
}

type client struct {
	token       string
	timeout     time.Duration
	retryConfig RetryConfig
}

func newClient(token string) *client {
	return &client{
		token:       token,
		timeout:     30 * time.Second,
		retryConfig: DefaultRetryConfig(),
	}
}

type jsonRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      string      `json:"id,omitempty"`
}

type jsonRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *jsonRPCError   `json:"error,omitempty"`
	ID      string          `json:"id,omitempty"`
}

type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// isRetryable determines if an error or HTTP status code is retryable
func isRetryable(err error, statusCode int) bool {
	// Network errors are generally retryable
	if err != nil {
		return true
	}

	// Server errors and rate limiting
	return statusCode >= 500 || statusCode == 429
}

// calculateBackoff calculates the delay before the next retry attempt
func calculateBackoff(config RetryConfig, attempt int) time.Duration {
	// Calculate basic exponential backoff
	delay := float64(config.BaseDelay) * math.Pow(2, float64(attempt))

	// Apply jitter to avoid thundering herd
	jitter := rand.Float64() * config.RandomFactor * delay
	delay += jitter

	// Cap at max delay
	if delay > float64(config.MaxDelay) {
		delay = float64(config.MaxDelay)
	}

	return time.Duration(delay)
}

func (c *client) call(ctx context.Context, method string, params interface{}, result interface{}) error {
	req := jsonRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      "1", // Any ID will work
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("error marshaling request: %w", err)
	}

	var lastErr error
	var statusCode int

	// Try the request with retries
	for attempt := 0; attempt <= c.retryConfig.MaxRetries; attempt++ {
		// Skip delay on the first attempt
		if attempt > 0 {
			// Calculate backoff duration
			backoff := calculateBackoff(c.retryConfig, attempt-1)

			// Create a timer using the context
			timer := time.NewTimer(backoff)
			select {
			case <-ctx.Done():
				timer.Stop()
				return fmt.Errorf("context canceled during retry backoff: %w", ctx.Err())
			case <-timer.C:
				// Continue with the retry
			}
		}

		// Make a new request object for each attempt
		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, apiEndpoint, bytes.NewBuffer(reqBody))
		if err != nil {
			return fmt.Errorf("error creating request: %w", err)
		}

		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Accept", "application/json")
		httpReq.Header.Set("Authorization", "Njalla "+c.token)

		httpClient := &http.Client{
			Timeout: c.timeout,
		}

		resp, err := httpClient.Do(httpReq)
		if err != nil {
			lastErr = fmt.Errorf("error making request: %w", err)
			// Retry on network errors
			continue
		}

		statusCode = resp.StatusCode

		// If we got a response, read the body then close it
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("error reading response body: %w", err)
			// Retry on read errors
			continue
		}

		// Handle HTTP status errors
		if statusCode >= 400 {
			lastErr = fmt.Errorf("HTTP error: %d", statusCode)
			if isRetryable(nil, statusCode) {
				continue
			}
			return lastErr
		}

		var jsonResp jsonRPCResponse
		if err := json.Unmarshal(body, &jsonResp); err != nil {
			lastErr = fmt.Errorf("error unmarshaling response: %w", err)
			// Generally retry on JSON parsing errors
			continue
		}

		if jsonResp.Error != nil {
			lastErr = fmt.Errorf("API error: %d - %s", jsonResp.Error.Code, jsonResp.Error.Message)
			// Don't retry on API logic errors
			return lastErr
		}

		if result != nil {
			if err := json.Unmarshal(jsonResp.Result, result); err != nil {
				return fmt.Errorf("error unmarshaling result: %w", err)
			}
		}

		// If we got here, the request succeeded
		return nil
	}

	// If we exhausted all retries
	return fmt.Errorf("max retries exceeded, last error: %w", lastErr)
}
