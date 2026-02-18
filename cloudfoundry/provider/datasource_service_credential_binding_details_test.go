package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type DatasourceServiceCredentialBindingDetailsModelPtr struct {
	HclType         string
	HclObjectName   string
	Name            *string
	App             *string
	ServiceInstance *string
	Type            *string
}

func hclDatasourceServiceCredentialBindingDetails(sip *DatasourceServiceCredentialBindingDetailsModelPtr) string {
	if sip != nil {
		s := `
		{{.HclType}} "cloudfoundry_service_credential_binding_details" {{.HclObjectName}} {
			{{- if .Name}}
				name  = "{{.Name}}"
			{{- end -}}
			{{if .App}}
				app = "{{.App}}"
			{{- end }}
			{{if .ServiceInstance}}
				service_instance = "{{.ServiceInstance}}"
			{{- end }}
			{{if .Type}}
				type = "{{.Type}}"
			{{- end }}
		}`
		tmpl, err := template.New("service_credential_binding_details").Parse(s)
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
	return sip.HclType + ` "cloudfoundry_service_credential_binding_details"  "` + sip.HclObjectName + ` {}`
}

func TestServiceCredentialBindingDetailsDataSource(t *testing.T) {
	var (
		// in canary -> PerformanceTeamBLR -> tf-space-1
		testManagedServiceInstanceGUID      = "e9ec29ca-993d-42e2-9c5b-cb17b1972cce"
		testUserProvidedServiceInstanceGUID = "5e2976bb-332e-41e1-8be3-53baafea9296"
		testAppGUID                         = "00fcc2ff-7377-4bdf-8217-f1f28f199a89"
		testServiceKeyName                  = "hifi"
	)
	t.Parallel()
	t.Run("happy path - read credential details of a managed service instance", func(t *testing.T) {
		cfg := getCFHomeConf()
		dataSourceName := "data.cloudfoundry_service_credential_binding_details.ds"
		rec := cfg.SetupVCR(t, "fixtures/datasource_service_credential_binding_details_managed_service")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclDatasourceServiceCredentialBindingDetails(&DatasourceServiceCredentialBindingDetailsModelPtr{
						HclType:         hclObjectDataSource,
						HclObjectName:   "ds",
						Name:            new(testServiceKeyName),
						ServiceInstance: new(testManagedServiceInstanceGUID),
						Type:            new(keyServiceCredentialBinding),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(dataSourceName, "type", keyServiceCredentialBinding),
						resource.TestCheckResourceAttr(dataSourceName, "service_instance", testManagedServiceInstanceGUID),
						resource.TestMatchResourceAttr(dataSourceName, "credential_binding", regexp.MustCompile(`"credentials"`)),
						resource.TestMatchResourceAttr(dataSourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(dataSourceName, "created_at", regexpValidRFC3999Format),
						resource.TestMatchResourceAttr(dataSourceName, "updated_at", regexpValidRFC3999Format),
					),
				},
			},
		})
	})
	t.Run("happy path - read credential details of a user-provided service instance", func(t *testing.T) {
		cfg := getCFHomeConf()
		dataSourceName := "data.cloudfoundry_service_credential_binding_details.ds"
		rec := cfg.SetupVCR(t, "fixtures/datasource_service_credential_binding_details_user_service")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclDatasourceServiceCredentialBindingDetails(&DatasourceServiceCredentialBindingDetailsModelPtr{
						HclType:         hclObjectDataSource,
						HclObjectName:   "ds",
						ServiceInstance: new(testUserProvidedServiceInstanceGUID),
						App:             new(testAppGUID),
						Type:            new(appServiceCredentialBinding),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(dataSourceName, "type", appServiceCredentialBinding),
						resource.TestCheckResourceAttr(dataSourceName, "app", testAppGUID),
						resource.TestCheckResourceAttr(dataSourceName, "service_instance", testUserProvidedServiceInstanceGUID),
						resource.TestMatchResourceAttr(dataSourceName, "credential_binding", regexp.MustCompile(`"credentials"`)),
						resource.TestMatchResourceAttr(dataSourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(dataSourceName, "created_at", regexpValidRFC3999Format),
						resource.TestMatchResourceAttr(dataSourceName, "updated_at", regexpValidRFC3999Format),
					),
				},
			},
		})
	})
	t.Run("error path - get unavailable binding", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_service_credential_binding_details_invalid_binding")
		defer stopQuietly(rec)
		resource.UnitTest(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclDatasourceServiceCredentialBindingDetails(&DatasourceServiceCredentialBindingDetailsModelPtr{
						HclType:         hclObjectDataSource,
						HclObjectName:   "ds",
						ServiceInstance: new(testUserProvidedServiceInstanceGUID),
						Name:            new("invalid-binding-name"),
						Type:            new(keyServiceCredentialBinding),
					}),
					ExpectError: regexp.MustCompile(`Request failed with expected exactly 1 result, but got less or more than 1.`),
				},
			},
		})
	})
}
