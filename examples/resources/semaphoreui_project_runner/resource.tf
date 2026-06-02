resource "semaphoreui_project" "project" {
  name = "Example Project"
}

resource "semaphoreui_project_runner" "runner" {
  project_id         = semaphoreui_project.project.id
  name               = "Example Runner"
  max_parallel_tasks = 1
  active             = true
  tags               = ["linux", "production"]
}
