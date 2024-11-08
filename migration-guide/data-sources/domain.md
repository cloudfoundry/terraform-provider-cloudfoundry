# cloudfoundry_domain

Gets information on a Cloud Foundry domain.

| cloudfoundry/cloudfoundry | cloudfoundry-community/cloudfoundry |
| -- | -- |
| <pre>data "cloudfoundry_domain" "mydomain" {</br>  name = "test.cfapps.stagingazure.hanavlab.ondemand.com"</br>}</br></pre>|<pre>data "cloudfoundry_domain" "mydomain" {</br>    sub_domain = "test"</br>    domain = "cfapps.stagingazure.hanavlab.ondemand.com"</br>}</br></pre> |  

## Differences

> [!NOTE]  
> 🔵 Required  🟢 Optional 🟠 Computed  🔴 Not present

| Attribute name | Cloud Foundry Provider|  Community Cloud Foundry Provider (old) | Description |
| --- | --- | --- | --- |
| sub_domain | 🔴 | 🟢 | `sub_domain` and `domain` attributes from community provider values need to be combined to a FQDN and value should be set in `name` attribute in the format of 'sub_domain.domain' |
| domain | 🔴| 🟢 | Refer above |
| supported_protocols | 🟠 | 🔴 | - |
| shared_orgs | 🟠 | 🔴 | - |
| labels | 🟠 | 🔴 | - |
| annotations | 🟠 | 🔴 | - |
| router_group | 🟠 | 🔴 | - |
