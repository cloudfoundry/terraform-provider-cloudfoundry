package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type ServiceRouteBindingsModelPtr struct {
	HclType         string
	HclObjectName   string
	Route           *string
	Routes          *string
	ServiceInstance *string
}

func hclServiceRouteBindings(sip *ServiceRouteBindingsModelPtr) string {
	if sip != nil {
		s := `
		{{.HclType}} "cloudfoundry_service_route_bindings" {{.HclObjectName}} {
			{{- if .Route}}
				route  = "{{.Route}}"
			{{- end -}}
			{{if .ServiceInstance}}
				service_instance = "{{.ServiceInstance}}"
			{{- end }}
			{{if .Routes}}
				routes = "{{.Routes}}"
			{{- end }}
		}`
		tmpl, err := template.New("service_route_bindings").Parse(s)
		if err != nil {
			panic(err)
		}
		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, sip)
		if err != nil {
			panic(err)
		}
		return buf.String()
	}
	return sip.HclType + ` "cloudfoundry_service_route_bindings"  "` + sip.HclObjectName + ` {}`
}

func TestServiceRouteBindingsDataSource(t *testing.T) {
	var (
		ServiceInstanceGUID = "ab65cad9-73fa-4dd4-9c09-87f89b2e77ec"
	)
	t.Parallel()
	t.Run("happy path - read route bindings", func(t *testing.T) {
		cfg := getCFHomeConf()
		dataSourceName := "data.cloudfoundry_service_route_bindings.ds"
		rec := cfg.SetupVCR(t, "fixtures/datasource_service_route_bindings")
		defer stopQuietly(rec)
		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclServiceRouteBindings(&ServiceRouteBindingsModelPtr{
						HclType:         hclObjectDataSource,
						HclObjectName:   "ds",
						ServiceInstance: &ServiceInstanceGUID,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(dataSourceName, "route_bindings.#", "1"),
					),
				},
			},
		})
	})
	t.Run("error path - get unavailable route bindings", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_service_route_bindings_invalid")
		defer stopQuietly(rec)
		resource.UnitTest(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclServiceRouteBindings(&ServiceRouteBindingsModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Route:         &ServiceInstanceGUID,
					}),
					ExpectError: regexp.MustCompile(`Unable to find any route bindings in list`),
				},
			},
		})
	})
}
