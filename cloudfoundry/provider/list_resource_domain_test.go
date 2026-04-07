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

const testDomainListOrgGUID = "7c8b9705-bc7c-4a96-9af0-6674c8290615"

func TestDomainListResource(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/list_resource_domain")
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
					Config: hclProvider(nil) + listDomainQueryConfig(
						"domain_list",
						"cloudfoundry",
						testDomainListOrgGUID,
					),
					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("cloudfoundry_domain.domain_list", 1),
						querycheck.ExpectIdentity(
							"cloudfoundry_domain.domain_list",
							map[string]knownvalue.Check{
								"domain_guid": knownvalue.StringRegexp(regexpValidUUID),
							},
						),
					},
				},
				{
					// List Query with include_resource = true
					Query: true,
					Config: hclProvider(nil) + listDomainQueryConfigWithIncludeResource(
						"domain_list",
						"cloudfoundry",
						testDomainListOrgGUID,
					),
					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("cloudfoundry_domain.domain_list", 1),
						// Verify full resource data is populated
						querycheck.ExpectResourceKnownValues(
							"cloudfoundry_domain.domain_list",
							queryfilter.ByResourceIdentity(map[string]knownvalue.Check{
								"domain_guid": knownvalue.StringExact("da0c442d-50e6-4be1-85f9-63945e1b680b"),
							}),
							[]querycheck.KnownValueCheck{
								{
									Path:       tfjsonpath.New("name"),
									KnownValue: knownvalue.StringExact("test.cfapps.stagingazure.hanavlab.ondemand.com"),
								},
								{
									Path:       tfjsonpath.New("id"),
									KnownValue: knownvalue.StringRegexp(regexpValidUUID),
								},
								{
									Path:       tfjsonpath.New("org"),
									KnownValue: knownvalue.StringExact(testDomainListOrgGUID),
								},
							},
						),
					},
				},
			},
		})
	})
}

func listDomainQueryConfig(label, providerName, orgGUID string) string {
	return fmt.Sprintf(`
list "cloudfoundry_domain" "%s" {
  provider = %s
  config {
    org = "%s"
  }
}`, label, providerName, orgGUID)
}

func listDomainQueryConfigWithIncludeResource(label, providerName, orgGUID string) string {
	return fmt.Sprintf(`
list "cloudfoundry_domain" "%s" {
  provider         = %s
  include_resource = true
  config {
    org = "%s"
  }
}`, label, providerName, orgGUID)
}
