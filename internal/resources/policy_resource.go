package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
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
	_ resource.Resource                = &PolicyResource{}
	_ resource.ResourceWithImportState = &PolicyResource{}
)

type PolicyResource struct {
	client *client.Client
}

type PolicyResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Type       types.String `tfsdk:"type"`
	RetryCount types.Int64  `tfsdk:"retry_count"`
	RetryDelay types.Int64  `tfsdk:"retry_delay"`
	Steps      types.List   `tfsdk:"steps"`
	CreatedAt  types.String `tfsdk:"created_at"`
	UpdatedAt  types.String `tfsdk:"updated_at"`
}

type PolicyStepModel struct {
	WaitBefore       types.Int64 `tfsdk:"wait_before"`
	Call             types.Bool  `tfsdk:"call"`
	PushNotification types.Bool  `tfsdk:"push_notification"`
	SMS              types.Bool  `tfsdk:"sms"`
	Email            types.Bool  `tfsdk:"email"`
	Members          types.List  `tfsdk:"members"`
}

type PolicyStepMemberModel struct {
	Type types.String `tfsdk:"type"`
	ID   types.String `tfsdk:"id"`
}

func NewPolicyResource() resource.Resource {
	return &PolicyResource{}
}

func (r *PolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy"
}

func policyStepMemberAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type": types.StringType,
		"id":   types.StringType,
	}
}

func policyStepAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"wait_before":       types.Int64Type,
		"call":              types.BoolType,
		"push_notification": types.BoolType,
		"sms":               types.BoolType,
		"email":             types.BoolType,
		"members":           types.ListType{ElemType: types.ObjectType{AttrTypes: policyStepMemberAttrTypes()}},
	}
}

func (r *PolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Mihari alert policy with escalation steps.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the policy.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the alert policy.",
			},
			"type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Policy type: template or default.",
				Validators: []validator.String{
					stringvalidator.OneOf("template", "default"),
				},
			},
			"retry_count": schema.Int64Attribute{
				Required:    true,
				Description: "Number of retries before escalating (1-20).",
				Validators: []validator.Int64{
					int64validator.Between(1, 20),
				},
			},
			"retry_delay": schema.Int64Attribute{
				Required:    true,
				Description: "Delay between retries in minutes (1-60).",
				Validators: []validator.Int64{
					int64validator.Between(1, 60),
				},
			},
			"steps": schema.ListNestedAttribute{
				Required:    true,
				Description: "Escalation steps (1-10).",
				Validators: []validator.List{
					listvalidator.SizeBetween(1, 10),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"wait_before": schema.Int64Attribute{
							Required:    true,
							Description: "Minutes to wait before this step (0-60).",
							Validators: []validator.Int64{
								int64validator.Between(0, 60),
							},
						},
						"call": schema.BoolAttribute{
							Required:    true,
							Description: "Enable phone call notification.",
						},
						"push_notification": schema.BoolAttribute{
							Required:    true,
							Description: "Enable push notification.",
						},
						"sms": schema.BoolAttribute{
							Required:    true,
							Description: "Enable SMS notification.",
						},
						"email": schema.BoolAttribute{
							Required:    true,
							Description: "Enable email notification.",
						},
						"members": schema.ListNestedAttribute{
							Required:    true,
							Description: "Members to notify (1-20).",
							Validators: []validator.List{
								listvalidator.SizeBetween(1, 20),
							},
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{
										Required:    true,
										Description: "Member type: user, current_persons_on_call, or teams.",
										Validators: []validator.String{
											stringvalidator.OneOf("user", "current_persons_on_call", "teams"),
										},
									},
									"id": schema.StringAttribute{
										Optional:    true,
										Description: "Member UUID. Required when type is 'user'.",
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

func (r *PolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq, diags := buildPolicyRequest(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := r.client.CreatePolicy(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating policy", err.Error())
		return
	}

	d := mapPolicyToState(ctx, policy, &plan)
	resp.Diagnostics.Append(d...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := r.client.GetPolicy(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading policy", err.Error())
		return
	}
	if policy == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	d := mapPolicyToState(ctx, policy, &state)
	resp.Diagnostics.Append(d...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *PolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state PolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq, diags := buildPolicyRequest(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := r.client.UpdatePolicy(ctx, state.ID.ValueString(), apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating policy", err.Error())
		return
	}

	d := mapPolicyToState(ctx, policy, &plan)
	resp.Diagnostics.Append(d...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeletePolicy(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting policy", err.Error())
	}
}

func (r *PolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func buildPolicyRequest(ctx context.Context, plan *PolicyResourceModel) (client.PolicyRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	apiReq := client.PolicyRequest{
		Name:       plan.Name.ValueString(),
		RetryCount: int(plan.RetryCount.ValueInt64()),
		RetryDelay: int(plan.RetryDelay.ValueInt64()),
	}

	if !plan.Type.IsNull() && !plan.Type.IsUnknown() {
		apiReq.Type = plan.Type.ValueString()
	}

	var steps []PolicyStepModel
	diags.Append(plan.Steps.ElementsAs(ctx, &steps, false)...)
	if diags.HasError() {
		return apiReq, diags
	}

	for _, step := range steps {
		stepReq := client.PolicyStepRequest{
			WaitBefore:       int(step.WaitBefore.ValueInt64()),
			Call:             step.Call.ValueBool(),
			PushNotification: step.PushNotification.ValueBool(),
			SMS:              step.SMS.ValueBool(),
			Email:            step.Email.ValueBool(),
		}

		var members []PolicyStepMemberModel
		diags.Append(step.Members.ElementsAs(ctx, &members, false)...)
		if diags.HasError() {
			return apiReq, diags
		}

		for _, member := range members {
			memberReq := client.PolicyStepMemberRequest{
				Type: member.Type.ValueString(),
			}
			if !member.ID.IsNull() && !member.ID.IsUnknown() {
				v := member.ID.ValueString()
				memberReq.ID = &v
			}
			stepReq.Members = append(stepReq.Members, memberReq)
		}

		apiReq.Steps = append(apiReq.Steps, stepReq)
	}

	return apiReq, diags
}

func mapPolicyToState(ctx context.Context, policy *client.Policy, state *PolicyResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(policy.ID)
	state.Name = types.StringValue(policy.Name)
	state.Type = types.StringValue(policy.Type)
	state.RetryCount = types.Int64Value(int64(policy.RetryCount))
	state.RetryDelay = types.Int64Value(int64(policy.RetryDelay))
	state.CreatedAt = types.StringValue(policy.CreatedAt)
	state.UpdatedAt = types.StringValue(policy.UpdatedAt)

	var stepValues []attr.Value
	for _, step := range policy.Steps {
		var memberValues []attr.Value
		for _, member := range step.Members {
			memberID := types.StringNull()
			if member.ID != nil {
				memberID = types.StringValue(*member.ID)
			}
			memberObj, d := types.ObjectValue(policyStepMemberAttrTypes(), map[string]attr.Value{
				"type": types.StringValue(member.Type),
				"id":   memberID,
			})
			diags.Append(d...)
			memberValues = append(memberValues, memberObj)
		}

		membersList, d := types.ListValue(types.ObjectType{AttrTypes: policyStepMemberAttrTypes()}, memberValues)
		diags.Append(d...)

		stepObj, d := types.ObjectValue(policyStepAttrTypes(), map[string]attr.Value{
			"wait_before":       types.Int64Value(int64(step.WaitBefore)),
			"call":              types.BoolValue(step.Call),
			"push_notification": types.BoolValue(step.PushNotification),
			"sms":               types.BoolValue(step.SMS),
			"email":             types.BoolValue(step.Email),
			"members":           membersList,
		})
		diags.Append(d...)
		stepValues = append(stepValues, stepObj)
	}

	stepsList, d := types.ListValue(types.ObjectType{AttrTypes: policyStepAttrTypes()}, stepValues)
	diags.Append(d...)
	state.Steps = stepsList

	return diags
}
