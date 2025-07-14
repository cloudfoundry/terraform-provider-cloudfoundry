package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type RemoteFileHashDataSourceType struct {
	Url types.String `tfsdk:"url"`
	Id  types.String `tfsdk:"id"`
}
