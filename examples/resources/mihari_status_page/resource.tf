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
