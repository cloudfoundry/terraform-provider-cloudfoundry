# Application Security Group Spaces


## Resource

Application Security Group `asg`  is now identified as `security_group`. The newer resource also exposes some additional parameters that are introduced as part of the v3 specification.


|  SAP Cloudfoundry Provider |Community Cloudfoundry Provider |
| -- | -- |
|  <pre>resource "cloudfoundry_security_group_spaces" "space_bindings1" {</br>  security_group = "38109de6-8059-41dd-b9fa-d802d8a35271"</br>  running_spaces = ["e4ccb84e-5d8b-4ca2-b59a-012f4cf45c5d"]</br>}</br></br>resource "cloudfoundry_security_group_spaces" "space_bindings2" {</br>  security_group = "531dd667-0fcf-44a0-8c6a-a541a062750d"</br>  staging_spaces = ["e4ccb84e-5d8b-4ca2-b59a-012f4cf45c5d"]</br>}</br></pre> |<pre>resource "cloudfoundry_space_asgs" "spaceasgs" {</br>    space        = "e4ccb84e-5d8b-4ca2-b59a-012f4cf45c5d"</br>    running_asgs = ["38109de6-8059-41dd-b9fa-d802d8a35271"]</br>    staging_asgs = ["531dd667-0fcf-44a0-8c6a-a541a062750d"]</br>}</br></pre> |

<br/>
