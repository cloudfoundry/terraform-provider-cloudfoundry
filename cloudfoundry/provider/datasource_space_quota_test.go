package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type SpaceQuotaModelPtr struct {
	HclType               string
	HclObjectName         string
	Name                  *string
	Id                    *string
	AllowPaidServicePlans *bool
	TotalServices         *int
	TotalServiceKeys      *int
	TotalRoutes           *int
	TotalRoutePorts       *int
	TotalMemory           *int
	InstanceMemory        *int
	TotalAppInstances     *int
	TotalAppTasks         *int
	TotalAppLogRateLimit  *int
	Spaces                *string
	Org                   *string
	CreatedAt             *string
	UpdatedAt             *string
}

func hclSpaceQuota(sqdsmp *SpaceQuotaModelPtr) string {
	if sqdsmp != nil {
		s := `
			{{.HclType}} "cloudfoundry_space_quota" "{{.HclObjectName}}" {
			{{- if .Name}}
				name  = "{{.Name}}"
			{{- end -}}
			{{if .Id}}
				id = {{.Id}}
			{{- end -}}
			{{if .AllowPaidServicePlans}}
				allow_paid_service_plans = {{.AllowPaidServicePlans}}
			{{- end -}}
			{{if .TotalServices}}
				total_services = {{.TotalServices}}
			{{- end }}
			{{if .TotalServiceKeys}}
				total_service_keys = {{.TotalServiceKeys}}
			{{- end }}
			{{if .TotalRoutes}}
				total_routes = {{.TotalRoutes}}
			{{- end }}
			{{if .TotalRoutePorts}}
				total_route_ports = {{.TotalRoutePorts}}
			{{- end }}
			{{if .TotalMemory}}
				total_memory = {{.TotalMemory}}
			{{- end }}
			{{if .InstanceMemory}}
				instance_memory = {{.InstanceMemory}}
			{{- end }}
			{{if .TotalAppInstances}}
				total_app_instances = {{.TotalAppInstances}}
			{{- end }}
			{{if .TotalAppTasks}}
				total_app_tasks = {{.TotalAppTasks}}
			{{- end }}
			{{if .TotalAppLogRateLimit}}
				total_app_log_rate_limit = {{.TotalAppLogRateLimit}}
			{{- end }}
			{{if .Org}}
				org = "{{.Org}}"
			{{- end }}
			{{if .Spaces}}
				spaces = {{.Spaces}}
			{{- end }}
			{{if .CreatedAt}}
				created_at = {{.CreatedAt}}
			{{- end }}
			{{if .UpdatedAt}}
				updated_at = {{.UpdatedAt}}
			{{- end }}
			}`
		tmpl, err := template.New("space_quota").Parse(s)
		if err != nil {
			panic(err)
		}
		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, sqdsmp)
		if err != nil {
			panic(err)
		}
		return buf.String()
	}
	return sqdsmp.HclType + ` cloudfoundry_space_quota  "` + sqdsmp.HclObjectName + `  {}`
}
func TestSpaceQuotaDataSource_Configure(t *testing.T) {
	t.Parallel()
	t.Run("error path - get unavailable datasource space quota", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_space_quota_invalid_name")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclSpaceQuota(&SpaceQuotaModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Name:          strtostrptr("testunavailablespacequota"),
					}),
					ExpectError: regexp.MustCompile(`Error: Unable to find space quota data in list`),
				},
			},
		})
	})
	t.Run("get available datasource space quota", func(t *testing.T) {
		cfg := getCFHomeConf()
		resourceName := "data.cloudfoundry_space_quota.ds"
		rec := cfg.SetupVCR(t, "fixtures/datasource_space_quota")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclSpaceQuota(&SpaceQuotaModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Name:          strtostrptr(testSpaceQuota),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestCheckResourceAttr(resourceName, "instance_memory", "2048"),
						resource.TestCheckResourceAttr(resourceName, "name", testSpaceQuota),
						resource.TestCheckResourceAttr(resourceName, "total_app_instances", "110"),
						resource.TestCheckResourceAttr(resourceName, "total_app_log_rate_limit", "1000"),
						resource.TestCheckResourceAttr(resourceName, "total_app_tasks", "10"),
						resource.TestCheckResourceAttr(resourceName, "total_memory", "51200"),
						resource.TestCheckResourceAttr(resourceName, "instance_memory", "2048"),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestMatchResourceAttr(resourceName, "updated_at", regexpValidRFC3999Format),
						resource.TestMatchResourceAttr(resourceName, "spaces.0", regexpValidUUID),
					),
				},
			},
		})
	})
}
