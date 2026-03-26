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

func TestOrgQuotaListResource(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/list_resource_org_quota")
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
					Config: hclProvider(nil) + listOrgQuotaQueryConfig(
						"org_quota_list",
						"cloudfoundry",
					),
					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("cloudfoundry_org_quota.org_quota_list", 1),
						querycheck.ExpectIdentity(
							"cloudfoundry_org_quota.org_quota_list",
							map[string]knownvalue.Check{
								"org_quota_guid": knownvalue.StringRegexp(regexpValidUUID),
							},
						),
					},
				},
				{
					Query: true,
					Config: hclProvider(nil) + listOrgQuotaQueryConfigWithIncludeResource(
						"org_quota_list",
						"cloudfoundry",
					),
					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("cloudfoundry_org_quota.org_quota_list", 1),
						querycheck.ExpectResourceKnownValues(
							"cloudfoundry_org_quota.org_quota_list",
							queryfilter.ByResourceIdentity(map[string]knownvalue.Check{
								"org_quota_guid": knownvalue.StringExact("16b266a3-d9b1-4c53-aab7-3526ba9aa636"),
							}),
							[]querycheck.KnownValueCheck{
								{
									Path:       tfjsonpath.New("name"),
									KnownValue: knownvalue.StringExact("default"),
								},
								{
									Path:       tfjsonpath.New("id"),
									KnownValue: knownvalue.StringRegexp(regexpValidUUID),
								},
							},
						),
					},
				},
			},
		})
	})

	t.Run("error path - configure", func(t *testing.T) {
		r := NewOrgQuotaListResource().(list.ListResourceWithConfigure)
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

func listOrgQuotaQueryConfig(label, providerName string) string {
	return fmt.Sprintf(`
list "cloudfoundry_org_quota" "%s" {
  provider = %s
}`, label, providerName)
}

func listOrgQuotaQueryConfigWithIncludeResource(label, providerName string) string {
	return fmt.Sprintf(`
list "cloudfoundry_org_quota" "%s" {
  provider         = %s
  include_resource = true
}`, label, providerName)
}
