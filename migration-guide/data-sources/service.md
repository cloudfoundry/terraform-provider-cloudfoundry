# cloudfoundry_service

Get Service Offering and its related plans.

| cloudfoundry/cloudfoundry | cloudfoundry-community/cloudfoundry |
| -- | -- |
| <pre>data "cloudfoundry_service" "redis" {</br>  name = "p-redis"</br>}</br></pre>|<pre>data "cloudfoundry_service" "redis" {</br>    name = "p-redis"    </br>}</br></pre> |

## Differences

> [!NOTE]  
> 🔵 Required  🟢 Optional 🟠 Computed  🔴 Not present

| Attribute name | SAP Cloud Foundry Provider (new)|  Community Cloud Foundry Provider (old) | Description |
| --- | --- | --- | --- |
| service_broker_guid | 🔴 | 🟢 | `service_broker_guid`attribute has been changed to `service_broker` |
| service_broker | 🟢 | 🔴 | - |
| service_broker_name | 🔴 | 🟠 | - |
| space | 🔴 | 🟢 | - |
