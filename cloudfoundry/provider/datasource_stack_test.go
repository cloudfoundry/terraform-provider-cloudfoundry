package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type StackModelPtr struct {
	HclType          string
	HclObjectName    string
	Name             *string
	Id               *string
	Description      *string
	BuildRootfsImage *string
	RunRootfsImage   *string
	Default          *bool
	Labels           *string
	Annotations      *string
	CreatedAt        *string
	UpdatedAt        *string
}

func hclStack(smp *StackModelPtr) string {
	if smp != nil {
		s := `
		{{.HclType}} "cloudfoundry_stack" {{.HclObjectName}} {
			{{- if .Name}}
				name = "{{.Name}}"
			{{- end -}}
			{{if .Id}}
				id = "{{.Id}}"
			{{- end -}}
			{{if .Description}}
				description = "{{.Description}}"
			{{- end -}}
			{{if .BuildRootfsImage}}
				build_rootfs_image = "{{.BuildRootfsImage}}"
			{{- end -}}
			{{if .Default}}
				default = {{.Default}}
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
		tmpl, err := template.New("datasource_stack").Parse(s)
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
	return smp.HclType + ` "cloudfoundry_stack" ` + smp.HclObjectName + ` {}`
}

func TestStackDataSource_Configure(t *testing.T) {
	var (
		dataSourceName   = "data.cloudfoundry_stack.ds"
		stackName        = "cflinuxfs4"
		stackDescription = "Cloud Foundry Linux-based filesystem (Ubuntu 22.04)"
		stackDefault     = "true"
	)
	t.Parallel()
	t.Run("happy path - read stack", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_stack")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclStack(&StackModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Name:          &stackName,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(dataSourceName, "id", regexpValidUUID),
						resource.TestCheckResourceAttr(dataSourceName, "name", stackName),
						resource.TestMatchResourceAttr(dataSourceName, "created_at", regexpValidRFC3999Format),
						resource.TestMatchResourceAttr(dataSourceName, "updated_at", regexpValidRFC3999Format),
						resource.TestCheckResourceAttr(dataSourceName, "description", stackDescription),
						resource.TestCheckResourceAttr(dataSourceName, "default", stackDefault),
					),
				},
			},
		})
	})
	t.Run("error path - stack does not exist", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_stack_invalid")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclStack(&StackModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Name:          &invalidOrgGUID,
					}),
					ExpectError: regexp.MustCompile(`Unable to find stack in list`),
				},
			},
		})
	})
}
