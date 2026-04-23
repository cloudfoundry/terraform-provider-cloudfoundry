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

const testServiceBrokerListSpaceGUID = "6243642c-232b-486b-beec-bb13771b2546"

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
								"service_broker_guid": knownvalue.StringExact("0ca93e74-0682-4e06-ac4b-949d3362a146"),
							}),
							[]querycheck.KnownValueCheck{
								{
									Path:       tfjsonpath.New("name"),
									KnownValue: knownvalue.StringExact("my-broker1"),
								},
								{
									Path:       tfjsonpath.New("url"),
									KnownValue: knownvalue.StringExact("https://simple-service-broker.apps.127-0-0-1.nip.io"),
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
