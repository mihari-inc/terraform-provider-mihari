package resources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccHeartbeatResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "mihari_heartbeat" "test" {
  name         = "test-heartbeat"
  period       = 60
  grace_period = 10
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mihari_heartbeat.test", "name", "test-heartbeat"),
					resource.TestCheckResourceAttr("mihari_heartbeat.test", "period", "60"),
					resource.TestCheckResourceAttr("mihari_heartbeat.test", "grace_period", "10"),
					resource.TestCheckResourceAttrSet("mihari_heartbeat.test", "id"),
				),
			},
			{
				ResourceName:      "mihari_heartbeat.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: `
resource "mihari_heartbeat" "test" {
  name         = "updated-heartbeat"
  period       = 120
  grace_period = 20
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mihari_heartbeat.test", "name", "updated-heartbeat"),
					resource.TestCheckResourceAttr("mihari_heartbeat.test", "period", "120"),
				),
			},
		},
	})
}
