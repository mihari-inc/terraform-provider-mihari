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

var _ datasource.DataSource = &StatusPageDataSource{}

type StatusPageDataSource struct {
	client *client.Client
}

type StatusPageDataSourceModel struct {
	ID                        types.String `tfsdk:"id"`
	CompanyName               types.String `tfsdk:"company_name"`
	Subdomain                 types.String `tfsdk:"subdomain"`
	CustomDomain              types.String `tfsdk:"custom_domain"`
	PasswordProtectionEnabled types.Bool   `tfsdk:"password_protection_enabled"`
	IPAllowlistEnabled        types.Bool   `tfsdk:"ip_allowlist_enabled"`
	IPAllowlist               types.List   `tfsdk:"ip_allowlist"`
	Sections                  types.List   `tfsdk:"sections"`
	CreatedAt                 types.String `tfsdk:"created_at"`
	UpdatedAt                 types.String `tfsdk:"updated_at"`
}

func spSectionResourceAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":            types.StringType,
		"resource_id":   types.StringType,
		"resource_type": types.StringType,
		"title":         types.StringType,
		"description":   types.StringType,
	}
}

func spSectionAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":        types.StringType,
		"name":      types.StringType,
		"resources": types.ListType{ElemType: types.ObjectType{AttrTypes: spSectionResourceAttrTypes()}},
	}
}

func NewStatusPageDataSource() datasource.DataSource {
	return &StatusPageDataSource{}
}

func (d *StatusPageDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_status_page"
}

func (d *StatusPageDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a single Mihari status page by ID.",
		Attributes: map[string]schema.Attribute{
			"id":                          schema.StringAttribute{Required: true, Description: "UUID of the status page."},
			"company_name":                schema.StringAttribute{Computed: true},
			"subdomain":                   schema.StringAttribute{Computed: true},
			"custom_domain":               schema.StringAttribute{Computed: true},
			"password_protection_enabled": schema.BoolAttribute{Computed: true},
			"ip_allowlist_enabled":        schema.BoolAttribute{Computed: true},
			"ip_allowlist": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
			"sections": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":   schema.StringAttribute{Computed: true},
						"name": schema.StringAttribute{Computed: true},
						"resources": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"id":            schema.StringAttribute{Computed: true},
									"resource_id":   schema.StringAttribute{Computed: true},
									"resource_type": schema.StringAttribute{Computed: true},
									"title":         schema.StringAttribute{Computed: true},
									"description":   schema.StringAttribute{Computed: true},
								},
							},
						},
					},
				},
			},
			"created_at": schema.StringAttribute{Computed: true},
			"updated_at": schema.StringAttribute{Computed: true},
		},
	}
}

func (d *StatusPageDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *StatusPageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config StatusPageDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	page, err := d.client.GetStatusPage(ctx, config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading status page", err.Error())
		return
	}
	if page == nil {
		resp.Diagnostics.AddError("Status page not found", fmt.Sprintf("Status page with ID %s not found", config.ID.ValueString()))
		return
	}

	config.CompanyName = types.StringValue(page.CompanyName)
	config.Subdomain = types.StringValue(page.Subdomain)
	config.PasswordProtectionEnabled = types.BoolValue(page.PasswordProtectionEnabled)
	config.IPAllowlistEnabled = types.BoolValue(page.IPAllowlistEnabled)
	config.CreatedAt = types.StringValue(page.CreatedAt)
	config.UpdatedAt = types.StringValue(page.UpdatedAt)

	if page.CustomDomain != nil {
		config.CustomDomain = types.StringValue(*page.CustomDomain)
	} else {
		config.CustomDomain = types.StringNull()
	}

	if len(page.IPAllowlist) > 0 {
		ipValues := make([]attr.Value, len(page.IPAllowlist))
		for i, ip := range page.IPAllowlist {
			ipValues[i] = types.StringValue(ip)
		}
		ipList, diags := types.ListValue(types.StringType, ipValues)
		resp.Diagnostics.Append(diags...)
		config.IPAllowlist = ipList
	} else {
		config.IPAllowlist = types.ListNull(types.StringType)
	}

	if len(page.Sections) > 0 {
		var sectionValues []attr.Value
		for _, section := range page.Sections {
			var resValues []attr.Value
			for _, res := range section.Resources {
				desc := types.StringNull()
				if res.Description != nil {
					desc = types.StringValue(*res.Description)
				}
				resObj, diags := types.ObjectValue(spSectionResourceAttrTypes(), map[string]attr.Value{
					"id":            types.StringValue(res.ID),
					"resource_id":   types.StringValue(res.ResourceID),
					"resource_type": types.StringValue(res.ResourceType),
					"title":         types.StringValue(res.Title),
					"description":   desc,
				})
				resp.Diagnostics.Append(diags...)
				resValues = append(resValues, resObj)
			}

			resList, diags := types.ListValue(types.ObjectType{AttrTypes: spSectionResourceAttrTypes()}, resValues)
			resp.Diagnostics.Append(diags...)

			sectionObj, diags2 := types.ObjectValue(spSectionAttrTypes(), map[string]attr.Value{
				"id":        types.StringValue(section.ID),
				"name":      types.StringValue(section.Name),
				"resources": resList,
			})
			resp.Diagnostics.Append(diags2...)
			sectionValues = append(sectionValues, sectionObj)
		}

		sectionsList, diags := types.ListValue(types.ObjectType{AttrTypes: spSectionAttrTypes()}, sectionValues)
		resp.Diagnostics.Append(diags...)
		config.Sections = sectionsList
	} else {
		config.Sections = types.ListNull(types.ObjectType{AttrTypes: spSectionAttrTypes()})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
