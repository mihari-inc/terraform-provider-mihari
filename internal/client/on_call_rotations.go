package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// OnCallRotation represents an on-call rotation response from the API.
type OnCallRotation struct {
	ID               string                 `json:"id"`
	OnCallCalendarID string                 `json:"on_call_calendar_id"`
	Rrule            string                 `json:"rrule"`
	StartDate        string                 `json:"start_date"`
	EndDate          string                 `json:"end_date"`
	Duration         int                    `json:"duration"`
	Members          []OnCallRotationMember `json:"members"`
	CreatedAt        string                 `json:"created_at"`
	UpdatedAt        string                 `json:"updated_at"`
}

// OnCallRotationMember represents a member in an on-call rotation.
type OnCallRotationMember struct {
	ID       string `json:"id"`
	MemberID string `json:"member_id"`
	Position int    `json:"position"`
	IsActive bool   `json:"is_active"`
}

// OnCallRotationRequest is the request body for creating/updating a rotation.
// Uses the UI-format that the API expects (event/repeat structure).
type OnCallRotationRequest struct {
	OnCallCalendarID string                        `json:"on_call_calendar_id"`
	Event            OnCallRotationEventRequest    `json:"event"`
	Repeat           OnCallRotationRepeatRequest   `json:"repeat"`
	Members          []OnCallRotationMemberRequest `json:"members"`
}

// OnCallRotationEventRequest represents the event part of a rotation request.
type OnCallRotationEventRequest struct {
	Start    OnCallRotationTimeRequest `json:"start"`
	Duration string                    `json:"duration"`
}

// OnCallRotationTimeRequest represents a time within the event.
type OnCallRotationTimeRequest struct {
	Date string `json:"date"`
	Hour string `json:"hour"`
}

// OnCallRotationRepeatRequest represents the repeat configuration.
type OnCallRotationRepeatRequest struct {
	Days []int  `json:"days"`
	End  string `json:"end"`
}

// OnCallRotationMemberRequest represents a member in a rotation request.
type OnCallRotationMemberRequest struct {
	ID    string `json:"id"`
	Value string `json:"value"`
	Label string `json:"label"`
	Type  string `json:"type"`
}

func (c *Client) GetOnCallRotation(ctx context.Context, id string) (*OnCallRotation, error) {
	body, statusCode, err := c.doRequest(ctx, "GET", fmt.Sprintf("on-call-rotations/%s?include=members", id), nil)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.IsNotFound() {
			return nil, nil
		}
		return nil, err
	}
	if statusCode == 404 {
		return nil, nil
	}

	var resp APIResponse[OnCallRotation]
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing on-call rotation response: %w", err)
	}
	return &resp.Data, nil
}

func (c *Client) CreateOnCallRotation(ctx context.Context, req OnCallRotationRequest) (*OnCallRotation, error) {
	body, _, err := c.doRequest(ctx, "POST", "on-call-rotations", req)
	if err != nil {
		return nil, err
	}

	var resp APIResponse[OnCallRotation]
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing on-call rotation create response: %w", err)
	}
	return &resp.Data, nil
}

func (c *Client) UpdateOnCallRotation(ctx context.Context, id string, req OnCallRotationRequest) (*OnCallRotation, error) {
	body, _, err := c.doRequest(ctx, "PUT", fmt.Sprintf("on-call-rotations/%s", id), req)
	if err != nil {
		return nil, err
	}

	var resp APIResponse[OnCallRotation]
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing on-call rotation update response: %w", err)
	}
	return &resp.Data, nil
}

func (c *Client) DeleteOnCallRotation(ctx context.Context, id string) error {
	_, _, err := c.doRequest(ctx, "DELETE", fmt.Sprintf("on-call-rotations/%s", id), nil)
	return err
}
