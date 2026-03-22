data "mihari_status_page" "example" {
  id = "existing-status-page-uuid"
}

output "status_page_subdomain" {
  value = data.mihari_status_page.example.subdomain
}
