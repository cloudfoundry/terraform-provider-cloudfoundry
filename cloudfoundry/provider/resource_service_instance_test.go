package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestResourceServiceInstance(t *testing.T) {
	var (
		// in canary -> PerformanceTeamBLR -> tf-space-1
		testSpaceGUID                         = "02c0cc92-6ecc-44b1-b7b2-096ca19ee143"
		testServiceInstanceManagedCreate      = "test-si-managed"
		testServiceInstanceUserProvidedCreate = "test-si-user-provided"
		testServiceInstanceUserProvidedUpdate = "test-si-user-provided1"
		// canary --> XSUAA --> application
		testServicePanGUID     = "432bd9db-20e2-4997-825f-e4a937705b87"
		testParameters         = `{"xsappname":"tf-unit-test","tenant-mode":"dedicated","description":"tf test1","foreign-scope-references":["user_attributes"],"scopes":[{"name":"uaa.user","description":"UAA"}],"role-templates":[{"name":"Token_Exchange","description":"UAA","scope-references":["uaa.user"]}]}`
		testParametersUpdated  = `{"xsappname":"tf-unit-test","tenant-mode":"dedicated","description":"tf test1-update","foreign-scope-references":["user_attributes"],"scopes":[{"name":"uaa.user","description":"UAA"}],"role-templates":[{"name":"Token_Exchange","description":"UAA","scope-references":["uaa.user"]}]}`
		testTags               = `["test-tag"]`
		testCredentials        = `{"user" : "test","password": "hello"}`
		testInvalidCredentials = `{"hello"}`
	)
	t.Parallel()
	t.Run("happy path - create service instance managed", func(t *testing.T) {
		resourceName := "cloudfoundry_service_instance.si"
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_service_instance_managed")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclServiceInstance(&ServiceInstanceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "si",
						Name:          strtostrptr(testServiceInstanceManagedCreate),
						Type:          strtostrptr(managedSerivceInstance),
						Space:         strtostrptr(testSpaceGUID),
						ServicePlan:   strtostrptr(testServicePanGUID),
						Parameters:    strtostrptr(testParameters),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "name", testServiceInstanceManagedCreate),
						resource.TestCheckResourceAttr(resourceName, "type", managedSerivceInstance),
						resource.TestCheckResourceAttr(resourceName, "space", testSpaceGUID),
						resource.TestCheckResourceAttr(resourceName, "service_plan", testServicePanGUID),
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestMatchResourceAttr(resourceName, "updated_at", regexpValidRFC3999Format),
					),
				},
				{
					Config: hclProvider(nil) + hclServiceInstance(&ServiceInstanceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "si",
						Name:          strtostrptr(testServiceInstanceManagedCreate),
						Type:          strtostrptr(managedSerivceInstance),
						Space:         strtostrptr(testSpaceGUID),
						ServicePlan:   strtostrptr(testServicePanGUID),
						Parameters:    strtostrptr(testParametersUpdated),
						Tags:          strtostrptr(testTags),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "name", testServiceInstanceManagedCreate),
						resource.TestCheckResourceAttr(resourceName, "type", managedSerivceInstance),
						resource.TestCheckResourceAttr(resourceName, "space", testSpaceGUID),
						resource.TestCheckResourceAttr(resourceName, "service_plan", testServicePanGUID),
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(resourceName, "parameters", regexp.MustCompile(`"tf test1-update"`)),
						resource.TestCheckResourceAttr(resourceName, "tags.0", "test-tag"),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestMatchResourceAttr(resourceName, "updated_at", regexpValidRFC3999Format),
					),
				},
				{
					ResourceName:            resourceName,
					ImportStateIdFunc:       getIdForImport(resourceName),
					ImportState:             true,
					ImportStateVerifyIgnore: []string{"parameters"},
					ImportStateVerify:       true,
				},
			},
		})
	})
	t.Run("happy path - create service instance user provided", func(t *testing.T) {
		resourceName := "cloudfoundry_service_instance.si_user_provided"
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_service_instance_user_provided")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclServiceInstance(&ServiceInstanceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "si_user_provided",
						Name:          strtostrptr(testServiceInstanceUserProvidedCreate),
						Type:          strtostrptr(userProvidedServiceInstance),
						Space:         strtostrptr(testSpaceGUID),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "name", testServiceInstanceUserProvidedCreate),
						resource.TestCheckResourceAttr(resourceName, "type", userProvidedServiceInstance),
						resource.TestCheckResourceAttr(resourceName, "space", testSpaceGUID),
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestMatchResourceAttr(resourceName, "updated_at", regexpValidRFC3999Format),
					),
				},
				{
					Config: hclProvider(nil) + hclServiceInstance(&ServiceInstanceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "si_user_provided",
						Name:          strtostrptr(testServiceInstanceUserProvidedUpdate),
						Type:          strtostrptr(userProvidedServiceInstance),
						Space:         strtostrptr(testSpaceGUID),
						Credentials:   strtostrptr(testCredentials),
						Labels:        strtostrptr(testUpdateLabel),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "name", testServiceInstanceUserProvidedUpdate),
						resource.TestCheckResourceAttr(resourceName, "type", userProvidedServiceInstance),
						resource.TestCheckResourceAttr(resourceName, "space", testSpaceGUID),
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestMatchResourceAttr(resourceName, "updated_at", regexpValidRFC3999Format),
						resource.TestMatchResourceAttr(resourceName, "credentials", regexp.MustCompile(`"password"`)),
						resource.TestCheckResourceAttr(resourceName, "labels.purpose", "production"),
						resource.TestCheckResourceAttr(resourceName, "space", testSpaceGUID),
					),
				},
				{
					ResourceName:            resourceName,
					ImportStateIdFunc:       getIdForImport(resourceName),
					ImportState:             true,
					ImportStateVerifyIgnore: []string{"credentials"},
					ImportStateVerify:       true,
				},
			},
		})
	})
	t.Run("error path - create service instance with invalid service plan", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_service_instance_invalid_service_plan")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclServiceInstance(&ServiceInstanceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "si_wrong_service_plan",
						Name:          strtostrptr("test-si-wrong-service-plan"),
						Type:          strtostrptr(managedSerivceInstance),
						Space:         strtostrptr(testSpaceGUID),
						ServicePlan:   strtostrptr(invalidOrgGUID),
					}),
					ExpectError: regexp.MustCompile(`Invalid service plan`),
				},
			},
		})
	})
	t.Run("error path - create service instance with invalid space", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_service_instance_invalid_space")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclServiceInstance(&ServiceInstanceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "si_wrong_space",
						Name:          strtostrptr("test-si-wrong-space"),
						Type:          strtostrptr(managedSerivceInstance),
						Space:         strtostrptr(invalidOrgGUID),
						ServicePlan:   strtostrptr(testServicePanGUID),
					}),
					ExpectError: regexp.MustCompile(`Invalid space`),
				},
			},
		})
	})
	t.Run("error path - create service instance with invalid credentials", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_service_instance_invalid_credentials")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclServiceInstance(&ServiceInstanceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "si_wrong_credentials",
						Name:          strtostrptr("test-si-wrong-credentials"),
						Type:          strtostrptr(userProvidedServiceInstance),
						Space:         strtostrptr(testSpaceGUID),
						Credentials:   strtostrptr(testInvalidCredentials),
					}),
					ExpectError: regexp.MustCompile(`Error: Invalid JSON String Value`),
				},
			},
		})
	})
	t.Run("error path - create service instance managed with already existing name ", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_service_instance_managed_exists_already")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclServiceInstance(&ServiceInstanceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "si_managed_already_exists",
						Name:          strtostrptr("tf-test-do-not-delete-managed"),
						Space:         strtostrptr(testSpaceGUID),
						Type:          strtostrptr(managedSerivceInstance),
						ServicePlan:   strtostrptr(testServicePanGUID),
					}),
					ExpectError: regexp.MustCompile(`Error: API Error in creating managed service instance`),
				},
			},
		})
	})
	t.Run("error path - create service instance user provided with already existing name ", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_service_instance_user_provided_exists_already")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclServiceInstance(&ServiceInstanceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "si_user_provided_already_exists",
						Name:          strtostrptr("tf-test-do-not-delete"),
						Space:         strtostrptr(testSpaceGUID),
						Type:          strtostrptr(userProvidedServiceInstance),
					}),
					ExpectError: regexp.MustCompile(`Error: API Error in creating user-provided service instance`),
				},
			},
		})
	})

	// This test verifies that updating a space's allow_ssh attribute does not cause
	// service instances in that space to be replaced. This is a regression test for
	// an issue where changing space attributes caused the service instance's space
	// reference to appear as "known after apply", triggering unwanted replacement.
	// The service instance references cloudfoundry_space.test.id so that when the
	// space is updated, the space ID flows through as "known after apply" during
	// planning. The UseStateForUnknown plan modifier prevents this from triggering
	// a replacement.
	t.Run("happy path - service instance not replaced when space allow_ssh changes", func(t *testing.T) {
		resourceName := "cloudfoundry_service_instance.si_space_update"
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_service_instance_space_allow_ssh_update")
		defer stopQuietly(rec)

		var serviceInstanceID string

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					// Step 1: Create a space with allow_ssh=false and a user-provided
					// service instance that references cloudfoundry_space.test.id
					// (using user-provided since it doesn't require waiting for provisioning)
					Config: hclProvider(nil) + `
resource "cloudfoundry_space" "test" {
	name      = "test-space-stability"
	org       = "` + testOrgGUID + `"
	allow_ssh = false
}
resource "cloudfoundry_service_instance" "si_space_update" {
	name        = "test-si-space-update"
	type        = "user-provided"
	space       = cloudfoundry_space.test.id
	credentials = ` + testCredentials + `
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						func(s *terraform.State) error {
							rs, ok := s.RootModule().Resources[resourceName]
							if !ok {
								return fmt.Errorf("resource not found: %s", resourceName)
							}
							serviceInstanceID = rs.Primary.ID
							return nil
						},
					),
				},
				{
					// Step 2: Change allow_ssh on the space. This causes the space
					// resource to be updated, making cloudfoundry_space.test.id
					// appear as "known after apply" during planning. Without
					// UseStateForUnknown on the service instance's space attribute,
					// this would trigger a replacement.
					Config: hclProvider(nil) + `
resource "cloudfoundry_space" "test" {
	name      = "test-space-stability"
	org       = "` + testOrgGUID + `"
	allow_ssh = true
}
resource "cloudfoundry_service_instance" "si_space_update" {
	name        = "test-si-space-update"
	type        = "user-provided"
	space       = cloudfoundry_space.test.id
	credentials = ` + testCredentials + `
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						func(s *terraform.State) error {
							rs, ok := s.RootModule().Resources[resourceName]
							if !ok {
								return fmt.Errorf("resource not found: %s", resourceName)
							}
							if rs.Primary.ID != serviceInstanceID {
								return fmt.Errorf("service instance was unexpectedly replaced: old ID %s, new ID %s", serviceInstanceID, rs.Primary.ID)
							}
							return nil
						},
					),
				},
			},
		})
	})

}
