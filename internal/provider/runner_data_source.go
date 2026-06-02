package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	apiclient "terraform-provider-semaphoreui/semaphoreui/client"
	"terraform-provider-semaphoreui/semaphoreui/client/runner"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &runnerDataSource{}
)

func NewRunnerDataSource() datasource.DataSource {
	return &runnerDataSource{}
}

type runnerDataSource struct {
	client *apiclient.SemaphoreUI
}

func (d *runnerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *runnerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_runner"
}

func (d *runnerDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = RunnerSchema().GetDataSource(ctx)
}

func (d *runnerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config RunnerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.ID.IsNull() && !config.ID.IsUnknown() {
		response, err := d.client.Runner.GetRunnersRunnerID(&runner.GetRunnersRunnerIDParams{
			RunnerID: config.ID.ValueInt64(),
		}, nil)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading SemaphoreUI Runner",
				"Could not read runner, unexpected error: "+err.Error(),
			)
			return
		}
		model, diags := convertRunnerResponseToRunnerModel(ctx, response.Payload)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
		return
	}

	response, err := d.client.Runner.GetRunners(&runner.GetRunnersParams{}, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading SemaphoreUI Runners",
			"Could not read runners, unexpected error: "+err.Error(),
		)
		return
	}
	for _, item := range response.Payload {
		if item.Name == config.Name.ValueString() {
			model, diags := convertRunnerResponseToRunnerModel(ctx, item)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
			return
		}
	}
	resp.Diagnostics.AddError(
		"Error Reading SemaphoreUI Runner",
		fmt.Sprintf("runner with name %q not found", config.Name.ValueString()),
	)
}
