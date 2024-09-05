# cloudfoundry_space_quota

Gets information on a Cloud Foundry space quota.

| cloudfoundry/cloudfoundry | cloudfoundry-community/cloudfoundry |
| -- | -- |
| <pre>data "cloudfoundry_space_quota" "my_space_quota" {</br>  name = "tf-test-do-not-delete"</br>  org  = "ca721b24-e24d-4171-83e1-1ef6bd836b38"</br>}</br></pre>|<pre>data "cloudfoundry_space_quota" "my_space_quota" {</br>  name = "tf-test-do-not-delete"</br>  org  = "ca721b24-e24d-4171-83e1-1ef6bd836b38"</br>}</br></pre> |  

## Differences

> [!NOTE]  
> 🔵 Required  🟢 Optional 🟠 Computed  🔴 Not present

| Attribute name | SAP Cloud Foundry Provider (new)|  Community Cloud Foundry Provider (old) | Description |
| --- | --- | --- | --- |
| spaces | 🟠 | 🔴 | - |
| total_app_log_rate_limit | 🟠 | 🔴 | - |
| total_memory | 🟠 | 🔴 | - |
| total_routes | 🟠 | 🔴 | - |
| total_services | 🟠 | 🔴 | - |
| allow_paid_service_plans | 🟠 | 🔴 | - |
| instance_memory | 🟠 | 🔴 | - |
| total_app_instances | 🟠 | 🔴 | - |
| total_services | 🟠 | 🔴 | - |
| total_app_tasks | 🟠 | 🔴 | - |
| total_route_ports | 🟠 | 🔴 | - |
| total_service_keys | 🟠 | 🔴 | - |
