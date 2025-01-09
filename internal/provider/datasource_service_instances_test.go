package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type datasourceServiceInstancesModelPtr struct {
	HclType          string
	HclObjectName    string
	Name             *string
	Org              *string
	Space            *string
	ServiceInstances *string
}

func hclServiceInstances(sip *datasourceServiceInstancesModelPtr) string {
	if sip != nil {
		s := `
		{{.HclType}} "cloudfoundry_service_instances" {{.HclObjectName}} {
			{{- if .Name}}
				name  = "{{.Name}}"
			{{- end -}}
			{{if .Org}}
				org = "{{.Org}}"
			{{- end -}}
			{{if .Space}}
				space = "{{.Space}}"
			{{- end -}}
			{{if .ServiceInstances}}
				service_instances = {{.ServiceInstances}}
			{{- end }}
		}`
		tmpl, err := template.New("service_instances").Parse(s)
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
	return sip.HclType + ` "cloudfoundry_service_instances"  "` + sip.HclObjectName + ` {}`
}

func TestServiceInstancesDataSource(t *testing.T) {
	var (
		// in canary -> PerformanceTeamBLR -> tf-space-1
		testSpaceGUID       = "02c0cc92-6ecc-44b1-b7b2-096ca19ee143"
		testServiceInstance = "tf-test-do-not-delete"
	)
	t.Parallel()
	t.Run("happy path - read service instances", func(t *testing.T) {
		cfg := getCFHomeConf()
		dataSourceName := "data.cloudfoundry_service_instances.ds"
		rec := cfg.SetupVCR(t, "fixtures/datasource_service_instances")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclServiceInstances(&datasourceServiceInstancesModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Space:         &testSpaceGUID,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(dataSourceName, "service_instances.#", "14"),
					),
				},
				{
					Config: hclProvider(nil) + hclServiceInstances(&datasourceServiceInstancesModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Org:           &testOrg2GUID,
						Name:          &testServiceInstance,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(dataSourceName, "service_instances.#", "2"),
					),
				},
			},
		})
	})
	t.Run("error path - get unavailable service instance", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_service_instances_invalid")
		dataSourceName := "data.cloudfoundry_service_instances.ds"
		defer stopQuietly(rec)
		// Create a Terraform configuration that uses the data source
		// and run `terraform apply`. The data source should not be found.
		resource.UnitTest(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclServiceInstances(&datasourceServiceInstancesModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Org:           &invalidOrgGUID,
					}),
					ExpectError: regexp.MustCompile(`API Error Fetching Organization`),
				},
				{
					Config: hclProvider(nil) + hclServiceInstances(&datasourceServiceInstancesModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Org:           &testOrg2GUID,
						Name:          strtostrptr("hi"),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(dataSourceName, "service_instances.#", "0"),
					),
				},
			},
		})
	})
}
