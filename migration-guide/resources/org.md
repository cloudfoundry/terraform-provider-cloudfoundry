# cloudfoundry_org (Resource)

Provides a Cloud Foundry resource for managing Cloud Foundry organizations


| cloudfoundry/cloudfoundry | cloudfoundry-community/cloudfoundry |
| -- | -- |
|  <pre>resource "cloudfoundry_org" "org" {</br>  name      = "tf-test"</br>  suspended = false</br>}</br></pre> |<pre>resource "cloudfoundry_org" "org" {</br>    name = "tf-test"</br>    quota = cloudfoundry_quota.runaway.id</br>}</br></pre> |

## Differences

> [!NOTE]  
> 🔵 Required  🟢 Optional 🟠 Computed  🔴 Not present

| Attribute name | SAP Cloud Foundry Provider (new)|  Community Cloud Foundry Provider (old) | Description |
| --- | --- | --- | --- |
| quota | 🟠| 🟢 | One cannot set quota as it is a read-only attribute in the current provider. For setting quota use resource [`cloudfoundry_org_quota`](/docs/resources/org_quota.md). |
| suspended | 🟢 | 🔴 | - |
| delete_recursive_allowed | 🔴 | 🟢 | V3 API by default follows recursive deletion. |
