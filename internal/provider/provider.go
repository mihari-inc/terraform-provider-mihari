package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mihari-io/terraform-provider-mihari/internal/client"
	"github.com/mihari-io/terraform-provider-mihari/internal/datasources"
	"github.com/mihari-io/terraform-provider-mihari/internal/resources"
)

var _ provider.Provider = &MihariProvider{}

// MihariProvider implements the Mihari Terraform provider.
type MihariProvider struct {
	version string
}

// MihariProviderModel describes the provider configuration data model.
type MihariProviderModel struct {
	APIURL         types.String `tfsdk:"api_url"`
	APIToken       types.String `tfsdk:"api_token"`
	OrganizationID types.String `tfsdk:"organization_id"`
}

// New returns a new provider factory function.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &MihariProvider{version: version}
	}
}

func (p *MihariProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "mihari"
	resp.Version = p.version
}

func (p *MihariProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for managing Mihari observability platform resources.",
		Attributes: map[string]schema.Attribute{
			"api_url": schema.StringAttribute{
				Description: "Base URL of the Mihari API. Can also be set via MIHARI_API_URL environment variable.",
				Optional:    true,
			},
			"api_token": schema.StringAttribute{
				Description: "Bearer token for API authentication. Can also be set via MIHARI_API_TOKEN environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"organization_id": schema.StringAttribute{
				Description: "Organization UUID. Can also be set via MIHARI_ORGANIZATION_ID environment variable.",
				Optional:    true,
			},
		},
	}
}

func (p *MihariProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config MihariProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiURL := os.Getenv("MIHARI_API_URL")
	apiToken := os.Getenv("MIHARI_API_TOKEN")
	orgID := os.Getenv("MIHARI_ORGANIZATION_ID")

	if !config.APIURL.IsNull() && !config.APIURL.IsUnknown() {
		apiURL = config.APIURL.ValueString()
	}
	if !config.APIToken.IsNull() && !config.APIToken.IsUnknown() {
		apiToken = config.APIToken.ValueString()
	}
	if !config.OrganizationID.IsNull() && !config.OrganizationID.IsUnknown() {
		orgID = config.OrganizationID.ValueString()
	}

	if apiURL == "" {
		resp.Diagnostics.AddError("Missing API URL", "The Mihari API URL must be set via the api_url attribute or MIHARI_API_URL environment variable.")
		return
	}
	if apiToken == "" {
		resp.Diagnostics.AddError("Missing API Token", "The Mihari API token must be set via the api_token attribute or MIHARI_API_TOKEN environment variable.")
		return
	}
	if orgID == "" {
		resp.Diagnostics.AddError("Missing Organization ID", "The organization ID must be set via the organization_id attribute or MIHARI_ORGANIZATION_ID environment variable.")
		return
	}

	c := client.NewClient(client.ClientConfig{
		BaseURL:        apiURL,
		APIToken:       apiToken,
		OrganizationID: orgID,
	})

	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *MihariProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewMonitorResource,
		resources.NewHeartbeatResource,
		resources.NewPolicyResource,
		resources.NewStatusPageResource,
		resources.NewOnCallCalendarResource,
		resources.NewOnCallRotationResource,
	}
}

func (p *MihariProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasources.NewMonitorDataSource,
		datasources.NewMonitorsDataSource,
		datasources.NewPolicyDataSource,
		datasources.NewStatusPageDataSource,
	}
}
