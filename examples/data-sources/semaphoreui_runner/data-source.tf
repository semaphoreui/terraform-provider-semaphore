# Look up a global runner by ID.
data "semaphoreui_runner" "by_id" {
  id = 1
}

# Look up a global runner by name.
data "semaphoreui_runner" "by_name" {
  name = "Example Global Runner"
}
