package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
