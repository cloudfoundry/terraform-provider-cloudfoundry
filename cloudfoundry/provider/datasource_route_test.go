package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type RouteDataSourceModelPtr struct {
	HclType       string
	HclObjectName string
	Host          *string
	Path          *string
	Port          *int
	Space         *string
	Domain        *string
	Org           *string
	Routes        *string
}

type RouteResourceModelPtr struct {
	HclType       string
	HclObjectName string
	Protocol      *string
	Id            *string
	Host          *string
	Path          *string
	Port          *int
	Url           *string
	Destinations  *string
	Space         *string
	Domain        *string
	Labels        *string
	Annotations   *string
	CreatedAt     *string
	UpdatedAt     *string
}

func hclResourceRoute(rrmp *RouteResourceModelPtr) string {
	if rrmp != nil {
		s := `
		{{.HclType}} "cloudfoundry_route" {{.HclObjectName}} {
			{{- if .Protocol}}
				protocol = "{{.Protocol}}"
			{{- end -}}
			{{if .Id}}
				id = "{{.Id}}"
			{{- end -}}
			{{if .Host}}
				host = "{{.Host}}"
			{{- end -}}
			{{if .Path}}
				path = "{{.Path}}"
			{{- end -}}
			{{if .Port}}
				port = {{.Port}}
			{{- end -}}
			{{if .Url}}
				url = "{{.Url}}"
			{{- end -}}
			{{if .Destinations}}
				destinations = {{.Destinations}}
			{{- end -}}
			{{if .Space}}
				space = "{{.Space}}"
			{{- end -}}
			{{if .Domain}}
				domain = "{{.Domain}}"
			{{- end -}}
			{{if .Labels}}
				labels = {{.Labels}}
			{{- end -}}
			{{if .Annotations}}
				annotations = {{.Annotations}}
			{{- end -}}
			{{if .UpdatedAt}}
				updated_at = "{{.UpdatedAt}}"
			{{- end -}}
			{{if .CreatedAt}}
				created_at = "{{.CreatedAt}}"
			{{- end }}
			}`
		tmpl, err := template.New("resource_route").Parse(s)
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
	return rrmp.HclType + ` "cloudfoundry_route" ` + rrmp.HclObjectName + ` {}`
}

func hclDataSourceRoute(rrmp *RouteDataSourceModelPtr) string {
	if rrmp != nil {
		s := `
		{{.HclType}} "cloudfoundry_route" {{.HclObjectName}} {
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
		tmpl, err := template.New("datasource_route").Parse(s)
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
	return rrmp.HclType + ` "cloudfoundry_route" ` + rrmp.HclObjectName + ` {}`
}

func TestRouteDataSource_Configure(t *testing.T) {
	t.Parallel()
	dataSourceName := "data.cloudfoundry_route.ds"
	t.Run("happy path - read route", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_route")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclDataSourceRoute(&RouteDataSourceModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Domain:        strtostrptr(testDomainRouteGUID),
						Space:         strtostrptr(testSpaceRouteGUID),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(dataSourceName, "routes.0.id", regexpValidUUID),
						resource.TestCheckResourceAttr(dataSourceName, "routes.0.protocol", "http"),
						resource.TestCheckResourceAttr(dataSourceName, "routes.0.host", "hahadban2-accountable-jackal-fn"),
						resource.TestCheckNoResourceAttr(dataSourceName, "routes.0.port"),
						resource.TestMatchResourceAttr(dataSourceName, "routes.0.created_at", regexpValidRFC3999Format),
						resource.TestMatchResourceAttr(dataSourceName, "routes.0.updated_at", regexpValidRFC3999Format),
						resource.TestMatchResourceAttr(dataSourceName, "routes.0.destinations.0.id", regexpValidUUID),
						resource.TestMatchResourceAttr(dataSourceName, "routes.0.destinations.0.app_id", regexpValidUUID),
						resource.TestCheckResourceAttr(dataSourceName, "routes.0.destinations.0.app_process_type", "web"),
						resource.TestCheckResourceAttr(dataSourceName, "routes.0.destinations.0.port", "80"),
						resource.TestCheckNoResourceAttr(dataSourceName, "routes.0.destinations.0.weight"),
						resource.TestCheckResourceAttr(dataSourceName, "routes.0.destinations.0.protocol", "http1"),
					),
				},
			},
		})
	})
	t.Run("error path - route not exist", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_route_invalid")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclDataSourceRoute(&RouteDataSourceModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Domain:        strtostrptr(testDomainRouteGUID),
						Space:         strtostrptr(testDomainRouteGUID),
					}),
					ExpectError: regexp.MustCompile(`Unable to find route in list`),
				},
			},
		})
	})
}
