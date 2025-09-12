# Configure the SemaphoreUI provider using required_providers.
terraform {
  required_providers {
    semaphoreui = {
      source  = "CruGlobal/semaphoreui"
      version = "~> 1.0"
    }
  }
}

provider "semaphoreui" {
  api_base_url = "http://localhost:3000/api" # Default: "http://localhost:3000/api"
  api_token = "your token"
}
