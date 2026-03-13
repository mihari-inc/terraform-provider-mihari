# Examples

## Complete Monitoring Setup

Full example with monitors, alert policy, status page, and on-call schedule.

```hcl
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

# ──────────────────────────────────────────────
# Variables
# ──────────────────────────────────────────────

variable "mihari_api_url" {
  type    = string
  default = "https://app.mihari.io"
}

variable "mihari_api_token" {
  type      = string
  sensitive = true
}

variable "mihari_organization_id" {
  type = string
}

# ──────────────────────────────────────────────
# Alert Policy
# ──────────────────────────────────────────────

resource "mihari_policy" "production" {
  name        = "Production Alert Policy"
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
      wait_before       = 10
      call              = true
      push_notification = true
      sms               = true
      email             = true
      members = [
        { type = "current_persons_on_call" }
      ]
    }
  ]
}

# ──────────────────────────────────────────────
# Monitors
# ──────────────────────────────────────────────

resource "mihari_monitor" "api" {
  name           = "Production API"
  type           = "http_status"
  url            = "https://api.myapp.com/health"
  check_interval = 3
  timeout        = 30
  check_ssl      = true
  is_active      = true
  policy_id      = mihari_policy.production.id
}

resource "mihari_monitor" "website" {
  name           = "Marketing Website"
  type           = "url_contains"
  url            = "https://www.myapp.com"
  keyword        = "Welcome"
  check_interval = 5
  timeout        = 30
  is_active      = true
  policy_id      = mihari_policy.production.id
}

resource "mihari_monitor" "database" {
  name           = "PostgreSQL Primary"
  type           = "tcp_port"
  host           = "db-primary.internal"
  port           = 5432
  check_interval = 2
  timeout        = 10
  is_active      = true
  policy_id      = mihari_policy.production.id
}

resource "mihari_monitor" "redis" {
  name           = "Redis Cache"
  type           = "tcp_port"
  host           = "redis.internal"
  port           = 6379
  check_interval = 2
  timeout        = 10
  is_active      = true
  policy_id      = mihari_policy.production.id
}

resource "mihari_monitor" "dns" {
  name           = "DNS Resolution"
  type           = "dns"
  host           = "myapp.com"
  check_interval = 10
  timeout        = 15
  is_active      = true
}

# ──────────────────────────────────────────────
# Heartbeats
# ──────────────────────────────────────────────

resource "mihari_heartbeat" "backup" {
  name         = "Nightly Database Backup"
  period       = 1440    # 24 hours
  grace_period = 60      # 1 hour grace
  is_active    = true
  policy_id    = mihari_policy.production.id
}

resource "mihari_heartbeat" "email_queue" {
  name         = "Email Queue Processor"
  period       = 5       # Every 5 minutes
  grace_period = 10
  is_active    = true
  policy_id    = mihari_policy.production.id
}

# ──────────────────────────────────────────────
# Status Page
# ──────────────────────────────────────────────

resource "mihari_status_page" "public" {
  company_name                = "MyApp"
  subdomain                   = "status-myapp"
  password_protection_enabled = false
  ip_allowlist_enabled        = false

  sections = [
    {
      name = "Web Services"
      resources = [
        {
          resource_id   = mihari_monitor.api.id
          resource_type = "monitor"
          title         = "API"
          description   = "REST API endpoints"
        },
        {
          resource_id   = mihari_monitor.website.id
          resource_type = "monitor"
          title         = "Website"
          description   = "Public marketing website"
        }
      ]
    },
    {
      name = "Infrastructure"
      resources = [
        {
          resource_id   = mihari_monitor.database.id
          resource_type = "monitor"
          title         = "Database"
          description   = "Primary PostgreSQL"
        },
        {
          resource_id   = mihari_monitor.redis.id
          resource_type = "monitor"
          title         = "Cache"
          description   = "Redis cache cluster"
        }
      ]
    },
    {
      name = "Background Jobs"
      resources = [
        {
          resource_id   = mihari_heartbeat.backup.id
          resource_type = "heartbeat"
          title         = "Backups"
          description   = "Nightly database backups"
        },
        {
          resource_id   = mihari_heartbeat.email_queue.id
          resource_type = "heartbeat"
          title         = "Email Queue"
          description   = "Email processing pipeline"
        }
      ]
    }
  ]
}

# ──────────────────────────────────────────────
# On-Call
# ──────────────────────────────────────────────

resource "mihari_on_call_calendar" "engineering" {
  name        = "Engineering On-Call"
  description = "24/7 engineering rotation"
  is_active   = true
}

# ──────────────────────────────────────────────
# Outputs
# ──────────────────────────────────────────────

output "status_page_url" {
  value = "https://status-myapp.mihari.io"
}

output "monitor_ids" {
  value = {
    api      = mihari_monitor.api.id
    website  = mihari_monitor.website.id
    database = mihari_monitor.database.id
    redis    = mihari_monitor.redis.id
    dns      = mihari_monitor.dns.id
  }
}
```

---

## Using Data Sources to Reference Existing Resources

```hcl
# Read monitors created outside Terraform
data "mihari_monitors" "all_active" {
  is_active = true
}

# Read a specific policy
data "mihari_policy" "default" {
  id = "existing-policy-uuid"
}

# Create a new monitor using the existing policy
resource "mihari_monitor" "new_service" {
  name           = "New Microservice"
  type           = "http_status"
  url            = "https://new-service.internal/health"
  check_interval = 3
  timeout        = 15
  policy_id      = data.mihari_policy.default.id
}
```

---

## Multi-Environment Setup

Use Terraform workspaces or variable files for different environments.

### variables.tf

```hcl
variable "environment" {
  type = string
}

variable "api_base_url" {
  type = string
}

variable "check_interval" {
  type    = number
  default = 5
}
```

### environments/production.tfvars

```hcl
environment    = "production"
api_base_url   = "https://api.myapp.com"
check_interval = 2
```

### environments/staging.tfvars

```hcl
environment    = "staging"
api_base_url   = "https://api.staging.myapp.com"
check_interval = 10
```

### main.tf

```hcl
resource "mihari_monitor" "api" {
  name           = "${var.environment} - API Health"
  type           = "http_status"
  url            = "${var.api_base_url}/health"
  check_interval = var.check_interval
  timeout        = 30
  is_active      = true
}
```

### Usage

```bash
terraform apply -var-file=environments/production.tfvars
terraform apply -var-file=environments/staging.tfvars
```

---

## Private Status Page with IP Restriction

```hcl
resource "mihari_status_page" "internal" {
  company_name                = "MyApp Internal"
  subdomain                   = "internal-status"
  password_protection_enabled = true
  password                    = var.status_page_password
  ip_allowlist_enabled        = true
  ip_allowlist = [
    "10.0.0.0/8",
    "172.16.0.0/12",
    "192.168.0.0/16",
    "203.0.113.50/32"   # Office IP
  ]

  sections = [
    {
      name = "All Services"
      resources = [
        {
          resource_id   = mihari_monitor.api.id
          resource_type = "monitor"
          title         = "API"
        }
      ]
    }
  ]
}
```
