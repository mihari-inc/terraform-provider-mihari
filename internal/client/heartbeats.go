package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// Heartbeat represents a heartbeat response from the API.
type Heartbeat struct {
	ID             string  `json:"id"`
	OrganizationID string  `json:"organization_id"`
	Name           string  `json:"name"`
	Period         int     `json:"period"`
	GracePeriod    int     `json:"grace_period"`
	IsActive       bool    `json:"is_active"`
	PolicyID       *string `json:"policy_id"`
	Status         string  `json:"status"`
	LastPingAt     *string `json:"last_ping_at"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

// HeartbeatRequest is the request body for creating/updating a heartbeat.
type HeartbeatRequest struct {
	Name        string  `json:"name"`
	Period      int     `json:"period"`
	GracePeriod int     `json:"grace_period"`
	IsActive    *bool   `json:"is_active,omitempty"`
	PolicyID    *string `json:"policy_id,omitempty"`
}

func (c *Client) GetHeartbeat(ctx context.Context, id string) (*Heartbeat, error) {
	body, statusCode, err := c.doRequest(ctx, "GET", fmt.Sprintf("heartbeats/%s", id), nil)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.IsNotFound() {
			return nil, nil
		}
		return nil, err
	}
	if statusCode == 404 {
		return nil, nil
	}

	var resp APIResponse[Heartbeat]
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing heartbeat response: %w", err)
	}
	return &resp.Data, nil
}

func (c *Client) CreateHeartbeat(ctx context.Context, req HeartbeatRequest) (*Heartbeat, error) {
	body, _, err := c.doRequest(ctx, "POST", "heartbeats", req)
	if err != nil {
		return nil, err
	}

	var resp APIResponse[Heartbeat]
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing heartbeat create response: %w", err)
	}
	return &resp.Data, nil
}

func (c *Client) UpdateHeartbeat(ctx context.Context, id string, req HeartbeatRequest) (*Heartbeat, error) {
	body, _, err := c.doRequest(ctx, "PUT", fmt.Sprintf("heartbeats/%s", id), req)
	if err != nil {
		return nil, err
	}

	var resp APIResponse[Heartbeat]
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing heartbeat update response: %w", err)
	}
	return &resp.Data, nil
}

func (c *Client) DeleteHeartbeat(ctx context.Context, id string) error {
	_, _, err := c.doRequest(ctx, "DELETE", fmt.Sprintf("heartbeats/%s", id), nil)
	return err
}
