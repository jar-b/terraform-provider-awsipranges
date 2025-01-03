package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
	Filters    types.List `tfsdk:"filters"`
	IPPrefixes types.List `tfsdk:"ip_prefixes"`
}

// FilterModel stores the filter type and value used to filter results.
type FilterModel struct {
	Type   types.String `tfsdk:"type"`
	Values types.List   `tfsdk:"values"`
}

var ipPrefixAttrType = map[string]attr.Type{
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
							MarkdownDescription: "Filter type. Valid values are: `ip`, `region`, " +
								"`network_border_group`, and `service`.",
						},
						"values": schema.ListAttribute{
							ElementType:         types.StringType,
							Required:            true,
							MarkdownDescription: "Filter values.",
						},
					},
				},
			},
			"ip_prefixes": schema.ListNestedAttribute{
				MarkdownDescription: "A list of IP address prefixes matching the filter criteria.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"ip_prefix": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Public IPv4 address range, in CIDR notation.",
						},
						"network_border_group": schema.StringAttribute{
							Computed: true,
							MarkdownDescription: "Name of the network border group, which is a unique set of " +
								"Availability Zones or Local Zones from which AWS advertises IP addresses, or `GLOBAL`.",
						},
						"region": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "AWS Region or `GLOBAL`.",
						},
						"service": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Subset of IP address ranges.",
						},
					},
				},
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
		var values []string
		resp.Diagnostics.Append(f.Values.ElementsAs(ctx, &values, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		ipFilters = append(ipFilters, awsipranges.Filter{
			Type:   awsipranges.FilterType(f.Type.ValueString()),
			Values: values,
		})
	}
	tflog.Debug(ctx, fmt.Sprintf("applying filters: %s", ipFilters))

	ipPrefixes, err := d.ranges.Filter(ipFilters)
	if err != nil {
		resp.Diagnostics.AddError("Filter Error", err.Error())
		return
	}

	elemType := types.ObjectType{AttrTypes: ipPrefixAttrType}
	if len(ipPrefixes) == 0 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ip_prefixes"), types.ListNull(elemType))...)
		return
	}

	elems := []attr.Value{}
	for _, p := range ipPrefixes {
		obj := map[string]attr.Value{
			"ip_prefix":            types.StringValue(p.IPPrefix),
			"region":               types.StringValue(p.Region),
			"network_border_group": types.StringValue(p.NetworkBorderGroup),
			"service":              types.StringValue(p.Service),
		}

		e, diag := types.ObjectValue(ipPrefixAttrType, obj)

		resp.Diagnostics.Append(diag...)
		elems = append(elems, e)
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ip_prefixes"), elems)...)
}
