package provider

import (
	"context"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-semaphoreui/semaphoreui/models"
)

func TestConvertProjectTemplateModelToTemplateRequest_environment(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name       string
		envID      types.Int64
		envIDs     types.Set
		wantEnvID  int64
		wantEnvIDs []int64
	}{
		{
			name:       "environment_id is sent in both fields",
			envID:      types.Int64Value(7),
			envIDs:     types.SetNull(types.Int64Type),
			wantEnvID:  7,
			wantEnvIDs: []int64{7},
		},
		{
			name:       "environment_ids is sent as the array only",
			envID:      types.Int64Null(),
			envIDs:     types.SetValueMust(types.Int64Type, []attr.Value{types.Int64Value(3), types.Int64Value(5)}),
			wantEnvIDs: []int64{3, 5},
		},
		{
			name:       "single-element environment_ids",
			envID:      types.Int64Null(),
			envIDs:     types.SetValueMust(types.Int64Type, []attr.Value{types.Int64Value(3)}),
			wantEnvIDs: []int64{3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request, diags := convertProjectTemplateModelToTemplateRequest(ctx, ProjectTemplateModel{
				EnvironmentID:  tt.envID,
				EnvironmentIDs: tt.envIDs,
				Arguments:      types.ListNull(types.StringType),
				SurveyVars:     types.ListNull(ProjectTemplateSurveyVarType),
				Vaults:         types.ListNull(ProjectTemplateVaultType),
			})
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}
			if request.EnvironmentID != tt.wantEnvID {
				t.Errorf("EnvironmentID = %d, want %d", request.EnvironmentID, tt.wantEnvID)
			}
			got := slices.Sorted(slices.Values(request.EnvironmentIds))
			if !slices.Equal(got, tt.wantEnvIDs) {
				t.Errorf("EnvironmentIds = %v, want %v", got, tt.wantEnvIDs)
			}
		})
	}
}

func TestConvertTemplateResponseToProjectTemplateModel_environment(t *testing.T) {
	ctx := context.Background()
	idSet := func(ids ...int64) types.Set {
		values := make([]attr.Value, len(ids))
		for i, id := range ids {
			values[i] = types.Int64Value(id)
		}
		return types.SetValueMust(types.Int64Type, values)
	}
	tests := []struct {
		name       string
		response   models.Template
		prevEnvIDs types.Set
		wantEnvID  types.Int64
		wantEnvIDs types.Set
	}{
		{
			name:       "single ID with environment_id config",
			response:   models.Template{EnvironmentIds: []int64{7}},
			prevEnvIDs: types.SetNull(types.Int64Type),
			wantEnvID:  types.Int64Value(7),
			wantEnvIDs: types.SetNull(types.Int64Type),
		},
		{
			name:       "single ID with environment_ids config",
			response:   models.Template{EnvironmentIds: []int64{7}},
			prevEnvIDs: idSet(7),
			wantEnvID:  types.Int64Null(),
			wantEnvIDs: idSet(7),
		},
		{
			name:       "single ID with unknown environment_ids plan",
			response:   models.Template{EnvironmentIds: []int64{7}},
			prevEnvIDs: types.SetUnknown(types.Int64Type),
			wantEnvID:  types.Int64Null(),
			wantEnvIDs: idSet(7),
		},
		{
			name:       "multiple IDs without prior state (import)",
			response:   models.Template{EnvironmentIds: []int64{5, 3}},
			prevEnvIDs: types.SetNull(types.Int64Type),
			wantEnvID:  types.Int64Null(),
			wantEnvIDs: idSet(3, 5),
		},
		{
			name:       "legacy environment_id fallback",
			response:   models.Template{EnvironmentID: 9},
			prevEnvIDs: types.SetNull(types.Int64Type),
			wantEnvID:  types.Int64Value(9),
			wantEnvIDs: types.SetNull(types.Int64Type),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, diags := convertTemplateResponseToProjectTemplateModel(ctx, &tt.response, &ProjectTemplateModel{
				EnvironmentIDs: tt.prevEnvIDs,
				SurveyVars:     types.ListNull(ProjectTemplateSurveyVarType),
				Vaults:         types.ListNull(ProjectTemplateVaultType),
			})
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}
			if !model.EnvironmentID.Equal(tt.wantEnvID) {
				t.Errorf("EnvironmentID = %v, want %v", model.EnvironmentID, tt.wantEnvID)
			}
			if !model.EnvironmentIDs.Equal(tt.wantEnvIDs) {
				t.Errorf("EnvironmentIDs = %v, want %v", model.EnvironmentIDs, tt.wantEnvIDs)
			}
		})
	}
}
