package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mihari-io/terraform-provider-mihari/internal/client"
)

var _ datasource.DataSource = &MonitorsDataSource{}

type MonitorsDataSource struct {
	client *client.Client
}

type MonitorsDataSourceModel struct {
	NameFilter     types.String `tfsdk:"name"`
	TypeFilter     types.String `tfsdk:"type"`
	IsActiveFilter types.Bool   `tfsdk:"is_active"`
	Monitors       types.List   `tfsdk:"monitors"`
}

func monitorItemAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":             types.StringType,
		"name":           types.StringType,
		"url":            types.StringType,
		"type":           types.StringType,
		"host":           types.StringType,
		"protocol":       types.StringType,
		"check_interval": types.Int64Type,
		"is_active":      types.BoolType,
		"status":         types.StringType,
	}
}

func NewMonitorsDataSource() datasource.DataSource {
	return &MonitorsDataSource{}
}

func (d *MonitorsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitors"
}

func (d *MonitorsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists Mihari monitors with optional filters.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by monitor name.",
			},
			"type": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by monitor type.",
			},
			"is_active": schema.BoolAttribute{
				Optional:    true,
				Description: "Filter by active status.",
			},
			"monitors": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of monitors.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":             schema.StringAttribute{Computed: true},
						"name":           schema.StringAttribute{Computed: true},
						"url":            schema.StringAttribute{Computed: true},
						"type":           schema.StringAttribute{Computed: true},
						"host":           schema.StringAttribute{Computed: true},
						"protocol":       schema.StringAttribute{Computed: true},
						"check_interval": schema.Int64Attribute{Computed: true},
						"is_active":      schema.BoolAttribute{Computed: true},
						"status":         schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *MonitorsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *MonitorsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config MonitorsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	filters := make(map[string]string)
	if !config.NameFilter.IsNull() {
		filters["name"] = config.NameFilter.ValueString()
	}
	if !config.TypeFilter.IsNull() {
		filters["type"] = config.TypeFilter.ValueString()
	}
	if !config.IsActiveFilter.IsNull() {
		if config.IsActiveFilter.ValueBool() {
			filters["is_active"] = "1"
		} else {
			filters["is_active"] = "0"
		}
	}

	monitors, err := d.client.ListMonitors(ctx, filters)
	if err != nil {
		resp.Diagnostics.AddError("Error listing monitors", err.Error())
		return
	}

	var monitorValues []attr.Value
	for _, m := range monitors {
		monitorObj, diags := types.ObjectValue(monitorItemAttrTypes(), map[string]attr.Value{
			"id":             types.StringValue(m.ID),
			"name":           types.StringValue(m.Title),
			"url":            types.StringValue(m.URL),
			"type":           types.StringValue(m.Type),
			"host":           types.StringValue(m.Host),
			"protocol":       types.StringValue(m.Protocol),
			"check_interval": types.Int64Value(int64(m.CheckInterval)),
			"is_active":      types.BoolValue(m.IsActive),
			"status":         types.StringValue(m.Status),
		})
		resp.Diagnostics.Append(diags...)
		monitorValues = append(monitorValues, monitorObj)
	}

	monitorsList, diags := types.ListValue(types.ObjectType{AttrTypes: monitorItemAttrTypes()}, monitorValues)
	resp.Diagnostics.Append(diags...)
	config.Monitors = monitorsList

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
