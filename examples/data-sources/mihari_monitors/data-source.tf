data "mihari_monitors" "active_http" {
  type      = "http_status"
  is_active = true
}

output "active_http_monitors" {
  value = data.mihari_monitors.active_http.monitors[*].name
}
