# cloudfoundry_domain

Provides a Cloud Foundry resource for managing shared or private domains in Cloud Foundry.

| cloudfoundry/cloudfoundry | cloudfoundry-community/cloudfoundry |
| -- | -- |
|  <pre>resource "cloudfoundry_domain" "sample" {</br>  name  = "test.cfapps.stagingazure.hanavlab.ondemand.com"</br>  org         = "23919ba5-f9b6-4128-a1fb-69890818d25c"</br>  shared_orgs = ["537e7b58-b3e0-4464-9cad-2deae6120a80", "30edf44a-2d4c-432c-9680-9a61123edcf1"]</br>}</br></pre> |<pre>resource "cloudfoundry_domain" "sample" {</br>  sub_domain = "test"</br>  domain = "cfapps.stagingazure.hanavlab.ondemand.com"</br>  org         = "23919ba5-f9b6-4128-a1fb-69890818d25c"</br>}</br></pre> |

## Differences

> [!NOTE]  
> 🔵 Required  🟢 Optional 🟠 Computed  🔴 Not present

|| Attribute name | Cloud Foundry Provider|  Community Cloud Foundry Provider (old) | Description |
| --- | --- | --- | --- |
| sub_domain | 🔴 | 🟢 | `sub_domain` and `domain` attributes from community provider values need to be combined to a FQDN and value should be set in `name` attribute. |
| domain | 🔴| 🟢 | Refer above |
| supported_protocols |  🟠| 🔴 | - |
| shared_orgs | 🟢 | 🔴 | Accomplishes the feature of [`cloudfoundry_private_domain_access`](https://github.com/cloudfoundry-community/terraform-provider-cloudfoundry/blob/main/docs/resources/private_domain_access.md) resource in the community provider .Allows specifying the organizations to share the private domain with. |
| labels | 🟢 | 🔴 | - |
| annotations | 🟢 | 🔴 | - |
