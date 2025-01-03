package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRangesDataSourceIP(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRangesDataSourceConfig("ip", "3.5.12.4"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.awsipranges_ranges.test", "ip_prefixes.#", "3"),
					resource.TestMatchTypeSetElemNestedAttrs("data.awsipranges_ranges.test", "ip_prefixes.*", map[string]*regexp.Regexp{
						"ip_prefix": regexp.MustCompile(`^3\.5\..*`), // Ex. 3.5.0.0/19
					}),
				),
			},
		},
	})
}

func TestAccRangesDataSourceNetworkBorderGroup(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRangesDataSourceConfig("network-border-group", "us-east-1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs("data.awsipranges_ranges.test", "ip_prefixes.*", map[string]string{
						"network_border_group": "us-east-1",
					}),
				),
			},
			{
				Config: testAccRangesDataSourceConfig("network-border-group", "US-EAST-1"), // verify case insensitivity
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs("data.awsipranges_ranges.test", "ip_prefixes.*", map[string]string{
						"network_border_group": "us-east-1",
					}),
				),
			},
		},
	})
}

func TestAccRangesDataSourceRegion(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRangesDataSourceConfig("region", "us-east-1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs("data.awsipranges_ranges.test", "ip_prefixes.*", map[string]string{
						"region": "us-east-1",
					}),
				),
			},
			{
				Config: testAccRangesDataSourceConfig("region", "US-EAST-1"), // verify case insensitivity
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs("data.awsipranges_ranges.test", "ip_prefixes.*", map[string]string{
						"region": "us-east-1",
					}),
				),
			},
		},
	})
}

func TestAccRangesDataSourceService(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRangesDataSourceConfig("service", "DYNAMODB"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs("data.awsipranges_ranges.test", "ip_prefixes.*", map[string]string{
						"service": "DYNAMODB",
					}),
				),
			},
			{
				Config: testAccRangesDataSourceConfig("service", "dynamodb"), // verify case insensitivity
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs("data.awsipranges_ranges.test", "ip_prefixes.*", map[string]string{
						"service": "DYNAMODB",
					}),
				),
			},
		},
	})
}

func TestAccRangesDataSourceNoMatch(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRangesDataSourceConfig("ip", "1.1.1.1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("data.awsipranges_ranges.test", "ip_prefixes"),
				),
			},
		},
	})
}

func testAccRangesDataSourceConfig(filterType, value string) string {
	return fmt.Sprintf(`
data "awsipranges_ranges" "test" {
  filters = [
    {
      type   = %1q
      values = [%2q]
    }
  ]
}
`, filterType, value)
}
