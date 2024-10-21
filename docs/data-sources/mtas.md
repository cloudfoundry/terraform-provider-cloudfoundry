---
page_title: "cloudfoundry_mtas Data Source - terraform-provider-cloudfoundry"
subcategory: ""
description: |-
  Gets information on Multi Target Applications present in a space.
---

# cloudfoundry_mtas (Data Source)

Gets information on Multi Target Applications present in a space.

## Example Usage

```terraform
data "cloudfoundry_mtas" "mtas" {
  space     = "02c0cc92-6ecc-44b1-b7b2-096ca19ee143"
  namespace = "test"
}

output "data" {
  value = data.cloudfoundry_mtas.mtas
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `space` (String) The GUID of the space where the MTA's are deployed

### Optional

- `deploy_url` (String) The URL of the deploy service, if a custom one has been used(should be present in the same landscape). By default 'deploy-service.<system-domain>'
- `id` (String) The MTA ID to filter by
- `namespace` (String) The namespace of the MTA to filter by

### Read-Only

- `mtas` (Attributes List) The list of MTA's (see [below for nested schema](#nestedatt--mtas))

<a id="nestedatt--mtas"></a>
### Nested Schema for `mtas`

Read-Only:

- `metadata` (Attributes) an identifier, version and namespace that uniquely identify the MTA (see [below for nested schema](#nestedatt--mtas--metadata))
- `modules` (Attributes List) the deployable parts contained in the MTA deployment archive, most commonly Cloud Foundry applications or content (see [below for nested schema](#nestedatt--mtas--modules))
- `services` (List of String)

<a id="nestedatt--mtas--metadata"></a>
### Nested Schema for `mtas.metadata`

Read-Only:

- `id` (String)
- `namespace` (String)
- `version` (String)


<a id="nestedatt--mtas--modules"></a>
### Nested Schema for `mtas.modules`

Read-Only:

- `app_name` (String)
- `created_on` (String)
- `module_name` (String)
- `provided_dendency_names` (List of String)
- `services` (List of String)
- `updated_on` (String)
- `uris` (List of String)