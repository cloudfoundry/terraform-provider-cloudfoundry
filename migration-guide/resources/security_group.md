# cloudfoundry_security_group

Provides an application security group resource for Cloud Foundry. Named as [`cloudfoundry_asg`](https://github.com/cloudfoundry-community/terraform-provider-cloudfoundry/blob/main/docs/resources/asg.md) in the community provider.

| cloudfoundry/cloudfoundry | cloudfoundry-community/cloudfoundry |
| -- | -- |
|  <pre>resource "cloudfoundry_security_group" "my_security_group" {</br>  name                     = "tf-test"</br>  globally_enabled_running = false</br>  globally_enabled_staging = false</br>  rules = [{</br>    protocol    = "tcp"</br>    destination = "192.168.1.100"</br>    ports       = "1883,8883"</br>    log         = true</br>    }, {</br>    protocol    = "udp"</br>    destination = "192.168.1.100"</br>    ports       = "1883,8883"</br>    log         = false</br>    },</br>    {</br>      protocol    = "icmp"</br>      type        = 0</br>      code        = 0</br>      destination = "192.168.1.100"</br>      log         = false</br>  }]</br>  staging_spaces = ["3bc20dc4-1870-4835-8308-dda2d766e61e", "e6886bba-e263-4b52-aaf1-85d410f15fc8"]</br>  running_spaces = ["e6886bba-e263-4b52-aaf1-85d410f15fc8"]</br></br>}</br></pre> |<pre>resource "cloudfoundry_asg" "my_security_group" {</br>  name = "rmq-service"</br></br>  rule {</br>    protocol = "tcp"</br>    destination = "192.168.1.100"</br>    ports = "1883,8883"</br>    log = true</br>  }</br>  rule {</br>    protocol = "tcp"</br>    destination = "192.168.1.101"</br>    ports = "5671-5672"</br>    log = true</br>  }</br>}</br></pre> |

## Differences

> [!NOTE]  
> 🔵 Required  🟢 Optional 🟠 Computed  🔴 Not present

| Attribute name | Cloud Foundry Provider|  Community Cloud Foundry Provider (old) | Description |
| --- | --- | --- | --- |
| globally_enabled_running | 🟢 | 🔴 | Achieves functionality of [`cloudfoundry_default_asg`](https://github.com/cloudfoundry-community/terraform-provider-cloudfoundry/blob/main/docs/resources/default_asg.md) in the community provider with attribute `name=running`. |
| globally_enabled_staging | 🟢 | 🔴 | Achieves functionality of [`cloudfoundry_default_asg`](https://github.com/cloudfoundry-community/terraform-provider-cloudfoundry/blob/main/docs/resources/default_asg.md) in the community provider with attribute `name=staging`. |
| running_spaces | 🟢 | 🔴 | - |
| staging_spaces | 🟢 | 🔴 | - |
| rule | 🔴 | 🟢 | `rule` has been changed to `rules`  to maintain conformity with V3 API. |
| rules | 🟢 | 🔴 | - |
