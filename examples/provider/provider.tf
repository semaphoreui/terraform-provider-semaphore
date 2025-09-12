# Configure the SemaphoreUI provider using required_providers.
terraform {
  required_providers {
    semaphore = {
      source  = "semaphoreui/semaphore"
      version = "~> 0.1"
    }
  }
}

provider "semaphore" {
  api_base_url = "http://localhost:3000/api" # Default: "http://localhost:3000/api"
  api_token = "your token"
}
