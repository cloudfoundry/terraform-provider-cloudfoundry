package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type SpaceModelPtr struct {
	HclType          string
	HclObjectName    string
	Name             *string
	Id               *string
	OrgId            *string
	Quota            *string
	AllowSSH         *bool
	IsolationSegment *string
	Labels           *string
	Annotations      *string
	CreatedAt        *string
	UpdatedAt        *string
}

func hclSpace(smp *SpaceModelPtr) string {
	if smp != nil {
		s := `
		{{.HclType}} "cloudfoundry_space" {{.HclObjectName}} {
			{{- if .Name}}
				name = "{{.Name}}"
			{{- end -}}
			{{if .Id}}
				id = "{{.Id}}"
			{{- end -}}
			{{if .OrgId}}
				org = "{{.OrgId}}"
			{{- end -}}
			{{if .Quota}}
				quota = "{{.Quota}}"
			{{- end -}}
			{{if .AllowSSH}}
				allow_ssh = {{.AllowSSH}}
			{{- end -}}
			{{if .IsolationSegment}}
				isolation_segment = "{{.IsolationSegment}}"
			{{- end -}}
			{{if .CreatedAt}}
				created_at = "{{.CreatedAt}}"
			{{- end -}}
			{{if .UpdatedAt}}
				updated_at = "{{.UpdatedAt}}"
			{{- end -}}
			{{if .Labels}}
				labels = {{.Labels}}
			{{- end -}}
			{{if .Annotations}}
				annotations = {{.Annotations}}
			{{- end }}
			}`
		tmpl, err := template.New("datasource_space").Parse(s)
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
	return smp.HclType + ` "cloudfoundry_space "` + smp.HclObjectName + ` {}`
}

func TestSpaceDataSource_Configure(t *testing.T) {
	t.Parallel()
	dataSourceName := "data.cloudfoundry_space.ds"
	t.Run("happy path - read space", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_space")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclSpace(&SpaceModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Name:          strtostrptr(testSpace),
						OrgId:         strtostrptr(testOrgGUID),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(dataSourceName, "id", testSpaceGUID),
						resource.TestCheckNoResourceAttr(dataSourceName, "quota"),
						resource.TestMatchResourceAttr(dataSourceName, "created_at", regexpValidRFC3999Format),
						resource.TestMatchResourceAttr(dataSourceName, "updated_at", regexpValidRFC3999Format),
						resource.TestCheckResourceAttr(dataSourceName, "allow_ssh", "true"),
						resource.TestCheckResourceAttr(dataSourceName, "labels.purpose", "prod"),
					),
				},
			},
		})
	})
	t.Run("error path - org does not exist", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_space_invalid_org")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclSpace(&SpaceModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Name:          strtostrptr(testSpace),
						OrgId:         strtostrptr(invalidOrgGUID),
					}),
					ExpectError: regexp.MustCompile(`API Error Fetching Organization`),
				},
			},
		})
	})
	t.Run("error path - space does not exist", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_space_invalid_space")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclSpace(&SpaceModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Name:          strtostrptr(testSpace + "x"),
						OrgId:         strtostrptr(testOrgGUID),
					}),
					ExpectError: regexp.MustCompile(`Unable to find space data in list`),
				},
			},
		})
	})

}
