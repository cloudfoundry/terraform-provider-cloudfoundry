# cloudfoundry_route (Resource)

Provides a Cloud Foundry resource for managing Cloud Foundry application routes.

| cloudfoundry/cloudfoundry | cloudfoundry-community/cloudfoundry |
| -- | -- |
|  <pre></br>resource "cloudfoundry_route" "bruh" {</br>  space  = "795a961c-6360-479a-9666-fff9cb906aad"</br>  domain = "440e24e5-ee11-41d9-a996-2ed0a1e2deea"</br></br>  #Optional </br></br>  host   = "myapp"</br>  path   = "/cart"</br>  destinations = [</br>    {</br>      app_id = "24a711f2-148b-4d48-b37a-90a66d6e842f"</br></br>    },</br>    {</br>      app_id           = "15a74002-cf3a-4bf2-b76f-fe96867c46ee"</br>      app_process_type = "web"</br>      port             = 36001</br>    },</br>  ]</br>}</br></br></pre> |<pre>resource "cloudfoundry_route" "bruh" {</br>    space = "795a961c-6360-479a-9666-fff9cb906aad"</br>    domain = "440e24e5-ee11-41d9-a996-2ed0a1e2deea"</br>    </br>    #Optional </br></br>    hostname = "myapp"</br>    path   = "/cart"</br>    target = [</br>    {</br>      app = "24a711f2-148b-4d48-b37a-90a66d6e842f"</br>    },</br>    {</br>      app  = "15a74002-cf3a-4bf2-b76f-fe96867c46ee"</br>      port = 36001</br>    },</br>  ]</br>}</br></pre> |

## Differences

> [!NOTE]  
> 🔵 Required  🟢 Optional 🟠 Computed  🔴 Not present

| Attribute name | SAP Cloud Foundry Provider (new)|  Community Cloud Foundry Provider (old) | Description |
| --- | --- | --- | --- |
| hostname | 🔴 | 🟢 | `hostname` has been changed to `host`  to maintain conformity with V3 API. |
| host | 🟢 | 🔴 |- |
| endpoint | 🔴 | 🟠 | `endpoint` has been changed to `url`  to maintain conformity with V3 API. |
| url | 🟠 | 🔴 | - |
| target | 🔴 | 🟢 | `target` has been changed to `destinations`  to maintain conformity with V3 API. |
| destinations | 🟢 | 🔴 |  Supports additional attributes such as app_process_type,protocol and weight. |
| labels | 🟢 | 🔴 | - |
| annotations | 🟢 | 🔴 | - |
| protocol | 🟠 | 🔴 | - |
