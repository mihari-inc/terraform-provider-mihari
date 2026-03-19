terraform {
  required_providers {
    mihari = {
      source  = "mihari-io/mihari"
      version = "~> 0.1"
    }
  }
}

provider "mihari" {
  api_url         = var.mihari_api_url
  api_token       = var.mihari_api_token
  organization_id = var.mihari_organization_id
}

variable "mihari_api_url" {
  description = "Mihari API base URL"
  type        = string
  default     = "https://platform.mihari.io"
}

variable "mihari_api_token" {
  description = "Mihari API token"
  type        = string
  sensitive   = true
}

variable "mihari_organization_id" {
  description = "Mihari organization UUID"
  type        = string
}

# ─── Alert Policy ────────────────────────────────────────────────────────────

resource "mihari_policy" "critical" {
  name        = "Critical Alert Policy"
  type        = "template"
  retry_count = 3
  retry_delay = 5

  steps = [
    {
      wait_before       = 0
      call              = false
      push_notification = true
      sms               = false
      email             = true
      members = [
        { type = "current_persons_on_call" }
      ]
    },
    {
      wait_before       = 5
      call              = true
      push_notification = true
      sms               = true
      email             = true
      members = [
        {
          type = "user"
          id   = "user-uuid-here"
        }
      ]
    }
  ]
}

# ─── Monitors ────────────────────────────────────────────────────────────────

resource "mihari_monitor" "api_health" {
  name           = "API Health Check"
  type           = "http_status"
  url            = "https://api.example.com/health"
  protocol       = "https"
  check_interval = 3
  timeout        = 30
  check_ssl      = true
  is_active      = true
  policy_id      = mihari_policy.critical.id

  headers = {
    "X-Custom-Header" = "monitoring"
  }
}

resource "mihari_monitor" "website_keyword" {
  name           = "Website Contains Login"
  type           = "url_contains"
  url            = "https://www.example.com"
  keyword        = "Sign In"
  check_interval = 5
  timeout        = 30
  is_active      = true
  policy_id      = mihari_policy.critical.id
}

resource "mihari_monitor" "tcp_database" {
  name           = "Database TCP Check"
  type           = "tcp_port"
  host           = "db.internal.example.com"
  port           = 5432
  check_interval = 2
  timeout        = 10
  is_active      = true
  policy_id      = mihari_policy.critical.id
}

resource "mihari_monitor" "dns_check" {
  name           = "DNS Resolution Check"
  type           = "dns"
  host           = "example.com"
  check_interval = 10
  timeout        = 15
  is_active      = true
}

resource "mihari_monitor" "ping_check" {
  name           = "Server Ping"
  type           = "ping"
  host           = "10.0.1.50"
  check_interval = 1
  timeout        = 10
  is_active      = true
}

# ─── Heartbeat ───────────────────────────────────────────────────────────────

resource "mihari_heartbeat" "nightly_backup" {
  name         = "Nightly Backup Job"
  period       = 1440
  grace_period = 30
  is_active    = true
  policy_id    = mihari_policy.critical.id
}

resource "mihari_heartbeat" "cron_sync" {
  name         = "Data Sync Cron"
  period       = 60
  grace_period = 10
  is_active    = true
}

# ─── Status Page ─────────────────────────────────────────────────────────────

resource "mihari_status_page" "public" {
  company_name                = "Example Corp"
  subdomain                   = "status-example"
  password_protection_enabled = false
  ip_allowlist_enabled        = false

  sections = [
    {
      name = "Core Services"
      resources = [
        {
          resource_id   = mihari_monitor.api_health.id
          resource_type = "monitor"
          title         = "API"
          description   = "Main API endpoint"
        },
        {
          resource_id   = mihari_monitor.website_keyword.id
          resource_type = "monitor"
          title         = "Website"
          description   = "Public website"
        }
      ]
    },
    {
      name = "Infrastructure"
      resources = [
        {
          resource_id   = mihari_monitor.tcp_database.id
          resource_type = "monitor"
          title         = "Database"
          description   = "PostgreSQL primary"
        }
      ]
    }
  ]
}

# ─── On-Call ─────────────────────────────────────────────────────────────────

resource "mihari_on_call_calendar" "engineering" {
  name        = "Engineering On-Call"
  description = "Primary engineering on-call rotation"
  is_active   = true
}

# ─── Data Sources ────────────────────────────────────────────────────────────

data "mihari_monitors" "active_http" {
  type      = "http_status"
  is_active = true
}

output "active_http_monitors" {
  value = data.mihari_monitors.active_http.monitors[*].name
}
