# cloudfoundry_org_role 

Provides a Cloud Foundry resource for assigning org roles. Roles of `type` are assigned to the individual user. V3 api has clear well-defined separate API's for roles and users and the innumerous endpoints for different roles at both space and org level have been removed. One just needs to specify the necessary `type` now for a particular user as a query while specifying the role. For specifying multiple role types for multiple users, one can create multiple resources now.

| cloudfoundry/cloudfoundry | cloudfoundry-community/cloudfoundry |
| -- | -- |
|  <pre>resource "cloudfoundry_org_role" "om1" {</br>  for_each =  tolist(["username1","username2"])</br>  org      = "organization-id"</br>  username = each.value</br>  type     = "organization_billing_manager"</br>}</br></br>resource "cloudfoundry_org_role" "oa1" {</br>  org              = "organization-id"</br>  username = "username"</br>  type     = "organization_auditor"</br>}</br></br></pre> |<pre>resource "cloudfoundry_org_users" "ou1" {</br>  org              = "organization-id"</br>  managers         = ["user-guid"]</br>  billing_managers = ["username1","username2"]</br>  auditors         = ["user-guid", "username"]</br>}</br></br></pre> |

