package provider

import (
	"bytes"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type ServiceBrokersModelPtr struct {
	HclType        string
	HclObjectName  string
	Name           *string
	Space          *string
	ServiceBrokers *string
}

func hclServiceBrokers(sbmp *ServiceBrokersModelPtr) string {
	if sbmp != nil {
		s := `
		{{.HclType}} "cloudfoundry_service_brokers" {{.HclObjectName}} {
			{{- if .Name}}
				name = "{{.Name}}"
			{{- end -}}
			{{if .Space}}
				space = "{{.Space}}"
			{{- end -}}
			{{if .ServiceBrokers}}
				service_brokers = "{{.ServiceBrokers}}"
			{{- end }}
			}`
		tmpl, err := template.New("datasource_service_brokers").Parse(s)
		if err != nil {
			panic(err)
		}
		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, sbmp)
		if err != nil {
			panic(err)
		}
		return buf.String()
	}
	return sbmp.HclType + ` "cloudfoundry_service_brokers" ` + sbmp.HclObjectName + ` {}`
}

func TestServiceBrokersDataSource(t *testing.T) {
	t.Parallel()
	t.Run("happy path - read service brokers", func(t *testing.T) {
		cfg := getCFHomeConf()
		dataSourceName := "data.cloudfoundry_service_brokers.ds"
		rec := cfg.SetupVCR(t, "fixtures/datasource_service_brokers")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclServiceBrokers(&ServiceBrokersModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(dataSourceName, "service_brokers.#", "3"),
					),
				},
				{
					Config: hclProvider(nil) + hclServiceBrokers(&ServiceBrokersModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Name:          new("hi"),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(dataSourceName, "service_brokers.#", "1"),
					),
				},
			},
		})
	})
	t.Run("error path - get unavailable service brokers", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_service_brokers_invalid")
		dataSourceName := "data.cloudfoundry_service_brokers.ds"
		defer stopQuietly(rec)
		// Create a Terraform configuration that uses the data source
		// and run `terraform apply`. The data source should not be found.
		resource.UnitTest(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclServiceBrokers(&ServiceBrokersModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Name:          new("invalid-service-broker-name"),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(dataSourceName, "service_brokers.#", "0"),
					),
				},
			},
		})
	})
}
