package provider

import (
	"bytes"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type ServicePlansModelPtr struct {
	HclType             string
	HclObjectName       string
	Name                *string
	ServiceOfferingName *string
	ServiceBrokerName   *string
	ServicePlans        *string
}

func hclServicePlans(smp *ServicePlansModelPtr) string {
	if smp != nil {
		s := `
		{{.HclType}} "cloudfoundry_service_plans" {{.HclObjectName}} {
			{{- if .Name}}
				name = "{{.Name}}"
			{{- end -}}
			{{if .ServiceOfferingName}}
				service_offering_name = "{{.ServiceOfferingName}}"
			{{- end -}}
			{{if .ServiceBrokerName}}
				service_broker_name = "{{.ServiceBrokerName}}"
			{{- end -}}
			{{if .ServicePlans}}
				service_plans = {{.ServicePlans}}
			{{- end }}
		}`
		tmpl, err := template.New("service").Parse(s)
		if err != nil {
			panic(err)
		}
		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, smp)
		if err != nil {
			panic(err)
		}
		return buf.String()
	}
	return smp.HclType + ` "cloudfoundry_service_plans" ` + smp.HclObjectName + ` {}`
}

func TestDatasourceServicePlans(t *testing.T) {
	t.Parallel()
	datasourceName := "data.cloudfoundry_service_plans.test"
	dataSourceName := "data.cloudfoundry_service_plans.ds"
	t.Run("error path - get unavailable service plan", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_service_plans_invalid")
		defer stopQuietly(rec)

		// Create a Terraform configuration that uses the data source
		// and run `terraform apply`. The data source should not be found.
		resource.UnitTest(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclServicePlans(&ServicePlansModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "test",
						Name:          strtostrptr("invalid-service-name"),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(dataSourceName, "service_plans.#", "0"),
					),
				},
			},
		})

	})

	t.Run("happy path - read service plans", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_service_plans")
		defer stopQuietly(rec)
		testServiceName := "xsuaa"                                // Canary service name
		testServiceGUID := "c1876449-7493-43dc-a36b-0a8215fa46ad" // Canary service GUID
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclServicePlans(&ServicePlansModelPtr{
						HclType:             hclObjectDataSource,
						HclObjectName:       "test",
						ServiceOfferingName: strtostrptr(testServiceName),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(datasourceName, "service_offering_name", testServiceName),
						resource.TestCheckResourceAttr(datasourceName, "service_plans.0.service_offering", testServiceGUID),
						resource.TestCheckResourceAttr(datasourceName, "service_plans.#", "7"),
					),
				},
			},
		})
	})

}
