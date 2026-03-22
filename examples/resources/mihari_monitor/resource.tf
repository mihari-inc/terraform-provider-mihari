# HTTP status monitor
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

# Keyword monitoring
resource "mihari_monitor" "website_keyword" {
  name           = "Website Contains Login"
  type           = "url_contains"
  url            = "https://www.example.com"
  keyword        = "Sign In"
  check_interval = 5
  timeout        = 30
  is_active      = true
}

# TCP port monitor
resource "mihari_monitor" "tcp_database" {
  name           = "Database TCP Check"
  type           = "tcp_port"
  host           = "db.internal.example.com"
  port           = 5432
  check_interval = 2
  timeout        = 10
  is_active      = true
}

# DNS monitor
resource "mihari_monitor" "dns_check" {
  name           = "DNS Resolution Check"
  type           = "dns"
  host           = "example.com"
  check_interval = 10
  timeout        = 15
  is_active      = true
}

# Ping monitor
resource "mihari_monitor" "ping_check" {
  name           = "Server Ping"
  type           = "ping"
  host           = "10.0.1.50"
  check_interval = 1
  timeout        = 10
  is_active      = true
}
