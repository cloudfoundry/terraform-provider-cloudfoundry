---
page_title: "cloudfoundry_service_instance_sharing Resource - terraform-provider-cloudfoundry"
subcategory: ""
description: |-
In Cloud Foundry, service instance sharing allows you to share a service instance with other spaces within your Cloud Foundry environment.
When you create a service instance, it is typically only accessible within the space where it was created. However, there may be cases where you want to make the service instance available to other spaces. This is where service instance sharing comes in.
This feature is useful in scenarios where you have a service instance that needs to be shared across different teams or projects within your Cloud Foundry environment. It promotes collaboration and allows multiple spaces to leverage the same service instance without the need for duplicating resources or creating separate instances.
---

# cloudfoundry_service_instance_sharing (Resource)

In Cloud Foundry, service instance sharing allows you to share a service instance with other spaces within your Cloud Foundry environment.
When you create a service instance, it is typically only accessible within the space where it was created. However, there may be cases where you want to make the service instance available to other spaces. This is where service instance sharing comes in.
This feature is useful in scenarios where you have a service instance that needs to be shared across different teams or projects within your Cloud Foundry environment. It promotes collaboration and allows multiple spaces to leverage the same service instance without the need for duplicating resources or creating separate instances.

# Example Usage

```terraform
resource "cloudfoundry_service_instance_sharing" "instance_sharing" {
  service_instance_id = "e9ec29ca-993d-42e2-9c5b-cb17b1972cce"
  space_id            = "ec6ac2b3-fb79-43c4-9734-000d4299bd59"
}
```

## Schema

### Required

- `service_instance_id` (String) The GUID of the service instance to be bound
- `space_id` (String) The GUID of the space to where the service instance should be shared