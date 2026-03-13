package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// Policy represents a policy response from the API.
type Policy struct {
	ID         string       `json:"id"`
	Name       string       `json:"name"`
	Type       string       `json:"type"`
	RetryCount int          `json:"retry_count"`
	RetryDelay int          `json:"retry_delay"`
	Steps      []PolicyStep `json:"steps"`
	CreatedAt  string       `json:"created_at"`
	UpdatedAt  string       `json:"updated_at"`
}

// PolicyStep represents a step within a policy.
type PolicyStep struct {
	ID               string             `json:"id"`
	WaitBefore       int                `json:"wait_before"`
	Order            int                `json:"order"`
	Call             bool               `json:"call"`
	PushNotification bool               `json:"push_notification"`
	SMS              bool               `json:"sms"`
	Email            bool               `json:"email"`
	Members          []PolicyStepMember `json:"members"`
}

// PolicyStepMember represents a member within a policy step.
type PolicyStepMember struct {
	ID   *string `json:"id"`
	Type string  `json:"type"`
}

// PolicyRequest is the request body for creating/updating a policy.
type PolicyRequest struct {
	Name       string              `json:"name"`
	Type       string              `json:"type,omitempty"`
	RetryCount int                 `json:"retry_count"`
	RetryDelay int                 `json:"retry_delay"`
	Steps      []PolicyStepRequest `json:"steps"`
}

// PolicyStepRequest is a step in a policy create/update request.
type PolicyStepRequest struct {
	WaitBefore       int                       `json:"wait_before"`
	Call             bool                      `json:"call"`
	PushNotification bool                      `json:"push_notification"`
	SMS              bool                      `json:"sms"`
	Email            bool                      `json:"email"`
	Members          []PolicyStepMemberRequest `json:"members"`
}

// PolicyStepMemberRequest is a member in a policy step request.
type PolicyStepMemberRequest struct {
	Type string  `json:"type"`
	ID   *string `json:"id"`
}

func (c *Client) GetPolicy(ctx context.Context, id string) (*Policy, error) {
	body, statusCode, err := c.doRequest(ctx, "GET", fmt.Sprintf("policies/%s?include=steps.members", id), nil)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.IsNotFound() {
			return nil, nil
		}
		return nil, err
	}
	if statusCode == 404 {
		return nil, nil
	}

	var resp APIResponse[Policy]
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing policy response: %w", err)
	}
	return &resp.Data, nil
}

func (c *Client) CreatePolicy(ctx context.Context, req PolicyRequest) (*Policy, error) {
	body, _, err := c.doRequest(ctx, "POST", "policies", req)
	if err != nil {
		return nil, err
	}

	var resp APIResponse[Policy]
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing policy create response: %w", err)
	}
	return &resp.Data, nil
}

func (c *Client) UpdatePolicy(ctx context.Context, id string, req PolicyRequest) (*Policy, error) {
	body, _, err := c.doRequest(ctx, "PUT", fmt.Sprintf("policies/%s", id), req)
	if err != nil {
		return nil, err
	}

	var resp APIResponse[Policy]
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing policy update response: %w", err)
	}
	return &resp.Data, nil
}

func (c *Client) DeletePolicy(ctx context.Context, id string) error {
	_, _, err := c.doRequest(ctx, "DELETE", fmt.Sprintf("policies/%s", id), nil)
	return err
}
