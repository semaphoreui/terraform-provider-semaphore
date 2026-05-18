resource "semaphoreui_project" "project" {
  name = "Example Project"
}

resource "semaphoreui_project_environment" "environment" {
  project_id = semaphoreui_project.project.id
  name       = "Example Environment"

  # extraVars
  variables = {
    key1 = "value1"
    key2 = "value2"
  }

  # environment variables
  environment = {
    KEY1 = "value1"
    KEY2 = "value2"
  }

  # secrets
  secrets = [{
    # extraVar Secret
    name  = "key3"
    type  = "var"
    value = "value3"
    }, {
    # environment Secret
    name  = "KEY4"
    type  = "env"
    value = "value4"
  }]
}

# Write-only secret values. The value_wo cleartext is never written to Terraform state
# or plan output; only value_wo_version is stored. Bumping the version pushes a new
# value to Semaphore on the next apply. Use this when sourcing secrets ephemerally from
# Vault, Infisical, or similar secret managers.
variable "api_token" {
  type      = string
  sensitive = true
}

resource "semaphoreui_project_environment" "environment_write_only" {
  project_id = semaphoreui_project.project.id
  name       = "Write-Only Environment"

  secrets = [{
    name             = "API_TOKEN"
    type             = "env"
    value_wo         = var.api_token
    value_wo_version = 1
  }]
}
