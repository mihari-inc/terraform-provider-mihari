package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mihari-io/terraform-provider-mihari/internal/client"
)

var (
	_ resource.Resource                = &OnCallCalendarResource{}
	_ resource.ResourceWithImportState = &OnCallCalendarResource{}
)

type OnCallCalendarResource struct {
	client *client.Client
}

type OnCallCalendarResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	IsActive    types.Bool   `tfsdk:"is_active"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func NewOnCallCalendarResource() resource.Resource {
	return &OnCallCalendarResource{}
}

func (r *OnCallCalendarResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_on_call_calendar"
}

func (r *OnCallCalendarResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Mihari on-call calendar.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the on-call calendar.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the on-call calendar.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Description of the on-call calendar.",
			},
			"is_active": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the calendar is active.",
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

func (r *OnCallCalendarResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *OnCallCalendarResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan OnCallCalendarResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := client.OnCallCalendarRequest{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	}
	if !plan.IsActive.IsNull() && !plan.IsActive.IsUnknown() {
		v := plan.IsActive.ValueBool()
		apiReq.IsActive = &v
	}

	calendar, err := r.client.CreateOnCallCalendar(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating on-call calendar", err.Error())
		return
	}

	mapOnCallCalendarToState(calendar, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *OnCallCalendarResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state OnCallCalendarResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	calendar, err := r.client.GetOnCallCalendar(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading on-call calendar", err.Error())
		return
	}
	if calendar == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	mapOnCallCalendarToState(calendar, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *OnCallCalendarResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan OnCallCalendarResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state OnCallCalendarResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := client.OnCallCalendarRequest{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	}
	if !plan.IsActive.IsNull() && !plan.IsActive.IsUnknown() {
		v := plan.IsActive.ValueBool()
		apiReq.IsActive = &v
	}

	calendar, err := r.client.UpdateOnCallCalendar(ctx, state.ID.ValueString(), apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating on-call calendar", err.Error())
		return
	}

	mapOnCallCalendarToState(calendar, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *OnCallCalendarResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state OnCallCalendarResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteOnCallCalendar(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting on-call calendar", err.Error())
	}
}

func (r *OnCallCalendarResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func mapOnCallCalendarToState(calendar *client.OnCallCalendar, state *OnCallCalendarResourceModel) {
	state.ID = types.StringValue(calendar.ID)
	state.Name = types.StringValue(calendar.Name)
	state.Description = types.StringValue(calendar.Description)
	state.IsActive = types.BoolValue(calendar.IsActive)
	state.CreatedAt = types.StringValue(calendar.CreatedAt)
	state.UpdatedAt = types.StringValue(calendar.UpdatedAt)
}
