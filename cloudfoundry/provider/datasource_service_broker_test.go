package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestServiceBrokerDataSource(t *testing.T) {
	t.Parallel()
	serviceBrokerName := "hi"
	t.Run("happy path - read service broker", func(t *testing.T) {
		cfg := getCFHomeConf()
		dataSourceName := "data.cloudfoundry_service_broker.ds"
		rec := cfg.SetupVCR(t, "fixtures/datasource_service_broker")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclServiceBroker(&ServiceBrokerModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Name:          &serviceBrokerName,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(dataSourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(dataSourceName, "space", regexpValidUUID),
						resource.TestCheckResourceAttr(dataSourceName, "name", serviceBrokerName),
						resource.TestMatchResourceAttr(dataSourceName, "created_at", regexpValidRFC3999Format),
						resource.TestMatchResourceAttr(dataSourceName, "updated_at", regexpValidRFC3999Format),
					),
				},
			},
		})
	})
	t.Run("error path - get unavailable service broker", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_service_broker_invalid")
		defer stopQuietly(rec)
		// Create a Terraform configuration that uses the data source
		// and run `terraform apply`. The data source should not be found.
		resource.UnitTest(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclServiceBroker(&ServiceBrokerModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Name:          new("invalid-service-instance-name"),
					}),
					ExpectError: regexp.MustCompile(`Unable to find service broker in list`),
				},
			},
		})
	})
}
