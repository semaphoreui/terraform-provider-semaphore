package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// mergeWriteOnlyFromConfig reads write-only project key values from config.
func mergeWriteOnlyFromConfig(ctx context.Context, cfg tfsdk.Config, plan *ProjectKeyModel, diags *diag.Diagnostics) {
	if plan.LoginPassword != nil {
		var pw types.String
		diags.Append(cfg.GetAttribute(ctx, path.Root("login_password").AtName("password_wo"), &pw)...)
		plan.LoginPassword.PasswordWo = pw
	}
	if plan.SSH != nil {
		var passphrase, privateKey types.String
		diags.Append(cfg.GetAttribute(ctx, path.Root("ssh").AtName("passphrase_wo"), &passphrase)...)
		diags.Append(cfg.GetAttribute(ctx, path.Root("ssh").AtName("private_key_wo"), &privateKey)...)
		plan.SSH.PassphraseWo = passphrase
		plan.SSH.PrivateKeyWo = privateKey
	}
}

func extractEnvironmentSecrets(ctx context.Context, list types.List, diags *diag.Diagnostics) []ProjectEnvironmentSecretModel {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	var secrets []ProjectEnvironmentSecretModel
	diags.Append(list.ElementsAs(ctx, &secrets, false)...)
	if diags.HasError() {
		return nil
	}

	return secrets
}

// mergeEnvironmentSecretsWriteOnly reads value_wo from config.
func mergeEnvironmentSecretsWriteOnly(ctx context.Context, cfg tfsdk.Config, planSecrets []ProjectEnvironmentSecretModel, diags *diag.Diagnostics) []ProjectEnvironmentSecretModel {
	if len(planSecrets) == 0 {
		return planSecrets
	}
	var cfgSecrets []ProjectEnvironmentSecretModel
	d := cfg.GetAttribute(ctx, path.Root("secrets"), &cfgSecrets)
	diags.Append(d...)
	if d.HasError() {
		return planSecrets
	}
	if len(cfgSecrets) != len(planSecrets) {
		return planSecrets
	}
	for i := range planSecrets {
		planSecrets[i].ValueWo = cfgSecrets[i].ValueWo
	}
	return planSecrets
}

func missingStringValue(value, writeOnly types.String) bool {
	valueSet := !value.IsNull() && !value.IsUnknown()
	writeOnlySet := !writeOnly.IsNull() && !writeOnly.IsUnknown()
	return !valueSet && !writeOnlySet
}

func hasEnvironmentSecretWithoutValue(secrets []ProjectEnvironmentSecretModel) bool {
	for _, secret := range secrets {
		if missingStringValue(secret.Value, secret.ValueWo) {
			return true
		}
	}
	return false
}
