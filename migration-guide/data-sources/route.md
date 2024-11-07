# cloudfoundry_route 

Gets information on a Cloud Foundry route.

| cloudfoundry/cloudfoundry | cloudfoundry-community/cloudfoundry |
| -- | -- |
| <pre>data "cloudfoundry_route" "bruh" {</br>  domain = "a25ca0c1-353a-40f9-bcf4-d2a0adf4112b"</br>  host = "my-host"</br>  space  = "b45da1f2-353a-40f9-bcf4-d2a0adf4112b"</br>}</br></pre>|<pre></br>data "cloudfoundry_route" "my-route" {</br>    domain   = "a25ca0c1-353a-40f9-bcf4-d2a0adf4112b"</br>    hostname = "my-host"</br>}</br></pre> |  

## Differences

> [!NOTE]  
> 🔵 Required  🟢 Optional 🟠 Computed  🔴 Not present

| Attribute name | Cloud Foundry Provider|  Community Cloud Foundry Provider (old) | Description |
| --- | --- | --- | --- |
| id | 🔴 | 🟠 |  The current provider returns multiple routes if available with same domain name in the `routes` attribute unlike the community provider. Therefore the id is present in the respective route resources in `routes` output |
| hostname | 🔴 | 🟢 | `hostname` has been changed to `host`  to maintain conformity with V3 API |
| host | 🟢 | 🔴 | Refer above |
| space | 🟢 | 🔴 | - |
| routes | 🟠 | 🔴 | - |
