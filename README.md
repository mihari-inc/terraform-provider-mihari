# Terraform Provider Mihari

Custom Terraform provider for managing [Mihari](https://mihari.io) observability platform resources as Infrastructure as Code.

## Features

- **Monitors**: HTTP status, keyword search, TCP/UDP ports, DNS, ping, SMTP, IMAP, POP3, Playwright
- **Heartbeats**: Cron job and periodic task monitoring
- **Alert Policies**: Multi-step escalation with phone, SMS, email, and push notifications
- **Status Pages**: Public/private status pages with sections and resource grouping
- **On-Call Management**: Calendars and rotations with team member scheduling

## Quick Start

### 1. Build & Install

```bash
cd terraform-provider-mihari
make build     # Compile the binary
make install   # Install to ~/.terraform.d/plugins/ for local use
```

### 2. Configure the Provider

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
  api_url         = "https://platform.mihari.io"
  api_token       = var.mihari_api_token
  organization_id = var.mihari_organization_id
}
```

Or use environment variables:

```bash
export MIHARI_API_URL="https://platform.mihari.io"
export MIHARI_API_TOKEN="your-api-token"
export MIHARI_ORGANIZATION_ID="your-org-uuid"
```

### 3. Declare Resources

```hcl
resource "mihari_monitor" "api" {
  name           = "API Health"
  type           = "http_status"
  url            = "https://api.example.com/health"
  check_interval = 3
  timeout        = 30
  is_active      = true
}
```

### 4. Apply

```bash
terraform init      # Download/locate the provider
terraform plan      # Preview changes
terraform apply     # Create/update resources
```

## Resources

| Resource | Description |
|----------|-------------|
| `mihari_monitor` | Website, API, port, DNS, ping monitors |
| `mihari_heartbeat` | Heartbeat/cron monitoring |
| `mihari_policy` | Alert policies with escalation steps |
| `mihari_status_page` | Public/private status pages |
| `mihari_on_call_calendar` | On-call calendars |
| `mihari_on_call_rotation` | On-call rotations with members |

## Data Sources

| Data Source | Description |
|-------------|-------------|
| `mihari_monitor` | Read a single monitor by ID |
| `mihari_monitors` | List monitors with filters |
| `mihari_policy` | Read a single policy by ID |
| `mihari_status_page` | Read a single status page by ID |

## Documentation

See the [`docs/`](docs/) directory for detailed guides:

- [Getting Started](docs/getting-started.md) - Installation and first configuration
- [Resources Reference](docs/resources.md) - All resources with their attributes
- [Data Sources Reference](docs/data-sources.md) - All data sources
- [Examples](docs/examples.md) - Complete real-world examples

## Development

### Commands

```bash
make build       # Build the provider binary
make install     # Install locally for Terraform
make test        # Run unit tests (17 tests)
make testacc     # Run acceptance tests (requires running Mihari API)
make fmt         # Format Go code
make tidy        # Tidy Go modules
make clean       # Remove build artifacts
```

### Running Acceptance Tests

```bash
export TF_ACC=1
export MIHARI_API_URL=http://localhost:8000
export MIHARI_API_TOKEN=your-test-token
export MIHARI_ORGANIZATION_ID=your-test-org-uuid
make testacc
```

### Project Structure

```
terraform-provider-mihari/
├── main.go                         # Provider entry point
├── go.mod / go.sum                 # Go module dependencies
├── GNUmakefile                     # Build, test, install targets
├── examples/main.tf                # Complete HCL example
├── docs/                           # Documentation
├── internal/
│   ├── provider/                   # Provider configuration
│   ├── client/                     # Mihari API HTTP client
│   │   ├── client.go               # Auth, error handling, base HTTP
│   │   ├── monitors.go             # Monitor CRUD
│   │   ├── heartbeats.go           # Heartbeat CRUD
│   │   ├── policies.go             # Policy CRUD (nested steps/members)
│   │   ├── status_pages.go         # StatusPage CRUD (nested sections)
│   │   ├── on_call_calendars.go    # Calendar CRUD
│   │   ├── on_call_rotations.go    # Rotation CRUD
│   │   └── client_test.go          # Unit tests (httptest mocks)
│   ├── resources/                  # Terraform resource implementations
│   └── datasources/                # Terraform data source implementations
```

## Known Limitations

- **StatusPage update**: The API update endpoint only supports `company_name` and `subdomain`. Changes to sections, password protection, or IP allowlist require `terraform destroy` + `terraform apply` (destroy and recreate).
- **Monitor name mapping**: The API returns `title` instead of `name` in responses. The provider handles this mapping transparently.
- **On-call rotation**: Uses the API's UI-format input structure (`event.start.date`, `repeat.days`). The provider transforms flat Terraform attributes into this format.
