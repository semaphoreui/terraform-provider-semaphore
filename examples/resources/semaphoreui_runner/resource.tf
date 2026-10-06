resource "semaphoreui_runner" "runner" {
  name               = "Example Global Runner"
  max_parallel_tasks = 1
  active             = true
  tags               = ["linux", "production"]
}
