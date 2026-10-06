# Import ID is specified by the string "runner/{runner_id}".
# - {runner_id} is the ID of the global runner in SemaphoreUI.
terraform import semaphoreui_runner.example runner/1
```
Or using `import {}` block in the configuration file:
```hcl
import {
  to = semaphoreui_runner.example
  id = "runner/1"
}
