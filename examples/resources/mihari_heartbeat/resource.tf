resource "mihari_heartbeat" "nightly_backup" {
  name         = "Nightly Backup Job"
  period       = 1440
  grace_period = 30
  is_active    = true
  policy_id    = mihari_policy.critical.id
}

resource "mihari_heartbeat" "cron_sync" {
  name         = "Data Sync Cron"
  period       = 60
  grace_period = 10
  is_active    = true
}
