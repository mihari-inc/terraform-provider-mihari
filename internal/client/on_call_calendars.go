package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// OnCallCalendar represents an on-call calendar response from the API.
type OnCallCalendar struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organization_id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	IsActive       bool   `json:"is_active"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

// OnCallCalendarRequest is the request body for creating/updating an on-call calendar.
type OnCallCalendarRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	IsActive    *bool  `json:"is_active,omitempty"`
}

func (c *Client) GetOnCallCalendar(ctx context.Context, id string) (*OnCallCalendar, error) {
	body, statusCode, err := c.doRequest(ctx, "GET", fmt.Sprintf("on-call-calendars/%s", id), nil)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.IsNotFound() {
			return nil, nil
		}
		return nil, err
	}
	if statusCode == 404 {
		return nil, nil
	}

	var resp APIResponse[OnCallCalendar]
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing on-call calendar response: %w", err)
	}
	return &resp.Data, nil
}

func (c *Client) CreateOnCallCalendar(ctx context.Context, req OnCallCalendarRequest) (*OnCallCalendar, error) {
	body, _, err := c.doRequest(ctx, "POST", "on-call-calendars", req)
	if err != nil {
		return nil, err
	}

	var resp APIResponse[OnCallCalendar]
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing on-call calendar create response: %w", err)
	}
	return &resp.Data, nil
}

func (c *Client) UpdateOnCallCalendar(ctx context.Context, id string, req OnCallCalendarRequest) (*OnCallCalendar, error) {
	body, _, err := c.doRequest(ctx, "PUT", fmt.Sprintf("on-call-calendars/%s", id), req)
	if err != nil {
		return nil, err
	}

	var resp APIResponse[OnCallCalendar]
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing on-call calendar update response: %w", err)
	}
	return &resp.Data, nil
}

func (c *Client) DeleteOnCallCalendar(ctx context.Context, id string) error {
	_, _, err := c.doRequest(ctx, "DELETE", fmt.Sprintf("on-call-calendars/%s", id), nil)
	return err
}
