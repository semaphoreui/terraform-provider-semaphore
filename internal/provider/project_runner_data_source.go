package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	apiclient "terraform-provider-semaphoreui/semaphoreui/client"
	"terraform-provider-semaphoreui/semaphoreui/client/runner"
	"terraform-provider-semaphoreui/semaphoreui/models"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &projectRunnerDataSource{}
)

func NewProjectRunnerDataSource() datasource.DataSource {
	return &projectRunnerDataSource{}
}

type projectRunnerDataSource struct {
	client *apiclient.SemaphoreUI
}

func (d *projectRunnerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.client = client
}

func (d *projectRunnerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_runner"
}

func (d *projectRunnerDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = ProjectRunnerSchema().GetDataSource(ctx)
}

func (d *projectRunnerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ProjectRunnerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.ID.IsNull() && !config.ID.IsUnknown() {
		response, err := d.client.Runner.GetProjectProjectIDRunnersRunnerID(&runner.GetProjectProjectIDRunnersRunnerIDParams{
			ProjectID: config.ProjectID.ValueInt64(),
			RunnerID:  config.ID.ValueInt64(),
		}, nil)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading SemaphoreUI Project Runner",
				"Could not read project runner, unexpected error: "+err.Error(),
			)
			return
		}
		model, diags := convertRunnerResponseToProjectRunnerModel(ctx, response.Payload, config.ProjectID)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
		return
	}

	response, err := d.client.Runner.GetProjectProjectIDRunners(&runner.GetProjectProjectIDRunnersParams{
		ProjectID: config.ProjectID.ValueInt64(),
	}, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading SemaphoreUI Project Runners",
			"Could not read project runners, unexpected error: "+err.Error(),
		)
		return
	}
	for _, item := range response.Payload {
		if item.Name == config.Name.ValueString() {
			// The list endpoint returns bare Runner objects (no token/private
			// key); wrap so the shared converter leaves those fields null.
			model, diags := convertRunnerResponseToProjectRunnerModel(ctx, &models.RunnerWithToken{Runner: *item}, config.ProjectID)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
			return
		}
	}
	resp.Diagnostics.AddError(
		"Error Reading SemaphoreUI Project Runner",
		fmt.Sprintf("project runner with name %q not found in project %d", config.Name.ValueString(), config.ProjectID.ValueInt64()),
	)
}
