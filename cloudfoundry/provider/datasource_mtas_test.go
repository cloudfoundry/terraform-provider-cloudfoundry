package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type MtasDataSourceModelPtr struct {
	HclType       string
	HclObjectName string
	Space         *string
	Id            *string
	Namespace     *string
	Mtas          *string
	DeployUrl     *string
}

type MtaResourceModelPtr struct {
	HclType                    string
	HclObjectName              string
	MtarPath                   *string
	MtarUrl                    *string
	ExtensionDescriptors       *string
	ExtensionDescriptorsString *string
	DeployUrl                  *string
	Space                      *string
	Mta                        *string
	Namespace                  *string
	Id                         *string
	SourceCodeHash             *string
	DeployStrategy             *string
	VersionRule                *string
	Modules                    *string
}

func hclDataSourceMtas(mdsmp *MtasDataSourceModelPtr) string {
	if mdsmp != nil {
		s := `
		{{.HclType}} "cloudfoundry_mtas" {{.HclObjectName}} {
			{{- if .Space}}
				space = "{{.Space}}"
			{{- end -}}
			{{if .Id}}
				id = "{{.Id}}"
			{{- end -}}
			{{if .Namespace}}
				namespace = "{{.Namespace}}"
			{{- end -}}
			{{if .Mtas}}
				mtas = "{{.Mtas}}"
			{{- end -}}
			{{if .DeployUrl}}
				deploy_url = "{{.DeployUrl}}"
			{{- end }}
			}`
		tmpl, err := template.New("datasource_mtar").Parse(s)
		if err != nil {
			panic(err)
		}
		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, mdsmp)
		if err != nil {
			panic(err)
		}
		return buf.String()
	}
	return mdsmp.HclType + ` "cloudfoundry_mtas "` + mdsmp.HclObjectName + ` {}`
}

func hclResourceMta(mrmp *MtaResourceModelPtr) string {
	if mrmp != nil {
		s := `
		{{.HclType}} "cloudfoundry_mta" {{.HclObjectName}} {
			{{- if .Space}}
				space = "{{.Space}}"
			{{- end -}}
			{{if .Id}}
				id = "{{.Id}}"
			{{- end -}}
			{{if .Namespace}}
				namespace = "{{.Namespace}}"
			{{- end -}}
			{{if .Mta}}
				mta = "{{.Mta}}"
			{{- end -}}
			{{if .MtarPath}}
				mtar_path = "{{.MtarPath}}"
			{{- end -}}
			{{if .MtarUrl}}
				mtar_url = "{{.MtarUrl}}"
			{{- end -}}
			{{if .ExtensionDescriptors}}
				extension_descriptors = {{.ExtensionDescriptors}}
			{{- end -}}
			{{if .ExtensionDescriptorsString}}
				extension_descriptors_string = {{.ExtensionDescriptorsString}}
			{{- end -}}
			{{if .SourceCodeHash}}
				source_code_hash = "{{.SourceCodeHash}}"
			{{- end -}}
			{{if .DeployStrategy}}
				deploy_strategy = "{{.DeployStrategy}}"
			{{- end -}}
			{{if .VersionRule}}
				version_rule = "{{.VersionRule}}"
			{{- end -}}
			{{if .Modules}}
				modules = {{.Modules}}
			{{- end -}}
			{{if .DeployUrl}}
				deploy_url = "{{.DeployUrl}}"
			{{- end }}
			}`
		tmpl, err := template.New("resource_mtar").Parse(s)
		if err != nil {
			panic(err)
		}
		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, mrmp)
		if err != nil {
			panic(err)
		}
		return buf.String()
	}
	return mrmp.HclType + ` "cloudfoundry_mta" ` + mrmp.HclObjectName + ` {}`
}

func TestMtasDataSource_Configure(t *testing.T) {
	var (
		//canary->tf-space-1
		mtaId     = "a.cf.app"
		spaceGuid = "02c0cc92-6ecc-44b1-b7b2-096ca19ee143"
	)
	t.Parallel()
	dataSourceName := "data.cloudfoundry_mtas.ds"
	t.Run("happy path - read mtas", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_mtas")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclDataSourceMtas(&MtasDataSourceModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Space:         strtostrptr(spaceGuid),
						Id:            strtostrptr(mtaId),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(dataSourceName, "mtas.0.metadata.id", mtaId),
						resource.TestCheckResourceAttr(dataSourceName, "space", spaceGuid),
						resource.TestCheckNoResourceAttr(dataSourceName, "deploy_url"),
					),
				},
			},
		})
	})
	t.Run("error path - mtas do not exist", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_mtas_invalid")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclDataSourceMtas(&MtasDataSourceModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Space:         strtostrptr(spaceGuid),
						Id:            strtostrptr(invalidOrgGUID),
					}),
					ExpectError: regexp.MustCompile(`Unable to fetch MTA details`),
				},
			},
		})
	})
}
