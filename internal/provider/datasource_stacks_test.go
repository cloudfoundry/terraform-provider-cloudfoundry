package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type StacksModelPtr struct {
	HclType       string
	HclObjectName string
	Name          *string
	Stacks        *string
}

func hclStacks(sdsmp *StacksModelPtr) string {
	if sdsmp != nil {
		s := `
			{{.HclType}} "cloudfoundry_stacks" "{{.HclObjectName}}" {
			{{- if .Name}}
				name  = "{{.Name}}"
			{{- end -}}
			{{if .Stacks}}
				stacks = {{.Stacks}}
			{{- end }}
			}`
		tmpl, err := template.New("stacks").Parse(s)
		if err != nil {
			panic(err)
		}
		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, sdsmp)
		if err != nil {
			panic(err)
		}
		return buf.String()
	}
	return sdsmp.HclType + ` "cloudfoundry_stacks" ` + sdsmp.HclObjectName + `  {}`
}
func TestStacksDataSource_Configure(t *testing.T) {
	resourceName := "data.cloudfoundry_stacks.ds"
	stackName := "cflinuxfs4"
	t.Parallel()
	t.Run("error path - get stacks", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_stacks_invalid")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclStacks(&StacksModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Name:          &invalidOrgGUID,
					}),
					ExpectError: regexp.MustCompile(`No stack present with mentioned criteria`),
				},
			},
		})
	})
	t.Run("happy path - get stacks", func(t *testing.T) {
		cfg := getCFHomeConf()

		rec := cfg.SetupVCR(t, "fixtures/datasource_stacks")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclStacks(&StacksModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Name:          &stackName,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "stacks.#", "1"),
					),
				},
			},
		})
	})
}
