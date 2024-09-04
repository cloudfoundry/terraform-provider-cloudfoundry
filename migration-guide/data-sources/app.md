# cloudfoundry_app

Gets information on a Cloud Foundry application.

| cloudfoundry/cloudfoundry | cloudfoundry-community/cloudfoundry |
| -- | -- |
| <pre>data "cloudfoundry_app" "my-app" {</br>  name  = "my-app"</br>  space_name = "tf-space-1"</br>  org_name   = "PerformanceTeamBLR"</br>}</br></pre>|<pre>data "cloudfoundry_app" "my-app" {</br>    name_or_id = "my-app"</br>    space      = "106a411e-3ea3-4d1b-a30d-54a6802bed27"</br>}</br></pre> |  

## Differences

> [!NOTE]  
> 🔵 Required  🟢 Optional 🟠 Computed  🔴 Not present

| Attribute name | SAP Cloud Foundry Provider (new)|  Community Cloud Foundry Provider (old) | Description |
| --- | --- | --- | --- |
| name_or_id |  🔴 | 🟢 | Only application name has to be specified in `name` |
| name | 🔵 | 🟠 | - |
| org_name | 🔵 |  🔴  | Organization name where space is present has to be specified |
| space_name | 🔵 |  🔴 | Instead of specifying guid for `space` attribute in the old community provider, user should specify space name in `space_name` attribute for the new provider |
| space |  🔴 | 🔵  | Refer above |
| buildpack |  🔴 | 🟠 | - |
| enable_ssh |  🔴 | 🟠 | - |
| state |  🔴 | 🟠 | - |
| docker_credentials | 🟠 | 🔴 | - |
| docker_image | 🟠 | 🔴 | - |
| health_check_interval | 🟠 | 🔴 | - |
| health_check_timeout | 🔴 | 🟠 | `health_check_timeout` has been changed to `timeout` to maintain conformity with V3 API |
| timeout | 🟠 | 🔴 | Refer above |
| health_check_invocation_timeout | 🟠 | 🔴 | - |
| log_rate_limit_per_second | 🟠 | 🔴 | - |
| memory | 🟠 | 🔴 | - |
| processes | 🟠 | 🔴 | - |
| readiness_health_check_http_endpoint | 🟠 | 🔴 | - |
| readiness_health_check_interval | 🟠 | 🔴 | - |
| readiness_health_check_invocation_timeout | 🟠 | 🔴 | - |
| readiness_health_check_type | 🟠 | 🔴 | - |
| routes | 🟠 | 🔴 | - |
| service_bindings | 🟠 | 🔴 | - |
| sidecars | 🟠 | 🔴 | - |
| stack| 🟠 | 🔴 | - |
| readiness_health_check_invocation_timeout | 🟠 | 🔴 | - |
