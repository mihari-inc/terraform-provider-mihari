package resources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOnCallCalendarResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "mihari_on_call_calendar" "test" {
  name        = "test-calendar"
  description = "Test on-call calendar"
  is_active   = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mihari_on_call_calendar.test", "name", "test-calendar"),
					resource.TestCheckResourceAttr("mihari_on_call_calendar.test", "description", "Test on-call calendar"),
					resource.TestCheckResourceAttr("mihari_on_call_calendar.test", "is_active", "true"),
					resource.TestCheckResourceAttrSet("mihari_on_call_calendar.test", "id"),
				),
			},
			{
				ResourceName:      "mihari_on_call_calendar.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: `
resource "mihari_on_call_calendar" "test" {
  name        = "updated-calendar"
  description = "Updated description"
  is_active   = false
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mihari_on_call_calendar.test", "name", "updated-calendar"),
					resource.TestCheckResourceAttr("mihari_on_call_calendar.test", "is_active", "false"),
				),
			},
		},
	})
}
