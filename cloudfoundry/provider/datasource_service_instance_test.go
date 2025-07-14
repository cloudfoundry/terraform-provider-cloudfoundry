package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type ServiceInstanceModelPtr struct {
	HclType          string
	HclObjectName    string
	Name             *string
	Id               *string
	Labels           *string
	Annotations      *string
	Type             *string
	Space            *string
	ServicePlan      *string
	Parameters       *string
	Credentials      *string
	Tags             *string
	SyslogDrainURL   *string
	RouteServiceURL  *string
	MaintenanceInfo  *string
	UpgradeAvailable *string
	DashboardURL     *string
	LastOperation    *string
	CreatedAt        *string
	UpdatedAt        *string
}

func hclServiceInstance(sip *ServiceInstanceModelPtr) string {
	if sip != nil {
		s := `
		{{.HclType}} "cloudfoundry_service_instance" {{.HclObjectName}} {
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
			{{if .Space}}
				space = "{{.Space}}"
			{{- end }}
			{{if .ServicePlan}}
				service_plan = "{{.ServicePlan}}"
			{{- end -}}
			{{if .Parameters}}
				parameters = <<EOT
				{{.Parameters}}
				EOT
			{{- end -}}
			{{if .Credentials}}
				credentials = <<EOT
				{{.Credentials}}
				EOT
			{{ end }}
			{{if .Tags}}
				tags = {{.Tags}}
			{{- end }}
		}`
		tmpl, err := template.New("service_instance").Parse(s)
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
	return sip.HclType + ` "cloudfoundry_service_instance"  "` + sip.HclObjectName + ` {}`
}

func TestServiceInstanceDataSource(t *testing.T) {
	var (
		// in canary -> PerformanceTeamBLR -> tf-space-1
		testSpaceGUID                       = "02c0cc92-6ecc-44b1-b7b2-096ca19ee143"
		testServiceInstanceUserProvided     = "tf-test-do-not-delete"
		testServiceInstanceManaged          = "tf-test-do-not-delete-managed"
		testServiceInstanceUserProvidedGUID = "5e2976bb-332e-41e1-8be3-53baafea9296" // in canary -> PerformanceTeamBLR -> tf-space-1
		testServiceInstanceManagedGUID      = "68fea1b6-11b9-4737-ad79-74e49832533f" // in canary -> PerformanceTeamBLR -> tf-space-1
	)
	t.Parallel()
	t.Run("happy path - read service instance user-provided", func(t *testing.T) {
		cfg := getCFHomeConf()
		dataSourceName := "data.cloudfoundry_service_instance.ds"
		rec := cfg.SetupVCR(t, "fixtures/datasource_service_instance")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclServiceInstance(&ServiceInstanceModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Name:          strtostrptr(testServiceInstanceUserProvided),
						Space:         strtostrptr(testSpaceGUID),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(dataSourceName, "id", testServiceInstanceUserProvidedGUID),
						resource.TestCheckResourceAttr(dataSourceName, "name", testServiceInstanceUserProvided),
						resource.TestCheckResourceAttr(dataSourceName, "type", "user-provided"),
						resource.TestCheckResourceAttr(dataSourceName, "space", testSpaceGUID),
					),
				},
			},
		})
	})
	t.Run("happy path - read service instance managed", func(t *testing.T) {
		cfg := getCFHomeConf()
		dataSourceName := "data.cloudfoundry_service_instance.ds"
		rec := cfg.SetupVCR(t, "fixtures/datasource_service_instance_managed")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclServiceInstance(&ServiceInstanceModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Name:          strtostrptr(testServiceInstanceManaged),
						Space:         strtostrptr(testSpaceGUID),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(dataSourceName, "id", testServiceInstanceManagedGUID),
						resource.TestCheckResourceAttr(dataSourceName, "name", testServiceInstanceManaged),
						resource.TestCheckResourceAttr(dataSourceName, "type", "managed"),
						resource.TestCheckResourceAttr(dataSourceName, "space", testSpaceGUID),
					),
				},
			},
		})
	})
	t.Run("error path - get unavailable service instance", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_service_instance_invalid_servicename")
		defer stopQuietly(rec)
		// Create a Terraform configuration that uses the data source
		// and run `terraform apply`. The data source should not be found.
		resource.UnitTest(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclServiceInstance(&ServiceInstanceModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Name:          strtostrptr("invalid-service-instance-name"),
						Space:         strtostrptr(testSpaceGUID),
					}),
					ExpectError: regexp.MustCompile(`Unable to find service instance in list`),
				},
			},
		})
	})
}
