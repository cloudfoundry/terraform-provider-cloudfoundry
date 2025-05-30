package provider

import (
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDatasourceServicePlan(t *testing.T) {

	datasourceName := "data.cloudfoundry_service_plan.test"
	endpoint := strtostrptr(os.Getenv("TEST_CF_API_URL"))
	user := strtostrptr(os.Getenv("TEST_CF_USER"))
	password := strtostrptr(os.Getenv("TEST_CF_PASSWORD"))
	origin := strtostrptr(os.Getenv("TEST_CF_ORIGIN"))
	if *endpoint == "" || *user == "" || *password == "" || *origin == "" {
		t.Logf("\nATTENTION: Using redacted user credentials since credentials not set as env \n Make sure you are not triggering a recording else test will fail")
		endpoint = redactedTestUser.Endpoint
		user = redactedTestUser.User
		password = redactedTestUser.Password
		origin = redactedTestUser.Origin
	}
	cfg := CloudFoundryProviderConfigPtr{
		Endpoint: endpoint,
		User:     user,
		Password: password,
		Origin:   origin,
	}

	t.Parallel()
	t.Run("error path - get unavailable service plan", func(t *testing.T) {

		rec := cfg.SetupVCR(t, "fixtures/datasource_service_plan_invalid")
		defer stopQuietly(rec)

		// Create a Terraform configuration that uses the data source
		// and run `terraform apply`. The data source should not be found.
		resource.UnitTest(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(&cfg) + `
data "cloudfoundry_service_plan" "test" {
	name = "invalid"
	service_offering_name = "invalid"
}`,
					ExpectError: regexp.MustCompile(`API Error fetching service plans.`),
				},
			},
		})

	})

	t.Run("happy path - read service plan", func(t *testing.T) {
		rec := cfg.SetupVCR(t, "fixtures/datasource_service_plan")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(&cfg) + `
data "cloudfoundry_service_plan" "test" {
	name = "application"
	service_offering_name = "xsuaa"
}`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(datasourceName, "name", "application"),
						resource.TestCheckResourceAttr(datasourceName, "service_offering_name", "xsuaa"),
						resource.TestCheckResourceAttr(datasourceName, "visibility_type", "organization"),
						resource.TestCheckResourceAttr(datasourceName, "free", "true"),
						resource.TestCheckResourceAttr(datasourceName, "available", "true"),
					),
				},
			},
		})
	})

	t.Run("error path - service offering name is required", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      `data "cloudfoundry_service_plan" "test" {name = "application"}`,
					ExpectError: regexp.MustCompile(`Error: Missing required argument`),
				},
			},
		})
	})

	t.Run("error path - plan name is required", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(nil),
			Steps: []resource.TestStep{
				{
					Config:      `data "cloudfoundry_service_plan" "test" {service_offering_name = "xsuaa"}`,
					ExpectError: regexp.MustCompile(`Error: Missing required argument`),
				},
			},
		})
	})
}
