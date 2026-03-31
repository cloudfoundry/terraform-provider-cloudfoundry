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

const testSpaceRoleSpaceGUID = "10b8f9bd-0731-49f3-9e88-a17b6eec090e"

func TestSpaceRoleListResource(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/list_resource_space_role")
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
					Config: hclProvider(nil) + listSpaceRoleQueryConfig(
						"space_role_list",
						"cloudfoundry",
						testSpaceRoleSpaceGUID,
					),
					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("cloudfoundry_space_role.space_role_list", 2),
						querycheck.ExpectIdentity(
							"cloudfoundry_space_role.space_role_list",
							map[string]knownvalue.Check{
								"role_guid": knownvalue.StringRegexp(regexpValidUUID),
							},
						),
					},
				},
				{
					Query: true,
					Config: hclProvider(nil) + listSpaceRoleQueryConfigWithIncludeResource(
						"space_role_list",
						"cloudfoundry",
						testSpaceRoleSpaceGUID,
					),
					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("cloudfoundry_space_role.space_role_list", 2),
						querycheck.ExpectResourceKnownValues(
							"cloudfoundry_space_role.space_role_list",
							queryfilter.ByResourceIdentity(map[string]knownvalue.Check{
								"role_guid": knownvalue.StringExact("c3986a71-0713-4c83-8237-4120051e9339"),
							}),
							[]querycheck.KnownValueCheck{
								{
									Path:       tfjsonpath.New("space"),
									KnownValue: knownvalue.StringExact(testSpaceRoleSpaceGUID),
								},
								{
									Path:       tfjsonpath.New("type"),
									KnownValue: knownvalue.StringExact("space_developer"),
								},
							},
						),
					},
				},
			},
		})
	})
}

func listSpaceRoleQueryConfig(label, providerName, spaceGUID string) string {
	return fmt.Sprintf(`
list "cloudfoundry_space_role" "%s" {
  provider = %s
  config {
    space = "%s"
  }
}`, label, providerName, spaceGUID)
}

func listSpaceRoleQueryConfigWithIncludeResource(label, providerName, spaceGUID string) string {
	return fmt.Sprintf(`
list "cloudfoundry_space_role" "%s" {
  provider         = %s
  include_resource = true
  config {
    space = "%s"
  }
}`, label, providerName, spaceGUID)
}
