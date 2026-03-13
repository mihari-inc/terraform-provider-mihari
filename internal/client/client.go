package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client is the HTTP client for the Mihari API.
type Client struct {
	BaseURL        string
	APIToken       string
	OrganizationID string
	HTTPClient     *http.Client
	UserAgent      string
}

// ClientConfig holds configuration for creating a new Client.
type ClientConfig struct {
	BaseURL        string
	APIToken       string
	OrganizationID string
}

// NewClient creates a new Mihari API client.
func NewClient(config ClientConfig) *Client {
	return &Client{
		BaseURL:        strings.TrimRight(config.BaseURL, "/"),
		APIToken:       config.APIToken,
		OrganizationID: config.OrganizationID,
		HTTPClient:     &http.Client{Timeout: 30 * time.Second},
		UserAgent:      "terraform-provider-mihari",
	}
}

// APIError represents an error response from the Mihari API.
type APIError struct {
	StatusCode int
	Message    string
	Errors     map[string][]string `json:"errors,omitempty"`
}

func (e *APIError) Error() string {
	if len(e.Errors) > 0 {
		var parts []string
		for field, msgs := range e.Errors {
			for _, msg := range msgs {
				parts = append(parts, fmt.Sprintf("%s: %s", field, msg))
			}
		}
		return fmt.Sprintf("API error (HTTP %d): %s - %s", e.StatusCode, e.Message, strings.Join(parts, "; "))
	}
	return fmt.Sprintf("API error (HTTP %d): %s", e.StatusCode, e.Message)
}

// IsNotFound returns true if the error is a 404 Not Found.
func (e *APIError) IsNotFound() bool {
	return e.StatusCode == 404
}

// apiErrorResponse is the JSON shape returned by Laravel on errors.
type apiErrorResponse struct {
	Message string              `json:"message"`
	Errors  map[string][]string `json:"errors,omitempty"`
}

// APIResponse wraps the standard Laravel {"data": ...} response.
type APIResponse[T any] struct {
	Data T `json:"data"`
}

// APIListResponse wraps paginated list responses.
type APIListResponse[T any] struct {
	Data  []T            `json:"data"`
	Links map[string]any `json:"links,omitempty"`
	Meta  map[string]any `json:"meta,omitempty"`
}

// doRequest executes an HTTP request against the Mihari API.
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, int, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	fullURL := fmt.Sprintf("%s/api/v1/%s", c.BaseURL, strings.TrimLeft(path, "/"))
	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, 0, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.APIToken)
	req.Header.Set("x-organization-id", c.OrganizationID)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		apiErr := &APIError{StatusCode: resp.StatusCode}
		var errResp apiErrorResponse
		if json.Unmarshal(respBody, &errResp) == nil {
			apiErr.Message = errResp.Message
			apiErr.Errors = errResp.Errors
		} else {
			apiErr.Message = string(respBody)
		}
		return nil, resp.StatusCode, apiErr
	}

	return respBody, resp.StatusCode, nil
}
