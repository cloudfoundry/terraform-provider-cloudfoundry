package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type SecurityGroupModelPtr struct {
	HclType                string
	HclObjectName          string
	Name                   *string
	Id                     *string
	Rules                  *string
	GloballyEnabledRunning *bool
	GloballyEnabledStaging *bool
	RunningSpaces          *string
	StagingSpaces          *string
	CreatedAt              *string
	UpdatedAt              *string
}

func hclSecurityGroup(sgmp *SecurityGroupModelPtr) string {
	if sgmp != nil {
		s := `
		{{.HclType}} "cloudfoundry_security_group" {{.HclObjectName}} {
			{{- if .Name}}
				name = "{{.Name}}"
			{{- end -}}
			{{if .Id}}
				id = "{{.Id}}"
			{{- end -}}
			{{if .Rules}}
				rules = {{.Rules}}
			{{- end -}}
			{{if .GloballyEnabledRunning}}
				globally_enabled_running = {{.GloballyEnabledRunning}}
			{{- end -}}
			{{if .GloballyEnabledStaging}}
				globally_enabled_staging = {{.GloballyEnabledStaging}}
			{{- end -}}
			{{if .RunningSpaces}}
				running_spaces = {{.RunningSpaces}}
			{{- end -}}
			{{if .StagingSpaces}}
				staging_spaces = {{.StagingSpaces}}
			{{- end -}}
			{{if .UpdatedAt}}
				updated_at = "{{.UpdatedAt}}"
			{{- end -}}
			{{if .CreatedAt}}
				created_at = "{{.CreatedAt}}"
			{{- end }}
			}`
		tmpl, err := template.New("datasource_security_group").Parse(s)
		if err != nil {
			panic(err)
		}
		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, sgmp)
		if err != nil {
			panic(err)
		}
		return buf.String()
	}
	return sgmp.HclType + ` "cloudfoundry_security_group" ` + sgmp.HclObjectName + ` {}`
}

func TestSecurityGroupDataSource_Configure(t *testing.T) {
	t.Parallel()
	dataSourceName := "data.cloudfoundry_security_group.ds"
	t.Run("happy path - read security group", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_security_group")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclSecurityGroup(&SecurityGroupModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Name:          strtostrptr(testSpace),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(dataSourceName, "id", regexpValidUUID),
						resource.TestCheckResourceAttr(dataSourceName, "rules.#", "3"),
						resource.TestCheckResourceAttr(dataSourceName, "rules.2.code", "0"),
						resource.TestMatchResourceAttr(dataSourceName, "created_at", regexpValidRFC3999Format),
						resource.TestMatchResourceAttr(dataSourceName, "updated_at", regexpValidRFC3999Format),
						resource.TestCheckResourceAttr(dataSourceName, "globally_enabled_running", "false"),
						resource.TestCheckTypeSetElemAttr(dataSourceName, "staging_spaces.*", testSpaceGUID),
						resource.TestCheckResourceAttr(dataSourceName, "running_spaces.#", "1"),
					),
				},
			},
		})
	})
	t.Run("error path - security group not exist", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_security_group_invalid")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclSecurityGroup(&SecurityGroupModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Name:          strtostrptr(invalidOrgGUID),
					}),
					ExpectError: regexp.MustCompile(`Unable to find security group in list`),
				},
			},
		})
	})
}
