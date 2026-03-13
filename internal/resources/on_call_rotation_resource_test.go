package resources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOnCallRotationResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "mihari_on_call_calendar" "test" {
  name      = "rotation-test-calendar"
  is_active = true
}

resource "mihari_on_call_rotation" "test" {
  on_call_calendar_id = mihari_on_call_calendar.test.id
  start_date          = "2024-01-15"
  start_hour          = "09:00"
  duration            = "08:00"
  repeat_days         = [1, 2, 3, 4, 5]
  repeat_end          = "2024-12-31"

  members = [
    { member_id = "00000000-0000-0000-0000-000000000001" }
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("mihari_on_call_rotation.test", "id"),
					resource.TestCheckResourceAttr("mihari_on_call_rotation.test", "start_date", "2024-01-15"),
					resource.TestCheckResourceAttr("mihari_on_call_rotation.test", "start_hour", "09:00"),
					resource.TestCheckResourceAttr("mihari_on_call_rotation.test", "duration", "08:00"),
				),
			},
		},
	})
}
