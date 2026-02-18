package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type RemoteMtarHashModelPtr struct {
	HclType       string
	HclObjectName string
	Url           *string
	Id            *string
}

func hclRemoteMtarHash(rfhmp *RemoteMtarHashModelPtr) string {
	if rfhmp != nil {
		s := `
		{{.HclType}} "cloudfoundry_remote_mtar_hash" {{.HclObjectName}} {
			{{- if .Url}}
				url = "{{.Url}}"
			{{- end -}}
			{{if .Id}}
				id = "{{.Id}}"
			{{- end }}
			}`
		tmpl, err := template.New("datasource_remote_file_hash").Parse(s)
		if err != nil {
			panic(err)
		}
		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, rfhmp)
		if err != nil {
			panic(err)
		}
		return buf.String()
	}
	return rfhmp.HclType + ` "cloudfoundry_remote_mtar_hash" ` + rfhmp.HclObjectName + ` {}`
}

func TestRemoteMtarHashDataSource_Configure(t *testing.T) {
	remoteUrl := "https://github.com/Dray56/mtar-archive/releases/download/v1.0.0/a.cf.app.mtar"
	dataSourceName := "data.cloudfoundry_remote_mtar_hash.ds"
	t.Parallel()
	t.Run("happy path - read remote file", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_remote_mtar_hash")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclRemoteMtarHash(&RemoteMtarHashModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Url:           &remoteUrl,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(dataSourceName, "url", remoteUrl),
						resource.TestMatchResourceAttr(dataSourceName, "id", regexpValidSHA),
					),
				},
			},
		})
	})
	t.Run("error path - invalid url/file content", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/datasource_remote_mtar_hash_invalid")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclRemoteMtarHash(&RemoteMtarHashModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Url:           new("https://www.google.com"),
					}),
					ExpectError: regexp.MustCompile(`Invalid content type`),
				},
				{
					Config: hclProvider(nil) + hclRemoteMtarHash(&RemoteMtarHashModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Url:           new("http://example.com invalid"),
					}),
					ExpectError: regexp.MustCompile(`Unable to create HEAD request`),
				},
				{
					Config: hclProvider(nil) + hclRemoteMtarHash(&RemoteMtarHashModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Url:           new("http://localhost:8080/"),
					}),
					ExpectError: regexp.MustCompile(`Unable to download file`),
				},
				{
					Config: hclProvider(nil) + hclRemoteMtarHash(&RemoteMtarHashModelPtr{
						HclType:       hclObjectDataSource,
						HclObjectName: "ds",
						Url:           new("hello"),
					}),
					ExpectError: regexp.MustCompile(`Invalid URL`),
				},
				{
					Config: hclProvider(nil) + `
									data "cloudfoundry_remote_mtar_hash" "ds" {
										url = ["cf-nodejs"]
									}
					`,
					ExpectError: regexp.MustCompile(`Incorrect attribute value type`),
				},
			},
		})
	})
}
