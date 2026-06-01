package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	apiclient "terraform-provider-semaphoreui/semaphoreui/client"
	"terraform-provider-semaphoreui/semaphoreui/client/runner"
	"terraform-provider-semaphoreui/semaphoreui/models"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &runnerResource{}
	_ resource.ResourceWithConfigure   = &runnerResource{}
	_ resource.ResourceWithImportState = &runnerResource{}
)

func NewRunnerResource() resource.Resource {
	return &runnerResource{}
}

type runnerResource struct {
	client *apiclient.SemaphoreUI
}

func (r *runnerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*apiclient.SemaphoreUI)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			"Expected *client.SemaphoreUI, got %T. Please report this issue to the provider developers.",
		)
		return
	}
	r.client = client
}

func (r *runnerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_runner"
}

func (r *runnerResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = RunnerSchema().GetResource(ctx)
}

func convertRunnerModelToRunnerRequest(ctx context.Context, model RunnerModel) (*models.RunnerRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	request := models.RunnerRequest{
		Name:             model.Name.ValueString(),
		Webhook:          model.Webhook.ValueString(),
		MaxParallelTasks: model.MaxParallelTasks.ValueInt64(),
		Active:           model.Active.ValueBool(),
	}
	if !model.Tags.IsNull() && !model.Tags.IsUnknown() {
		var tags []string
		diags.Append(model.Tags.ElementsAs(ctx, &tags, false)...)
		request.Tags = tags
	}
	return &request, diags
}

// convertRunnerResponseToRunnerModel maps an API response onto the model.
// registrationToken is supplied by the caller: the API only returns the token
// once, in the creation response, so on subsequent reads we pass the value
// carried over from prior state rather than losing it.
func convertRunnerResponseToRunnerModel(ctx context.Context, response *models.Runner, registrationToken types.String) (RunnerModel, diag.Diagnostics) {
	tagsSource := response.Tags
	if tagsSource == nil {
		tagsSource = []string{}
	}
	tags, diags := types.SetValueFrom(ctx, types.StringType, tagsSource)
	model := RunnerModel{
		ID:                types.Int64Value(response.ID),
		Name:              types.StringValue(response.Name),
		Webhook:           types.StringValue(response.Webhook),
		MaxParallelTasks:  types.Int64Value(response.MaxParallelTasks),
		Active:            types.BoolValue(response.Active),
		Tags:              tags,
		RegistrationToken: registrationToken,
		IsDefault:         types.BoolValue(response.IsDefault),
	}
	return model, diags
}

func (r *runnerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RunnerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request, diags := convertRunnerModelToRunnerRequest(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.client.Runner.PostRunners(&runner.PostRunnersParams{
		Runner: request,
	}, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating SemaphoreUI Runner",
			"Could not create runner, unexpected error: "+err.Error(),
		)
		return
	}

	if err := r.ensureActive(response.Payload.ID, plan.Active.ValueBool(), response.Payload.Active); err != nil {
		resp.Diagnostics.AddError(
			"Error Setting SemaphoreUI Runner Active State",
			"Could not set runner active state, unexpected error: "+err.Error(),
		)
		return
	}

	model, diags := convertRunnerResponseToRunnerModel(ctx, &response.Payload.Runner, resolveRegistrationToken(response.Payload))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Creation may ignore the `active` flag (it is applied via the dedicated
	// endpoint above), so reflect the planned value rather than the response.
	model.Active = plan.Active

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// resolveRegistrationToken returns the one-time token from a create response,
// reading the `registration_token` field first and falling back to `token`
// (older SemaphoreUI versions populate the latter).
func resolveRegistrationToken(payload *models.RunnerWithToken) types.String {
	if payload.RegistrationToken != nil && *payload.RegistrationToken != "" {
		return types.StringValue(*payload.RegistrationToken)
	}
	return types.StringValue(payload.Token)
}

// ensureActive sets the runner active state via the dedicated endpoint when the
// current state differs from the desired one. Some SemaphoreUI versions ignore
// the `active` field on create/update and only honor this endpoint.
func (r *runnerResource) ensureActive(runnerID int64, desired, current bool) error {
	if desired == current {
		return nil
	}
	_, err := r.client.Runner.PostRunnersRunnerIDActive(&runner.PostRunnersRunnerIDActiveParams{
		RunnerID: runnerID,
		Active:   &models.RunnerActive{Active: desired},
	}, nil)
	return err
}

func (r *runnerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RunnerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.client.Runner.GetRunnersRunnerID(&runner.GetRunnersRunnerIDParams{
		RunnerID: state.ID.ValueInt64(),
	}, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading SemaphoreUI Runner",
			"Could not read runner, unexpected error: "+err.Error(),
		)
		return
	}

	model, diags := convertRunnerResponseToRunnerModel(ctx, response.Payload, state.RegistrationToken)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *runnerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state RunnerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request, diags := convertRunnerModelToRunnerRequest(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Runner.PutRunnersRunnerID(&runner.PutRunnersRunnerIDParams{
		RunnerID: plan.ID.ValueInt64(),
		Runner:   request,
	}, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating SemaphoreUI Runner",
			"Could not update runner, unexpected error: "+err.Error(),
		)
		return
	}

	response, err := r.client.Runner.GetRunnersRunnerID(&runner.GetRunnersRunnerIDParams{
		RunnerID: plan.ID.ValueInt64(),
	}, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading SemaphoreUI Runner",
			"Could not read runner, unexpected error: "+err.Error(),
		)
		return
	}

	if err := r.ensureActive(plan.ID.ValueInt64(), plan.Active.ValueBool(), response.Payload.Active); err != nil {
		resp.Diagnostics.AddError(
			"Error Setting SemaphoreUI Runner Active State",
			"Could not set runner active state, unexpected error: "+err.Error(),
		)
		return
	}

	model, diags := convertRunnerResponseToRunnerModel(ctx, response.Payload, state.RegistrationToken)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	model.Active = plan.Active

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *runnerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RunnerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Runner.DeleteRunnersRunnerID(&runner.DeleteRunnersRunnerIDParams{
		RunnerID: state.ID.ValueInt64(),
	}, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Removing SemaphoreUI Runner",
			"Could not remove runner, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *runnerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	fields, err := parseImportFields(req.ID, []string{"runner"})
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Runner Import ID",
			"Could not parse import ID: "+err.Error(),
		)
		return
	}

	response, err := r.client.Runner.GetRunnersRunnerID(&runner.GetRunnersRunnerIDParams{
		RunnerID: fields["runner"],
	}, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading SemaphoreUI Runner",
			"Could not read runner, unexpected error: "+err.Error(),
		)
		return
	}

	// The registration token is only returned when the runner is created, so it
	// is not available on import and remains an empty string in state.
	model, diags := convertRunnerResponseToRunnerModel(ctx, response.Payload, types.StringValue(""))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
