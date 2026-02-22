package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
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

		testSpaceGUIDNew      = "bb77a8bc-00f9-4cca-9df1-2e63641ff1a2"
		testServicePanGUIDNew = "75d7c3f6-e629-4db0-abf7-bc9c804d379d"
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
						Name:          new(testServiceInstanceManagedCreate),
						Type:          new(managedSerivceInstance),
						Space:         new(testSpaceGUID),
						ServicePlan:   new(testServicePanGUID),
						Parameters:    new(testParameters),
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
						Name:          new(testServiceInstanceManagedCreate),
						Type:          new(managedSerivceInstance),
						Space:         new(testSpaceGUID),
						ServicePlan:   new(testServicePanGUID),
						Parameters:    new(testParametersUpdated),
						Tags:          new(testTags),
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

	t.Run("happy path - import service instance using resource identity", func(t *testing.T) {
		resourceName := "cloudfoundry_service_instance.si"
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_service_instance_import_identity")
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
						Space:         strtostrptr(testSpaceGUIDNew),
						ServicePlan:   strtostrptr(testServicePanGUIDNew),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "name", testServiceInstanceManagedCreate),
						resource.TestCheckResourceAttr(resourceName, "type", managedSerivceInstance),
						resource.TestCheckResourceAttr(resourceName, "space", testSpaceGUIDNew),
						resource.TestCheckResourceAttr(resourceName, "service_plan", testServicePanGUIDNew),
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestMatchResourceAttr(resourceName, "updated_at", regexpValidRFC3999Format),
					),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectIdentity("cloudfoundry_service_instance.si", map[string]knownvalue.Check{
							"service_instance_guid": knownvalue.NotNull(),
						}),
					},
				},
				{
					ResourceName:    resourceName,
					ImportState:     true,
					ImportStateKind: resource.ImportBlockWithResourceIdentity,
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
						Name:          new(testServiceInstanceUserProvidedCreate),
						Type:          new(userProvidedServiceInstance),
						Space:         new(testSpaceGUID),
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
						Name:          new(testServiceInstanceUserProvidedUpdate),
						Type:          new(userProvidedServiceInstance),
						Space:         new(testSpaceGUID),
						Credentials:   new(testCredentials),
						Labels:        new(testUpdateLabel),
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
						Name:          new("test-si-wrong-service-plan"),
						Type:          new(managedSerivceInstance),
						Space:         new(testSpaceGUID),
						ServicePlan:   new(invalidOrgGUID),
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
						Name:          new("test-si-wrong-space"),
						Type:          new(managedSerivceInstance),
						Space:         new(invalidOrgGUID),
						ServicePlan:   new(testServicePanGUID),
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
						Name:          new("test-si-wrong-credentials"),
						Type:          new(userProvidedServiceInstance),
						Space:         new(testSpaceGUID),
						Credentials:   new(testInvalidCredentials),
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
						Name:          new("tf-test-do-not-delete-managed"),
						Space:         new(testSpaceGUID),
						Type:          new(managedSerivceInstance),
						ServicePlan:   new(testServicePanGUID),
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
						Name:          new("tf-test-do-not-delete"),
						Space:         new(testSpaceGUID),
						Type:          new(userProvidedServiceInstance),
					}),
					ExpectError: regexp.MustCompile(`Error: API Error in creating user-provided service instance`),
				},
			},
		})
	})

}
