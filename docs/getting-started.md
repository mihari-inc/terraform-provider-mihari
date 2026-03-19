# Getting Started

## Prerequisites

- [Terraform](https://www.terraform.io/downloads) >= 1.0
- [Go](https://go.dev/dl/) >= 1.21 (for building from source)
- A Mihari account with an API token and organization ID

## Installation

### From Source

```bash
git clone <repo-url>
cd terraform-provider-mihari
make install
```

This compiles the provider and places it in `~/.terraform.d/plugins/` so Terraform can find it locally.

### Verify Installation

Create a test file `main.tf`:

```hcl
terraform {
  required_providers {
    mihari = {
      source  = "mihari-io/mihari"
      version = "0.1.0"
    }
  }
}

provider "mihari" {}
```

Run:

```bash
terraform init
```

You should see: `Terraform has been successfully initialized!`

## Configuration

The provider needs three values to connect to the Mihari API:

| Attribute | Environment Variable | Description |
|-----------|---------------------|-------------|
| `api_url` | `MIHARI_API_URL` | Base URL of the Mihari API |
| `api_token` | `MIHARI_API_TOKEN` | Bearer token for authentication |
| `organization_id` | `MIHARI_ORGANIZATION_ID` | Organization UUID |

### Option 1: Environment Variables (recommended for CI/CD)

```bash
export MIHARI_API_URL="https://platform.mihari.io"
export MIHARI_API_TOKEN="your-token-here"
export MIHARI_ORGANIZATION_ID="your-org-uuid"
```

```hcl
provider "mihari" {}
```

### Option 2: HCL Configuration (with variables)

```hcl
variable "mihari_api_token" {
  type      = string
  sensitive = true
}

provider "mihari" {
  api_url         = "https://platform.mihari.io"
  api_token       = var.mihari_api_token
  organization_id = "your-org-uuid"
}
```

Pass the token via CLI:

```bash
terraform apply -var="mihari_api_token=your-token"
```

Or via a `terraform.tfvars` file (add to `.gitignore`!):

```hcl
mihari_api_token = "your-token-here"
```

## How Terraform Works (Key Concepts)

### Declarative, Not Imperative

You describe the **desired state**, not the steps to get there. If you declare 5 monitors, Terraform ensures exactly 5 exist.

### Idempotent

Running `terraform apply` multiple times with the same configuration changes nothing after the first run. Terraform compares the current state with the desired state and only makes necessary changes.

### State File

Terraform stores the mapping between your HCL declarations and real API resources in `terraform.tfstate`. This file tracks resource IDs so Terraform knows which API resource corresponds to which HCL block.

**Important**: Never delete `terraform.tfstate` unless you want Terraform to "forget" about all managed resources.

### Lifecycle

```
terraform init      # Install providers
terraform plan      # Preview what will change
terraform apply     # Apply changes (create/update/delete)
terraform destroy   # Delete all managed resources
```

## Your First Resource

Create a file `main.tf`:

```hcl
terraform {
  required_providers {
    mihari = {
      source  = "mihari-io/mihari"
      version = "0.1.0"
    }
  }
}

provider "mihari" {
  api_url         = "http://localhost:8000"
  api_token       = "your-token"
  organization_id = "your-org-uuid"
}

# Create a simple HTTP monitor
resource "mihari_monitor" "website" {
  name           = "My Website"
  type           = "http_status"
  url            = "https://example.com"
  check_interval = 5    # Check every 5 minutes
  timeout        = 30   # 30 second timeout
  is_active      = true
}

# Output the monitor ID
output "monitor_id" {
  value = mihari_monitor.website.id
}
```

Apply it:

```bash
terraform init
terraform plan    # Review what will be created
terraform apply   # Type "yes" to confirm
```

You should see output like:

```
mihari_monitor.website: Creating...
mihari_monitor.website: Creation complete after 1s [id=abc-123-def]

Apply complete! Resources: 1 added, 0 changed, 0 destroyed.

Outputs:
  monitor_id = "abc-123-def"
```

### Modify

Change the name in `main.tf`:

```hcl
  name = "My Updated Website"
```

```bash
terraform apply
# ~ mihari_monitor.website will be updated in-place
#   ~ name: "My Website" -> "My Updated Website"
```

### Destroy

```bash
terraform destroy
# mihari_monitor.website: Destroying... [id=abc-123-def]
# Destroy complete! Resources: 1 destroyed.
```

## Import Existing Resources

If you have resources already created in Mihari's UI, you can import them into Terraform:

```bash
terraform import mihari_monitor.existing abc-123-def-456
```

Then add the corresponding HCL block to your `.tf` file with matching attributes.

## Next Steps

- [Resources Reference](resources.md) - All available resources and their attributes
- [Examples](examples.md) - Complete real-world configuration examples
