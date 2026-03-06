# cloudfoundry_app

Provides a Cloud Foundry resource to manage applications.

| cloudfoundry/cloudfoundry | cloudfoundry-community/cloudfoundry |
| -- | -- |
|  <pre>resource "cloudfoundry_app" "my-app" {</br>  name       = "my-app"</br>  space_name = "tf-space-1"</br>  org_name   = "PerformanceTeamBLR"</br>  buildpacks = ["nodejs_buildpack"]</br>  memory     = "512M"</br>  path       = "something.zip"</br>  service_bindings = [</br>    {</br>      service_instance = "xsuaa-tf"</br>    }</br>  ]</br>  routes = [</br>    {</br>      route = my-app.hello.world.example..com"</br>    }</br>  ]</br>}</br></pre> |<pre>resource "cloudfoundry_app" "my-app" {</br>  name       = "my-app"</br>  space      = "e6886bba-e263-4b52-aaf1-85d410f15fc8"</br>  buildpack = "nodejs_buildpack"</br>  memory     = 512</br>  path       = "something.zip"</br>  service_binding {</br>      service_instance = "xsuaa-tf"</br>  }</br>  routes = {</br>      route = my-app.hello.world.example..com"</br>  }</br>}</br></pre> |

## Differences

> [!NOTE]  
> 🔵 Required  🟢 Optional 🟠 Computed  🔴 Not present

| Attribute name | Cloud Foundry Provider|  Community Cloud Foundry Provider (old) | Description |
| --- | --- | --- | --- |
| org_name| 🔵 | 🔴 | Organization name where space is present has to be specified. |
| space_name | 🔵 | 🔴 | Instead of specifying guid for `space` attribute in the old community provider, user should specify space name in `space_name` attribute for the new provider. |
| space | 🔴 | 🔵 | Refer above |
| buildpack | 🔴 | 🟢 | `buildpack` attribute functionality can be achieved by `buildpacks` attribute. |
| enable_ssh | 🔴 | 🟢 | It can be enabled on a space level. For further details, refer [here](https://docs.cloudfoundry.org/devguide/deploy-apps/ssh-apps.html#config-ssh-access-apps). |
| stopped | 🔴 | 🟢 | `stopped` attribute functionality can be achieved by setting `instances` to 0. |
| routes.route | 🟢 | 🟢 | In the new provider, FQDN needs to be specified instead of the route GUID in the community provider. Route resource is automatically created if not present. |
| routes.port | 🔴 | 🟢 | Not present in V3 manifest schema. Can be set in `port` attribute of  [`cloudfoundry_route`](route.md) resource. |
| routes.protocol | 🟢 | 🔴 | - |
| health_check_interval | 🟢 | 🔴 | - |
| log_rate_limit_per_second | 🟢 | 🔴 | - |
| random_route | 🟢 | 🔴 | - |
| no_route | 🟢 | 🔴 | - |
| processes | 🟢 | 🔴 | - |
| sidecars | 🟢 | 🔴 | - |
| readiness_health_check_http_endpoint | 🟢 | 🔴 | - |
| readiness_health_check_interval | 🟢 | 🔴 | - |
| readiness_health_check_invocation_timeout | 🟢 | 🔴 | - |
| readiness_health_check_type | 🟢 | 🔴 | - |
| health_check_timeout | 🔴 | 🟢 | `health_check_timeout` has been changed to `timeout`  to maintain conformity with V3 API. |
| timeout | 🟢 | 🟢 | `timeout` attribute in the current provider is for health check timeout and not for starting the app initially. |
| service_binding | 🔴 | 🟢 | - |
| service_bindings | 🟢 | 🔴 | `service_binding` has been changed to `service_bindings` to maintain conformity with V3 API. |
