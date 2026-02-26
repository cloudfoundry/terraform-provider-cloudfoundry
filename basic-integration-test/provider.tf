terraform {
  required_providers {
    cloudfoundry = {
      source = "cloudfoundry/cloudfoundry"
    }
    zipper = {
      source = "ArthurHlt/zipper"
    }
  }
}

provider "cloudfoundry" {
#  api_url             = var.cf_api_url
#  user                = var.cf_user
#  password            = var.cf_password
#  skip_ssl_validation = var.cf_skip_ssl_validation
}

provider "zipper" {}
