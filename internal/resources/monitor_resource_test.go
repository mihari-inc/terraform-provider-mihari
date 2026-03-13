package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/mihari-io/terraform-provider-mihari/internal/provider"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"mihari": providerserver.NewProtocol6WithError(provider.New("test")()),
}

func TestAccMonitorResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorResourceConfig("test-monitor", "http_status", "https://example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mihari_monitor.test", "name", "test-monitor"),
					resource.TestCheckResourceAttr("mihari_monitor.test", "type", "http_status"),
					resource.TestCheckResourceAttr("mihari_monitor.test", "url", "https://example.com"),
					resource.TestCheckResourceAttrSet("mihari_monitor.test", "id"),
				),
			},
			// ImportState
			{
				ResourceName:      "mihari_monitor.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update
			{
				Config: testAccMonitorResourceConfig("updated-monitor", "http_status", "https://example.com/updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mihari_monitor.test", "name", "updated-monitor"),
					resource.TestCheckResourceAttr("mihari_monitor.test", "url", "https://example.com/updated"),
				),
			},
		},
	})
}

func TestAccMonitorResource_tcp(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorTCPConfig("tcp-test", "db.example.com", 5432),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mihari_monitor.test", "name", "tcp-test"),
					resource.TestCheckResourceAttr("mihari_monitor.test", "type", "tcp_port"),
					resource.TestCheckResourceAttr("mihari_monitor.test", "host", "db.example.com"),
					resource.TestCheckResourceAttr("mihari_monitor.test", "port", "5432"),
				),
			},
		},
	})
}

func testAccMonitorResourceConfig(name, monitorType, url string) string {
	return fmt.Sprintf(`
resource "mihari_monitor" "test" {
  name           = %q
  type           = %q
  url            = %q
  check_interval = 5
  timeout        = 30
}
`, name, monitorType, url)
}

func testAccMonitorTCPConfig(name, host string, port int) string {
	return fmt.Sprintf(`
resource "mihari_monitor" "test" {
  name           = %q
  type           = "tcp_port"
  host           = %q
  port           = %d
  check_interval = 5
  timeout        = 10
}
`, name, host, port)
}
