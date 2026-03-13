package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
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
	_ resource.Resource                = &HeartbeatResource{}
	_ resource.ResourceWithImportState = &HeartbeatResource{}
)

type HeartbeatResource struct {
	client *client.Client
}

type HeartbeatResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Period      types.Int64  `tfsdk:"period"`
	GracePeriod types.Int64  `tfsdk:"grace_period"`
	IsActive    types.Bool   `tfsdk:"is_active"`
	PolicyID    types.String `tfsdk:"policy_id"`
	Status      types.String `tfsdk:"status"`
	LastPingAt  types.String `tfsdk:"last_ping_at"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func NewHeartbeatResource() resource.Resource {
	return &HeartbeatResource{}
}

func (r *HeartbeatResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_heartbeat"
}

func (r *HeartbeatResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Mihari heartbeat monitor.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the heartbeat.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the heartbeat.",
			},
			"period": schema.Int64Attribute{
				Required:    true,
				Description: "Expected period between pings in minutes.",
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"grace_period": schema.Int64Attribute{
				Required:    true,
				Description: "Grace period in minutes before alerting.",
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"is_active": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the heartbeat is active.",
			},
			"policy_id": schema.StringAttribute{
				Optional:    true,
				Description: "UUID of the alert policy to attach.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Current heartbeat status.",
			},
			"last_ping_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of the last ping.",
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

func (r *HeartbeatResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *HeartbeatResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan HeartbeatResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := client.HeartbeatRequest{
		Name:        plan.Name.ValueString(),
		Period:      int(plan.Period.ValueInt64()),
		GracePeriod: int(plan.GracePeriod.ValueInt64()),
	}
	if !plan.IsActive.IsNull() && !plan.IsActive.IsUnknown() {
		v := plan.IsActive.ValueBool()
		apiReq.IsActive = &v
	}
	if !plan.PolicyID.IsNull() && !plan.PolicyID.IsUnknown() {
		v := plan.PolicyID.ValueString()
		apiReq.PolicyID = &v
	}

	heartbeat, err := r.client.CreateHeartbeat(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating heartbeat", err.Error())
		return
	}

	mapHeartbeatToState(heartbeat, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *HeartbeatResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state HeartbeatResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	heartbeat, err := r.client.GetHeartbeat(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading heartbeat", err.Error())
		return
	}
	if heartbeat == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	mapHeartbeatToState(heartbeat, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *HeartbeatResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan HeartbeatResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state HeartbeatResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := client.HeartbeatRequest{
		Name:        plan.Name.ValueString(),
		Period:      int(plan.Period.ValueInt64()),
		GracePeriod: int(plan.GracePeriod.ValueInt64()),
	}
	if !plan.IsActive.IsNull() && !plan.IsActive.IsUnknown() {
		v := plan.IsActive.ValueBool()
		apiReq.IsActive = &v
	}
	if !plan.PolicyID.IsNull() && !plan.PolicyID.IsUnknown() {
		v := plan.PolicyID.ValueString()
		apiReq.PolicyID = &v
	}

	heartbeat, err := r.client.UpdateHeartbeat(ctx, state.ID.ValueString(), apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating heartbeat", err.Error())
		return
	}

	mapHeartbeatToState(heartbeat, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *HeartbeatResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state HeartbeatResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteHeartbeat(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting heartbeat", err.Error())
	}
}

func (r *HeartbeatResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func mapHeartbeatToState(heartbeat *client.Heartbeat, state *HeartbeatResourceModel) {
	state.ID = types.StringValue(heartbeat.ID)
	state.Name = types.StringValue(heartbeat.Name)
	state.Period = types.Int64Value(int64(heartbeat.Period))
	state.GracePeriod = types.Int64Value(int64(heartbeat.GracePeriod))
	state.IsActive = types.BoolValue(heartbeat.IsActive)
	state.Status = types.StringValue(heartbeat.Status)
	state.CreatedAt = types.StringValue(heartbeat.CreatedAt)
	state.UpdatedAt = types.StringValue(heartbeat.UpdatedAt)
	if heartbeat.PolicyID != nil {
		state.PolicyID = types.StringValue(*heartbeat.PolicyID)
	} else {
		state.PolicyID = types.StringNull()
	}
	if heartbeat.LastPingAt != nil {
		state.LastPingAt = types.StringValue(*heartbeat.LastPingAt)
	} else {
		state.LastPingAt = types.StringNull()
	}
}
