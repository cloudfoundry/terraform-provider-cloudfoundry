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

const testServiceInstanceListSpaceGUID = "10b8f9bd-0731-49f3-9e88-a17b6eec090e"

func TestServiceInstanceListResource(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/list_resource_service_instance")
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
					Config: hclProvider(nil) + listServiceInstanceQueryConfig(
						"si_list",
						"cloudfoundry",
						testServiceInstanceListSpaceGUID,
					),
					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("cloudfoundry_service_instance.si_list", 1),
						querycheck.ExpectIdentity(
							"cloudfoundry_service_instance.si_list",
							map[string]knownvalue.Check{
								"service_instance_guid": knownvalue.StringRegexp(regexpValidUUID),
							},
						),
					},
				},
				{
					// List Query with include_resource = true
					Query: true,
					Config: hclProvider(nil) + listServiceInstanceQueryConfigWithIncludeResource(
						"si_list",
						"cloudfoundry",
						testServiceInstanceListSpaceGUID,
					),
					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("cloudfoundry_service_instance.si_list", 1),
						querycheck.ExpectResourceKnownValues(
							"cloudfoundry_service_instance.si_list",
							queryfilter.ByResourceIdentity(map[string]knownvalue.Check{
								"service_instance_guid": knownvalue.StringExact("c02f9dec-f627-4b89-9542-b04edbedd95a"),
							}),
							[]querycheck.KnownValueCheck{
								{
									Path:       tfjsonpath.New("name"),
									KnownValue: knownvalue.StringExact("test-service"),
								},
								{
									Path:       tfjsonpath.New("type"),
									KnownValue: knownvalue.StringExact("managed"),
								},
								{
									Path:       tfjsonpath.New("space"),
									KnownValue: knownvalue.StringExact("10b8f9bd-0731-49f3-9e88-a17b6eec090e"),
								},
							},
						),
					},
				},
			},
		})
	})
}

func listServiceInstanceQueryConfig(label, providerName, spaceGUID string) string {
	return fmt.Sprintf(`
list "cloudfoundry_service_instance" "%s" {
  provider = %s
  config {
    space = "%s"
  }
}`, label, providerName, spaceGUID)
}

func listServiceInstanceQueryConfigWithIncludeResource(label, providerName, spaceGUID string) string {
	return fmt.Sprintf(`
list "cloudfoundry_service_instance" "%s" {
  provider         = %s
  include_resource = true
  config {
    space = "%s"
  }
}`, label, providerName, spaceGUID)
}
