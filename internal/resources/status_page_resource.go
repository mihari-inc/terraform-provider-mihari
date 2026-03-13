package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mihari-io/terraform-provider-mihari/internal/client"
)

var (
	_ resource.Resource                = &StatusPageResource{}
	_ resource.ResourceWithImportState = &StatusPageResource{}
)

type StatusPageResource struct {
	client *client.Client
}

type StatusPageResourceModel struct {
	ID                        types.String `tfsdk:"id"`
	CompanyName               types.String `tfsdk:"company_name"`
	Subdomain                 types.String `tfsdk:"subdomain"`
	CustomDomain              types.String `tfsdk:"custom_domain"`
	PasswordProtectionEnabled types.Bool   `tfsdk:"password_protection_enabled"`
	Password                  types.String `tfsdk:"password"`
	IPAllowlistEnabled        types.Bool   `tfsdk:"ip_allowlist_enabled"`
	IPAllowlist               types.List   `tfsdk:"ip_allowlist"`
	Sections                  types.List   `tfsdk:"sections"`
	CreatedAt                 types.String `tfsdk:"created_at"`
	UpdatedAt                 types.String `tfsdk:"updated_at"`
}

func sectionResourceAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":            types.StringType,
		"resource_id":   types.StringType,
		"resource_type": types.StringType,
		"title":         types.StringType,
		"description":   types.StringType,
	}
}

func sectionAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":        types.StringType,
		"name":      types.StringType,
		"resources": types.ListType{ElemType: types.ObjectType{AttrTypes: sectionResourceAttrTypes()}},
	}
}

func NewStatusPageResource() resource.Resource {
	return &StatusPageResource{}
}

func (r *StatusPageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_status_page"
}

func (r *StatusPageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Mihari status page.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the status page.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"company_name": schema.StringAttribute{
				Required:    true,
				Description: "Company name displayed on the status page.",
			},
			"subdomain": schema.StringAttribute{
				Required:    true,
				Description: "Subdomain for the status page URL.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"custom_domain": schema.StringAttribute{
				Optional:    true,
				Description: "Custom domain for the status page.",
			},
			"password_protection_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Enable password protection.",
			},
			"password": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Password for protected status pages (write-only).",
			},
			"ip_allowlist_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Enable IP allowlist restriction.",
			},
			"ip_allowlist": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "List of allowed IP addresses or CIDR ranges.",
				Validators: []validator.List{
					listvalidator.SizeAtMost(50),
				},
			},
			"sections": schema.ListNestedAttribute{
				Optional:    true,
				Description: "Status page sections (1-50).",
				Validators: []validator.List{
					listvalidator.SizeBetween(1, 50),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "UUID of the section.",
						},
						"name": schema.StringAttribute{
							Required:    true,
							Description: "Section name.",
						},
						"resources": schema.ListNestedAttribute{
							Required:    true,
							Description: "Resources in this section (1-100).",
							Validators: []validator.List{
								listvalidator.SizeBetween(1, 100),
							},
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										Computed:    true,
										Description: "UUID of the resource section entry.",
									},
									"resource_id": schema.StringAttribute{
										Required:    true,
										Description: "UUID of the monitor or heartbeat.",
									},
									"resource_type": schema.StringAttribute{
										Required:    true,
										Description: "Type: monitor, heartbeat, service, database, api.",
										Validators: []validator.String{
											stringvalidator.OneOf("monitor", "heartbeat", "service", "database", "api"),
										},
									},
									"title": schema.StringAttribute{
										Required:    true,
										Description: "Display title.",
									},
									"description": schema.StringAttribute{
										Optional:    true,
										Description: "Description of the resource.",
									},
								},
							},
						},
					},
				},
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Creation timestamp.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Last update timestamp.",
			},
		},
	}
}

func (r *StatusPageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *StatusPageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan StatusPageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq, diags := buildStatusPageRequest(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	page, err := r.client.CreateStatusPage(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating status page", err.Error())
		return
	}

	d := mapStatusPageToState(ctx, page, &plan)
	resp.Diagnostics.Append(d...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *StatusPageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state StatusPageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	page, err := r.client.GetStatusPage(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading status page", err.Error())
		return
	}
	if page == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Preserve password from state since API never returns it
	password := state.Password

	d := mapStatusPageToState(ctx, page, &state)
	resp.Diagnostics.Append(d...)
	state.Password = password

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *StatusPageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan StatusPageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state StatusPageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq, diags := buildStatusPageRequest(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	page, err := r.client.UpdateStatusPage(ctx, state.ID.ValueString(), apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating status page", err.Error())
		return
	}

	d := mapStatusPageToState(ctx, page, &plan)
	resp.Diagnostics.Append(d...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *StatusPageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state StatusPageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteStatusPage(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting status page", err.Error())
	}
}

func (r *StatusPageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func buildStatusPageRequest(ctx context.Context, plan *StatusPageResourceModel) (client.StatusPageRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	apiReq := client.StatusPageRequest{
		CompanyName: plan.CompanyName.ValueString(),
		Subdomain:   plan.Subdomain.ValueString(),
	}

	if !plan.CustomDomain.IsNull() && !plan.CustomDomain.IsUnknown() {
		v := plan.CustomDomain.ValueString()
		apiReq.CustomDomain = &v
	}
	if !plan.PasswordProtectionEnabled.IsNull() && !plan.PasswordProtectionEnabled.IsUnknown() {
		v := plan.PasswordProtectionEnabled.ValueBool()
		apiReq.PasswordProtectionEnabled = &v
	}
	if !plan.Password.IsNull() && !plan.Password.IsUnknown() {
		v := plan.Password.ValueString()
		apiReq.Password = &v
	}
	if !plan.IPAllowlistEnabled.IsNull() && !plan.IPAllowlistEnabled.IsUnknown() {
		v := plan.IPAllowlistEnabled.ValueBool()
		apiReq.IPAllowlistEnabled = &v
	}
	if !plan.IPAllowlist.IsNull() && !plan.IPAllowlist.IsUnknown() {
		var ips []string
		diags.Append(plan.IPAllowlist.ElementsAs(ctx, &ips, false)...)
		apiReq.IPAllowlist = ips
	}

	if !plan.Sections.IsNull() && !plan.Sections.IsUnknown() {
		type SectionPlan struct {
			ID        types.String `tfsdk:"id"`
			Name      types.String `tfsdk:"name"`
			Resources types.List   `tfsdk:"resources"`
		}
		type ResourcePlan struct {
			ID           types.String `tfsdk:"id"`
			ResourceID   types.String `tfsdk:"resource_id"`
			ResourceType types.String `tfsdk:"resource_type"`
			Title        types.String `tfsdk:"title"`
			Description  types.String `tfsdk:"description"`
		}

		var sections []SectionPlan
		diags.Append(plan.Sections.ElementsAs(ctx, &sections, false)...)

		for _, section := range sections {
			sectionReq := client.StatusSectionRequest{
				Name: section.Name.ValueString(),
			}

			var resources []ResourcePlan
			diags.Append(section.Resources.ElementsAs(ctx, &resources, false)...)

			for _, res := range resources {
				resReq := client.StatusSectionResourceRequest{
					ResourceID:   res.ResourceID.ValueString(),
					ResourceType: res.ResourceType.ValueString(),
					Title:        res.Title.ValueString(),
				}
				if !res.Description.IsNull() && !res.Description.IsUnknown() {
					v := res.Description.ValueString()
					resReq.Description = &v
				}
				sectionReq.Resources = append(sectionReq.Resources, resReq)
			}

			apiReq.Sections = append(apiReq.Sections, sectionReq)
		}
	}

	return apiReq, diags
}

func mapStatusPageToState(ctx context.Context, page *client.StatusPage, state *StatusPageResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(page.ID)
	state.CompanyName = types.StringValue(page.CompanyName)
	state.Subdomain = types.StringValue(page.Subdomain)
	state.PasswordProtectionEnabled = types.BoolValue(page.PasswordProtectionEnabled)
	state.IPAllowlistEnabled = types.BoolValue(page.IPAllowlistEnabled)
	state.CreatedAt = types.StringValue(page.CreatedAt)
	state.UpdatedAt = types.StringValue(page.UpdatedAt)

	if page.CustomDomain != nil {
		state.CustomDomain = types.StringValue(*page.CustomDomain)
	} else {
		state.CustomDomain = types.StringNull()
	}

	if len(page.IPAllowlist) > 0 {
		ipValues := make([]attr.Value, len(page.IPAllowlist))
		for i, ip := range page.IPAllowlist {
			ipValues[i] = types.StringValue(ip)
		}
		ipList, d := types.ListValue(types.StringType, ipValues)
		diags.Append(d...)
		state.IPAllowlist = ipList
	} else {
		state.IPAllowlist = types.ListNull(types.StringType)
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
				resObj, d := types.ObjectValue(sectionResourceAttrTypes(), map[string]attr.Value{
					"id":            types.StringValue(res.ID),
					"resource_id":   types.StringValue(res.ResourceID),
					"resource_type": types.StringValue(res.ResourceType),
					"title":         types.StringValue(res.Title),
					"description":   desc,
				})
				diags.Append(d...)
				resValues = append(resValues, resObj)
			}

			resList, d := types.ListValue(types.ObjectType{AttrTypes: sectionResourceAttrTypes()}, resValues)
			diags.Append(d...)

			sectionObj, d := types.ObjectValue(sectionAttrTypes(), map[string]attr.Value{
				"id":        types.StringValue(section.ID),
				"name":      types.StringValue(section.Name),
				"resources": resList,
			})
			diags.Append(d...)
			sectionValues = append(sectionValues, sectionObj)
		}

		sectionsList, d := types.ListValue(types.ObjectType{AttrTypes: sectionAttrTypes()}, sectionValues)
		diags.Append(d...)
		state.Sections = sectionsList
	} else {
		state.Sections = types.ListNull(types.ObjectType{AttrTypes: sectionAttrTypes()})
	}

	return diags
}
