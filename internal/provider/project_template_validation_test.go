package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestPlaybookRequiredValidator(t *testing.T) {
	ctx := context.Background()
	schema := ProjectTemplateSchema().GetResource(ctx)
	configType := schema.Type().TerraformType(ctx).(tftypes.Object)
	tests := []struct {
		name     string
		app      any
		playbook any
		wantErr  bool
	}{
		{name: "unknown playbook", app: "ansible", playbook: tftypes.UnknownValue},
		{name: "unknown app", app: tftypes.UnknownValue, playbook: nil},
		{name: "unknown app and empty playbook", app: tftypes.UnknownValue, playbook: ""},
		{name: "both unknown", app: tftypes.UnknownValue, playbook: tftypes.UnknownValue},
		{name: "default app and unknown playbook", app: nil, playbook: tftypes.UnknownValue},
		{name: "ansible with playbook", app: "ansible", playbook: "playbook.yml"},
		{name: "ansible without playbook", app: "ansible", playbook: nil, wantErr: true},
		{name: "ansible with empty playbook", app: "ansible", playbook: "", wantErr: true},
		{name: "default app without playbook", app: nil, playbook: nil, wantErr: true},
		{name: "bash without playbook", app: "bash", playbook: nil, wantErr: true},
		{name: "terraform without playbook", app: "terraform", playbook: nil},
		{name: "tofu without playbook", app: "tofu", playbook: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attributes := make(map[string]tftypes.Value, len(configType.AttributeTypes))
			for name, attributeType := range configType.AttributeTypes {
				attributes[name] = tftypes.NewValue(attributeType, nil)
			}
			attributes["app"] = tftypes.NewValue(tftypes.String, tt.app)
			attributes["playbook"] = tftypes.NewValue(tftypes.String, tt.playbook)
			request := resource.ValidateConfigRequest{Config: tfsdk.Config{
				Schema: schema,
				Raw:    tftypes.NewValue(configType, attributes),
			}}
			var response resource.ValidateConfigResponse
			playbookRequiredValidator{}.ValidateResource(ctx, request, &response)
			if response.Diagnostics.HasError() != tt.wantErr {
				t.Fatalf("expected error=%t, got diagnostics: %v", tt.wantErr, response.Diagnostics)
			}
			if tt.wantErr && response.Diagnostics.Errors()[0].Summary() != "Missing playbook" {
				t.Fatalf("unexpected validation error: %v", response.Diagnostics)
			}
		})
	}
}
