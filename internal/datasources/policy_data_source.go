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

var _ datasource.DataSource = &PolicyDataSource{}

type PolicyDataSource struct {
	client *client.Client
}

type PolicyDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Type       types.String `tfsdk:"type"`
	RetryCount types.Int64  `tfsdk:"retry_count"`
	RetryDelay types.Int64  `tfsdk:"retry_delay"`
	Steps      types.List   `tfsdk:"steps"`
	CreatedAt  types.String `tfsdk:"created_at"`
	UpdatedAt  types.String `tfsdk:"updated_at"`
}

func policyStepMemberDSAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type": types.StringType,
		"id":   types.StringType,
	}
}

func policyStepDSAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"wait_before":       types.Int64Type,
		"call":              types.BoolType,
		"push_notification": types.BoolType,
		"sms":               types.BoolType,
		"email":             types.BoolType,
		"members":           types.ListType{ElemType: types.ObjectType{AttrTypes: policyStepMemberDSAttrTypes()}},
	}
}

func NewPolicyDataSource() datasource.DataSource {
	return &PolicyDataSource{}
}

func (d *PolicyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy"
}

func (d *PolicyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a single Mihari alert policy by ID.",
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Required: true, Description: "UUID of the policy."},
			"name":        schema.StringAttribute{Computed: true},
			"type":        schema.StringAttribute{Computed: true},
			"retry_count": schema.Int64Attribute{Computed: true},
			"retry_delay": schema.Int64Attribute{Computed: true},
			"steps": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"wait_before":       schema.Int64Attribute{Computed: true},
						"call":              schema.BoolAttribute{Computed: true},
						"push_notification": schema.BoolAttribute{Computed: true},
						"sms":               schema.BoolAttribute{Computed: true},
						"email":             schema.BoolAttribute{Computed: true},
						"members": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{Computed: true},
									"id":   schema.StringAttribute{Computed: true},
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

func (d *PolicyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config PolicyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := d.client.GetPolicy(ctx, config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading policy", err.Error())
		return
	}
	if policy == nil {
		resp.Diagnostics.AddError("Policy not found", fmt.Sprintf("Policy with ID %s not found", config.ID.ValueString()))
		return
	}

	config.Name = types.StringValue(policy.Name)
	config.Type = types.StringValue(policy.Type)
	config.RetryCount = types.Int64Value(int64(policy.RetryCount))
	config.RetryDelay = types.Int64Value(int64(policy.RetryDelay))
	config.CreatedAt = types.StringValue(policy.CreatedAt)
	config.UpdatedAt = types.StringValue(policy.UpdatedAt)

	var stepValues []attr.Value
	for _, step := range policy.Steps {
		var memberValues []attr.Value
		for _, member := range step.Members {
			memberID := types.StringNull()
			if member.ID != nil {
				memberID = types.StringValue(*member.ID)
			}
			memberObj, diags := types.ObjectValue(policyStepMemberDSAttrTypes(), map[string]attr.Value{
				"type": types.StringValue(member.Type),
				"id":   memberID,
			})
			resp.Diagnostics.Append(diags...)
			memberValues = append(memberValues, memberObj)
		}

		membersList, diags := types.ListValue(types.ObjectType{AttrTypes: policyStepMemberDSAttrTypes()}, memberValues)
		resp.Diagnostics.Append(diags...)

		stepObj, diags2 := types.ObjectValue(policyStepDSAttrTypes(), map[string]attr.Value{
			"wait_before":       types.Int64Value(int64(step.WaitBefore)),
			"call":              types.BoolValue(step.Call),
			"push_notification": types.BoolValue(step.PushNotification),
			"sms":               types.BoolValue(step.SMS),
			"email":             types.BoolValue(step.Email),
			"members":           membersList,
		})
		resp.Diagnostics.Append(diags2...)
		stepValues = append(stepValues, stepObj)
	}

	stepsList, diags := types.ListValue(types.ObjectType{AttrTypes: policyStepDSAttrTypes()}, stepValues)
	resp.Diagnostics.Append(diags...)
	config.Steps = stepsList

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
