# Migration Guide

 Although the definitions look similar the Terraform providers of the [cloudfoundry-community](https://github.com/cloudfoundry-community/terraform-provider-cloudfoundry) and [cloudfoundry](https://github.com/cloudfoundry/terraform-provider-cloudfoundry) are not the same. They differ in the definition as well as the state structure especially due to the usage if the V3 API in the newer provider.

Therefore, the providers cannot be simply interchanged. If you want to switch you must rewrite your configuration. Additionally you must import the resources to create a state consistent with the new provider.

> [!WARNING]
> Keep a backup of the existing configuration and state. Also, before one starts the migration process, please understand the differences in the structure of the resources and data sources as we support the latest V3 API's from CloudFoundry API. Refer our documentation for more details.

The newer V3 APIs have brought in changes to certain parameters of the existing resources and data sources to enhance the functionality and optimize the way the APIs behave. This impacts the parameters available in the resources i.e., some of the parameters might have been removed and newer parameters might have been added. Please refer to the [V3 API Specifications](https://v3-apidocs.cloudfoundry.org/version/3.166.0/index.html) for details.

This migration guide helps one in transitioning from the existing [community CF provider](https://github.com/cloudfoundry-community/terraform-provider-cloudfoundry) to the [current CF provider](https://github.com/cloudfoundry/terraform-provider-cloudfoundry) and states the difference between them.

## Maintaining both providers

A one shot rewrite of your entire configuration may not be practical especially when you have a large configuration and some of the resources/data sources are not (yet) available in the newer provider. In this case it makes sense to keep both the providers and iteratively rewrite the configuration moving from the old to the new provider.

We recommend the following iterative procedure for the migration:

1. Obtain details (required fields) of resource/data source from state
1. Remove the resource/data source from the state
1. Update  resource/data source in the configuration to the new providers structure.
1. Import the resource into the state file
1. Validate the new configuration

### Example

Consider the following Terraform configuration:

```terraform
terraform {
  required_providers {
    cloudfoundry = {
     source = "cloudfoundry-community/cloudfoundry"
      version = "0.53.1"
    }
  }
}

provider "cloudfoundry" { 
  api_url = "https://api.example.com"   
  username = "your-username"            
  password = "your-password"                                 
}

resource "cloudfoundry_org" "org" {
  name = "tf-test"
}

resource "cloudfoundry_space" "space" {
  name = "my-space"                     
  org  =  cloudfoundry_org.org.id                 
}

resource "cloudfoundry_app" "my-app" {
  name  = "my-app"                      
  space = cloudfoundry_space.my_space.id
  path = "/path/to/your/app"            
}
```

Let's assume that we want to migrate the resource `cloudfoundry_app` - `my-app` to the new provider. To do so we execute the steps in the following sections

#### Step 1 - Fetch the details of the resource from the state

Get details of `my-app` from the state via `terraform state show`.

```bash
terraform state show cloudfoundry_app.my-app | grep -m 1 id
```

This will return the technical id of the resource which is required for the import.

#### Step 2 - Remove the resource from the state

> [!IMPORTANT]
> Make sure that the state is backed up before manipulating it.

Remove the resource `my-app` from the Terraform state via:

```bash
terraform state rm cloudfoundry_app.my-app
```

#### Step 3 - Update the configuration

Replace the configuration of the `cloudfoundry_app` resource with the new provider and initialize your setup via `terraform init -upgrade` to fetch the new Cloud Foundry provider.

```terraform
terraform {
  required_providers {
    cloudfoundry = {
     source = "cloudfoundry-community/cloudfoundry"
      version = "0.53.1"
    }
    
    #Add new provider
    cloudfoundry-v3-new = {
      source = "cloudfoundry/cloudfoundry"
      version = "0.2.0-beta"
    }
  }
}

#Old provider remains to handle non migrated resources
provider "cloudfoundry" { 
  api_url = "https://api.example.com"   
  username = "your-username"            
  password = "your-password"            
}

#Configure the new Terraform Cloud foundry provider with alias
provider "cloudfoundry-v3-new" {  
  api_url = "https://api.example.com"   
  username = "your-username"            
  password = "your-password"      
}

resource "cloudfoundry_org" "org" {
  name = "tf-test"
}

resource "cloudfoundry_space" "space" {
  name = "my-space"                     
  org  =  cloudfoundry_org.org.id                 
}

#  --- Use provider meta-argument to differentiate the providers ---
 
resource "cloudfoundry_app" "my-app" {
  provider = cloudfoundry-v3-new
  name = "my-app"                     
  org_name  = cloudfoundry_org.org.name
  space_name = cloudfoundry_space.my_space.name               
}
```

#### Step 4 - Import the reconfigured resource

Import the state using the updated resource definition via:

```terraform
terraform import cloudfoundry_app.my-app d0348ed0-6e89-4836-80db-16479526a748
```

#### Step 5 - Validate the new configuration

After the successful import run `terraform plan` to verify. It might prompt to reapply the configuration to populate some attribute values. After that the new configuration is ready for use.

## Overview of Provider Differences

The following sections and the documents referenced within provide a detailed overview of the changes/similarities in the resources and data sources between the two providers.

> [!IMPORTANT]
> For every resource and datasource, the computed attributes `created_at` and `updated_at` have been added in the new provider in line with V3.

### Similar Resources

The below mentioned resources replicate the schema and functionality of those present in the existing community provider.

- [Isolation Segment](../docs/resources/isolation_segment.md)
- [Isolation Segment Entitlement](../docs/resources/isolation_segment_entitlement.md)

### Changed Resources

The below mentioned resources have been newly added in the current provider.

- [Multi Target Application Deployment](../docs/resources/mta.md)
- [User Groups Binding](../docs/resources/user_groups.md)

While most resources have maintained the same structure, some resources needed minor changes in schema to follow the V3 API structure. Following is a list of resources whose schema have changed.

- [Application](./resources/app.md)
- [Buildpack](./resources/buildpack.md)
- [Domain](./resources/domain.md)
- [Org Quota](./resources/org_quota.md)
- [Organisation](./resources/org.md)
- [Route](./resources/route.md)
- [Security Group](./resources/security_group.md)
- [Service Broker](./resources/service_broker.md)
- [Service Credential Binding](./resources/service_credential_binding.md)
- [Service Route Binding](./resources/service_route_binding.md)
- [Service Instance](./resources/service_instance.md)
- [Space Quota](./resources/space_quota.md)
- [Space](./resources/space.md)
- [User](./resources/user.md)

Few resources required a major change in functionality or the way the resources were created which are mentioned below.

- [Org Role](./resources/org_role.md)
- [Space Role](./resources/space_role.md)

### Changed DataSources

The below mentioned dataSources have been newly added in the current provider.

- [Applications](../docs/data-sources/apps.md)
- [Buildpacks](../docs/data-sources/buildpacks.md)
- [Domains](../docs/data-sources/domains.md)
- [Multi Target Application](../docs/data-sources/mta.md)
- [Multi Target Applications](../docs/data-sources/mtas.md)
- [Isolation Segment Entitlement](../docs/data-sources/isolation_segment_entitlement.md)
- [Isolation Segments](../docs/data-sources/isolation_segments.md)
- [Org Role](../docs/data-sources/org_role.md)
- [Org Roles](../docs/data-sources/org_roles.md)
- [Organizations](../docs/data-sources/orgs.md)
- [Routes](../docs/data-sources/routes.md)
- [Security Groups](../docs/data-sources/security_groups.md)
- [Service Broker](../docs/data-sources/service_broker.md)
- [Service Brokers](../docs/data-sources/service_brokers.md)
- [Service Instances](../docs/data-sources/service_instances.md)
- [Service Plans](../docs/data-sources/service_plans.md)
- [Service Route Binding](../docs/data-sources/service_route_binding.md)
- [Service Route Bindings](../docs/data-sources/service_route_bindings.md)
- [Space Role](../docs/data-sources/space_role.md)
- [Spaces](../docs/data-sources/spaces.md)
- [Space Roles](../docs/data-sources/space_roles.md)
- [Users](../docs/data-sources/users.md)

While most dataSources have  maintained the same structure, some dataSources needed minor changes in schema to follow the V3 API structure. Following is a list of datasources whose schema have changed

- [Application](./data-sources/app.md)
- [Domain](./data-sources/domain.md)
- [Org Quota](./data-sources/org_quota.md)
- [Organisation](./data-sources/org.md)
- [Route](./data-sources/route.md)
- [Security Group](./data-sources/security_group.md)
- [Service credential Binding](./data-sources/service_credential_binding.md)
- [Service Instance](./data-sources/service_instance.md)
- [Space Quota](./data-sources/space_quota.md)
- [Space](./data-sources/space.md)
- [Stack](./data-sources/stack.md)
- [User](./data-sources/user.md)
