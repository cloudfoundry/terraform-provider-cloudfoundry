# cloudfoundry_space

Provides a Cloud Foundry resource for managing Cloud Foundry spaces within organizations.

| cloudfoundry/cloudfoundry | cloudfoundry-community/cloudfoundry |
| -- | -- |
|  <pre>resource "cloudfoundry_space" "space" {</br>  name      = "space"</br>  org       = "ca721b24-e24d-4171-83e1-1ef6bd836b38"</br>  allow_ssh = "true"</br>}</br></pre> |<pre>resource "cloudfoundry_space" "space" {</br>    name = "space"</br>    org  = "ca721b24-e24d-4171-83e1-1ef6bd836b38"</br>    quota = "dd457c79-f7c9-4828-862b-35843d3b646d"</br>    asgs = [ "ba10cc63-cc43-46b1-a00c-5f2a0d7d992e" ]</br>    allow_ssh = true</br>}</br></pre> |

## Differences

> [!NOTE]  
> 🔵 Required  🟢 Optional 🟠 Computed  🔴 Not present

| Attribute name | Cloud Foundry Provider|  Community Cloud Foundry Provider (old) | Description |
| --- | --- | --- | --- |
| quota | 🟠 | 🟢 | One cannot set quota as it is a read-only attribute in the current provider. For setting quota  use resource [`cloudfoundry_space_quota`](space_quota.md). |
| asgs| 🔴 | 🟢 | Security groups not present in space resource as part of V3 API Spec. One can however set it with `running_spaces` attribute from resource [`cloudfoundry_security_group`](security_group.md). |
| staging_asgs| 🔴 | 🟢 | Staging Security groups not present in space resource as part of V3 API Spec. One can however set it with `staging_spaces` attribute from resource [`cloudfoundry_security_group`](security_group.md). |
| delete_recursive_allowed | 🔴 | 🟢 | V3 API by default follows recursive deletion. |
