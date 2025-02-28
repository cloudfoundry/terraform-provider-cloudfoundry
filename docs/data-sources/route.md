---
page_title: "cloudfoundry_route Data Source - terraform-provider-cloudfoundry"
subcategory: ""
description: |-
  Gets information on a Cloud Foundry route.
---

# cloudfoundry_route (Data Source)

Gets information on a Cloud Foundry route.

## Example Usage

```terraform
data "cloudfoundry_route" "route" {
  domain = "a25ca0c1-353a-40f9-bcf4-d2a0adf4112b"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `domain` (String) The domain guid associated to the route.

### Optional

- `host` (String) The hostname associated to the route to lookup.
- `org` (String) The org guid associated to the route to lookup.
- `path` (String) The path associated to the route to lookup.
- `port` (Number) The port associated to the route to lookup.
- `space` (String) The space guid associated to the route.

### Read-Only

- `routes` (Attributes List) The list of routes. (see [below for nested schema](#nestedatt--routes))

<a id="nestedatt--routes"></a>
### Nested Schema for `routes`

Read-Only:

- `annotations` (Map of String) The annotations associated with Cloud Foundry resources.
- `created_at` (String) The date and time when the resource was created in [RFC3339](https://www.ietf.org/rfc/rfc3339.txt) format.
- `destinations` (Attributes Set) A destination represents the relationship between a route and a resource that can serve traffic. (see [below for nested schema](#nestedatt--routes--destinations))
- `domain` (String) The domain guid associated to the route.
- `host` (String) The hostname associated to the route to lookup.
- `id` (String) The GUID of the object.
- `labels` (Map of String) The labels associated with Cloud Foundry resources.
- `path` (String) The path associated to the route to lookup.
- `port` (Number) The port associated to the route to lookup.
- `protocol` (String) The protocol supported by the route, based on the route's domain configuration.
- `space` (String) The space guid associated to the route.
- `updated_at` (String) The date and time when the resource was updated in [RFC3339](https://www.ietf.org/rfc/rfc3339.txt) format.
- `url` (String) The URL for the route.

<a id="nestedatt--routes--destinations"></a>
### Nested Schema for `routes.destinations`

Read-Only:

- `app_id` (String) The GUID of the app to route traffic to.
- `app_process_type` (String) Type of the process belonging to the app to route traffic to.
- `id` (String) The GUID of the object.
- `port` (Number) Port on the destination process to route traffic to.
- `protocol` (String) Protocol to use for this destination.
- `weight` (Number) Percentage of traffic which will be routed to this destination.