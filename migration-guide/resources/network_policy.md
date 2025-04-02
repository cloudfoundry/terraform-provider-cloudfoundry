# cloudfoundry_network_policy

Provides a Cloud Foundry resource for managing Cloud Foundry Network Policies

| cloudfoundry/cloudfoundry | cloudfoundry-community/cloudfoundry |
| -- | -- |
| <pre>resource "cloudfoundry_network_policy" "policy" {<br/>  policies = [<br/>    {<br/>      source_app      = "16b53647-9709-44bf-91b2-116de83ffd3d"<br/>      destination_app = "41048361-adc7-4686-9115-36b16d8df12c"<br/>      port            = "61443"<br/>      protocol        = "tcp"<br/>    }<br/>  ]<br/>}</pre> | <pre>resource "cloudfoundry_network_policy" "policy" {<br/>  policy {<br/>    source_app      = "16b53647-9709-44bf-91b2-116de83ffd3d"<br/>    destination_app = "41048361-adc7-4686-9115-36b16d8df12c"<br/>    port            = "61443"<br/>    protocol        = "tcp"<br/>  }<br/>}</pre> |

## Differences

> [!NOTE]
> 🔵 Required  🟢 Optional 🟠 Computed  🔴 Not present

| Attribute name | Cloud Foundry Provider|  Community Cloud Foundry Provider (old) | Description |
| --- | --- | --- | --- |
| policy | 🔴 | 🔵 | |
| policies | 🔵 | 🔴 | Moved from the block-style `policy` to a list of `policies`. The individual entries maintain compatibility |
| policies.source_app | 🔵 | 🔵 | |
| policies.destination_app | 🔵 | 🔵 | |
| policies.port | 🔵 | 🔵 | |
| policies.protocol | 🟢 | 🟢 | |
