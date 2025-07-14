package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type DatasourceServiceCredentialBindingModelPtr struct {
	HclType         string
	HclObjectName   string
	Name            *string
	App             *string
	ServiceInstance *string
}

func hclDatasourceServiceCredentialBinding(sip *DatasourceServiceCredentialBindingModelPtr) string {
	if sip != nil {
		s := `
		{{.HclType}} "cloudfoundry_service_credential_binding" {{.HclObjectName}} {
			{{- if .Name}}
				name  = "{{.Name}}"
			{{- end -}}
			{{if .App}}
				app = "{{.App}}"
			{{- end }}
			{{if .ServiceInstance}}
				service_instance = "{{.ServiceInstance}}"
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

func TestServiceCredentialBindingDataSource(t *testing.T) {
	var (
		// in canary -> PerformanceTeamBLR -> tf-space-1
		testManagedServiceInstanceGUID      = "e9ec29ca-993d-42e2-9c5b-cb17b1972cce"
		testUserProvidedServiceInstanceGUID = "5e2976bb-332e-41e1-8be3-53baafea9296"
		testAppGUID                         = "e61536de-f863-4da3-9772-a8538e1bad3d"
		testServiceKeyName                  = "hifi"
	)
	t.Parallel()
	t.Run("happy path - read credentials of a managed service instance", func(t *testing.T) {
		cfg := getCFHomeConf()
		dataSourceName := "data.cloudfoundry_service_credential_binding.ds"
		rec := cfg.SetupVCR(t, "fixtures/datasource_service_credential_binding_managed_service")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclDatasourceServiceCredentialBinding(&DatasourceServiceCredentialBindingModelPtr{
						HclType:         hclObjectDataSource,
						HclObjectName:   "ds",
						Name:            strtostrptr(testServiceKeyName),
						ServiceInstance: strtostrptr(testManagedServiceInstanceGUID),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(dataSourceName, "credential_bindings.0.service_instance", testManagedServiceInstanceGUID),
						resource.TestMatchResourceAttr(dataSourceName, "credential_bindings.0.credential_binding", regexp.MustCompile(`"credentials"`)),
						resource.TestMatchResourceAttr(dataSourceName, "credential_bindings.0.id", regexpValidUUID),
						resource.TestMatchResourceAttr(dataSourceName, "credential_bindings.0.created_at", regexpValidRFC3999Format),
						resource.TestMatchResourceAttr(dataSourceName, "credential_bindings.0.updated_at", regexpValidRFC3999Format),
					),
				},
			},
		})
	})
	t.Run("happy path - read credentials of a user-provided service instance", func(t *testing.T) {
		cfg := getCFHomeConf()
		dataSourceName := "data.cloudfoundry_service_credential_binding.ds"
		rec := cfg.SetupVCR(t, "fixtures/datasource_service_credential_binding_user_service")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclDatasourceServiceCredentialBinding(&DatasourceServiceCredentialBindingModelPtr{
						HclType:         hclObjectDataSource,
						HclObjectName:   "ds",
						ServiceInstance: strtostrptr(testUserProvidedServiceInstanceGUID),
						App:             strtostrptr(testAppGUID),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(dataSourceName, "credential_bindings.0.type", appServiceCredentialBinding),
						resource.TestCheckResourceAttr(dataSourceName, "credential_bindings.0.app", testAppGUID),
						resource.TestCheckResourceAttr(dataSourceName, "credential_bindings.0.service_instance", testUserProvidedServiceInstanceGUID),
						resource.TestMatchResourceAttr(dataSourceName, "credential_bindings.0.credential_binding", regexp.MustCompile(`"credentials"`)),
						resource.TestMatchResourceAttr(dataSourceName, "credential_bindings.0.id", regexpValidUUID),
						resource.TestMatchResourceAttr(dataSourceName, "credential_bindings.0.created_at", regexpValidRFC3999Format),
						resource.TestMatchResourceAttr(dataSourceName, "credential_bindings.0.updated_at", regexpValidRFC3999Format),
					),
				},
			},
		})
	})
	t.Run("error path - get unavailable binding", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_service_credential_binding_invalid_binding")
		defer stopQuietly(rec)
		resource.UnitTest(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclDatasourceServiceCredentialBinding(&DatasourceServiceCredentialBindingModelPtr{
						HclType:         hclObjectDataSource,
						HclObjectName:   "ds",
						ServiceInstance: strtostrptr(invalidOrgGUID),
					}),
					ExpectError: regexp.MustCompile(`Unable to find any credential bindings in list`),
				},
			},
		})
	})
}
