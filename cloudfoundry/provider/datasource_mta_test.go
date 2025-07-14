package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type MtaDataSourceModelPtr struct {
	HclType       string
	HclObjectName string
	Space         *string
	Id            *string
	Namespace     *string
	Mta           *string
	DeployUrl     *string
}

func hclDataSourceMta(mdsmp *MtaDataSourceModelPtr) string {
	if mdsmp != nil {
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
	return mdsmp.HclType + ` "cloudfoundry_mta" ` + mdsmp.HclObjectName + ` {}`
}

func TestMtaDataSource_Configure(t *testing.T) {
	var (
		//canary->tf-space-1
		mtaId     = "a.cf.app"
		spaceGuid = "02c0cc92-6ecc-44b1-b7b2-096ca19ee143"
		namespace = "test"
	)
	t.Parallel()
	dataSourceName := "data.cloudfoundry_mta.ds"
	t.Run("happy path - read mta", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_mta")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclDataSourceMta(&MtaDataSourceModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Space:         &spaceGuid,
						Id:            &mtaId,
						Namespace:     &namespace,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(dataSourceName, "mta.metadata.id", mtaId),
						resource.TestCheckResourceAttr(dataSourceName, "space", spaceGuid),
						resource.TestCheckNoResourceAttr(dataSourceName, "deploy_url"),
					),
				},
			},
		})
	})
	t.Run("error path - mta does not exist", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_mta_invalid")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclDataSourceMta(&MtaDataSourceModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Space:         &invalidOrgGUID,
						Id:            &mtaId,
					}),
					ExpectError: regexp.MustCompile(`Unable to fetch Space details`),
				},
				{
					Config: hclProvider(nil) + hclDataSourceMta(&MtaDataSourceModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Space:         &spaceGuid,
						Id:            &mtaId,
					}),
					ExpectError: regexp.MustCompile(`Unable to fetch MTA details`),
				},
			},
		})
	})
}
