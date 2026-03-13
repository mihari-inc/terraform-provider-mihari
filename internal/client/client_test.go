package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	c := NewClient(ClientConfig{
		BaseURL:        "https://api.mihari.io/",
		APIToken:       "test-token",
		OrganizationID: "org-123",
	})

	if c.BaseURL != "https://api.mihari.io" {
		t.Errorf("expected trailing slash to be trimmed, got %s", c.BaseURL)
	}
	if c.APIToken != "test-token" {
		t.Errorf("expected token test-token, got %s", c.APIToken)
	}
	if c.OrganizationID != "org-123" {
		t.Errorf("expected org org-123, got %s", c.OrganizationID)
	}
}

func TestDoRequest_AuthHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("expected Authorization header, got %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("x-organization-id") != "org-123" {
			t.Errorf("expected x-organization-id header, got %s", r.Header.Get("x-organization-id"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("expected Accept application/json, got %s", r.Header.Get("Accept"))
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"data":{}}`))
	}))
	defer server.Close()

	c := NewClient(ClientConfig{
		BaseURL:        server.URL,
		APIToken:       "test-token",
		OrganizationID: "org-123",
	})

	_, _, err := c.doRequest(context.Background(), "GET", "test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDoRequest_404ReturnsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(map[string]string{"message": "Not found"})
	}))
	defer server.Close()

	c := NewClient(ClientConfig{BaseURL: server.URL, APIToken: "t", OrganizationID: "o"})

	_, _, err := c.doRequest(context.Background(), "GET", "test", nil)
	if err == nil {
		t.Fatal("expected error for 404")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if !apiErr.IsNotFound() {
		t.Errorf("expected IsNotFound() true, got false")
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("expected status 404, got %d", apiErr.StatusCode)
	}
}

func TestDoRequest_422ValidationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(422)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "The given data was invalid.",
			"errors": map[string][]string{
				"name": {"The name field is required."},
				"type": {"The type field is required."},
			},
		})
	}))
	defer server.Close()

	c := NewClient(ClientConfig{BaseURL: server.URL, APIToken: "t", OrganizationID: "o"})

	_, _, err := c.doRequest(context.Background(), "POST", "test", map[string]string{})
	if err == nil {
		t.Fatal("expected error for 422")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 422 {
		t.Errorf("expected status 422, got %d", apiErr.StatusCode)
	}
	if len(apiErr.Errors) != 2 {
		t.Errorf("expected 2 validation errors, got %d", len(apiErr.Errors))
	}
}

func TestDoRequest_500ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(map[string]string{"message": "Internal Server Error"})
	}))
	defer server.Close()

	c := NewClient(ClientConfig{BaseURL: server.URL, APIToken: "t", OrganizationID: "o"})

	_, _, err := c.doRequest(context.Background(), "GET", "test", nil)
	if err == nil {
		t.Fatal("expected error for 500")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 500 {
		t.Errorf("expected status 500, got %d", apiErr.StatusCode)
	}
}

// --- Monitor Tests ---

func TestGetMonitor_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/monitors/mon-123" {
			t.Errorf("expected path /api/v1/monitors/mon-123, got %s", r.URL.Path)
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id":             "mon-123",
				"title":          "My Monitor",
				"type":           "http_status",
				"url":            "https://example.com",
				"check_interval": 5,
				"timeout":        30,
				"is_active":      true,
				"check_ssl":      true,
				"status":         "up",
				"created_at":     "2024-01-01T00:00:00Z",
				"updated_at":     "2024-01-01T00:00:00Z",
			},
		})
	}))
	defer server.Close()

	c := NewClient(ClientConfig{BaseURL: server.URL, APIToken: "t", OrganizationID: "o"})

	monitor, err := c.GetMonitor(context.Background(), "mon-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if monitor == nil {
		t.Fatal("expected monitor, got nil")
	}
	if monitor.ID != "mon-123" {
		t.Errorf("expected ID mon-123, got %s", monitor.ID)
	}
	if monitor.Title != "My Monitor" {
		t.Errorf("expected title My Monitor, got %s", monitor.Title)
	}
	if monitor.Type != "http_status" {
		t.Errorf("expected type http_status, got %s", monitor.Type)
	}
	if monitor.CheckInterval != 5 {
		t.Errorf("expected check_interval 5, got %d", monitor.CheckInterval)
	}
}

func TestGetMonitor_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(map[string]string{"message": "Not found"})
	}))
	defer server.Close()

	c := NewClient(ClientConfig{BaseURL: server.URL, APIToken: "t", OrganizationID: "o"})

	monitor, err := c.GetMonitor(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("expected nil error for 404, got: %v", err)
	}
	if monitor != nil {
		t.Errorf("expected nil monitor for 404, got: %+v", monitor)
	}
}

func TestCreateMonitor_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		if body["name"] != "Test Monitor" {
			t.Errorf("expected name Test Monitor, got %v", body["name"])
		}
		if body["type"] != "http_status" {
			t.Errorf("expected type http_status, got %v", body["type"])
		}

		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id":             "new-mon-id",
				"title":          "Test Monitor",
				"type":           "http_status",
				"url":            "https://example.com",
				"check_interval": 5,
				"timeout":        30,
				"is_active":      true,
				"check_ssl":      false,
				"status":         "up",
				"created_at":     "2024-01-01T00:00:00Z",
				"updated_at":     "2024-01-01T00:00:00Z",
			},
		})
	}))
	defer server.Close()

	c := NewClient(ClientConfig{BaseURL: server.URL, APIToken: "t", OrganizationID: "o"})

	interval := 5
	timeout := 30
	monitor, err := c.CreateMonitor(context.Background(), MonitorRequest{
		Name:          "Test Monitor",
		Type:          "http_status",
		URL:           "https://example.com",
		CheckInterval: &interval,
		Timeout:       &timeout,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if monitor.ID != "new-mon-id" {
		t.Errorf("expected ID new-mon-id, got %s", monitor.ID)
	}
}

func TestDeleteMonitor_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(204)
	}))
	defer server.Close()

	c := NewClient(ClientConfig{BaseURL: server.URL, APIToken: "t", OrganizationID: "o"})

	err := c.DeleteMonitor(context.Background(), "mon-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- Heartbeat Tests ---

func TestGetHeartbeat_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id":           "hb-123",
				"name":         "Nightly Backup",
				"period":       1440,
				"grace_period": 30,
				"is_active":    true,
				"status":       "up",
				"created_at":   "2024-01-01T00:00:00Z",
				"updated_at":   "2024-01-01T00:00:00Z",
			},
		})
	}))
	defer server.Close()

	c := NewClient(ClientConfig{BaseURL: server.URL, APIToken: "t", OrganizationID: "o"})

	hb, err := c.GetHeartbeat(context.Background(), "hb-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hb.Name != "Nightly Backup" {
		t.Errorf("expected name Nightly Backup, got %s", hb.Name)
	}
	if hb.Period != 1440 {
		t.Errorf("expected period 1440, got %d", hb.Period)
	}
}

// --- Policy Tests ---

func TestGetPolicy_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id":          "pol-123",
				"name":        "Critical Policy",
				"type":        "template",
				"retry_count": 3,
				"retry_delay": 5,
				"steps": []map[string]interface{}{
					{
						"id":                "step-1",
						"wait_before":       0,
						"call":              false,
						"push_notification": true,
						"sms":               false,
						"email":             true,
						"members": []map[string]interface{}{
							{"type": "current_persons_on_call", "id": nil},
						},
					},
				},
				"created_at": "2024-01-01T00:00:00Z",
				"updated_at": "2024-01-01T00:00:00Z",
			},
		})
	}))
	defer server.Close()

	c := NewClient(ClientConfig{BaseURL: server.URL, APIToken: "t", OrganizationID: "o"})

	policy, err := c.GetPolicy(context.Background(), "pol-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.Name != "Critical Policy" {
		t.Errorf("expected name Critical Policy, got %s", policy.Name)
	}
	if len(policy.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(policy.Steps))
	}
	if len(policy.Steps[0].Members) != 1 {
		t.Fatalf("expected 1 member, got %d", len(policy.Steps[0].Members))
	}
	if policy.Steps[0].Members[0].Type != "current_persons_on_call" {
		t.Errorf("expected member type current_persons_on_call, got %s", policy.Steps[0].Members[0].Type)
	}
}

func TestCreatePolicy_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		if body["name"] != "New Policy" {
			t.Errorf("expected name New Policy, got %v", body["name"])
		}

		steps, ok := body["steps"].([]interface{})
		if !ok || len(steps) != 1 {
			t.Fatalf("expected 1 step in request")
		}

		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id":          "pol-new",
				"name":        "New Policy",
				"type":        "template",
				"retry_count": 3,
				"retry_delay": 5,
				"steps":       []map[string]interface{}{},
				"created_at":  "2024-01-01T00:00:00Z",
				"updated_at":  "2024-01-01T00:00:00Z",
			},
		})
	}))
	defer server.Close()

	c := NewClient(ClientConfig{BaseURL: server.URL, APIToken: "t", OrganizationID: "o"})

	policy, err := c.CreatePolicy(context.Background(), PolicyRequest{
		Name:       "New Policy",
		Type:       "template",
		RetryCount: 3,
		RetryDelay: 5,
		Steps: []PolicyStepRequest{
			{
				WaitBefore:       0,
				Call:             false,
				PushNotification: true,
				SMS:              false,
				Email:            true,
				Members: []PolicyStepMemberRequest{
					{Type: "current_persons_on_call"},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.ID != "pol-new" {
		t.Errorf("expected ID pol-new, got %s", policy.ID)
	}
}

// --- StatusPage Tests ---

func TestGetStatusPage_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id":                          "sp-123",
				"company_name":                "Acme Corp",
				"subdomain":                   "status-acme",
				"password_protection_enabled": false,
				"ip_allowlist_enabled":        false,
				"ip_allowlist":                []string{},
				"sections": []map[string]interface{}{
					{
						"id":   "sec-1",
						"name": "Core Services",
						"resources": []map[string]interface{}{
							{
								"id":            "res-1",
								"resource_id":   "mon-123",
								"resource_type": "monitor",
								"title":         "API",
								"description":   "Main API",
							},
						},
					},
				},
				"created_at": "2024-01-01T00:00:00Z",
				"updated_at": "2024-01-01T00:00:00Z",
			},
		})
	}))
	defer server.Close()

	c := NewClient(ClientConfig{BaseURL: server.URL, APIToken: "t", OrganizationID: "o"})

	page, err := c.GetStatusPage(context.Background(), "sp-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.CompanyName != "Acme Corp" {
		t.Errorf("expected company Acme Corp, got %s", page.CompanyName)
	}
	if len(page.Sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(page.Sections))
	}
	if len(page.Sections[0].Resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(page.Sections[0].Resources))
	}
}

// --- OnCallCalendar Tests ---

func TestGetOnCallCalendar_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id":          "cal-123",
				"name":        "Engineering On-Call",
				"description": "Primary rotation",
				"is_active":   true,
				"created_at":  "2024-01-01T00:00:00Z",
				"updated_at":  "2024-01-01T00:00:00Z",
			},
		})
	}))
	defer server.Close()

	c := NewClient(ClientConfig{BaseURL: server.URL, APIToken: "t", OrganizationID: "o"})

	cal, err := c.GetOnCallCalendar(context.Background(), "cal-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cal.Name != "Engineering On-Call" {
		t.Errorf("expected name Engineering On-Call, got %s", cal.Name)
	}
}

// --- OnCallRotation Tests ---

func TestCreateOnCallRotation_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		if body["on_call_calendar_id"] != "cal-123" {
			t.Errorf("expected calendar_id cal-123, got %v", body["on_call_calendar_id"])
		}

		event, ok := body["event"].(map[string]interface{})
		if !ok {
			t.Fatal("expected event in request body")
		}
		start, _ := event["start"].(map[string]interface{})
		if start["date"] != "2024-01-15" {
			t.Errorf("expected start date 2024-01-15, got %v", start["date"])
		}

		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id":                  "rot-new",
				"on_call_calendar_id": "cal-123",
				"rrule":               "FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR;UNTIL=20241231",
				"start_date":          "2024-01-15T09:00:00Z",
				"end_date":            "2024-12-31T09:00:00Z",
				"duration":            480,
				"members":             []map[string]interface{}{},
				"created_at":          "2024-01-01T00:00:00Z",
				"updated_at":          "2024-01-01T00:00:00Z",
			},
		})
	}))
	defer server.Close()

	c := NewClient(ClientConfig{BaseURL: server.URL, APIToken: "t", OrganizationID: "o"})

	rotation, err := c.CreateOnCallRotation(context.Background(), OnCallRotationRequest{
		OnCallCalendarID: "cal-123",
		Event: OnCallRotationEventRequest{
			Start:    OnCallRotationTimeRequest{Date: "2024-01-15", Hour: "09:00"},
			Duration: "08:00",
		},
		Repeat: OnCallRotationRepeatRequest{
			Days: []int{1, 2, 3, 4, 5},
			End:  "2024-12-31",
		},
		Members: []OnCallRotationMemberRequest{
			{ID: "mem-1", Value: "mem-1", Label: "Member 1", Type: "user"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rotation.ID != "rot-new" {
		t.Errorf("expected ID rot-new, got %s", rotation.ID)
	}
}

// --- APIError Tests ---

func TestAPIError_Error(t *testing.T) {
	err := &APIError{
		StatusCode: 422,
		Message:    "Validation failed",
		Errors: map[string][]string{
			"name": {"required"},
		},
	}

	errStr := err.Error()
	if errStr == "" {
		t.Error("expected non-empty error string")
	}
}

func TestAPIError_IsNotFound(t *testing.T) {
	err404 := &APIError{StatusCode: 404}
	err500 := &APIError{StatusCode: 500}

	if !err404.IsNotFound() {
		t.Error("expected 404 to be not found")
	}
	if err500.IsNotFound() {
		t.Error("expected 500 to not be not found")
	}
}
