# cloudfoundry_service_instance

Provides a data source for fetching information of a service instance in Cloud Foundry. Combines functionality of [`cloudfoundry_service_instance`](https://github.com/cloudfoundry-community/terraform-provider-cloudfoundry/blob/main/docs/data-sources/service_instance.md) and [`cloudfoundry_user_provided_service`](https://github.com/cloudfoundry-community/terraform-provider-cloudfoundry/blob/main/docs/data-sources/user_provided_service.md#cloudfoundry_user_provided_service) in the community provider.

| cloudfoundry/cloudfoundry | cloudfoundry-community/cloudfoundry |
| -- | -- |
| <pre>data "cloudfoundry_service_instance" "my-instance" {</br>  name  = "my-service-instance"</br>  space = "02c0cc92-6ecc-44b1-b7b2-096ca19ee143"</br>}</br></br>data "cloudfoundry_service_instance" "svc" {</br>  name  = "managed-service-instance"</br>  space = "02c0cc92-6ecc-44b1-b7b2-096ca19ee143"</br>}</br></pre>|<pre>data "cloudfoundry_user_provided_service" "my-instance" {</br>    name  = "my-service-instance"</br>    space = "02c0cc92-6ecc-44b1-b7b2-096ca19ee143"</br>}</br></br>data "cloudfoundry_service_instance" "svc" {</br>    name_or_id = "managed-service-instance"</br>    space      = "02c0cc92-6ecc-44b1-b7b2-096ca19ee143"</br>}</br></pre> |

## Differences

> [!NOTE]  
> 🔵 Required  🟢 Optional 🟠 Computed  🔴 Not present

| Attribute name | Cloud Foundry Provider|  Community Cloud Foundry Provider (old) | Description |
| --- | --- | --- | --- |
| name_or_id | 🔴 | 🟢 | Only service instance name has to be specified in `name` |
| labels | 🟠 | 🔴 | - |
| annotations | 🟠 | 🔴 | - |
| service_plan_id | 🔴 | 🟠 | `service_plan_id` has been changed to `service_plan`  to maintain conformity with V3 API |
| service_plan | 🟠 | 🔴 | - |
| last_operation | 🟠 | 🔴 | - |
| maintenance_info | 🟠 | 🔴 | - |
| upgrade_available | 🟠 | 🔴 | - |
| dashboard_url | 🟠 | 🔴 | - |
| type | 🟠 | 🔴 | - |
| credentials | 🔴 | 🟠 | - |
