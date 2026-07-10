# Switch to Service Instance Configuration by Service offering and Service Plan Name

> [!IMPORTANT]
> This guide is relevant from release 1.17.0 of the Terraform provider for Cloudfoundry.

## Overview

Until release 1.17.0 of the Terraform provider for cloudfoundry, the resource `cloudfoundry_service_instance` required the technical ID of the service plan (attribute `serviceplan`) to be specified in the resource configuration. This technical ID is not easily accessible for users and requires additional effort to retrieve it i.e., by calling the data source `cloudfoundry_service_plan`. To improve the user experience, the resource `cloudfoundry_service_instance` has been enhanced to allow users to specify the service instance configuration by using the service offering name and the service plan name instead of the technical ID. The provider will then automatically resolve the technical ID based on the provided names.

This guide should help you to switch your existing configuration of `cloudfoundry_service_instance` to the new configuration by using the service offering name and the service plan name.

## Prerequisite

Make sure that you have a backup of your Terraform state and your Terraform configuration before you start with the switch. This will allow you to easily revert back to the previous configuration in case of any issues.

## Procedure

For this sample we assume that you have an existing configuration of `cloudfoundry_service_instance` using a Terraform provider for SAP BTP release prior to 1.17.0 and you want to switch to the new configuration by using the service offering name and the service plan name.

Here is an example of such a setup:

```terraform
terraform {
  required_providers {
    cloudfoundry = {
      source  = "cloudfoundry/cloudfoundry"
      version = "1.16.0"
    }
  }
}
provider "cloudfoundry" {
}

data "cloudfoundry_org" "team_org" {
  name = "PerformanceTeamBLR"
}

data "cloudfoundry_space" "team_space" {
  name = "tf-space-1"
  org  = data.cloudfoundry_org.team_org.id
}

data "cloudfoundry_service_plan" "by_name" {
  name          = "lite"
  service_offering_name = "destination"
}

resource "cloudfoundry_service_instance" "destination" {
  service_plan = data.cloudfoundry_service_plan.by_name.id
  name           = "my-destination-instance"
  type = "managed"
  space = data.cloudfoundry_space.team_space.id
}
```

To switch to the new configuration, execute the following steps:

1. Update the provider version to 1.17.0 or later in your Terraform configuration.
2. Execute the command `terraform init -upgrade` to update the provider.
3. Execute the command `terraform plan --refresh-only` to plan a refresh of the state. This will show the values of the new attributes `service_offering_name` and `serviceplan_name` that are planned to be added to the state based on the existing configuration. You will see a output like this for the resource `cloudfoundry_service_instance`:

   ![screenshot of output of terraform plan --refresh-only](../assets/service-instance-plan-refresh.png)

4. Execute the command `terraform apply --refresh-only` to apply the refresh. This will update the state with the new attributes `service_offering_name` and `service_plan_name`. You will see a output like this for the resource `cloudfoundry_service_instance`:

   ![screenshot of output of terraform apply](../assets/service-instance-apply-refresh.png)

5. Update the resource configuration of `cloudfoundry_service_instance` to use the new attributes `service_offering_name` and `service_plan_name` instead of `serviceplan` and remove the data source `cloudfoundry_service_plan`. The updated configuration should look like this:

  ```terraform
  terraform {
    required_providers {
      cloudfoundry = {
      source  = "cloudfoundry/cloudfoundry"
      version = "1.16.0"
      }
    }
  }
  
  provider "cloudfoundry" {
  }

  data "cloudfoundry_org" "team_org" {
    name = "PerformanceTeamBLR"
  }

  data "cloudfoundry_space" "team_space" {
    name = "tf-space-1"
    org  = data.cloudfoundry_org.team_org.id
  }

  resource "cloudfoundry_service_instance" "destination" {
    name           = "my-destination-instance"
    type = "managed"
    space = data.cloudfoundry_space.team_space.id
    service_offering_name = "destination"
    service_plan_name = "lite"
  }

  ```
  
6. Execute the command `terraform plan` to review the changes. You should see that Terraform does not plan any changes to be applied.
