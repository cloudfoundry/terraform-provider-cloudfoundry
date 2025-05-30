---
page_title: "cloudfoundry_service_plans Data Source - terraform-provider-cloudfoundry"
subcategory: ""
description: |-
  Fetches Service Plans based on the filters provided
---

# cloudfoundry_service_plans (Data Source)

Fetches Service Plans based on the filters provided

## Example Usage

```terraform
data "cloudfoundry_service_plans" "xsuaa-offering" {
  service_offering_name = "xsuaa"
}

output "serviceplans" {
  value = data.cloudfoundry_service_plans.xsuaa-offering.service_plans
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `name` (String) The name of the service plan to look up
- `service_broker_name` (String) The name of the service broker which offers the service. Use this to filter two equally named services from different brokers.
- `service_offering_name` (String) The name of the service offering for whose plans to look up

### Read-Only

- `service_plans` (Attributes List) The list of the service plans (see [below for nested schema](#nestedatt--service_plans))

<a id="nestedatt--service_plans"></a>
### Nested Schema for `service_plans`

Read-Only:

- `annotations` (Map of String) The annotations associated with Cloud Foundry resources.
- `available` (Boolean) Whether or not the service plan is available
- `broker_catalog` (Attributes) This object contains information obtained from the service broker catalog (see [below for nested schema](#nestedatt--service_plans--broker_catalog))
- `costs` (Attributes List) The cost of the service plan as obtained from the service broker catalog (see [below for nested schema](#nestedatt--service_plans--costs))
- `created_at` (String) The date and time when the resource was created in [RFC3339](https://www.ietf.org/rfc/rfc3339.txt) format.
- `description` (String) Description of the service plan
- `free` (Boolean) Whether or not the service plan is free of charge
- `id` (String) The GUID of the object.
- `labels` (Map of String) The labels associated with Cloud Foundry resources.
- `maintenance_info` (Attributes) Information about the version of this service plan (see [below for nested schema](#nestedatt--service_plans--maintenance_info))
- `name` (String) Name of the service plan
- `schemas` (Attributes) Schema definitions for service instances and service bindings for the service plan (see [below for nested schema](#nestedatt--service_plans--schemas))
- `service_offering` (String) The service offering that this service plan relates to
- `updated_at` (String) The date and time when the resource was updated in [RFC3339](https://www.ietf.org/rfc/rfc3339.txt) format.
- `visibility_type` (String) Denotes the visibility of the plan

<a id="nestedatt--service_plans--broker_catalog"></a>
### Nested Schema for `service_plans.broker_catalog`

Read-Only:

- `bindable` (Boolean) Specifies whether service instances of the service can be bound to applications
- `id` (String) The identifier that the service broker provided for this service plan
- `maximum_polling_duration` (Number) The maximum number of seconds that Cloud Foundry will wait for an asynchronous service broker operation
- `metadata` (String) Additional information provided by the service broker as specified by OSBAPI
- `plan_updateable` (Boolean) Whether the service plan supports upgrade/downgrade for service plans


<a id="nestedatt--service_plans--costs"></a>
### Nested Schema for `service_plans.costs`

Read-Only:

- `amount` (Number) Pricing amount
- `currency` (String) Currency code for the pricing amount, e.g. USD, GBP
- `unit` (String) Display name for type of cost, e.g. Monthly, Hourly, Request, GB


<a id="nestedatt--service_plans--maintenance_info"></a>
### Nested Schema for `service_plans.maintenance_info`

Read-Only:

- `description` (String) A textual explanation associated with this version
- `version` (String) The current semantic version of the service plan


<a id="nestedatt--service_plans--schemas"></a>
### Nested Schema for `service_plans.schemas`

Read-Only:

- `service_binding` (Attributes) (see [below for nested schema](#nestedatt--service_plans--schemas--service_binding))
- `service_instance` (Attributes) (see [below for nested schema](#nestedatt--service_plans--schemas--service_instance))

<a id="nestedatt--service_plans--schemas--service_binding"></a>
### Nested Schema for `service_plans.schemas.service_binding`

Read-Only:

- `create_parameters` (String) Schema definition for the input parameters for service Binding creation


<a id="nestedatt--service_plans--schemas--service_instance"></a>
### Nested Schema for `service_plans.schemas.service_instance`

Read-Only:

- `create_parameters` (String) Schema definition for the input parameters for service instance creation
- `update_parameters` (String) Schema definition for the input parameters for service instance update