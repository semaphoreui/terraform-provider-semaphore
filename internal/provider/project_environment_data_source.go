package provider

import (
	"context"
	"fmt"
	apiclient "terraform-provider-semaphoreui/semaphoreui/client"
	"terraform-provider-semaphoreui/semaphoreui/client/variable_group"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &projectEnvironmentDataSource{}
)

func NewProjectEnvironmentDataSource() datasource.DataSource {
	return &projectEnvironmentDataSource{}
}

type projectEnvironmentDataSource struct {
	client *apiclient.SemaphoreUI
}

func (d *projectEnvironmentDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Metadata returns the data source type name.
func (d *projectEnvironmentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_environment"
}

// Schema defines the schema for the data source.
func (d *projectEnvironmentDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = ProjectEnvironmentSchema().GetDataSource(ctx)
}

// getEnvironmentIDByName resolves a project environment name to its ID via the list endpoint.
func (d *projectEnvironmentDataSource) getEnvironmentIDByName(projectID int64, name string) (int64, error) {
	response, err := d.client.VariableGroup.GetProjectProjectIDEnvironment(&variable_group.GetProjectProjectIDEnvironmentParams{
		ProjectID: projectID,
	}, nil)
	if err != nil {
		return 0, fmt.Errorf("could not read project environments: %s", err.Error())
	}
	var matches []int64
	for _, environment := range response.Payload {
		if environment.Name == name {
			matches = append(matches, environment.ID)
		}
	}
	switch len(matches) {
	case 0:
		return 0, fmt.Errorf("project environment with name %q not found", name)
	case 1:
		return matches[0], nil
	default:
		return 0, fmt.Errorf("multiple project environments with name %q found (IDs %v); use id instead", name, matches)
	}
}

func (d *projectEnvironmentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ProjectEnvironmentModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	environmentID := config.ID.ValueInt64()
	if config.ID.IsNull() || config.ID.IsUnknown() {
		id, err := d.getEnvironmentIDByName(config.ProjectID.ValueInt64(), config.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading SemaphoreUI Project Environment",
				err.Error(),
			)
			return
		}
		environmentID = id
	}

	response, err := d.client.VariableGroup.GetProjectProjectIDEnvironmentEnvironmentID(&variable_group.GetProjectProjectIDEnvironmentEnvironmentIDParams{
		ProjectID:     config.ProjectID.ValueInt64(),
		EnvironmentID: environmentID,
	}, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading SemaphoreUI Project Environment",
			"Could not read project environment, unexpected error: "+err.Error(),
		)
		return
	}
	model := convertEnvironmentResponseToProjectEnvironmentDataSourceModel(ctx, response.Payload, &config)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
