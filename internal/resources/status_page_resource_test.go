package resources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccStatusPageResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "mihari_status_page" "test" {
  company_name                = "Test Corp"
  subdomain                   = "test-status-tf"
  password_protection_enabled = false
  ip_allowlist_enabled        = false
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mihari_status_page.test", "company_name", "Test Corp"),
					resource.TestCheckResourceAttr("mihari_status_page.test", "subdomain", "test-status-tf"),
					resource.TestCheckResourceAttrSet("mihari_status_page.test", "id"),
				),
			},
			{
				ResourceName:            "mihari_status_page.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func TestAccStatusPageResource_withSections(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "mihari_monitor" "test" {
  name           = "test-for-status-page"
  type           = "http_status"
  url            = "https://example.com"
  check_interval = 5
  timeout        = 30
}

resource "mihari_status_page" "test" {
  company_name                = "Test Corp"
  subdomain                   = "test-sections-tf"
  password_protection_enabled = false
  ip_allowlist_enabled        = false

  sections = [
    {
      name = "Core Services"
      resources = [
        {
          resource_id   = mihari_monitor.test.id
          resource_type = "monitor"
          title         = "API"
          description   = "Main API endpoint"
        }
      ]
    }
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mihari_status_page.test", "company_name", "Test Corp"),
					resource.TestCheckResourceAttrSet("mihari_status_page.test", "id"),
				),
			},
		},
	})
}
