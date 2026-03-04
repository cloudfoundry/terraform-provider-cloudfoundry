# cloudfoundry_service_credential_binding  

Provides a resource for managing service credential bindings in Cloud Foundry. Combines [`cloudfoundry_service_key`](https://github.com/cloudfoundry-community/terraform-provider-cloudfoundry/blob/main/docs/resources/service_key.md) in the community provider and app service binding to resemble service credential binding resource as provided in V3 API.

| cloudfoundry/cloudfoundry | cloudfoundry-community/cloudfoundry |
| -- | -- |
|  <pre>resource "cloudfoundry_service_credential_binding" "scb1" {</br>  type             = "key"</br>  name             = "hifi"</br>  service_instance = "e9ec29ca-993d-42e2-9c5b-cb17b1972cce"</br>}</br></br>resource "cloudfoundry_service_credential_binding" "scb7" {</br>  type             = "app"</br>  name             = "hifi"</br>  service_instance = "e9ec29ca-993d-42e2-9c5b-cb17b1972cce"</br>  app              = "ec6ac2b3-fb79-43c4-9734-000d4299bd59"</br>}</br></pre> |<pre>resource "cloudfoundry_service_key" "redis1-key1" {</br>  name = "hifi"</br>  service_instance = "e9ec29ca-993d-42e2-9c5b-cb17b1972cce"</br>}</br></pre> |

## Differences

> [!NOTE]  
> 🔵 Required  🟢 Optional 🟠 Computed  🔴 Not present

| Attribute name | Cloud Foundry Provider|  Community Cloud Foundry Provider (old) | Description |
| --- | --- | --- | --- |
| credentials | 🔴 | 🟠 | Attribute not returned as part of V3 API resource. However, it is obtainable from `credential_binding` attribute of data source [`cloudfoundry_service_credential_binding`](../data-sources/service_credential_binding.md). |
| type | 🔵 | 🔴 | Need to specify whether binding is of type app or key. |
| labels | 🟢 | 🔴 | - |
| annotations | 🟢 | 🔴 | - |
| app | 🟢 | 🔴 | App GUID needs to be specified if `type` binding is `app`. |
| last_operation | 🟠 | 🔴 | - |
| params_json | 🔴 | 🟢 |  `params_json` has been changed to `parameters`  to maintain conformity with V3 API. |
| parameters | 🟢 | 🔴 | - |
| params | 🔴 | 🟢 | `params` functionality can be achieved by `parameters`. |
