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

const testUserCfListOrgGUID = "aafd584c-623a-427c-83ea-182115b5f854"
const testUserCfListSpaceGUID = "d809bfcb-3c74-4687-9740-b4fc1fa4b3d2"

func TestUserfListResource(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/list_resource_user_cf")
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
					Config: hclProvider(nil) + listUserCfQueryConfig(
						"user_list1",
						"cloudfoundry",
						testUserCfListOrgGUID,
					),
					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("cloudfoundry_user_cf.user_list1", 4),

						querycheck.ExpectIdentity(
							"cloudfoundry_user_cf.user_list1",
							map[string]knownvalue.Check{
								"user_guid": knownvalue.StringRegexp(regexpValidUUID),
							},
						),
					},
				},
				{
					// List Query with include_resource = true
					Query: true,
					Config: hclProvider(nil) + listUserCfQueryConfigWithIncludeResource(
						"user_list2",
						"cloudfoundry",
						testUserCfListOrgGUID,
					),
					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("cloudfoundry_user_cf.user_list2", 4),

						querycheck.ExpectResourceKnownValues(
							"cloudfoundry_user_cf.user_list2",
							queryfilter.ByResourceIdentity(map[string]knownvalue.Check{
								"user_guid": knownvalue.StringExact("897d153b-c944-4997-ab89-dc8042c23a3c"),
							}),
							[]querycheck.KnownValueCheck{
								{
									Path:       tfjsonpath.New("username"),
									KnownValue: knownvalue.StringExact("test@example.com"),
								},
							},
						),
					},
				},
				{
					// List Query with include_resource = true for space
					Query: true,
					Config: hclProvider(nil) + listUserCfQueryConfigSpaceWithIncludeResource(
						"user_list3",
						"cloudfoundry",
						testUserCfListSpaceGUID,
					),
					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("cloudfoundry_user_cf.user_list3", 3),

						querycheck.ExpectResourceKnownValues(
							"cloudfoundry_user_cf.user_list3",
							queryfilter.ByResourceIdentity(map[string]knownvalue.Check{
								"user_guid": knownvalue.StringExact("897d153b-c944-4997-ab89-dc8042c23a3c"),
							}),
							[]querycheck.KnownValueCheck{
								{
									Path:       tfjsonpath.New("username"),
									KnownValue: knownvalue.StringExact("test@example.com"),
								},
							},
						),
					},
				},
			},
		})
	})
}

func listUserCfQueryConfig(label, providerName, orgGUID string) string {
	return fmt.Sprintf(`
list "cloudfoundry_user_cf" "%s" {
  provider = %s
  config {
    org = "%s"
  }
}`, label, providerName, orgGUID)
}

func listUserCfQueryConfigWithIncludeResource(label, providerName, orgGUID string) string {
	return fmt.Sprintf(`
list "cloudfoundry_user_cf" "%s" {
  provider         = %s
  include_resource = true
  config {
    org = "%s"
  }
}`, label, providerName, orgGUID)
}

func listUserCfQueryConfigSpaceWithIncludeResource(label, providerName, spaceGUID string) string {
	return fmt.Sprintf(`
list "cloudfoundry_user_cf" "%s" {
  provider         = %s
  include_resource = true
  config {
    space = "%s"
  }
}`, label, providerName, spaceGUID)
}
