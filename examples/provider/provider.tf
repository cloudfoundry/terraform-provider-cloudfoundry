terraform {
  required_providers {
    cloudfoundry = {
      source  = "cloudfoundry/cloudfoundry"
      version = "1.2.0"
    }
  }
}
provider "cloudfoundry" {
  api_url  = "https://api.cf.example.com"
  user     = "admin"
  password = "admin"
}