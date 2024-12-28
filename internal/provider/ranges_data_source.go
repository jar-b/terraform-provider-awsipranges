package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/jar-b/awsipranges"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &RangesDataSource{}

func NewRangesDataSource() datasource.DataSource {
	return &RangesDataSource{}
}

// RangesDataSource defines the data source implementation.
type RangesDataSource struct {
	ranges *awsipranges.AWSIPRanges
}

// RangesDataSourceModel describes the data source data model.
type RangesDataSourceModel struct {
	Filters types.List   `tfsdk:"filters"`
	Id      types.String `tfsdk:"id"`
}

// FilterModel stores the filter type and value used to filter results.
type FilterModel struct {
	Type  types.String `tfsdk:"type"`
	Value types.String `tfsdk:"value"`
}

var rangeAttrType = map[string]attr.Type{
	"ip_prefix":            types.StringType,
	"region":               types.StringType,
	"network_border_group": types.StringType,
	"service":              types.StringType,
}

func (d *RangesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ranges"
}

func (d *RangesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Checks whether an IP address is in an AWS range.",

		Attributes: map[string]schema.Attribute{
			"filters": schema.ListNestedAttribute{
				MarkdownDescription: "Filters to apply to the IP ranges data set. Filtering can " +
					"be done on IP address, network border group, region, and service.",
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Required: true,
						},
						"value": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Contains identifier",
				Computed:            true,
			},
		},
	}
}

func (d *RangesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	ranges, ok := req.ProviderData.(*awsipranges.AWSIPRanges)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *awsipranges.AWSIPRanges, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.ranges = ranges
}

func (d *RangesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RangesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var tfFilters []FilterModel
	resp.Diagnostics.Append(data.Filters.ElementsAs(ctx, &tfFilters, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var ipFilters []awsipranges.Filter
	for _, f := range tfFilters {
		ipFilters = append(ipFilters, awsipranges.Filter{
			Type:  awsipranges.FilterType(f.Type.ValueString()),
			Value: f.Value.ValueString(),
		})
	}

	_, err := d.ranges.Filter(ipFilters)
	if err != nil {
		resp.Diagnostics.AddError("Filter error", err.Error())
		return
	}

	// For the purposes of this example code, hardcoding a response value to
	// save into the Terraform state.
	data.Id = types.StringValue("example-id")

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read a data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
