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

const testServiceBrokerListSpaceGUID = "dd457c79-f7c9-4828-862b-35843d3b646d"

func TestServiceBrokerListResource(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/list_resource_service_broker")
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
					Config: hclProvider(nil) + listServiceBrokerQueryConfig(
						"broker_list",
						"cloudfoundry",
						testServiceBrokerListSpaceGUID,
					),
					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("cloudfoundry_service_broker.broker_list", 2),
						querycheck.ExpectIdentity(
							"cloudfoundry_service_broker.broker_list",
							map[string]knownvalue.Check{
								"service_broker_guid": knownvalue.StringRegexp(regexpValidUUID),
							},
						),
					},
				},
				{
					// List Query with include_resource = true
					Query: true,
					Config: hclProvider(nil) + listServiceBrokerQueryConfigWithIncludeResource(
						"broker_list",
						"cloudfoundry",
						testServiceBrokerListSpaceGUID,
					),
					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("cloudfoundry_service_broker.broker_list", 2),
						querycheck.ExpectResourceKnownValues(
							"cloudfoundry_service_broker.broker_list",
							queryfilter.ByResourceIdentity(map[string]knownvalue.Check{
								"service_broker_guid": knownvalue.StringExact("44faadbe-f56b-4ee2-bf40-e31113aa7324"),
							}),
							[]querycheck.KnownValueCheck{
								{
									Path:       tfjsonpath.New("name"),
									KnownValue: knownvalue.StringExact("url-broker"),
								},
								{
									Path:       tfjsonpath.New("url"),
									KnownValue: knownvalue.NotNull(),
								},
								{
									Path:       tfjsonpath.New("space"),
									KnownValue: knownvalue.StringExact(testServiceBrokerListSpaceGUID),
								},
							},
						),
					},
				},
			},
		})
	})
}

func listServiceBrokerQueryConfig(label, providerName, spaceGUID string) string {
	return fmt.Sprintf(`
list "cloudfoundry_service_broker" "%s" {
  provider = %s
  config {
    space = "%s"
  }
}`, label, providerName, spaceGUID)
}

func listServiceBrokerQueryConfigWithIncludeResource(label, providerName, spaceGUID string) string {
	return fmt.Sprintf(`
list "cloudfoundry_service_broker" "%s" {
  provider         = %s
  include_resource = true
  config {
    space = "%s"
  }
}`, label, providerName, spaceGUID)
}
