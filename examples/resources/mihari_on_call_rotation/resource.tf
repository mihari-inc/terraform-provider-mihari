resource "mihari_on_call_rotation" "weekly" {
  on_call_calendar_id = mihari_on_call_calendar.engineering.id
  start_date          = "2025-01-06"
  start_hour          = "09:00"
  duration            = "24:00"
  repeat_days         = [1, 2, 3, 4, 5]
  repeat_end          = "2025-12-31"

  members = [
    { member_id = "user-uuid-1" },
    { member_id = "user-uuid-2" },
    { member_id = "user-uuid-3" }
  ]
}
