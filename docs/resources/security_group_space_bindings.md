---
page_title: "cloudfoundry_security_group_space_bindings Resource - terraform-provider-cloudfoundry"
subcategory: ""
description: |-
  Provides a Cloud Foundry resource for binding and unbinding a security group from spaces. Only handles bindings managed through this resource and does not touch the existing space bindings with the security group. On deleting the resource, the security group will be unbound from the mentioned spaces.
---

# cloudfoundry_security_group_space_bindings (Resource)

Provides a Cloud Foundry resource for binding and unbinding a security group from spaces. Only handles bindings managed through this resource and does not touch the existing space bindings with the security group. On deleting the resource, the security group will be unbound from the mentioned spaces.

## Example Usage

```terraform
resource "cloudfoundry_security_group_space_bindings" "my_security_group_spaces" {
  security_group = "56eedab7-cb97-469b-a3e9-89521827c039"
  staging_spaces = ["02c0cc92-6ecc-44b1-b7b2-096ca19ee143"]
  running_spaces = ["121c3a95-0f82-45a6-8ff2-1920b2067edb", "02c0cc92-6ecc-44b1-b7b2-096ca19ee143"]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `security_group` (String) GUID of the isolation segment

### Optional

- `running_spaces` (Set of String) The spaces where the security_group is applied to applications during runtime
- `staging_spaces` (Set of String) The spaces where the security_group is applied to applications during staging
