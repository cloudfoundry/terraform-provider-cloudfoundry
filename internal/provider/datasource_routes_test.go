package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func hclDataSourceRoutes(rrmp *RouteDataSourceModelPtr) string {
	if rrmp != nil {
		s := `
		{{.HclType}} "cloudfoundry_routes" {{.HclObjectName}} {
			{{- if .Host}}
				host = "{{.Host}}"
			{{- end -}}
			{{if .Path}}
				path = "{{.Path}}"
			{{- end -}}
			{{if .Port}}
				port = {{.Port}}
			{{- end -}}
			{{if .Routes}}
				routes = {{.Routes}}
			{{- end -}}
			{{if .Space}}
				space = "{{.Space}}"
			{{- end -}}
			{{if .Org}}
				org = "{{.Org}}"
			{{- end -}}
			{{if .Domain}}
				domain = "{{.Domain}}"
			{{- end }}
			}`
		tmpl, err := template.New("datasource_routes").Parse(s)
		if err != nil {
			panic(err)
		}
		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, rrmp)
		if err != nil {
			panic(err)
		}
		return buf.String()
	}
	return rrmp.HclType + ` "cloudfoundry_routes" ` + rrmp.HclObjectName + ` {}`
}

func TestRoutesDataSource_Configure(t *testing.T) {
	t.Parallel()
	dataSourceName := "data.cloudfoundry_routes.ds"
	testSpaceRouteGUID := "02c0cc92-6ecc-44b1-b7b2-096ca19ee143"
	t.Run("happy path - read route", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_routes")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclDataSourceRoutes(&RouteDataSourceModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Space:         strtostrptr(testSpaceRouteGUID),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(dataSourceName, "routes.#", "23"),
					),
				},
			},
		})
	})
	t.Run("error path - route not exist", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_routes_invalid")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclDataSourceRoutes(&RouteDataSourceModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Domain:        strtostrptr(testDomainRouteGUID),
						Space:         strtostrptr(testSpaceRouteGUID),
					}),
					ExpectError: regexp.MustCompile(`Unable to find route in list`),
				},
			},
		})
	})
}
