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

const testMtaListSpaceGUID = "7c34d82b-5aec-4e66-b233-34d947459ad3"

func TestMtaListResource(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/list_resource_mta")
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
					Config: hclProvider(nil) + listMtaQueryConfig(
						"mta_list",
						"cloudfoundry",
						testMtaListSpaceGUID,
					),
					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("cloudfoundry_mta.mta_list", 1),
						querycheck.ExpectIdentity(
							"cloudfoundry_mta.mta_list",
							map[string]knownvalue.Check{
								"space_guid": knownvalue.StringExact(testMtaListSpaceGUID),
								"mta_id":     knownvalue.NotNull(),
								"namespace":  knownvalue.NotNull(),
							},
						),
					},
				},
				{
					// List Query with include_resource = true
					Query: true,
					Config: hclProvider(nil) + listMtaQueryConfigWithIncludeResource(
						"mta_list",
						"cloudfoundry",
						testMtaListSpaceGUID,
					),
					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("cloudfoundry_mta.mta_list", 1),
						// Verify full resource data is populated
						querycheck.ExpectResourceKnownValues(
							"cloudfoundry_mta.mta_list",
							queryfilter.ByResourceIdentity(map[string]knownvalue.Check{
								"space_guid": knownvalue.StringExact(testMtaListSpaceGUID),
								"mta_id":     knownvalue.StringExact("my-mta"),
								"namespace":  knownvalue.StringExact("test"),
							}),
							[]querycheck.KnownValueCheck{
								{
									Path:       tfjsonpath.New("space"),
									KnownValue: knownvalue.StringExact(testMtaListSpaceGUID),
								},
								{
									Path:       tfjsonpath.New("id"),
									KnownValue: knownvalue.NotNull(),
								},
							},
						),
					},
				},
			},
		})
	})
}

func listMtaQueryConfig(label, providerName, spaceGUID string) string {
	return fmt.Sprintf(`
list "cloudfoundry_mta" "%s" {
  provider = %s
  config {
    space = "%s"
  }
}`, label, providerName, spaceGUID)
}

func listMtaQueryConfigWithIncludeResource(label, providerName, spaceGUID string) string {
	return fmt.Sprintf(`
list "cloudfoundry_mta" "%s" {
  provider         = %s
  include_resource = true
  config {
    space = "%s"
  }
}`, label, providerName, spaceGUID)
}
