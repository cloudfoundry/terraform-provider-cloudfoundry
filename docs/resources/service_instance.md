---
page_title: "cloudfoundry_service_instance Resource - terraform-provider-cloudfoundry"
subcategory: ""
description: |-
  Creates a service instance in a cloudfoundry space.
  Further documentation:
  https://docs.cloudfoundry.org/devguide/services
---

# cloudfoundry_service_instance (Resource)

Creates a service instance in a cloudfoundry space.

__Further documentation:__
https://docs.cloudfoundry.org/devguide/services

## Example Usage

```terraform
data "cloudfoundry_org" "team_org" {
  name = "PerformanceTeamBLR"
}

data "cloudfoundry_space" "team_space" {
  name = "PerformanceTeamBLR"
  org  = data.cloudfoundry_org.team_org.id
}

data "cloudfoundry_service" "xsuaa_svc" {
  name = "xsuaa"
}
data "cloudfoundry_service" "autoscaler_svc" {
  name = "autoscaler"
}
resource "cloudfoundry_service_instance" "xsuaa_svc" {
  name         = "xsuaa_svc"
  type         = "managed"
  tags         = ["terraform-test", "test1"]
  space        = data.cloudfoundry_space.team_space.id
  service_plan = data.cloudfoundry_service.xsuaa_svc.service_plans["application"]
  parameters   = <<EOT
  {
  "xsappname": "tf-test23",
  "tenant-mode": "dedicated",
  "description": "tf test123",
  "foreign-scope-references": ["user_attributes"],
  "scopes": [
    {
      "name": "uaa.user",
      "description": "UAA"
    }
  ],
  "role-templates": [
    {
      "name": "Token_Exchange",
      "description": "UAA",
      "scope-references": ["uaa.user"]
    }
  ]
}
EOT
}

# Managed service instance without parameters
resource "cloudfoundry_service_instance" "dev-autoscaler" {
  name         = "tf-autoscaler-test"
  type         = "managed"
  tags         = ["terraform-test", "autoscaler"]
  space        = data.cloudfoundry_space.team_space.id
  service_plan = data.cloudfoundry_service.autoscaler_svc.service_plans["standard"]
  timeouts = {
    create = "10m"
  }
}
# User provided service instance
resource "cloudfoundry_service_instance" "dev-usp" {
  name        = "tf-usp-test"
  type        = "user-provided"
  tags        = ["terraform-test", "usp"]
  space       = data.cloudfoundry_space.team_space.id
  credentials = <<EOT
  {
    "user": "user1",
    "password": "demo122"
  }
  EOT
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The name of the service instance
- `space` (String) The ID of the space in which to create the service instance
- `type` (String) Type of the service instance. Either managed or user-provided.

### Optional

- `annotations` (Map of String) The annotations associated with Cloud Foundry resources. Add as described [here](https://docs.cloudfoundry.org/adminguide/metadata.html#-view-metadata-for-an-object).
- `credentials` (String, Sensitive) A JSON object that is made available to apps bound to this service instance of type user-provided.
- `labels` (Map of String) The labels associated with Cloud Foundry resources. Add as described [here](https://docs.cloudfoundry.org/adminguide/metadata.html#-view-metadata-for-an-object).
- `parameters` (String) A JSON object that is passed to the service broker for managed service instance.
- `route_service_url` (String) URL to which requests for bound routes will be forwarded; only shown when type is user-provided.
- `service_plan` (String) The ID of the service plan from which to create the service instance
- `syslog_drain_url` (String) URL to which logs for bound applications will be streamed; only shown when type is user-provided.
- `tags` (Set of String) Set of tags used by apps to identify service instances. They are shown in the app VCAP_SERVICES env.
- `timeouts` (Attributes) (see [below for nested schema](#nestedatt--timeouts))
- `upgrade_available` (Boolean) Whether or not an upgrade of this service instance is available on the current Service Plan; details are available in the maintenance_info object; Only shown when type is managed

### Read-Only

- `created_at` (String) The date and time when the resource was created in [RFC3339](https://www.ietf.org/rfc/rfc3339.txt) format.
- `dashboard_url` (String) The URL to the service instance dashboard (or null if there is none); only shown when type is managed.
- `id` (String) The GUID of the object.
- `last_operation` (Attributes) The last operation of this service instance. (see [below for nested schema](#nestedatt--last_operation))
- `maintenance_info` (Attributes) Information about the version of this service instance; only shown when type is managed (see [below for nested schema](#nestedatt--maintenance_info))
- `updated_at` (String) The date and time when the resource was updated in [RFC3339](https://www.ietf.org/rfc/rfc3339.txt) format.

<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String) Timeout for creating the service instance. Default is 40 minutes
- `delete` (String) Timeout for deleting the service instance. Default is 40 minutes
- `update` (String) Timeout for updating the service instance. Default is 40 minutes


<a id="nestedatt--last_operation"></a>
### Nested Schema for `last_operation`

Optional:

- `created_at` (String) The time at which the last operation was created

Read-Only:

- `description` (String) The description of the last operation
- `state` (String) The state of the last operation
- `type` (String) The type of the last operation
- `updated_at` (String) The time of the last operation


<a id="nestedatt--maintenance_info"></a>
### Nested Schema for `maintenance_info`

Optional:

- `description` (String) A description of the version of the service instance
- `version` (String) The version of the service instance

## Import

Import is supported using the following syntax:

```terraform
terraform import cloudfoundry_service_instance.xsuaa_svc 68fea1b6-11b9-4737-ad79-74e49832533f
```