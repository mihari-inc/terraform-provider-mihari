data "mihari_monitor" "example" {
  id = "existing-monitor-uuid"
}

output "monitor_name" {
  value = data.mihari_monitor.example.name
}
