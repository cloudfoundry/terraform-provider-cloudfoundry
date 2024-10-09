terraform {
  required_providers {
    cloudfoundry = {
      source  = "cloudfoundry/cloudfoundry"
      version = "1.0.0-rc1"
    }
  }
}
provider "cloudfoundry" {}