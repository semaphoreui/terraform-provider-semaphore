package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	schemaD "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	schemaR "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	superschema "github.com/orange-cloudavenue/terraform-plugin-framework-superschema"
)

type (
	ProjectRunnerModel struct {
		ID               types.Int64  `tfsdk:"id"`
		ProjectID        types.Int64  `tfsdk:"project_id"`
		Name             types.String `tfsdk:"name"`
		Webhook          types.String `tfsdk:"webhook"`
		MaxParallelTasks types.Int64  `tfsdk:"max_parallel_tasks"`
		Active           types.Bool   `tfsdk:"active"`
		Tags             types.Set    `tfsdk:"tags"`
		IsDefault        types.Bool   `tfsdk:"is_default"`
	}
)

func ProjectRunnerSchema() superschema.Schema {
	return superschema.Schema{
		Common: superschema.SchemaDetails{
			MarkdownDescription: "The project runner",
		},
		Resource: superschema.SchemaDetails{
			MarkdownDescription: "resource allows you to define a runner owned by a project. Runners execute the tasks scheduled by templates. Use the `semaphoreui_runner_registration_token` resource to generate the one-time token the runner uses to register.",
		},
		DataSource: superschema.SchemaDetails{
			MarkdownDescription: "data source allows you to read a runner owned by a project.",
		},
		Attributes: map[string]superschema.Attribute{
			"id": superschema.Int64Attribute{
				Common: &schemaR.Int64Attribute{
					MarkdownDescription: "The runner ID.",
				},
				Resource: &schemaR.Int64Attribute{
					Computed:      true,
					PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
				},
				DataSource: &schemaD.Int64Attribute{
					Optional: true,
					Computed: true,
					Validators: []validator.Int64{
						int64validator.ExactlyOneOf(
							path.MatchRoot("id"),
							path.MatchRoot("name"),
						),
					},
				},
			},
			"project_id": superschema.Int64Attribute{
				Common: &schemaR.Int64Attribute{
					MarkdownDescription: "The project ID that the runner belongs to.",
					Required:            true,
				},
				Resource: &schemaR.Int64Attribute{
					PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
				},
			},
			"name": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					MarkdownDescription: "The display name of the runner.",
				},
				Resource: &schemaR.StringAttribute{
					Optional: true,
					Computed: true,
				},
				DataSource: &schemaD.StringAttribute{
					Optional: true,
					Computed: true,
					Validators: []validator.String{
						stringvalidator.ExactlyOneOf(
							path.MatchRoot("id"),
							path.MatchRoot("name"),
						),
					},
				},
			},
			"webhook": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					MarkdownDescription: "URL called by the runner to report task events.",
				},
				Resource: &schemaR.StringAttribute{
					Optional: true,
					Computed: true,
				},
				DataSource: &schemaD.StringAttribute{
					Computed: true,
				},
			},
			"max_parallel_tasks": superschema.Int64Attribute{
				Common: &schemaR.Int64Attribute{
					MarkdownDescription: "The maximum number of tasks the runner may execute in parallel.",
				},
				Resource: &schemaR.Int64Attribute{
					Optional: true,
					Computed: true,
				},
				DataSource: &schemaD.Int64Attribute{
					Computed: true,
				},
			},
			"active": superschema.BoolAttribute{
				Common: &schemaR.BoolAttribute{
					MarkdownDescription: "Indicates whether the runner is allowed to pick up tasks.",
				},
				Resource: &schemaR.BoolAttribute{
					Optional: true,
					Computed: true,
					Default:  booldefault.StaticBool(true),
				},
				DataSource: &schemaD.BoolAttribute{
					Computed: true,
				},
			},
			"tags": superschema.SetAttribute{
				Common: &schemaR.SetAttribute{
					MarkdownDescription: "Tags used to route tasks to specific runners.",
					ElementType:         types.StringType,
				},
				Resource: &schemaR.SetAttribute{
					Optional: true,
					Computed: true,
				},
				DataSource: &schemaD.SetAttribute{
					Computed: true,
				},
			},
			"is_default": superschema.BoolAttribute{
				Common: &schemaR.BoolAttribute{
					MarkdownDescription: "Indicates whether this is the default runner.",
				},
				Resource: &schemaR.BoolAttribute{
					Optional: true,
					Computed: true,
				},
				DataSource: &schemaD.BoolAttribute{
					Computed: true,
				},
			},
		},
	}
}
