package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// Monitor represents a monitor response from the API.
type Monitor struct {
	ID                 string            `json:"id"`
	OrganizationID     string            `json:"organization_id"`
	Title              string            `json:"title"` // API returns "title" for the name field
	URL                string            `json:"url"`
	Type               string            `json:"type"`
	TypeLabel          string            `json:"type_label"`
	Keyword            string            `json:"keyword"`
	ExpectedStatusCode *int              `json:"expected_status_code"`
	Host               string            `json:"host"`
	Port               *int              `json:"port"`
	Protocol           string            `json:"protocol"`
	CheckInterval      int               `json:"check_interval"`
	Timeout            int               `json:"timeout"`
	Headers            map[string]string `json:"headers"`
	CheckSSL           bool              `json:"check_ssl"`
	IsActive           bool              `json:"is_active"`
	PolicyID           *string           `json:"policy_id"`
	Status             string            `json:"status"`
	LastCheckedAt      *string           `json:"last_checked_at"`
	CreatedAt          string            `json:"created_at"`
	UpdatedAt          string            `json:"updated_at"`
}

// MonitorRequest is the request body for creating/updating a monitor.
type MonitorRequest struct {
	Name               string            `json:"name"`
	URL                string            `json:"url,omitempty"`
	Type               string            `json:"type"`
	Keyword            string            `json:"keyword,omitempty"`
	ExpectedStatusCode *int              `json:"expected_status_code,omitempty"`
	Host               string            `json:"host,omitempty"`
	Port               *int              `json:"port,omitempty"`
	Protocol           string            `json:"protocol,omitempty"`
	CheckInterval      *int              `json:"check_interval,omitempty"`
	Timeout            *int              `json:"timeout,omitempty"`
	Headers            map[string]string `json:"headers,omitempty"`
	CheckSSL           *bool             `json:"check_ssl,omitempty"`
	IsActive           *bool             `json:"is_active,omitempty"`
	PolicyID           *string           `json:"policy_id,omitempty"`
}

func (c *Client) GetMonitor(ctx context.Context, id string) (*Monitor, error) {
	body, statusCode, err := c.doRequest(ctx, "GET", fmt.Sprintf("monitors/%s", id), nil)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.IsNotFound() {
			return nil, nil
		}
		return nil, err
	}
	if statusCode == 404 {
		return nil, nil
	}

	var resp APIResponse[Monitor]
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing monitor response: %w", err)
	}
	return &resp.Data, nil
}

func (c *Client) ListMonitors(ctx context.Context, filters map[string]string) ([]Monitor, error) {
	path := "monitors?per_page=100"
	for key, value := range filters {
		path += fmt.Sprintf("&filter[%s]=%s", key, value)
	}

	body, _, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var resp APIListResponse[Monitor]
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing monitors list response: %w", err)
	}
	return resp.Data, nil
}

func (c *Client) CreateMonitor(ctx context.Context, req MonitorRequest) (*Monitor, error) {
	body, _, err := c.doRequest(ctx, "POST", "monitors", req)
	if err != nil {
		return nil, err
	}

	var resp APIResponse[Monitor]
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing monitor create response: %w", err)
	}
	return &resp.Data, nil
}

func (c *Client) UpdateMonitor(ctx context.Context, id string, req MonitorRequest) (*Monitor, error) {
	body, _, err := c.doRequest(ctx, "PUT", fmt.Sprintf("monitors/%s", id), req)
	if err != nil {
		return nil, err
	}

	var resp APIResponse[Monitor]
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing monitor update response: %w", err)
	}
	return &resp.Data, nil
}

func (c *Client) DeleteMonitor(ctx context.Context, id string) error {
	_, _, err := c.doRequest(ctx, "DELETE", fmt.Sprintf("monitors/%s", id), nil)
	return err
}
