# Lookup by environment ID
data "semaphoreui_project_environment" "environment" {
  project_id = 1
  id         = 4
}

# Lookup by environment name
data "semaphoreui_project_environment" "by_name" {
  project_id = 1
  name       = "Production"
}
