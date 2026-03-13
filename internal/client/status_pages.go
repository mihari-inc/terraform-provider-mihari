package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// StatusPage represents a status page response from the API.
type StatusPage struct {
	ID                        string            `json:"id"`
	OrganizationID            string            `json:"organization_id"`
	CompanyName               string            `json:"company_name"`
	Subdomain                 string            `json:"subdomain"`
	CustomDomain              *string           `json:"custom_domain"`
	CustomDomainVerifiedAt    *string           `json:"custom_domain_verified_at"`
	PasswordProtectionEnabled bool              `json:"password_protection_enabled"`
	IPAllowlistEnabled        bool              `json:"ip_allowlist_enabled"`
	IPAllowlist               []string          `json:"ip_allowlist"`
	Sections                  []StatusSection   `json:"sections"`
	CreatedAt                 string            `json:"created_at"`
	UpdatedAt                 string            `json:"updated_at"`
}

// StatusSection represents a section within a status page.
type StatusSection struct {
	ID        string                  `json:"id"`
	Name      string                  `json:"name"`
	Resources []StatusSectionResource `json:"resources"`
}

// StatusSectionResource represents a resource within a status page section.
type StatusSectionResource struct {
	ID           string  `json:"id"`
	ResourceID   string  `json:"resource_id"`
	ResourceType string  `json:"resource_type"`
	Title        string  `json:"title"`
	Description  *string `json:"description"`
}

// StatusPageRequest is the request body for creating a status page.
type StatusPageRequest struct {
	CompanyName               string                  `json:"company_name"`
	Subdomain                 string                  `json:"subdomain"`
	CustomDomain              *string                 `json:"custom_domain,omitempty"`
	PasswordProtectionEnabled *bool                   `json:"password_protection_enabled,omitempty"`
	Password                  *string                 `json:"password,omitempty"`
	IPAllowlistEnabled        *bool                   `json:"ip_allowlist_enabled,omitempty"`
	IPAllowlist               []string                `json:"ip_allowlist,omitempty"`
	Sections                  []StatusSectionRequest  `json:"sections,omitempty"`
}

// StatusSectionRequest is a section in a status page create request.
type StatusSectionRequest struct {
	Name      string                         `json:"name"`
	Resources []StatusSectionResourceRequest `json:"resources"`
}

// StatusSectionResourceRequest is a resource in a section request.
type StatusSectionResourceRequest struct {
	ResourceID   string  `json:"resource_id"`
	ResourceType string  `json:"resource_type"`
	Title        string  `json:"title"`
	Description  *string `json:"description,omitempty"`
}

func (c *Client) GetStatusPage(ctx context.Context, id string) (*StatusPage, error) {
	body, statusCode, err := c.doRequest(ctx, "GET", fmt.Sprintf("status-pages/%s?include=sections.resourcesSections", id), nil)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.IsNotFound() {
			return nil, nil
		}
		return nil, err
	}
	if statusCode == 404 {
		return nil, nil
	}

	var resp APIResponse[StatusPage]
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing status page response: %w", err)
	}
	return &resp.Data, nil
}

func (c *Client) CreateStatusPage(ctx context.Context, req StatusPageRequest) (*StatusPage, error) {
	body, _, err := c.doRequest(ctx, "POST", "status-pages", req)
	if err != nil {
		return nil, err
	}

	var resp APIResponse[StatusPage]
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing status page create response: %w", err)
	}
	return &resp.Data, nil
}

func (c *Client) UpdateStatusPage(ctx context.Context, id string, req StatusPageRequest) (*StatusPage, error) {
	body, _, err := c.doRequest(ctx, "PUT", fmt.Sprintf("status-pages/%s", id), req)
	if err != nil {
		return nil, err
	}

	var resp APIResponse[StatusPage]
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing status page update response: %w", err)
	}
	return &resp.Data, nil
}

func (c *Client) DeleteStatusPage(ctx context.Context, id string) error {
	_, _, err := c.doRequest(ctx, "DELETE", fmt.Sprintf("status-pages/%s", id), nil)
	return err
}
