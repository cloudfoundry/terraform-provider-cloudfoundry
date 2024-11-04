terraform {
  required_providers {
    cloudfoundry = {
      source  = "cloudfoundry/cloudfoundry"
      version = "1.0.0"
    }
  }
}
provider "cloudfoundry" {
  api_url  = "https://api.cf.example.com"
  user     = "admin"
  password = "admin"
}