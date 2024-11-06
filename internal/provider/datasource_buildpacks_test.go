package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type BuildpacksModelPtr struct {
	HclType       string
	HclObjectName string
	Name          *string
	Stack         *string
	Buildpacks    *string
}

func hclBuildpacks(ddsmp *BuildpacksModelPtr) string {
	if ddsmp != nil {
		s := `
			{{.HclType}} "cloudfoundry_buildpacks" "{{.HclObjectName}}" {
			{{- if .Name}}
				name  = "{{.Name}}"
			{{- end -}}
			{{if .Stack}}
				stack = "{{.Stack}}"
			{{- end }}
			{{if .Buildpacks}}
				buildpacks = {{.Buildpacks}}
			{{- end }}
			}`
		tmpl, err := template.New("buildpacks").Parse(s)
		if err != nil {
			panic(err)
		}
		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, ddsmp)
		if err != nil {
			panic(err)
		}
		return buf.String()
	}
	return ddsmp.HclType + ` cloudfoundry_Buildpacks" ` + ddsmp.HclObjectName + `  {}`
}
func TestBuildpacksDataSource_Configure(t *testing.T) {
	resourceName := "data.cloudfoundry_buildpacks.ds"
	buildpackName := "java_buildpack"
	stackName := "cflinuxfs4"
	t.Parallel()
	t.Run("error path - get buildpacks", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_buildpacks_invalid")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclBuildpacks(&BuildpacksModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Name:          &invalidOrgGUID,
					}),
					ExpectError: regexp.MustCompile(`Unable to find any buildpack in the list`),
				},
			},
		})
	})
	t.Run("happy path - get buildpacks", func(t *testing.T) {
		cfg := getCFHomeConf()

		rec := cfg.SetupVCR(t, "fixtures/datasource_buildpacks")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclBuildpacks(&BuildpacksModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Name:          &buildpackName,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "buildpacks.#", "2"),
					),
				},
				{
					Config: hclProvider(nil) + hclBuildpacks(&BuildpacksModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Stack:         &stackName,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "buildpacks.#", "8"),
					),
				},
			},
		})
	})
}
