package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
	_ resource.Resource                = &MonitorResource{}
	_ resource.ResourceWithImportState = &MonitorResource{}
)

type MonitorResource struct {
	client *client.Client
}

type MonitorResourceModel struct {
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
	Headers            types.Map    `tfsdk:"headers"`
	CheckSSL           types.Bool   `tfsdk:"check_ssl"`
	IsActive           types.Bool   `tfsdk:"is_active"`
	PolicyID           types.String `tfsdk:"policy_id"`
	Status             types.String `tfsdk:"status"`
	LastCheckedAt      types.String `tfsdk:"last_checked_at"`
	CreatedAt          types.String `tfsdk:"created_at"`
	UpdatedAt          types.String `tfsdk:"updated_at"`
}

func NewMonitorResource() resource.Resource {
	return &MonitorResource{}
}

func (r *MonitorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitor"
}

func (r *MonitorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Mihari monitor.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the monitor.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Display name of the monitor.",
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "Monitor type: url_contains, url_not_contains, http_status, ping, tcp_port, udp_port, smtp, pop3, imap, dns, playwright.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						"url_contains", "url_not_contains", "http_status",
						"ping", "tcp_port", "udp_port", "smtp", "pop3",
						"imap", "dns", "playwright",
					),
				},
			},
			"url": schema.StringAttribute{
				Optional:    true,
				Description: "URL to monitor. Required for url_contains, url_not_contains, and http_status types.",
			},
			"keyword": schema.StringAttribute{
				Optional:    true,
				Description: "Keyword to search for. Required for url_contains and url_not_contains types.",
			},
			"expected_status_code": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Expected HTTP status code (100-599).",
				Validators: []validator.Int64{
					int64validator.Between(100, 599),
				},
			},
			"host": schema.StringAttribute{
				Optional:    true,
				Description: "Host to monitor. Required for ping, tcp_port, udp_port, smtp, pop3, imap, dns types.",
			},
			"port": schema.Int64Attribute{
				Optional:    true,
				Description: "Port number (1-65535). Required for tcp_port, udp_port, smtp, pop3, imap types.",
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
			"protocol": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Protocol: http or https.",
				Validators: []validator.String{
					stringvalidator.OneOf("http", "https"),
				},
			},
			"check_interval": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Check interval in minutes (1-1440).",
				Validators: []validator.Int64{
					int64validator.Between(1, 1440),
				},
			},
			"timeout": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Timeout in seconds (1-300).",
				Validators: []validator.Int64{
					int64validator.Between(1, 300),
				},
			},
			"headers": schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Custom HTTP headers.",
			},
			"check_ssl": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether to validate SSL certificates.",
			},
			"is_active": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the monitor is active.",
			},
			"policy_id": schema.StringAttribute{
				Optional:    true,
				Description: "UUID of the alert policy.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Current status: up, down, recovery, acknowledge.",
			},
			"last_checked_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of last check.",
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

func (r *MonitorResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *MonitorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan MonitorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := buildMonitorRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	monitor, err := r.client.CreateMonitor(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating monitor", err.Error())
		return
	}

	mapMonitorToState(ctx, monitor, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *MonitorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state MonitorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	monitor, err := r.client.GetMonitor(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading monitor", err.Error())
		return
	}
	if monitor == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	mapMonitorToState(ctx, monitor, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *MonitorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan MonitorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state MonitorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := buildMonitorRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	monitor, err := r.client.UpdateMonitor(ctx, state.ID.ValueString(), apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating monitor", err.Error())
		return
	}

	mapMonitorToState(ctx, monitor, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *MonitorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state MonitorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMonitor(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting monitor", err.Error())
	}
}

func (r *MonitorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func buildMonitorRequest(ctx context.Context, plan *MonitorResourceModel, diags *diag.Diagnostics) client.MonitorRequest {
	apiReq := client.MonitorRequest{
		Name: plan.Name.ValueString(),
		Type: plan.Type.ValueString(),
	}

	if !plan.URL.IsNull() && !plan.URL.IsUnknown() {
		apiReq.URL = plan.URL.ValueString()
	}
	if !plan.Keyword.IsNull() && !plan.Keyword.IsUnknown() {
		apiReq.Keyword = plan.Keyword.ValueString()
	}
	if !plan.ExpectedStatusCode.IsNull() && !plan.ExpectedStatusCode.IsUnknown() {
		v := int(plan.ExpectedStatusCode.ValueInt64())
		apiReq.ExpectedStatusCode = &v
	}
	if !plan.Host.IsNull() && !plan.Host.IsUnknown() {
		apiReq.Host = plan.Host.ValueString()
	}
	if !plan.Port.IsNull() && !plan.Port.IsUnknown() {
		v := int(plan.Port.ValueInt64())
		apiReq.Port = &v
	}
	if !plan.Protocol.IsNull() && !plan.Protocol.IsUnknown() {
		apiReq.Protocol = plan.Protocol.ValueString()
	}
	if !plan.CheckInterval.IsNull() && !plan.CheckInterval.IsUnknown() {
		v := int(plan.CheckInterval.ValueInt64())
		apiReq.CheckInterval = &v
	}
	if !plan.Timeout.IsNull() && !plan.Timeout.IsUnknown() {
		v := int(plan.Timeout.ValueInt64())
		apiReq.Timeout = &v
	}
	if !plan.CheckSSL.IsNull() && !plan.CheckSSL.IsUnknown() {
		v := plan.CheckSSL.ValueBool()
		apiReq.CheckSSL = &v
	}
	if !plan.IsActive.IsNull() && !plan.IsActive.IsUnknown() {
		v := plan.IsActive.ValueBool()
		apiReq.IsActive = &v
	}
	if !plan.PolicyID.IsNull() && !plan.PolicyID.IsUnknown() {
		v := plan.PolicyID.ValueString()
		apiReq.PolicyID = &v
	}
	if !plan.Headers.IsNull() && !plan.Headers.IsUnknown() {
		headers := make(map[string]string)
		plan.Headers.ElementsAs(ctx, &headers, false)
		apiReq.Headers = headers
	}

	return apiReq
}

func mapMonitorToState(ctx context.Context, monitor *client.Monitor, state *MonitorResourceModel) {
	state.ID = types.StringValue(monitor.ID)
	// Critical mapping: API returns "title" for the name field
	state.Name = types.StringValue(monitor.Title)
	state.Type = types.StringValue(monitor.Type)
	state.URL = types.StringValue(monitor.URL)
	state.Keyword = types.StringValue(monitor.Keyword)
	state.Host = types.StringValue(monitor.Host)
	state.Protocol = types.StringValue(monitor.Protocol)
	state.CheckInterval = types.Int64Value(int64(monitor.CheckInterval))
	state.Timeout = types.Int64Value(int64(monitor.Timeout))
	state.CheckSSL = types.BoolValue(monitor.CheckSSL)
	state.IsActive = types.BoolValue(monitor.IsActive)
	state.Status = types.StringValue(monitor.Status)
	state.CreatedAt = types.StringValue(monitor.CreatedAt)
	state.UpdatedAt = types.StringValue(monitor.UpdatedAt)

	if monitor.ExpectedStatusCode != nil {
		state.ExpectedStatusCode = types.Int64Value(int64(*monitor.ExpectedStatusCode))
	} else {
		state.ExpectedStatusCode = types.Int64Null()
	}
	if monitor.Port != nil {
		state.Port = types.Int64Value(int64(*monitor.Port))
	} else {
		state.Port = types.Int64Null()
	}
	if monitor.PolicyID != nil {
		state.PolicyID = types.StringValue(*monitor.PolicyID)
	} else {
		state.PolicyID = types.StringNull()
	}
	if monitor.LastCheckedAt != nil {
		state.LastCheckedAt = types.StringValue(*monitor.LastCheckedAt)
	} else {
		state.LastCheckedAt = types.StringNull()
	}

	if monitor.Headers != nil && len(monitor.Headers) > 0 {
		headersMap, _ := types.MapValueFrom(ctx, types.StringType, monitor.Headers)
		state.Headers = headersMap
	} else {
		state.Headers = types.MapNull(types.StringType)
	}
}
