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

const testServiceCredentialBindingListServiceInstanceGUID = "c02f9dec-f627-4b89-9542-b04edbedd95a"

func TestServiceCredentialBindingListResource(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/list_resource_service_credential_binding")
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
					Config: hclProvider(nil) + listServiceCredentialBindingQueryConfig(
						"binding_list",
						"cloudfoundry",
						testServiceCredentialBindingListServiceInstanceGUID,
					),
					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("cloudfoundry_service_credential_binding.binding_list", 1),
						querycheck.ExpectIdentity(
							"cloudfoundry_service_credential_binding.binding_list",
							map[string]knownvalue.Check{
								"service_credential_binding_guid": knownvalue.StringRegexp(regexpValidUUID),
							},
						),
					},
				},
				{
					// List Query with include_resource = true
					Query: true,
					Config: hclProvider(nil) + listServiceCredentialBindingQueryConfigWithIncludeResource(
						"binding_list",
						"cloudfoundry",
						testServiceCredentialBindingListServiceInstanceGUID,
					),
					QueryResultChecks: []querycheck.QueryResultCheck{
						querycheck.ExpectLength("cloudfoundry_service_credential_binding.binding_list", 1),
						querycheck.ExpectResourceKnownValues(
							"cloudfoundry_service_credential_binding.binding_list",
							queryfilter.ByResourceIdentity(map[string]knownvalue.Check{
								"service_credential_binding_guid": knownvalue.StringExact("6ed8d7e1-4464-46d6-960d-08462d66aac4"),
							}),
							[]querycheck.KnownValueCheck{
								{
									Path:       tfjsonpath.New("type"),
									KnownValue: knownvalue.StringExact("app"),
								},
								{
									Path:       tfjsonpath.New("service_instance"),
									KnownValue: knownvalue.StringExact(testServiceCredentialBindingListServiceInstanceGUID),
								},
							},
						),
					},
				},
			},
		})
	})
}

func listServiceCredentialBindingQueryConfig(label, providerName, serviceInstanceGUID string) string {
	return fmt.Sprintf(`
list "cloudfoundry_service_credential_binding" "%s" {
  provider = %s
  config {
    service_instance = "%s"
  }
}`, label, providerName, serviceInstanceGUID)
}

func listServiceCredentialBindingQueryConfigWithIncludeResource(label, providerName, serviceInstanceGUID string) string {
	return fmt.Sprintf(`
list "cloudfoundry_service_credential_binding" "%s" {
  provider         = %s
  include_resource = true
  config {
    service_instance = "%s"
  }
}`, label, providerName, serviceInstanceGUID)
}
