---
page_title: "cloudfoundry_space_role Data Source - terraform-provider-cloudfoundry"
subcategory: ""
description: |-
  Gets information on a Cloud Foundry space role with a given role ID.
---

# cloudfoundry_space_role (Data Source)

Gets information on a Cloud Foundry space role with a given role ID.

## Example Usage

```terraform
data "cloudfoundry_space_role" "my_role" {
  id = "4c6849f2-6407-4385-a556-0840369f336b"
}

output "role_object" {
  value = data.cloudfoundry_space_role.my_role
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `id` (String) The guid for the role

### Read-Only

- `created_at` (String) The date and time when the resource was created in [RFC3339](https://www.ietf.org/rfc/rfc3339.txt) format.
- `space` (String) The guid of the space the role is assigned to
- `type` (String) Role type; see [Valid role types](https://v3-apidocs.cloudfoundry.org/version/3.154.0/index.html#valid-role-types)
- `updated_at` (String) The date and time when the resource was updated in [RFC3339](https://www.ietf.org/rfc/rfc3339.txt) format.
- `user` (String) The guid of the cloudfoundry user the role is assigned to