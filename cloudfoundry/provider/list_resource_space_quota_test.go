package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/querycheck/queryfilter"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

const testSpaceQuotaOrgGUID = "7c8b9705-bc7c-4a96-9af0-6674c8290615"

func TestSpaceQuotaListResource(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/list_resource_space_quota")
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
					Config: hclProvider(nil) + listSpaceQuotaQueryConfig(
						"space_quota_list",
						"cloudfoundry",
						testSpaceQuotaOrgGUID,
					),
					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("cloudfoundry_space_quota.space_quota_list", 1),
						querycheck.ExpectIdentity(
							"cloudfoundry_space_quota.space_quota_list",
							map[string]knownvalue.Check{
								"space_quota_guid": knownvalue.StringRegexp(regexpValidUUID),
							},
						),
					},
				},
				{
					Query: true,
					Config: hclProvider(nil) + listSpaceQuotaQueryConfigWithIncludeResource(
						"space_quota_list",
						"cloudfoundry",
						testSpaceQuotaOrgGUID,
					),
					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("cloudfoundry_space_quota.space_quota_list", 1),
						querycheck.ExpectResourceKnownValues(
							"cloudfoundry_space_quota.space_quota_list",
							queryfilter.ByResourceIdentity(map[string]knownvalue.Check{
								"space_quota_guid": knownvalue.StringExact("92b11295-db0f-4b77-80d7-239df17ebd45"),
							}),
							[]querycheck.KnownValueCheck{
								{
									Path:       tfjsonpath.New("org"),
									KnownValue: knownvalue.StringExact(testSpaceQuotaOrgGUID),
								},
								{
									Path:       tfjsonpath.New("id"),
									KnownValue: knownvalue.StringRegexp(regexpValidUUID),
								},
								{
									Path:       tfjsonpath.New("total_routes"),
									KnownValue: knownvalue.Int64Exact(50),
								},
								{
									Path:       tfjsonpath.New("total_memory"),
									KnownValue: knownvalue.Int64Exact(10240),
								},
							},
						),
					},
				},
			},
		})
	})
}

func listSpaceQuotaQueryConfig(label, providerName, orgGUID string) string {
	return fmt.Sprintf(`
list "cloudfoundry_space_quota" "%s" {
  provider = %s
  config {
    org = "%s"
  }
}`, label, providerName, orgGUID)
}

func listSpaceQuotaQueryConfigWithIncludeResource(label, providerName, orgGUID string) string {
	return fmt.Sprintf(`
list "cloudfoundry_space_quota" "%s" {
  provider         = %s
  include_resource = true
  config {
    org = "%s"
  }
}`, label, providerName, orgGUID)
}
