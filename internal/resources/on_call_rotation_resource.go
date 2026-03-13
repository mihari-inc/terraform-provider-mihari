package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mihari-io/terraform-provider-mihari/internal/client"
)

var (
	_ resource.Resource                = &OnCallRotationResource{}
	_ resource.ResourceWithImportState = &OnCallRotationResource{}
)

type OnCallRotationResource struct {
	client *client.Client
}

type OnCallRotationResourceModel struct {
	ID               types.String `tfsdk:"id"`
	OnCallCalendarID types.String `tfsdk:"on_call_calendar_id"`
	StartDate        types.String `tfsdk:"start_date"`
	StartHour        types.String `tfsdk:"start_hour"`
	Duration         types.String `tfsdk:"duration"`
	RepeatDays       types.List   `tfsdk:"repeat_days"`
	RepeatEnd        types.String `tfsdk:"repeat_end"`
	Members          types.List   `tfsdk:"members"`
	Rrule            types.String `tfsdk:"rrule"`
	CreatedAt        types.String `tfsdk:"created_at"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
}

func rotationMemberAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"member_id": types.StringType,
	}
}

func NewOnCallRotationResource() resource.Resource {
	return &OnCallRotationResource{}
}

func (r *OnCallRotationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_on_call_rotation"
}

func (r *OnCallRotationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Mihari on-call rotation.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the rotation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"on_call_calendar_id": schema.StringAttribute{
				Required:    true,
				Description: "UUID of the parent on-call calendar.",
			},
			"start_date": schema.StringAttribute{
				Required:    true,
				Description: "Start date (YYYY-MM-DD).",
			},
			"start_hour": schema.StringAttribute{
				Required:    true,
				Description: "Start hour (HH:MM).",
			},
			"duration": schema.StringAttribute{
				Required:    true,
				Description: "Duration (HH:MM).",
			},
			"repeat_days": schema.ListAttribute{
				Required:    true,
				ElementType: types.Int64Type,
				Description: "Days to repeat (0=Sunday, 1=Monday, ..., 6=Saturday).",
			},
			"repeat_end": schema.StringAttribute{
				Required:    true,
				Description: "End date for repetition (YYYY-MM-DD).",
			},
			"members": schema.ListNestedAttribute{
				Required:    true,
				Description: "Rotation members.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"member_id": schema.StringAttribute{
							Required:    true,
							Description: "UUID of the team member.",
						},
					},
				},
			},
			"rrule": schema.StringAttribute{
				Computed:    true,
				Description: "Generated recurrence rule string.",
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

func (r *OnCallRotationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *OnCallRotationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan OnCallRotationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq, diags := buildOnCallRotationRequest(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rotation, err := r.client.CreateOnCallRotation(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating on-call rotation", err.Error())
		return
	}

	d := mapOnCallRotationToState(ctx, rotation, &plan)
	resp.Diagnostics.Append(d...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *OnCallRotationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state OnCallRotationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rotation, err := r.client.GetOnCallRotation(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading on-call rotation", err.Error())
		return
	}
	if rotation == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Preserve input-only fields from state since API returns storage format
	state.Rrule = types.StringValue(rotation.Rrule)
	state.CreatedAt = types.StringValue(rotation.CreatedAt)
	state.UpdatedAt = types.StringValue(rotation.UpdatedAt)
	state.ID = types.StringValue(rotation.ID)
	state.OnCallCalendarID = types.StringValue(rotation.OnCallCalendarID)

	// Map members from API response
	var memberValues []attr.Value
	for _, member := range rotation.Members {
		memberObj, d := types.ObjectValue(rotationMemberAttrTypes(), map[string]attr.Value{
			"member_id": types.StringValue(member.MemberID),
		})
		resp.Diagnostics.Append(d...)
		memberValues = append(memberValues, memberObj)
	}
	if len(memberValues) > 0 {
		membersList, d := types.ListValue(types.ObjectType{AttrTypes: rotationMemberAttrTypes()}, memberValues)
		resp.Diagnostics.Append(d...)
		state.Members = membersList
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *OnCallRotationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan OnCallRotationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state OnCallRotationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq, diags := buildOnCallRotationRequest(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rotation, err := r.client.UpdateOnCallRotation(ctx, state.ID.ValueString(), apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating on-call rotation", err.Error())
		return
	}

	d := mapOnCallRotationToState(ctx, rotation, &plan)
	resp.Diagnostics.Append(d...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *OnCallRotationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state OnCallRotationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteOnCallRotation(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting on-call rotation", err.Error())
	}
}

func (r *OnCallRotationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func buildOnCallRotationRequest(ctx context.Context, plan *OnCallRotationResourceModel) (client.OnCallRotationRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	var repeatDays []int
	var int64Days []int64
	diags.Append(plan.RepeatDays.ElementsAs(ctx, &int64Days, false)...)
	for _, d := range int64Days {
		repeatDays = append(repeatDays, int(d))
	}

	type MemberPlan struct {
		MemberID types.String `tfsdk:"member_id"`
	}
	var members []MemberPlan
	diags.Append(plan.Members.ElementsAs(ctx, &members, false)...)

	var memberReqs []client.OnCallRotationMemberRequest
	for _, m := range members {
		id := m.MemberID.ValueString()
		memberReqs = append(memberReqs, client.OnCallRotationMemberRequest{
			ID:    id,
			Value: id,
			Label: "Member",
			Type:  "user",
		})
	}

	apiReq := client.OnCallRotationRequest{
		OnCallCalendarID: plan.OnCallCalendarID.ValueString(),
		Event: client.OnCallRotationEventRequest{
			Start: client.OnCallRotationTimeRequest{
				Date: plan.StartDate.ValueString(),
				Hour: plan.StartHour.ValueString(),
			},
			Duration: plan.Duration.ValueString(),
		},
		Repeat: client.OnCallRotationRepeatRequest{
			Days: repeatDays,
			End:  plan.RepeatEnd.ValueString(),
		},
		Members: memberReqs,
	}

	return apiReq, diags
}

func mapOnCallRotationToState(ctx context.Context, rotation *client.OnCallRotation, state *OnCallRotationResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(rotation.ID)
	state.OnCallCalendarID = types.StringValue(rotation.OnCallCalendarID)
	state.Rrule = types.StringValue(rotation.Rrule)
	state.CreatedAt = types.StringValue(rotation.CreatedAt)
	state.UpdatedAt = types.StringValue(rotation.UpdatedAt)

	// Map members from response
	var memberValues []attr.Value
	for _, member := range rotation.Members {
		memberObj, d := types.ObjectValue(rotationMemberAttrTypes(), map[string]attr.Value{
			"member_id": types.StringValue(member.MemberID),
		})
		diags.Append(d...)
		memberValues = append(memberValues, memberObj)
	}

	if len(memberValues) > 0 {
		membersList, d := types.ListValue(types.ObjectType{AttrTypes: rotationMemberAttrTypes()}, memberValues)
		diags.Append(d...)
		state.Members = membersList
	}

	return diags
}
