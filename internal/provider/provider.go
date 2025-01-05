package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/jar-b/awsipranges"
)

const (
	cachefilePath     = ".aws/ip-ranges.json"
	defaultExpiration = "720h"

	// createDateFormat is the format of the `createDate` field in the
	// underlying JSON (YY-MM-DD-hh-mm-ss)
	//
	// Ref: https://docs.aws.amazon.com/vpc/latest/userguide/aws-ip-syntax.html
	createDateFormat = "2006-01-02-15-04-05"
)

var _ provider.Provider = &AWSIPRangesProvider{}
var _ provider.ProviderWithFunctions = &AWSIPRangesProvider{}

// AWSIPRangesProvider defines the provider implementation.
type AWSIPRangesProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// AWSIPRangesProviderModel describes the provider data model.
type AWSIPRangesProviderModel struct {
	Cachefile  types.String `tfsdk:"cachefile"`
	Expiration types.String `tfsdk:"expiration"`
}

func (p *AWSIPRangesProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "awsipranges"
	resp.Version = p.version
}

func (p *AWSIPRangesProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Terraform provider for working with public AWS IP range data.",
		Attributes: map[string]schema.Attribute{
			"cachefile": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("Location to cache the `ip-ranges.json` file. If no value is provided, "+
					"the provider will attempt to cache the ranges file in a default location (`%s` "+
					"inside the current user's home directory) and read from it on subsequent runs.", cachefilePath),
				Optional: true,
			},
			"expiration": schema.StringAttribute{
				MarkdownDescription: "Duration after which the cached `ip-ranges.json` file should be replaced. If " +
					"no value is provided, the provider will use a default value of `720h` (30 days). If the cache should " +
					"never be expired, set the value to an empty string. Cache expiration is triggered by a comparison " +
					"against the `createDate` field in the source `ip-ranges.json` file, not the time stamp when the " +
					"cache file was written, so setting this to a very low value is likely to result in the source " +
					"being fetched anew each time the provider is configured.",
				Optional: true,
			},
		},
	}
}

func (p *AWSIPRangesProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data AWSIPRangesProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cachefile := data.Cachefile.ValueString()
	if data.Cachefile.IsNull() {
		cachefile = defaultCachefilePath()
	}

	expiration := defaultExpiration
	if !data.Expiration.IsNull() {
		expiration = data.Expiration.ValueString()
	}

	ranges, err := loadRanges(ctx, cachefile, expiration)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed To Load IP Ranges",
			err.Error(),
		)
		return
	}

	resp.DataSourceData = ranges
}

func (p *AWSIPRangesProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}

func (p *AWSIPRangesProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewRangesDataSource,
	}
}

func (p *AWSIPRangesProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &AWSIPRangesProvider{
			version: version,
		}
	}
}

// defaultCachefilePath constructs a default path to the cachefile
func defaultCachefilePath() string {
	u, _ := user.Current()
	return filepath.Join(u.HomeDir, cachefilePath)
}

// isExpired checks whether the createDate of a cached ip-ranges.json
// file is older than the configured expiration duration
//
// If expiration is not set, always returns false.
func isExpired(ctx context.Context, createDate, expiration string) (bool, error) {
	if expiration == "" {
		return false, nil
	}

	created, err := time.Parse(createDateFormat, createDate)
	if err != nil {
		return false, err
	}
	expirationDuration, err := time.ParseDuration(expiration)
	if err != nil {
		return false, err
	}

	if expirationDuration > time.Since(created) {
		return false, nil
	}

	tflog.Debug(ctx, "cache is expired, refreshing")
	return true, nil
}

// loadRanges attempts to read ip-ranges data from cache, falling back
// to fetching the source file if os.ReadFile fails or the creation date
// exceeds the configured cache expiration time
func loadRanges(ctx context.Context, cachefile, expiration string) (*awsipranges.AWSIPRanges, error) {
	if b, err := os.ReadFile(cachefile); err == nil {
		var ranges awsipranges.AWSIPRanges
		if err := json.Unmarshal(b, &ranges); err == nil {
			if exp, err := isExpired(ctx, ranges.CreateDate, expiration); err != nil {
				return nil, err
			} else if !exp {
				return &ranges, nil
			}
		}
	}

	b, err := awsipranges.Get()
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(cachefile, b, 0644); err != nil {
		tflog.Debug(ctx, fmt.Sprintf("cache write failed: %s", err))
	}

	var ranges awsipranges.AWSIPRanges
	if err := json.Unmarshal(b, &ranges); err != nil {
		return nil, err
	}

	return &ranges, nil
}
