# cloudfoundry_service_instance

Provides a resource for managing service instances in Cloud Foundry. Combines [`cloudfoundry_service_instance`](https://github.com/cloudfoundry-community/terraform-provider-cloudfoundry/blob/main/docs/resources/service_instance.md) and [`cloudfoundry_user_provided_service`](https://github.com/cloudfoundry-community/terraform-provider-cloudfoundry/blob/main/docs/resources/user_provided_service.md#cloudfoundry_user_provided_service) in the community provider to resemble service instance resource as provided in V3 API.

| cloudfoundry/cloudfoundry | cloudfoundry-community/cloudfoundry |
| -- | -- |
|  <pre>resource "cloudfoundry_service_instance" "redis1" {</br>  name         = "pricing-grid"</br>  type         = "managed"</br>  space        = "e6886bba-e263-4b52-aaf1-85d410f15fc8"</br>  tags         = ["terraform-test", "test1"]</br>  service_plan = data.cloudfoundry_service.redis.service_plans["shared-vm"]</br></br>resource "cloudfoundry_service_instance" "mq" {</br>  name        = "mq"</br>  type        = "user-provided"</br>  space       = "e6886bba-e263-4b52-aaf1-85d410f15fc8"</br>  credentials = <<EOT</br>  {</br>    "url" = "mq://localhost:9000"</br>    "username" = "admin"</br>    "password" = "admin"</br>  }</br>  EOT</br>}</br></pre> |<pre>resource "cloudfoundry_service_instance" "redis1" {</br>  name         = "pricing-grid"</br>  space        = "e6886bba-e263-4b52-aaf1-85d410f15fc8"</br>  tags         = ["terraform-test", "test1"]</br>  service_plan = data.cloudfoundry_service.redis.service_plans["shared-vm"]</br>}</br></br>resource "cloudfoundry_user_provided_service" "mq" {</br>  name = "mq-server"</br>  space = "e6886bba-e263-4b52-aaf1-85d410f15fc8"</br>  credentials = {</br>    "url" = "mq://localhost:9000"</br>    "username" = "admin"</br>    "password" = "admin"</br>  }</br>}</br></pre> |

## Differences

> [!NOTE]  
> 🔵 Required  🟢 Optional 🟠 Computed  🔴 Not present

| Attribute name | Cloud Foundry Provider|  Community Cloud Foundry Provider (old) | Description |
| --- | --- | --- | --- |
| type | 🔵 | 🔴 | Need to specify whether instance is of type managed or user-provided. |
| labels | 🟢 | 🔴 | - |
| annotations | 🟢 | 🔴 | - |
| app | 🟢 | 🔴 | App GUID needs to be specified if `type` binding is `app`. |
| last_operation | 🟠 | 🔴 | - |
| maintenance_info | 🟠 | 🔴 | - |
| upgrade_available | 🟠 | 🔴 | - |
| dashboard_url | 🟠 | 🔴 | - |
| json_params |  🔴 | 🟢 | `json_params` has been changed to `parameters`  to maintain conformity with V3 API. |
| parameters | 🟢| 🔴 | - |
| credentials_json | 🔴 | 🟢 | `credentials_json` Functionality can be achieved by `credentials` attribute. |
| credentials | 🟢| 🔴 | - |
| recursive_delete | 🔴 | 🟢 | V3 API by default follows recursive deletion. |
| replace_on_params_change | 🔴 | 🟢 | - |
| replace_on_service_plan_change | 🔴 | 🟢 | - |
