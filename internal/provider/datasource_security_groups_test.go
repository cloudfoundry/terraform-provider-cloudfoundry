package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type SecurityGroupsModelPtr struct {
	HclType        string
	HclObjectName  string
	Name           *string
	RunningSpace   *string
	StagingSpace   *string
	SecurityGroups *string
}

func hclSecurityGroups(sgmp *SecurityGroupsModelPtr) string {
	if sgmp != nil {
		s := `
		{{.HclType}} "cloudfoundry_security_groups" {{.HclObjectName}} {
			{{- if .Name}}
				name = "{{.Name}}"
			{{- end -}}
			{{if .RunningSpace}}
				running_space = "{{.RunningSpace}}"
			{{- end -}}
			{{if .StagingSpace}}
				staging_space = "{{.StagingSpace}}"
			{{- end -}}
			{{if .SecurityGroups}}
				security_groups = "{{.SecurityGroups}}"
			{{- end }}
			}`
		tmpl, err := template.New("datasource_security_groups").Parse(s)
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
	return sgmp.HclType + ` "cloudfoundry_security_groups" ` + sgmp.HclObjectName + ` {}`
}

func TestSecurityGroupsDataSource_Configure(t *testing.T) {
	t.Parallel()
	dataSourceName := "data.cloudfoundry_security_groups.ds"
	name := "postgresql-db-a7dd32b7-f185-4c8d-9a94-d5fa7f1c6f33"
	runningSpace := "0668fb26-eebb-4ad6-92cb-2e11a1f11844"
	t.Run("happy path - read security groups", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_security_groups")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclSecurityGroups(&SecurityGroupsModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Name:          &name,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(dataSourceName, "security_groups.#", "1"),
					),
				},
				{
					Config: hclProvider(nil) + hclSecurityGroups(&SecurityGroupsModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						RunningSpace:  &runningSpace,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(dataSourceName, "security_groups.#", "1"),
					),
				},
			},
		})
	})
	t.Run("error path - security groups do not exist", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_security_groups_invalid")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclSecurityGroups(&SecurityGroupsModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Name:          strtostrptr(invalidOrgGUID),
					}),
					ExpectError: regexp.MustCompile(`Unable to find any security group in list`),
				},
			},
		})
	})
}
