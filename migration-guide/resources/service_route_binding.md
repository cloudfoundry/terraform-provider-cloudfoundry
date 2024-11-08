# cloudfoundry_service_route_binding  

Provides a resource for managing service route bindings in Cloud Foundry. Named as [`cloudfoundry_route_service_binding`](https://github.com/cloudfoundry-community/terraform-provider-cloudfoundry/blob/main/docs/resources/route_service_binding.md) in the community provider.

|  Cloud Foundry Provider |Community Cloudfoundry Provider |
| -- | -- |
|  <pre>resource "cloudfoundry_service_route_binding" "srb" {</br>  service_instance = "3a8588f9-f846-444f-ab9e-48282f06449b"</br>  route            = "3966c2fb-d84d-462d-82a5-a81cf7cdab20"</br>  labels           = { "hi" : "fi" }</br>}</br></pre> |<pre>resource "cloudfoundry_route_service_binding" "srb" {</br>  service_instance = "3a8588f9-f846-444f-ab9e-48282f06449b"</br>  route            = "3966c2fb-d84d-462d-82a5-a81cf7cdab20"</br>}</br></pre> |

## Differences

> [!NOTE]  
> 🔵 Required  🟢 Optional 🟠 Computed  🔴 Not present

| Attribute name | Cloud Foundry Provider|  Community Cloud Foundry Provider (old) | Description |
| --- | --- | --- | --- |
| json_params | 🔴 | 🟢 |  `json_params` has been changed to `parameters`  to maintain conformity with V3 API. |
| parameters | 🟢 | 🔴 | - |
| labels | 🟢 | 🔴 | - |
| annotations | 🟢 | 🔴 | - |
| last_operation | 🟠 | 🔴 | - |

