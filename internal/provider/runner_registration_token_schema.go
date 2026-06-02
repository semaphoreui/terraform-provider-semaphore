package provider

import (
	schemaR "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	superschema "github.com/orange-cloudavenue/terraform-plugin-framework-superschema"
)

type RunnerRegistrationTokenModel struct {
	ID                types.String `tfsdk:"id"`
	RunnerID          types.Int64  `tfsdk:"runner_id"`
	ProjectID         types.Int64  `tfsdk:"project_id"`
	Keepers           types.Map    `tfsdk:"keepers"`
	RegistrationToken types.String `tfsdk:"registration_token"`
}

func RunnerRegistrationTokenSchema() superschema.Schema {
	return superschema.Schema{
		Common: superschema.SchemaDetails{
			MarkdownDescription: "A one-time, short-lived registration token for an unregistered runner.",
		},
		Resource: superschema.SchemaDetails{
			MarkdownDescription: "resource generates a fresh one-time registration token for an existing, unregistered runner. " +
				"Regenerating invalidates the previous token. The token is returned only once, at creation, and stored " +
				"(sensitive) in Terraform state. The resource is immutable: changing `runner_id`, `project_id` or `keepers` " +
				"forces a new token to be generated. Use `keepers` to rotate the token on demand (e.g. bump a value to issue a new one). " +
				"The runner must not already be registered, otherwise the API returns an error. " +
				"Note: generating a token leaves the runner inactive until it registers, so a runner managed alongside this " +
				"resource should set `active = false` to avoid a permanent diff.",
		},
		Attributes: map[string]superschema.Attribute{
			"id": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					MarkdownDescription: "Synthetic identifier of the form `runner/{runner_id}` (or `project/{project_id}/runner/{runner_id}` for project runners).",
				},
				Resource: &schemaR.StringAttribute{
					Computed:      true,
					PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				},
			},
			"runner_id": superschema.Int64Attribute{
				Common: &schemaR.Int64Attribute{
					MarkdownDescription: "The ID of the runner to generate a registration token for.",
					Required:            true,
				},
				Resource: &schemaR.Int64Attribute{
					PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
				},
			},
			"project_id": superschema.Int64Attribute{
				Common: &schemaR.Int64Attribute{
					MarkdownDescription: "The project ID that owns the runner. Set this for project runners; omit it for global (admin) runners.",
				},
				Resource: &schemaR.Int64Attribute{
					Optional:      true,
					PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
				},
			},
			"keepers": superschema.MapAttribute{
				Common: &schemaR.MapAttribute{
					MarkdownDescription: "Arbitrary map of values that, when changed, forces a new registration token to be generated. " +
						"Use it to trigger rotation (the SemaphoreUI API exposes no other way to rotate a token in place).",
					ElementType: types.StringType,
				},
				Resource: &schemaR.MapAttribute{
					Optional:      true,
					PlanModifiers: []planmodifier.Map{mapplanmodifier.RequiresReplace()},
				},
			},
			"registration_token": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					MarkdownDescription: "The generated one-time registration token. Returned only at creation and persisted (sensitive) to Terraform state.",
					Sensitive:           true,
				},
				Resource: &schemaR.StringAttribute{
					Computed:      true,
					PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				},
			},
		},
	}
}
