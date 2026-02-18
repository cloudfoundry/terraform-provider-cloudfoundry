package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type OrgQuotaModelPtr struct {
	HclType               string
	HclObjectName         string
	Name                  *string
	Id                    *string
	AllowPaidServicePlans *bool
	TotalServices         *int
	TotalServiceKeys      *int
	TotalRoutes           *int
	TotalRoutePorts       *int
	TotalPrivateDomains   *int
	TotalMemory           *int
	InstanceMemory        *int
	TotalAppInstances     *int
	TotalAppTasks         *int
	TotalAppLogRateLimit  *int
	Organizations         *string
	CreatedAt             *string
	UpdatedAt             *string
}

func hclOrgQuota(oqdsmp *OrgQuotaModelPtr) string {
	if oqdsmp != nil {
		s := `
			{{.HclType}} "cloudfoundry_org_quota" "{{.HclObjectName}}" {
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
				total_routes_ports = {{.TotalRoutePorts}}
			{{- end }}
			{{if .TotalPrivateDomains}}
				total_private_domains = {{.TotalPrivateDomains}}
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
			{{if .Organizations}}
				orgs = {{.Organizations}}
			{{- end }}
			{{if .CreatedAt}}
				created_at = {{.CreatedAt}}
			{{- end }}
			{{if .UpdatedAt}}
				updated_at = {{.UpdatedAt}}
			{{- end }}
			}`
		tmpl, err := template.New("org_quota").Parse(s)
		if err != nil {
			panic(err)
		}
		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, oqdsmp)
		if err != nil {
			panic(err)
		}
		return buf.String()
	}
	return oqdsmp.HclType + ` cloudfoundry_org_quota  "` + oqdsmp.HclObjectName + `  {}`
}
func TestOrgQuotaDataSource_Configure(t *testing.T) {
	t.Parallel()
	t.Run("error path - get unavailable datasource org quota", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_org_invalid_org_quota_name")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclOrgQuota(&OrgQuotaModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Name:          new("testunavailableorgquota"),
					}),
					ExpectError: regexp.MustCompile(`Error: Unable to find org quota data in list`),
				},
			},
		})
	})
	t.Run("get available datasource org quota", func(t *testing.T) {
		cfg := getCFHomeConf()
		resourceName := "data.cloudfoundry_org_quota.ds"
		rec := cfg.SetupVCR(t, "fixtures/datasource_org_quota")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclOrgQuota(&OrgQuotaModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Name:          new(testOrgQuota),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestCheckResourceAttr(resourceName, "instance_memory", "2048"),
						resource.TestCheckResourceAttr(resourceName, "name", testOrgQuota),
						resource.TestCheckResourceAttr(resourceName, "total_app_instances", "100"),
						resource.TestCheckResourceAttr(resourceName, "total_app_log_rate_limit", "1000"),
						resource.TestCheckResourceAttr(resourceName, "total_app_tasks", "10"),
						resource.TestCheckResourceAttr(resourceName, "total_memory", "51200"),
						resource.TestCheckResourceAttr(resourceName, "instance_memory", "2048"),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestMatchResourceAttr(resourceName, "updated_at", regexpValidRFC3999Format),
						resource.TestMatchResourceAttr(resourceName, "orgs.0", regexpValidUUID),
					),
				},
			},
		})
	})
}
