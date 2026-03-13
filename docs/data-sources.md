# Data Sources Reference

Data sources allow you to read existing Mihari resources without managing them. Useful for referencing resources created outside Terraform or in another Terraform workspace.

## mihari_monitor

Read a single monitor by its UUID.

### Example

```hcl
data "mihari_monitor" "existing" {
  id = "abc-123-def-456"
}

output "monitor_name" {
  value = data.mihari_monitor.existing.name
}

output "monitor_status" {
  value = data.mihari_monitor.existing.status
}
```

### Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `id` | string | **(Required)** UUID of the monitor |
| `name` | string | Monitor name |
| `type` | string | Monitor type |
| `url` | string | Monitored URL |
| `keyword` | string | Search keyword |
| `expected_status_code` | number | Expected HTTP status |
| `host` | string | Monitored host |
| `port` | number | Monitored port |
| `protocol` | string | Protocol (http/https) |
| `check_interval` | number | Check interval (minutes) |
| `timeout` | number | Timeout (seconds) |
| `check_ssl` | bool | SSL validation enabled |
| `is_active` | bool | Active status |
| `policy_id` | string | Attached policy UUID |
| `status` | string | Current status (up/down/recovery/acknowledge) |
| `last_checked_at` | string | Last check timestamp |
| `created_at` | string | Creation timestamp |
| `updated_at` | string | Update timestamp |

---

## mihari_monitors

List monitors with optional filters. Returns all matching monitors.

### Example

```hcl
# Get all active HTTP monitors
data "mihari_monitors" "active_http" {
  type      = "http_status"
  is_active = true
}

output "active_count" {
  value = length(data.mihari_monitors.active_http.monitors)
}

output "monitor_names" {
  value = data.mihari_monitors.active_http.monitors[*].name
}
```

### Filter Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `name` | string | Filter by name (partial match) |
| `type` | string | Filter by monitor type |
| `is_active` | bool | Filter by active status |

### Result: `monitors`

List of objects with:

| Attribute | Type | Description |
|-----------|------|-------------|
| `id` | string | Monitor UUID |
| `name` | string | Monitor name |
| `url` | string | Monitored URL |
| `type` | string | Monitor type |
| `host` | string | Monitored host |
| `protocol` | string | Protocol |
| `check_interval` | number | Check interval |
| `is_active` | bool | Active status |
| `status` | string | Current status |

---

## mihari_policy

Read a single alert policy by its UUID, including all escalation steps and members.

### Example

```hcl
data "mihari_policy" "existing" {
  id = "pol-123-abc"
}

output "policy_name" {
  value = data.mihari_policy.existing.name
}

output "step_count" {
  value = length(data.mihari_policy.existing.steps)
}
```

### Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `id` | string | **(Required)** Policy UUID |
| `name` | string | Policy name |
| `type` | string | Policy type |
| `retry_count` | number | Retry count |
| `retry_delay` | number | Retry delay (minutes) |
| `steps` | list(object) | Escalation steps with members |
| `created_at` | string | Creation timestamp |
| `updated_at` | string | Update timestamp |

---

## mihari_status_page

Read a single status page by its UUID, including all sections and resources.

### Example

```hcl
data "mihari_status_page" "public" {
  id = "sp-123-abc"
}

output "subdomain" {
  value = data.mihari_status_page.public.subdomain
}

output "section_count" {
  value = length(data.mihari_status_page.public.sections)
}
```

### Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `id` | string | **(Required)** Status page UUID |
| `company_name` | string | Company name |
| `subdomain` | string | Subdomain |
| `custom_domain` | string | Custom domain |
| `password_protection_enabled` | bool | Password protection |
| `ip_allowlist_enabled` | bool | IP restriction |
| `ip_allowlist` | list(string) | Allowed IPs |
| `sections` | list(object) | Sections with resources |
| `created_at` | string | Creation timestamp |
| `updated_at` | string | Update timestamp |
