resource "mihari_policy" "critical" {
  name        = "Critical Alert Policy"
  type        = "template"
  retry_count = 3
  retry_delay = 5

  steps = [
    {
      wait_before       = 0
      call              = false
      push_notification = true
      sms               = false
      email             = true
      members = [
        { type = "current_persons_on_call" }
      ]
    },
    {
      wait_before       = 5
      call              = true
      push_notification = true
      sms               = true
      email             = true
      members = [
        {
          type = "user"
          id   = "user-uuid-here"
        }
      ]
    }
  ]
}
