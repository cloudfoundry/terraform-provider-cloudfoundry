package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/list"
	res "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/querycheck/queryfilter"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

const testOrgRoleOrgGUID = "7c8b9705-bc7c-4a96-9af0-6674c8290615"

func TestOrgRoleListResource(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/list_resource_org_role")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			TerraformVersionChecks: []tfversion.TerraformVersionCheck{
				tfversion.SkipBelow(tfversion.Version1_14_0),
			},
			Steps: []resource.TestStep{
				{
					Query: true,
					Config: hclProvider(nil) + listOrgRoleQueryConfig(
						"org_role_list",
						"cloudfoundry",
						testOrgRoleOrgGUID,
					),
					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("cloudfoundry_org_role.org_role_list", 1),
						querycheck.ExpectIdentity(
							"cloudfoundry_org_role.org_role_list",
							map[string]knownvalue.Check{
								"role_guid": knownvalue.StringRegexp(regexpValidUUID),
							},
						),
					},
				},
				{
					Query: true,
					Config: hclProvider(nil) + listOrgRoleQueryConfigWithIncludeResource(
						"org_role_list",
						"cloudfoundry",
						testOrgRoleOrgGUID,
					),
					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("cloudfoundry_org_role.org_role_list", 1),
						querycheck.ExpectResourceKnownValues(
							"cloudfoundry_org_role.org_role_list",
							queryfilter.ByResourceIdentity(map[string]knownvalue.Check{
								"role_guid": knownvalue.StringExact("a8320596-5c60-47f9-a7ba-2083eadae14e"),
							}),
							[]querycheck.KnownValueCheck{
								{
									Path:       tfjsonpath.New("type"),
									KnownValue: knownvalue.StringExact("organization_user"),
								},
								{
									Path:       tfjsonpath.New("org"),
									KnownValue: knownvalue.StringExact(testOrgRoleOrgGUID),
								},
							},
						),
					},
				},
			},
		})
	})

	t.Run("error path - configure", func(t *testing.T) {
		r := NewOrgRoleListResource().(list.ListResourceWithConfigure)
		resp := &res.ConfigureResponse{}
		req := res.ConfigureRequest{
			ProviderData: struct{}{},
		}

		r.Configure(context.Background(), req, resp)

		if !resp.Diagnostics.HasError() {
			t.Error("Expected error for invalid provider data type")
		}
	})
}

func listOrgRoleQueryConfig(label, providerName, orgGUID string) string {
	return fmt.Sprintf(`
list "cloudfoundry_org_role" "%s" {
  provider = %s
  config {
    org = "%s"
  }
}`, label, providerName, orgGUID)
}

func listOrgRoleQueryConfigWithIncludeResource(label, providerName, orgGUID string) string {
	return fmt.Sprintf(`
list "cloudfoundry_org_role" "%s" {
  provider         = %s
  include_resource = true
  config {
    org = "%s"
  }
}`, label, providerName, orgGUID)
}
