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

const testSecurityGroupListSpaceGUID = "10b8f9bd-0731-49f3-9e88-a17b6eec090e"

func TestSecurityGroupListResource(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/list_resource_security_group")
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
					Config: hclProvider(nil) + listSecurityGroupQueryConfig(
						"security_group_list",
						"cloudfoundry",
					),
					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("cloudfoundry_security_group.security_group_list", 2),
						querycheck.ExpectIdentity(
							"cloudfoundry_security_group.security_group_list",
							map[string]knownvalue.Check{
								"security_group_guid": knownvalue.StringRegexp(regexpValidUUID),
							},
						),
					},
				},
				{
					// List Query with include_resource = true
					Query: true,
					Config: hclProvider(nil) + listSecurityGroupQueryConfigWithIncludeResource(
						"security_group_list",
						"cloudfoundry",
						testSecurityGroupListSpaceGUID,
					),
					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("cloudfoundry_security_group.security_group_list", 1),
						// Verify full resource data is populated
						querycheck.ExpectResourceKnownValues(
							"cloudfoundry_security_group.security_group_list",
							queryfilter.ByResourceIdentity(map[string]knownvalue.Check{
								"security_group_guid": knownvalue.StringRegexp(regexpValidUUID),
							}),
							[]querycheck.KnownValueCheck{
								{
									Path:       tfjsonpath.New("id"),
									KnownValue: knownvalue.StringRegexp(regexpValidUUID),
								},
								{
									Path:       tfjsonpath.New("name"),
									KnownValue: knownvalue.StringExact("dummy1"),
								},
							},
						),
					},
				},
			},
		})
	})
}

func listSecurityGroupQueryConfig(label, providerName string) string {
	return fmt.Sprintf(`
list "cloudfoundry_security_group" "%s" {
  provider = %s
}`, label, providerName)
}

func listSecurityGroupQueryConfigWithIncludeResource(label, providerName, spaceGUID string) string {
	return fmt.Sprintf(`
list "cloudfoundry_security_group" "%s" {
  provider         = %s
  include_resource = true
  config {
    running_space = "%s"
  }
}`, label, providerName, spaceGUID)
}
