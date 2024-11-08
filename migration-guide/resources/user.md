# cloudfoundry_user

Provides a resource for creating users in the origin store and registering them in Cloud Foundry

|  Cloud Foundry Provider |Community Cloudfoundry Provider |
| -- | -- |
|  <pre>resource "cloudfoundry_user" "my_user" {</br>  username    = "test"</br>  email       = "test@gmail.com"</br>  password    = "test123"</br>  given_name  = "test"</br>  family_name = "test"</br>  annotations = { "purpose" : "testing", hi : "hello" }</br>}</br></pre> |<pre>resource "cloudfoundry_user" "my_user" {</br>  name    = "test"</br>  email       = "test@gmail.com"</br>  password    = "test123"</br>  given_name  = "test"</br>  family_name = "test"</br>}</br></pre> |

## Differences
> [!NOTE]  
> 🔵 Required  🟢 Optional 🟠 Computed  🔴 Not present

| Attribute name| Cloud Foundry Provider|  Community Provider(old) | Description
|---| ---| ---| ---| 
|name | 🔴|  🔵  | `name` has been changed to `username`  to maintain conformity with UAA API. |
|username  |   🔵 |🔴| - |
|groups  |   🟠 |🟢| Assigning groups to the user functionality can be achieved from the [`cloudfoundry_user_groups`](../../docs/resources/user_groups.md) resource | 
|labels  |  🟢 | 🔴| - | 
|annotations  |   🟢 | 🔴| - | 
