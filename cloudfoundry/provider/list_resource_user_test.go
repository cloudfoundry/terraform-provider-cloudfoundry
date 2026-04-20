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

const testUserListOrgGUID = "7c8b9705-bc7c-4a96-9af0-6674c8290615"

func TestUserListResource(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/list_resource_user")
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
					Config: hclProvider(nil) + listUserQueryConfig(
						"user_list",
						"cloudfoundry",
						testUserListOrgGUID,
					),
					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("cloudfoundry_user.user_list", 2),
						querycheck.ExpectIdentity(
							"cloudfoundry_user.user_list",
							map[string]knownvalue.Check{
								"user_guid": knownvalue.StringRegexp(regexpValidUUID),
							},
						),
					},
				},
				{
					// List Query with include_resource = true
					Query: true,
					Config: hclProvider(nil) + listUserQueryConfigWithIncludeResource(
						"user_list",
						"cloudfoundry",
						testUserListOrgGUID,
					),
					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("cloudfoundry_user.user_list", 2),
						// Verify full resource data is populated
						querycheck.ExpectResourceKnownValues(
							"cloudfoundry_user.user_list",
							queryfilter.ByResourceIdentity(map[string]knownvalue.Check{
								"user_guid": knownvalue.StringExact("e6962a20-ee9d-4384-b8f8-61c36af3c54d"),
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

func listUserQueryConfig(label, providerName, orgGUID string) string {
	return fmt.Sprintf(`
list "cloudfoundry_user" "%s" {
  provider = %s
  config {
    org = "%s"
  }
}`, label, providerName, orgGUID)
}

func listUserQueryConfigWithIncludeResource(label, providerName, orgGUID string) string {
	return fmt.Sprintf(`
list "cloudfoundry_user" "%s" {
  provider         = %s
  include_resource = true
  config {
    org = "%s"
  }
}`, label, providerName, orgGUID)
}
