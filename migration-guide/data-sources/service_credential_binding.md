# cloudfoundry_service_credential_binding

Gets information on Service Credential Bindings for a given service instance. Combines [`cloudfoundry_service_key`](https://github.com/cloudfoundry-community/terraform-provider-cloudfoundry/blob/main/docs/data-sources/service_key.md) in the community provider and querying for app app credential_bindings.

| cloudfoundry/cloudfoundry | cloudfoundry-community/cloudfoundry |
| -- | -- |
| <pre>data "cloudfoundry_service_credential_binding" "scbd" {</br>  service_instance = "5e2976bb-332e-41e1-8be3-53baafea9296"</br>  app              = "ec6ac2b3-fb79-43c4-9734-000d4299bd59"</br>}</br></br>data "cloudfoundry_service_credential_binding" "my-key" {</br>    name             = "my-service-key"</br>    service_instance = "e9ec29ca-993d-42e2-9c5b-cb17b1972cce"</br>}</br></pre>|<pre>data "cloudfoundry_service_key" "my-key" {</br>    name             = "my-service-key"</br>    service_instance = "e9ec29ca-993d-42e2-9c5b-cb17b1972cce"</br>}</br></pre> |

## Differences

> [!NOTE]  
> 🔵 Required  🟢 Optional 🟠 Computed  🔴 Not present

| Attribute name | SAP Cloud Foundry Provider (new)|  Community Cloud Foundry Provider (old) | Description |
| --- | --- | --- | --- |
| name | 🟢 | 🔵 | Since `cloudfoundry_service_key` resource in community provider can only query for service keys and not app bindings, name has been made optional now as app bindings need not have names. |
| app | 🟢 | 🔴 | One can query for a specific app credential_binding by providing its GUID |
| credential_bindings | 🟠 | 🔴 | - |
| id | 🔴 | 🟠 |  The current provider returns multiple credential bindings if available for the given `service_instance` unlike the community provider where we only search for a particular service key by its `name`. Therefore, one can refer the `id` of the respective binding in the `credential_bindings` attribute. |
| credentials |  🔴 |🟠| Since multiple credential bindings can be obtained, the credential details can be obtained in the respective bindings in the `credential_binding` attribute of `credential_bindings` output. |
