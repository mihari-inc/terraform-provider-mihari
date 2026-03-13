package resources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPolicyResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "mihari_policy" "test" {
  name        = "test-policy"
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
    }
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mihari_policy.test", "name", "test-policy"),
					resource.TestCheckResourceAttr("mihari_policy.test", "type", "template"),
					resource.TestCheckResourceAttr("mihari_policy.test", "retry_count", "3"),
					resource.TestCheckResourceAttr("mihari_policy.test", "retry_delay", "5"),
					resource.TestCheckResourceAttrSet("mihari_policy.test", "id"),
				),
			},
			{
				ResourceName:      "mihari_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccPolicyResource_multiStep(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "mihari_policy" "test" {
  name        = "multi-step-policy"
  type        = "template"
  retry_count = 5
  retry_delay = 10

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
      wait_before       = 10
      call              = true
      push_notification = true
      sms               = true
      email             = true
      members = [
        { type = "current_persons_on_call" }
      ]
    }
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mihari_policy.test", "name", "multi-step-policy"),
					resource.TestCheckResourceAttr("mihari_policy.test", "retry_count", "5"),
				),
			},
		},
	})
}
