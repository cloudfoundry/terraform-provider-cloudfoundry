package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type ServiceInstanceSharingType struct {
	Id                types.String `tfsdk:"id"`
	ServiceInstanceId types.String `tfsdk:"service_instance_id"`
	SpaceId           types.String `tfsdk:"space_id"`
}
