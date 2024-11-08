# cloudfoundry_stack

Gets information on a Cloud Foundry stack.

|  Cloud Foundry Provider | Community Cloudfoundry Provider  |
| -- | -- |
| <pre>data "cloudfoundry_stack" "mystack" {</br>    name = "my_custom_stack"</br>}</br></pre>|<pre>data "cloudfoundry_stack" "mystack" {</br>    name = "my_custom_stack"</br>}</br></pre> |  

## Differences
> [!NOTE]  
> 🔵 Required  🟢 Optional 🟠 Computed  🔴 Not present

| Attribute name| Cloud Foundry Provider| Community Provider(old) |Description
|---| ---| ---| ---| 
|build_rootfs_image | 🟠|🔴|  - |
|run_rootfs_image | 🟠|🔴|  - |
|default|  🟠|🔴 | - |