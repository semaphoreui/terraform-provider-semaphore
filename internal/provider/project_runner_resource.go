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
	_ resource.Resource                = &projectRunnerResource{}
	_ resource.ResourceWithConfigure   = &projectRunnerResource{}
	_ resource.ResourceWithImportState = &projectRunnerResource{}
)

func NewProjectRunnerResource() resource.Resource {
	return &projectRunnerResource{}
}

type projectRunnerResource struct {
	client *apiclient.SemaphoreUI
}

func (r *projectRunnerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *projectRunnerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_runner"
}

func (r *projectRunnerResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ProjectRunnerSchema().GetResource(ctx)
}

func convertProjectRunnerModelToRunnerRequest(ctx context.Context, model ProjectRunnerModel) (*models.RunnerRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	request := models.RunnerRequest{
		ProjectID:        model.ProjectID.ValueInt64(),
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

// convertRunnerResponseToProjectRunnerModel maps an API response onto the model.
// registrationToken is supplied by the caller: the API only returns the token
// once, in the creation response, so on subsequent reads we pass the value
// carried over from prior state rather than losing it.
func convertRunnerResponseToProjectRunnerModel(ctx context.Context, response *models.Runner, projectID types.Int64, registrationToken types.String) (ProjectRunnerModel, diag.Diagnostics) {
	tagsSource := response.Tags
	if tagsSource == nil {
		tagsSource = []string{}
	}
	tags, diags := types.SetValueFrom(ctx, types.StringType, tagsSource)
	model := ProjectRunnerModel{
		ID:                types.Int64Value(response.ID),
		ProjectID:         projectID,
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

func (r *projectRunnerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ProjectRunnerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request, diags := convertProjectRunnerModelToRunnerRequest(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.client.Runner.PostProjectProjectIDRunners(&runner.PostProjectProjectIDRunnersParams{
		ProjectID: plan.ProjectID.ValueInt64(),
		Runner:    request,
	}, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating SemaphoreUI Project Runner",
			"Could not create project runner, unexpected error: "+err.Error(),
		)
		return
	}

	model, diags := convertRunnerResponseToProjectRunnerModel(ctx, &response.Payload.Runner, plan.ProjectID, types.StringValue(response.Payload.Token))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *projectRunnerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ProjectRunnerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.client.Runner.GetProjectProjectIDRunnersRunnerID(&runner.GetProjectProjectIDRunnersRunnerIDParams{
		ProjectID: state.ProjectID.ValueInt64(),
		RunnerID:  state.ID.ValueInt64(),
	}, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading SemaphoreUI Project Runner",
			"Could not read project runner, unexpected error: "+err.Error(),
		)
		return
	}

	model, diags := convertRunnerResponseToProjectRunnerModel(ctx, response.Payload, state.ProjectID, state.RegistrationToken)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *projectRunnerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ProjectRunnerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request, diags := convertProjectRunnerModelToRunnerRequest(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Runner.PutProjectProjectIDRunnersRunnerID(&runner.PutProjectProjectIDRunnersRunnerIDParams{
		ProjectID: plan.ProjectID.ValueInt64(),
		RunnerID:  plan.ID.ValueInt64(),
		Runner:    request,
	}, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating SemaphoreUI Project Runner",
			"Could not update project runner, unexpected error: "+err.Error(),
		)
		return
	}

	response, err := r.client.Runner.GetProjectProjectIDRunnersRunnerID(&runner.GetProjectProjectIDRunnersRunnerIDParams{
		ProjectID: plan.ProjectID.ValueInt64(),
		RunnerID:  plan.ID.ValueInt64(),
	}, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading SemaphoreUI Project Runner",
			"Could not read project runner, unexpected error: "+err.Error(),
		)
		return
	}

	model, diags := convertRunnerResponseToProjectRunnerModel(ctx, response.Payload, plan.ProjectID, state.RegistrationToken)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *projectRunnerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ProjectRunnerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Runner.DeleteProjectProjectIDRunnersRunnerID(&runner.DeleteProjectProjectIDRunnersRunnerIDParams{
		ProjectID: state.ProjectID.ValueInt64(),
		RunnerID:  state.ID.ValueInt64(),
	}, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Removing SemaphoreUI Project Runner",
			"Could not remove project runner, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *projectRunnerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	fields, err := parseImportFields(req.ID, []string{"project", "runner"})
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Project Runner Import ID",
			"Could not parse import ID: "+err.Error(),
		)
		return
	}

	response, err := r.client.Runner.GetProjectProjectIDRunnersRunnerID(&runner.GetProjectProjectIDRunnersRunnerIDParams{
		ProjectID: fields["project"],
		RunnerID:  fields["runner"],
	}, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading SemaphoreUI Project Runner",
			"Could not read project runner, unexpected error: "+err.Error(),
		)
		return
	}

	// The registration token is only returned when the runner is created, so it
	// is not available on import and remains an empty string in state.
	model, diags := convertRunnerResponseToProjectRunnerModel(ctx, response.Payload, types.Int64Value(fields["project"]), types.StringValue(""))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
