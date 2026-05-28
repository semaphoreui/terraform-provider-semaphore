resource "semaphoreui_project" "project" {
  name = "Example Project"
}

resource "semaphoreui_project_key" "login_password" {
  project_id = semaphoreui_project.project.id
  name       = "Example Login"
  login_password = {
    login    = "username"
    password = "password"
  }
}

resource "semaphoreui_project_key" "ssh" {
  project_id = semaphoreui_project.project.id
  name       = "Example SSH"
  ssh = {
    passphrase  = "password"
    private_key = file("./id_rsa")
  }
}

# Write-only secrets — the value is sent to SemaphoreUI on create/update but is
# never stored in Terraform state. Bump `*_wo_version` to push a new value.
resource "semaphoreui_project_key" "login_password_write_only" {
  project_id = semaphoreui_project.project.id
  name       = "Example Login (write-only)"
  login_password = {
    login               = "username"
    password_wo         = var.password # e.g. data.infisical_secret.api.value
    password_wo_version = 1
  }
}

resource "semaphoreui_project_key" "ssh_write_only" {
  project_id = semaphoreui_project.project.id
  name       = "Example SSH (write-only)"
  ssh = {
    passphrase_wo          = var.ssh_passphrase
    passphrase_wo_version  = 1
    private_key_wo         = var.ssh_private_key
    private_key_wo_version = 1
  }
}

resource "semaphoreui_project_key" "none" {
  project_id = semaphoreui_project.project.id
  name       = "Example None"
  none       = {}
}
