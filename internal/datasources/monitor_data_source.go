package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mihari-io/terraform-provider-mihari/internal/client"
)

var _ datasource.DataSource = &MonitorDataSource{}

type MonitorDataSource struct {
	client *client.Client
}

type MonitorDataSourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	URL                types.String `tfsdk:"url"`
	Type               types.String `tfsdk:"type"`
	Keyword            types.String `tfsdk:"keyword"`
	ExpectedStatusCode types.Int64  `tfsdk:"expected_status_code"`
	Host               types.String `tfsdk:"host"`
	Port               types.Int64  `tfsdk:"port"`
	Protocol           types.String `tfsdk:"protocol"`
	CheckInterval      types.Int64  `tfsdk:"check_interval"`
	Timeout            types.Int64  `tfsdk:"timeout"`
	CheckSSL           types.Bool   `tfsdk:"check_ssl"`
	IsActive           types.Bool   `tfsdk:"is_active"`
	PolicyID           types.String `tfsdk:"policy_id"`
	Status             types.String `tfsdk:"status"`
	LastCheckedAt      types.String `tfsdk:"last_checked_at"`
	CreatedAt          types.String `tfsdk:"created_at"`
	UpdatedAt          types.String `tfsdk:"updated_at"`
}

func NewMonitorDataSource() datasource.DataSource {
	return &MonitorDataSource{}
}

func (d *MonitorDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitor"
}

func (d *MonitorDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a single Mihari monitor by ID.",
		Attributes: map[string]schema.Attribute{
			"id":                  schema.StringAttribute{Required: true, Description: "UUID of the monitor."},
			"name":                schema.StringAttribute{Computed: true},
			"url":                 schema.StringAttribute{Computed: true},
			"type":                schema.StringAttribute{Computed: true},
			"keyword":             schema.StringAttribute{Computed: true},
			"expected_status_code": schema.Int64Attribute{Computed: true},
			"host":                schema.StringAttribute{Computed: true},
			"port":                schema.Int64Attribute{Computed: true},
			"protocol":            schema.StringAttribute{Computed: true},
			"check_interval":      schema.Int64Attribute{Computed: true},
			"timeout":             schema.Int64Attribute{Computed: true},
			"check_ssl":           schema.BoolAttribute{Computed: true},
			"is_active":           schema.BoolAttribute{Computed: true},
			"policy_id":           schema.StringAttribute{Computed: true},
			"status":              schema.StringAttribute{Computed: true},
			"last_checked_at":     schema.StringAttribute{Computed: true},
			"created_at":          schema.StringAttribute{Computed: true},
			"updated_at":          schema.StringAttribute{Computed: true},
		},
	}
}

func (d *MonitorDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected DataSource Configure Type", fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *MonitorDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config MonitorDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	monitor, err := d.client.GetMonitor(ctx, config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading monitor", err.Error())
		return
	}
	if monitor == nil {
		resp.Diagnostics.AddError("Monitor not found", fmt.Sprintf("Monitor with ID %s not found", config.ID.ValueString()))
		return
	}

	config.Name = types.StringValue(monitor.Title)
	config.Type = types.StringValue(monitor.Type)
	config.URL = types.StringValue(monitor.URL)
	config.Keyword = types.StringValue(monitor.Keyword)
	config.Host = types.StringValue(monitor.Host)
	config.Protocol = types.StringValue(monitor.Protocol)
	config.CheckInterval = types.Int64Value(int64(monitor.CheckInterval))
	config.Timeout = types.Int64Value(int64(monitor.Timeout))
	config.CheckSSL = types.BoolValue(monitor.CheckSSL)
	config.IsActive = types.BoolValue(monitor.IsActive)
	config.Status = types.StringValue(monitor.Status)
	config.CreatedAt = types.StringValue(monitor.CreatedAt)
	config.UpdatedAt = types.StringValue(monitor.UpdatedAt)

	if monitor.ExpectedStatusCode != nil {
		config.ExpectedStatusCode = types.Int64Value(int64(*monitor.ExpectedStatusCode))
	} else {
		config.ExpectedStatusCode = types.Int64Null()
	}
	if monitor.Port != nil {
		config.Port = types.Int64Value(int64(*monitor.Port))
	} else {
		config.Port = types.Int64Null()
	}
	if monitor.PolicyID != nil {
		config.PolicyID = types.StringValue(*monitor.PolicyID)
	} else {
		config.PolicyID = types.StringNull()
	}
	if monitor.LastCheckedAt != nil {
		config.LastCheckedAt = types.StringValue(*monitor.LastCheckedAt)
	} else {
		config.LastCheckedAt = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
