# cloudfoundry_space

Gets information on a Cloud Foundry space.

| cloudfoundry/cloudfoundry | cloudfoundry-community/cloudfoundry |
| -- | -- |
| <pre>data "cloudfoundry_space" "space" {</br>  name = "myspace"</br>  org  = "ca721b24-e24d-4171-83e1-1ef6bd836b38"</br>}</br></pre>|<pre>data "cloudfoundry_space" "space" {</br>    name = "myspace"</br>    org_name = "org"</br>}</br></pre> | 

## Differences

> [!NOTE]  
> 🔵 Required  🟢 Optional 🟠 Computed  🔴 Not present

| Attribute name | SAP Cloud Foundry Provider (new)|  Community Cloud Foundry Provider (old) | Description |
| --- | --- | --- | --- |
| org_name | 🔴 | 🟢 | - |
| org | 🔵 | 🟢 | Space can now be queried only by `org` GUID and not by `org_name`. If one knows org_name and not org GUID, one can obtain the id value from [`cloudfoundry_org`](/docs/data-sources/org.md) data source by specifying `name`. |
| allow_ssh | 🟠 | 🔴 | - |
| isolation_segment | 🟠 | 🔴 | - |
