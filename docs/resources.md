# Resources Reference

## mihari_monitor

Manages a Mihari monitor for checking website, API, port, DNS, or ping availability.

### Example

```hcl
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
    "Authorization" = "Bearer monitoring-token"
    "X-Source"      = "terraform"
  }
}
```

### Attributes

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | Display name of the monitor |
| `type` | string | Yes | Monitor type (see below). **Changing this recreates the resource.** |
| `url` | string | Conditional | URL to monitor. Required for `url_contains`, `url_not_contains`, `http_status` |
| `keyword` | string | Conditional | Keyword to search. Required for `url_contains`, `url_not_contains` |
| `expected_status_code` | number | No | Expected HTTP status code (100-599) |
| `host` | string | Conditional | Host to check. Required for `ping`, `tcp_port`, `udp_port`, `smtp`, `pop3`, `imap`, `dns` |
| `port` | number | Conditional | Port number (1-65535). Required for `tcp_port`, `udp_port`, `smtp`, `pop3`, `imap` |
| `protocol` | string | No | `http` or `https` |
| `check_interval` | number | No | Check frequency in minutes (1-1440) |
| `timeout` | number | No | Timeout in seconds (1-300) |
| `headers` | map(string) | No | Custom HTTP headers |
| `check_ssl` | bool | No | Validate SSL certificates |
| `is_active` | bool | No | Whether the monitor is active |
| `policy_id` | string | No | UUID of the alert policy to attach |

**Computed attributes**: `id`, `status`, `last_checked_at`, `created_at`, `updated_at`

### Monitor Types

| Type | Description | Required Fields |
|------|-------------|-----------------|
| `http_status` | Check HTTP response status code | `url` |
| `url_contains` | Check if page contains a keyword | `url`, `keyword` |
| `url_not_contains` | Check if page does NOT contain a keyword | `url`, `keyword` |
| `ping` | ICMP ping check | `host` |
| `tcp_port` | TCP port connectivity | `host`, `port` |
| `udp_port` | UDP port connectivity | `host`, `port` |
| `smtp` | SMTP server check | `host`, `port` |
| `pop3` | POP3 server check | `host`, `port` |
| `imap` | IMAP server check | `host`, `port` |
| `dns` | DNS resolution check | `host` |
| `playwright` | Browser-based check | `url` |

---

## mihari_heartbeat

Manages a heartbeat monitor for cron jobs and periodic tasks. Your application pings a Mihari endpoint periodically; if the ping is missed, an alert is triggered.

### Example

```hcl
resource "mihari_heartbeat" "nightly_backup" {
  name         = "Nightly Database Backup"
  period       = 1440   # Expected every 24 hours
  grace_period = 30     # 30 min grace before alerting
  is_active    = true
  policy_id    = mihari_policy.critical.id
}
```

### Attributes

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | Display name |
| `period` | number | Yes | Expected interval between pings (minutes, >= 1) |
| `grace_period` | number | Yes | Grace period before alerting (minutes, >= 1) |
| `is_active` | bool | No | Whether the heartbeat is active |
| `policy_id` | string | No | UUID of the alert policy |

**Computed**: `id`, `status`, `last_ping_at`, `created_at`, `updated_at`

---

## mihari_policy

Manages an alert policy with multi-step escalation. Each step defines notification channels and who to notify.

### Example

```hcl
resource "mihari_policy" "critical" {
  name        = "Critical Alert Policy"
  type        = "template"
  retry_count = 3
  retry_delay = 5

  steps = [
    {
      wait_before       = 0     # Immediate
      call              = false
      push_notification = true
      sms               = false
      email             = true
      members = [
        { type = "current_persons_on_call" }
      ]
    },
    {
      wait_before       = 10    # 10 min after first step
      call              = true
      push_notification = true
      sms               = true
      email             = true
      members = [
        { type = "user", id = "user-uuid-here" }
      ]
    }
  ]
}
```

### Attributes

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | Policy name |
| `type` | string | No | `template` or `default` |
| `retry_count` | number | Yes | Retries before escalating (1-20) |
| `retry_delay` | number | Yes | Delay between retries in minutes (1-60) |
| `steps` | list(object) | Yes | Escalation steps (1-10) |

### Step Attributes

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `wait_before` | number | Yes | Minutes to wait before this step (0-60) |
| `call` | bool | Yes | Enable phone call |
| `push_notification` | bool | Yes | Enable push notification |
| `sms` | bool | Yes | Enable SMS |
| `email` | bool | Yes | Enable email |
| `members` | list(object) | Yes | Who to notify (1-20) |

### Member Attributes

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `type` | string | Yes | `user`, `current_persons_on_call`, or `teams` |
| `id` | string | Conditional | User UUID. Required when `type = "user"` |

**Computed**: `id`, `created_at`, `updated_at`

---

## mihari_status_page

Manages a public or private status page with sections grouping monitors and heartbeats.

### Example

```hcl
resource "mihari_status_page" "public" {
  company_name                = "Acme Corp"
  subdomain                   = "status-acme"
  custom_domain               = "status.acme.com"
  password_protection_enabled = false
  ip_allowlist_enabled        = false

  sections = [
    {
      name = "Core Services"
      resources = [
        {
          resource_id   = mihari_monitor.api.id
          resource_type = "monitor"
          title         = "API"
          description   = "Main REST API"
        },
        {
          resource_id   = mihari_monitor.web.id
          resource_type = "monitor"
          title         = "Website"
        }
      ]
    },
    {
      name = "Background Jobs"
      resources = [
        {
          resource_id   = mihari_heartbeat.backup.id
          resource_type = "heartbeat"
          title         = "Nightly Backup"
        }
      ]
    }
  ]
}
```

### Attributes

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `company_name` | string | Yes | Company name on the page |
| `subdomain` | string | Yes | Subdomain (e.g. `status-acme` -> `status-acme.mihari.io`). **Changing recreates.** |
| `custom_domain` | string | No | Custom domain |
| `password_protection_enabled` | bool | No | Enable password protection |
| `password` | string | No | Password (sensitive, write-only) |
| `ip_allowlist_enabled` | bool | No | Enable IP restriction |
| `ip_allowlist` | list(string) | No | Allowed IPs/CIDRs (max 50) |
| `sections` | list(object) | No | Page sections (1-50) |

### Section Attributes

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | Section name |
| `resources` | list(object) | Yes | Resources in section (1-100) |

### Section Resource Attributes

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `resource_id` | string | Yes | UUID of the monitor/heartbeat |
| `resource_type` | string | Yes | `monitor`, `heartbeat`, `service`, `database`, or `api` |
| `title` | string | Yes | Display title on status page |
| `description` | string | No | Description |

**Computed**: `id`, section `id`s, resource `id`s, `created_at`, `updated_at`

> **Note**: The API update endpoint only supports `company_name` and `subdomain`. To modify sections, password, or IP allowlist, Terraform will destroy and recreate the status page.

---

## mihari_on_call_calendar

Manages an on-call calendar that groups rotations.

### Example

```hcl
resource "mihari_on_call_calendar" "engineering" {
  name        = "Engineering On-Call"
  description = "Primary engineering rotation"
  is_active   = true
}
```

### Attributes

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | Calendar name |
| `description` | string | No | Description |
| `is_active` | bool | No | Whether active |

**Computed**: `id`, `created_at`, `updated_at`

---

## mihari_on_call_rotation

Manages a rotation within an on-call calendar, defining who is on-call and when.

### Example

```hcl
resource "mihari_on_call_rotation" "weekly" {
  on_call_calendar_id = mihari_on_call_calendar.engineering.id
  start_date          = "2024-01-15"
  start_hour          = "09:00"
  duration            = "24:00"
  repeat_days         = [1, 2, 3, 4, 5]   # Monday to Friday
  repeat_end          = "2024-12-31"

  members = [
    { member_id = "member-uuid-1" },
    { member_id = "member-uuid-2" },
    { member_id = "member-uuid-3" },
  ]
}
```

### Attributes

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `on_call_calendar_id` | string | Yes | Parent calendar UUID |
| `start_date` | string | Yes | Start date (YYYY-MM-DD) |
| `start_hour` | string | Yes | Start time (HH:MM) |
| `duration` | string | Yes | Shift duration (HH:MM) |
| `repeat_days` | list(number) | Yes | Days of week (0=Sun, 1=Mon, ..., 6=Sat) |
| `repeat_end` | string | Yes | End date (YYYY-MM-DD) |
| `members` | list(object) | Yes | Rotation members |

### Member Attributes

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `member_id` | string | Yes | Team member UUID |

**Computed**: `id`, `rrule`, `created_at`, `updated_at`
