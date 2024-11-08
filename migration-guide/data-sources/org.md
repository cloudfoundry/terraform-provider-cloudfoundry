# cloudfoundry_org

Gets information on a Cloud Foundry organization.

| cloudfoundry/cloudfoundry | cloudfoundry-community/cloudfoundry |
| -- | -- |
| <pre>data "cloudfoundry_org" "org" {</br>  name = "myorg"</br>}</br></pre>|<pre>data "cloudfoundry_org" "org" {</br>    name = "myorg"    </br>}</br></pre> |  

## Differences

> [!NOTE]  
> 🔵 Required  🟢 Optional 🟠 Computed  🔴 Not present

| Attribute name | Cloud Foundry Provider|  Community Cloud Foundry Provider (old) | Description |
| --- | --- | --- | --- |
| quota | 🟠| 🔴 | - |
| suspended | 🟠 | 🔴 | - |
