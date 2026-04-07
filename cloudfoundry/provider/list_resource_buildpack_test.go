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

func TestBuildpackListResource(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/list_resource_buildpack")
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
					Config: hclProvider(nil) + listBuildpackQueryConfig(
						"buildpack_list",
						"cloudfoundry",
					),
					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("cloudfoundry_buildpack.buildpack_list", 5),
						querycheck.ExpectIdentity(
							"cloudfoundry_buildpack.buildpack_list",
							map[string]knownvalue.Check{
								"buildpack_guid": knownvalue.StringRegexp(regexpValidUUID),
							},
						),
					},
				},
				{
					// List Query with include_resource = true
					Query: true,
					Config: hclProvider(nil) + listBuildpackQueryConfigWithIncludeResource(
						"buildpack_list",
						"cloudfoundry",
					),
					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("cloudfoundry_buildpack.buildpack_list", 5),
						// Verify full resource data is populated
						querycheck.ExpectResourceKnownValues(
							"cloudfoundry_buildpack.buildpack_list",
							queryfilter.ByResourceIdentity(map[string]knownvalue.Check{
								"buildpack_guid": knownvalue.StringExact("c4daabe6-aa72-47da-be17-c15769b0d34d"),
							}),
							[]querycheck.KnownValueCheck{
								{
									Path:       tfjsonpath.New("name"),
									KnownValue: knownvalue.StringExact("binary_buildpack"),
								},
								{
									Path:       tfjsonpath.New("id"),
									KnownValue: knownvalue.StringRegexp(regexpValidUUID),
								},
								{
									Path:       tfjsonpath.New("stack"),
									KnownValue: knownvalue.StringExact("cflinuxfs4"),
								},
							},
						),
					},
				},
			},
		})
	})

}

func listBuildpackQueryConfig(label, providerName string) string {
	return fmt.Sprintf(`
list "cloudfoundry_buildpack" "%s" {
  provider = %s
  config {
    stack = "cflinuxfs4"
  }
}`, label, providerName)
}

func listBuildpackQueryConfigWithIncludeResource(label, providerName string) string {
	return fmt.Sprintf(`
list "cloudfoundry_buildpack" "%s" {
  provider         = %s
  include_resource = true
  config {
    stack = "cflinuxfs4"
  }
}`, label, providerName)
}
