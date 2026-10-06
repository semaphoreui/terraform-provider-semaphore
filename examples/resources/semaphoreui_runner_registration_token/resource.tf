resource "semaphoreui_runner" "runner" {
  name = "Example Runner"

  # Generating a registration token leaves the runner unregistered and
  # inactive on the server until it registers with the new token. Keep
  # `active = false` here to avoid a permanent diff that would otherwise try
  # to re-activate the runner on every plan.
  active = false
}

# Generate a one-time registration token for the (unregistered) runner.
# Bump a value in `keepers` to rotate the token — it forces a new one to be
# generated, invalidating the previous one.
resource "semaphoreui_runner_registration_token" "token" {
  runner_id = semaphoreui_runner.runner.id

  keepers = {
    rotation = "1"
  }
}

# For a project runner, also set project_id:
#
# resource "semaphoreui_runner_registration_token" "token" {
#   project_id = semaphoreui_project.project.id
#   runner_id  = semaphoreui_project_runner.runner.id
# }
