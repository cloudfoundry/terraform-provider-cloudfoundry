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

const testIsolationSegmentListOrgGUID = "7c8b9705-bc7c-4a96-9af0-6674c8290615"

func TestIsolationSegmentListResource(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/list_resource_isolation_segment")
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
					Config: hclProvider(nil) + listIsolationSegmentQueryConfig(
						"isolation_segment_list",
						"cloudfoundry",
					),
					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("cloudfoundry_isolation_segment.isolation_segment_list", 2),
						querycheck.ExpectIdentity(
							"cloudfoundry_isolation_segment.isolation_segment_list",
							map[string]knownvalue.Check{
								"segment_guid": knownvalue.StringRegexp(regexpValidUUID),
							},
						),
					},
				},
				{
					// List Query with include_resource = true
					Query: true,
					Config: hclProvider(nil) + listIsolationSegmentQueryConfigWithIncludeResource(
						"isolation_segment_list",
						"cloudfoundry",
						testIsolationSegmentListOrgGUID,
					),
					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("cloudfoundry_isolation_segment.isolation_segment_list", 1),
						// Verify full resource data is populated
						querycheck.ExpectResourceKnownValues(
							"cloudfoundry_isolation_segment.isolation_segment_list",
							queryfilter.ByResourceIdentity(map[string]knownvalue.Check{
								"segment_guid": knownvalue.StringRegexp(regexpValidUUID),
							}),
							[]querycheck.KnownValueCheck{
								{
									Path:       tfjsonpath.New("name"),
									KnownValue: knownvalue.StringExact("my-isolation-segment"),
								},
								{
									Path:       tfjsonpath.New("id"),
									KnownValue: knownvalue.StringRegexp(regexpValidUUID),
								},
							},
						),
					},
				},
			},
		})
	})
}

func listIsolationSegmentQueryConfig(label, providerName string) string {
	return fmt.Sprintf(`
list "cloudfoundry_isolation_segment" "%s" {
  provider = %s
}`, label, providerName)
}

func listIsolationSegmentQueryConfigWithIncludeResource(label, providerName, orgGUID string) string {
	return fmt.Sprintf(`
list "cloudfoundry_isolation_segment" "%s" {
  provider         = %s
  include_resource = true
  config {
    org = "%s"
  }
}`, label, providerName, orgGUID)
}
