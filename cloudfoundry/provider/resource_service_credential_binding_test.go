package provider

import (
	"bytes"
	"fmt"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

type ResourceServiceCredentialBindingModelPtr struct {
	HclType         string
	HclObjectName   string
	Name            *string
	Id              *string
	Labels          *string
	Annotations     *string
	Type            *string
	App             *string
	Parameters      *string
	ServiceInstance *string
	LastOperation   *string
	CreatedAt       *string
	UpdatedAt       *string
}

func hclResourceServiceCredentialBinding(sip *ResourceServiceCredentialBindingModelPtr) string {
	if sip != nil {
		s := `
		{{.HclType}} "cloudfoundry_service_credential_binding" {{.HclObjectName}} {
			{{- if .Name}}
				name  = "{{.Name}}"
			{{- end -}}
			{{if .Id}}
				id = "{{.Id}}"
			{{- end -}}
			{{if .Labels}}
				labels = {{.Labels}}
			{{- end -}}
			{{if .Annotations}}
				annotations = {{.Annotations}}
			{{- end -}}
			{{if .Type}}
				type = "{{.Type}}"
			{{- end -}}
			{{if .App}}
				app = "{{.App}}"
			{{- end }}
			{{if .ServiceInstance}}
				service_instance = "{{.ServiceInstance}}"
			{{- end -}}
			{{if .Parameters}}
				parameters = <<EOT
				{{.Parameters}}
				EOT
			{{- end -}}
			{{if .LastOperation}}
				last_operation = "{{.LastOperation}}"
			{{- end -}}
			{{if .CreatedAt}}
				created_at = "{{.CreatedAt}}"
			{{- end -}}
			{{if .UpdatedAt}}
				updated_at = "{{.UpdatedAt}}"
			{{- end }}
		}`
		tmpl, err := template.New("service_credential_binding").Parse(s)
		if err != nil {
			panic(err)
		}
		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, sip)
		if err != nil {
			panic(err)
		}
		return buf.String()
	}
	return sip.HclType + ` "cloudfoundry_service_credential_binding"  "` + sip.HclObjectName + ` {}`
}
func TestResourceServiceCredentialBinding(t *testing.T) {
	var (
		// in canary -> PerformanceTeamBLR -> tf-space-1
		testServiceKeyManagedCreate         = "test-sk-managed"
		testAppBindingUserCreate            = "test-ab-user-provided"
		testAppBindingManagedCreate         = "test-ab-managed-provided1"
		testParameters                      = `{"xsappname":"tf-unit-test","tenant-mode":"dedicated","description":"tf test1","foreign-scope-references":["user_attributes"],"scopes":[{"name":"uaa.user","description":"UAA"}],"role-templates":[{"name":"Token_Exchange","description":"UAA","scope-references":["uaa.user"]}]}`
		testManagedServiceInstanceGUID      = "68fea1b6-11b9-4737-ad79-74e49832533f"
		testUserProvidedServiceInstanceGUID = "5e2976bb-332e-41e1-8be3-53baafea9296"
		testAppGUID                         = "e04e63c1-6e69-4537-b9e2-95ab6f3ebfcf"
		testApp2GUID                        = "ede9258f-5b3d-4837-906f-8b4b2c4cbe58"
		testApp3GUID                        = "f3d8f5c5-1bcb-489e-87a7-3011d042e4d0"
	)
	t.Parallel()
	t.Run("happy path - create service key for managed service instance", func(t *testing.T) {
		resourceName := "cloudfoundry_service_credential_binding.si"
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_service_credential_binding_managed_service_key")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclResourceServiceCredentialBinding(&ResourceServiceCredentialBindingModelPtr{
						HclType:         hclObjectResource,
						HclObjectName:   "si",
						Name:            strtostrptr(testServiceKeyManagedCreate),
						Type:            strtostrptr(keyServiceCredentialBinding),
						ServiceInstance: strtostrptr(testManagedServiceInstanceGUID),
						Parameters:      strtostrptr(testParameters),
						Labels:          strtostrptr(testCreateLabel),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "name", testServiceKeyManagedCreate),
						resource.TestCheckResourceAttr(resourceName, "type", keyServiceCredentialBinding),
						resource.TestCheckResourceAttr(resourceName, "service_instance", testManagedServiceInstanceGUID),
						resource.TestMatchResourceAttr(resourceName, "parameters", regexp.MustCompile(`"tf test1"`)),
						resource.TestCheckResourceAttr(resourceName, "labels.purpose", "testing"),
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestMatchResourceAttr(resourceName, "updated_at", regexpValidRFC3999Format),
					),
				},
				{
					Config: hclProvider(nil) + hclResourceServiceCredentialBinding(&ResourceServiceCredentialBindingModelPtr{
						HclType:         hclObjectResource,
						HclObjectName:   "si",
						Name:            strtostrptr(testServiceKeyManagedCreate),
						Type:            strtostrptr(keyServiceCredentialBinding),
						ServiceInstance: strtostrptr(testManagedServiceInstanceGUID),
						Parameters:      strtostrptr(testParameters),
						Labels:          strtostrptr(testUpdateLabel),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "name", testServiceKeyManagedCreate),
						resource.TestCheckResourceAttr(resourceName, "type", keyServiceCredentialBinding),
						resource.TestCheckResourceAttr(resourceName, "service_instance", testManagedServiceInstanceGUID),
						resource.TestMatchResourceAttr(resourceName, "parameters", regexp.MustCompile(`"tf test1"`)),
						resource.TestCheckResourceAttr(resourceName, "labels.purpose", "production"),
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
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
	t.Run("happy path - create app binding for user-provided service instance", func(t *testing.T) {
		resourceName := "cloudfoundry_service_credential_binding.si_user_provided"
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_service_credential_binding_user_app_binding")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclResourceServiceCredentialBinding(&ResourceServiceCredentialBindingModelPtr{
						HclType:         hclObjectResource,
						HclObjectName:   "si_user_provided",
						Name:            strtostrptr(testAppBindingUserCreate),
						Type:            strtostrptr(appServiceCredentialBinding),
						ServiceInstance: strtostrptr(testUserProvidedServiceInstanceGUID),
						App:             strtostrptr(testApp3GUID),
						Labels:          strtostrptr(testCreateLabel),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "name", testAppBindingUserCreate),
						resource.TestCheckResourceAttr(resourceName, "type", appServiceCredentialBinding),
						resource.TestCheckResourceAttr(resourceName, "service_instance", testUserProvidedServiceInstanceGUID),
						resource.TestCheckResourceAttr(resourceName, "labels.purpose", "testing"),
						resource.TestCheckResourceAttr(resourceName, "app", testApp3GUID),
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestMatchResourceAttr(resourceName, "updated_at", regexpValidRFC3999Format),
					),
				},
				{
					Config: hclProvider(nil) + hclResourceServiceCredentialBinding(&ResourceServiceCredentialBindingModelPtr{
						HclType:         hclObjectResource,
						HclObjectName:   "si_user_provided",
						Name:            strtostrptr(testAppBindingUserCreate),
						Type:            strtostrptr(appServiceCredentialBinding),
						ServiceInstance: strtostrptr(testUserProvidedServiceInstanceGUID),
						App:             strtostrptr(testApp3GUID),
						Labels:          strtostrptr(testUpdateLabel),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "name", testAppBindingUserCreate),
						resource.TestCheckResourceAttr(resourceName, "type", appServiceCredentialBinding),
						resource.TestCheckResourceAttr(resourceName, "service_instance", testUserProvidedServiceInstanceGUID),
						resource.TestCheckResourceAttr(resourceName, "labels.purpose", "production"),
						resource.TestCheckResourceAttr(resourceName, "app", testApp3GUID),
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestMatchResourceAttr(resourceName, "updated_at", regexpValidRFC3999Format),
					),
				},
				{
					ResourceName:      resourceName,
					ImportStateIdFunc: getIdForImport(resourceName),
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})

	t.Run("happy path - create app binding for managed service instance", func(t *testing.T) {
		resourceName := "cloudfoundry_service_credential_binding.si_managed"
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_service_credential_binding_managed_app_binding")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclResourceServiceCredentialBinding(&ResourceServiceCredentialBindingModelPtr{
						HclType:         hclObjectResource,
						HclObjectName:   "si_managed",
						Name:            strtostrptr(testAppBindingManagedCreate),
						Type:            strtostrptr(appServiceCredentialBinding),
						ServiceInstance: strtostrptr(testManagedServiceInstanceGUID),
						App:             strtostrptr(testApp2GUID),
						Parameters:      strtostrptr(testParameters),
						Labels:          strtostrptr(testCreateLabel),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "name", testAppBindingManagedCreate),
						resource.TestCheckResourceAttr(resourceName, "type", appServiceCredentialBinding),
						resource.TestCheckResourceAttr(resourceName, "service_instance", testManagedServiceInstanceGUID),
						resource.TestMatchResourceAttr(resourceName, "parameters", regexp.MustCompile(`"tf test1"`)),
						resource.TestCheckResourceAttr(resourceName, "labels.purpose", "testing"),
						resource.TestCheckResourceAttr(resourceName, "app", testApp2GUID),
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestMatchResourceAttr(resourceName, "updated_at", regexpValidRFC3999Format),
					),
				},
				{
					Config: hclProvider(nil) + hclResourceServiceCredentialBinding(&ResourceServiceCredentialBindingModelPtr{
						HclType:         hclObjectResource,
						HclObjectName:   "si_managed",
						Name:            strtostrptr(testAppBindingManagedCreate),
						Type:            strtostrptr(appServiceCredentialBinding),
						ServiceInstance: strtostrptr(testManagedServiceInstanceGUID),
						App:             strtostrptr(testApp2GUID),
						Parameters:      strtostrptr(testParameters),
						Labels:          strtostrptr(testUpdateLabel),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "name", testAppBindingManagedCreate),
						resource.TestCheckResourceAttr(resourceName, "type", appServiceCredentialBinding),
						resource.TestCheckResourceAttr(resourceName, "service_instance", testManagedServiceInstanceGUID),
						resource.TestMatchResourceAttr(resourceName, "parameters", regexp.MustCompile(`"tf test1"`)),
						resource.TestCheckResourceAttr(resourceName, "labels.purpose", "production"),
						resource.TestCheckResourceAttr(resourceName, "app", testApp2GUID),
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
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

	// This test verifies that updating a space's allow_ssh attribute does not cause
	// credential bindings to be replaced through a cascading dependency chain:
	// space -> service_instance -> credential_binding. When the space is updated,
	// cloudfoundry_space.test.id becomes "known after apply", which flows into the
	// service instance's space attribute, making cloudfoundry_service_instance.test.id
	// also "known after apply", which then flows into the credential binding's
	// service_instance attribute. UseStateForUnknown at each level prevents the cascade
	// from triggering replacements.
	t.Run("happy path - credential binding not replaced when space allow_ssh changes", func(t *testing.T) {
		resourceName := "cloudfoundry_service_credential_binding.si_stability"
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_service_credential_binding_space_allow_ssh_update")
		defer stopQuietly(rec)

		var bindingID string

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					// Step 1: Create the full dependency chain:
					// space -> service_instance -> credential_binding
					Config: hclProvider(nil) + `
resource "cloudfoundry_space" "test" {
	name      = "test-space-binding-stability"
	org       = "` + testOrgGUID + `"
	allow_ssh = false
}
resource "cloudfoundry_service_instance" "test" {
	name  = "test-si-binding-stability"
	type  = "user-provided"
	space = cloudfoundry_space.test.id
}
resource "cloudfoundry_service_credential_binding" "si_stability" {
	name             = "test-binding-stability"
	type             = "app"
	service_instance = cloudfoundry_service_instance.test.id
	app              = "` + testApp3GUID + `"
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						func(s *terraform.State) error {
							rs, ok := s.RootModule().Resources[resourceName]
							if !ok {
								return fmt.Errorf("resource not found: %s", resourceName)
							}
							bindingID = rs.Primary.ID
							return nil
						},
					),
				},
				{
					// Step 2: Change allow_ssh on the space. This causes a cascade:
					// space updated -> space.id "known after apply" ->
					// service_instance.space "known after apply" ->
					// service_instance.id "known after apply" ->
					// credential_binding.service_instance "known after apply".
					// Without UseStateForUnknown at each level, this cascade
					// would trigger replacements.
					Config: hclProvider(nil) + `
resource "cloudfoundry_space" "test" {
	name      = "test-space-binding-stability"
	org       = "` + testOrgGUID + `"
	allow_ssh = true
}
resource "cloudfoundry_service_instance" "test" {
	name  = "test-si-binding-stability"
	type  = "user-provided"
	space = cloudfoundry_space.test.id
}
resource "cloudfoundry_service_credential_binding" "si_stability" {
	name             = "test-binding-stability"
	type             = "app"
	service_instance = cloudfoundry_service_instance.test.id
	app              = "` + testApp3GUID + `"
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						func(s *terraform.State) error {
							rs, ok := s.RootModule().Resources[resourceName]
							if !ok {
								return fmt.Errorf("resource not found: %s", resourceName)
							}
							if rs.Primary.ID != bindingID {
								return fmt.Errorf("credential binding was unexpectedly replaced: old ID %s, new ID %s", bindingID, rs.Primary.ID)
							}
							return nil
						},
					),
				},
			},
		})
	})

	t.Run("error path - create app binding with existing name", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_service_credential_binding_invalid_name")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclResourceServiceCredentialBinding(&ResourceServiceCredentialBindingModelPtr{
						HclType:         hclObjectResource,
						HclObjectName:   "si_binding_already_exists",
						Name:            strtostrptr("test"),
						ServiceInstance: strtostrptr(testUserProvidedServiceInstanceGUID),
						Type:            strtostrptr(appServiceCredentialBinding),
						App:             strtostrptr(testAppGUID),
					}),
					ExpectError: regexp.MustCompile(`API Error in creating service Credential Binding`),
				},
			},
		})
	})
	t.Run("error path - create service key for user-provided service instance", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_service_credential_binding_invalid_service_key")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclResourceServiceCredentialBinding(&ResourceServiceCredentialBindingModelPtr{
						HclType:         hclObjectResource,
						HclObjectName:   "service_key",
						Name:            strtostrptr("tf-test-do-not-delete"),
						ServiceInstance: strtostrptr(testUserProvidedServiceInstanceGUID),
						Type:            strtostrptr(keyServiceCredentialBinding),
					}),
					ExpectError: regexp.MustCompile(`API Error in creating service Credential Binding`),
				},
			},
		})
	})

}
