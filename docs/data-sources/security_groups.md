---
page_title: "cloudfoundry_security_groups Data Source - terraform-provider-cloudfoundry"
subcategory: ""
description: |-
  Gets information on Cloud Foundry application security groups.
---

# cloudfoundry_security_groups (Data Source)

Gets information on Cloud Foundry application security groups.

## Example Usage

```terraform
data "cloudfoundry_security_groups" "sgroups" {
  name = "riemann"
}

output "sgroup" {
  value = data.cloudfoundry_security_groups.sgroups
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `name` (String) Name of the security group to filter by
- `running_space` (String) The GUID of the running space to filter by
- `staging_space` (String) The GUID of the staging space to filter by

### Read-Only

- `security_groups` (Attributes List) List of security groups (see [below for nested schema](#nestedatt--security_groups))

<a id="nestedatt--security_groups"></a>
### Nested Schema for `security_groups`

Read-Only:

- `created_at` (String) The date and time when the resource was created in [RFC3339](https://www.ietf.org/rfc/rfc3339.txt) format.
- `globally_enabled_running` (Boolean) Specifies whether the group should be applied globally to all running applications
- `globally_enabled_staging` (Boolean) Specifies whether the group should be applied globally to all staging applications
- `id` (String) The GUID of the object.
- `name` (String) Name of the security group
- `rules` (Attributes List) Rules that will be applied by this security group (see [below for nested schema](#nestedatt--security_groups--rules))
- `running_spaces` (Set of String) The spaces where the security_group is applied to applications during runtime
- `staging_spaces` (Set of String) The spaces where the security_group is applied to applications during stagingo
- `updated_at` (String) The date and time when the resource was updated in [RFC3339](https://www.ietf.org/rfc/rfc3339.txt) format.

<a id="nestedatt--security_groups--rules"></a>
### Nested Schema for `security_groups.rules`

Read-Only:

- `code` (Number) Code field of the ICMP type
- `description` (String) A description for the rule
- `destination` (String) Destinations that the rule applies to
- `log` (Boolean) Whether logging for the rule is enabled
- `ports` (String) Ports that the rule applies to
- `protocol` (String) Protocol type
- `type` (Number) Type number for ICMP protocol