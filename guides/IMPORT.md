# Import

## Overview

In general Terraform supports the *import* of resources into the Terraform state. You find the official documentation on how to achieve this [here](https://developer.hashicorp.com/terraform/cli/import).

The Terraform provider for Cloud Foundry supports the import of resources as well. [The documentation](https://registry.terraform.io/providers/cloudfoundry/cloudfoundry/latest/docs) of the Terraform provider for Cloud Foundry provides the necessary information on how to import a resource and which keys to use on the level of each resource.

To get a quick overview of the resources and if they support the import functionality, you can refer to the [Resource Overview](#resource-overview) section in this document.

## Resource Overview

The following list provides an overview of the resources and their support for the import functionality (state: 01.08.2025)



| Resource                                              | Import Support |
|---                                                    |---
| cloudfoundry_app                                      | Yes |
| cloudfoundry_buildpack                                | Yes |
| cloudfoundry_domain                                   | Yes |
| cloudfoundry_isolation_segment                        | Yes |
| cloudfoundry_isolation_segment_entitlement            | No  |
| cloudfoundry_mta                                      | Yes |
| cloudfoundry_network_policy                           | No  |
| cloudfoundry_org                                      | Yes |
| cloudfoundry_org_quota                                | Yes |
| cloudfoundry_org_role                                 | Yes |
| cloudfoundry_route                                    | Yes |
| cloudfoundry_security_group                           | Yes |
| cloudfoundry_security_group_space_binding             | No  |
| cloudfoundry_service_broker                           | Yes |
| cloudfoundry_service_credential_binding               | Yes |
| cloudfoundry_service_instance                         | Yes with restrictions (see [documentation](https://registry.terraform.io/providers/cloudfoundry/cloudfoundry/latest/docs/resources/service_instance#restriction)) |
| cloudfoundry_service_instance_sharing                 | No  |
| cloudfoundry_service_plan_visibility                  | Yes |
| cloudfoundry_service_route_binding                    | Yes |
| cloudfoundry_space                                    | Yes |
| cloudfoundry_space_quota                              | Yes |
| cloudfoundry_space_role                               | Yes |
| cloudfoundry_user                                     | Yes |
| cloudfoundry_user_cf                                  | Yes |
| cloudfoundry_user_cf_groups                           | No  |
