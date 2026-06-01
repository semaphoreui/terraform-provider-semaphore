# Import ID is specified by the string "project/{project_id}/runner/{runner_id}".
# - {project_id} is the ID of the project in SemaphoreUI.
# - {runner_id} is the ID of the runner in SemaphoreUI.
terraform import semaphoreui_project_runner.example project/1/runner/2
```
Or using `import {}` block in the configuration file:
```hcl
import {
  to = semaphoreui_project_runner.example
  id = "project/1/runner/2"
}
