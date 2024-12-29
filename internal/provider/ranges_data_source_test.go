package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRangesDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRangesDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.awsipranges_ranges.test", "ip_prefixes.#", "3"),
				),
			},
		},
	})
}

const testAccRangesDataSourceConfig = `
data "awsipranges_ranges" "test" {
  filters = [
    {
      type  = "ip"
      value = "3.5.12.4"
    }
  ]
}
`
