package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDomainDataSource_Configure(t *testing.T) {
	var (
		// in staging
		testDomainName = "cert.cfapps.stagingazure.hanavlab.ondemand.com"
	)
	t.Parallel()
	dataSourceName := "data.cloudfoundry_domain.ds"
	t.Run("happy path - read domain", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_domain")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclDomain(&DomainModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Name:          strtostrptr(testDomainName),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(dataSourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(dataSourceName, "created_at", regexpValidRFC3999Format),
						resource.TestMatchResourceAttr(dataSourceName, "updated_at", regexpValidRFC3999Format),
						resource.TestCheckResourceAttr(dataSourceName, "name", testDomainName),
						resource.TestCheckResourceAttr(dataSourceName, "internal", "false"),
						resource.TestCheckResourceAttr(dataSourceName, "supported_protocols.0", "http"),
						resource.TestCheckNoResourceAttr(dataSourceName, "labels"),
						resource.TestCheckNoResourceAttr(dataSourceName, "annotations"),
						resource.TestCheckNoResourceAttr(dataSourceName, "router_group"),
					),
				},
			},
		})
	})
	t.Run("error path - domain does not exist", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_domain_invalid")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclDomain(&DomainModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Name:          strtostrptr("test.com"),
					}),
					ExpectError: regexp.MustCompile(`Unable to find domain in list`),
				},
			},
		})
	})
}
