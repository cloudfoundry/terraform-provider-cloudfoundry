package provider

import (
	"bytes"
	"regexp"
	"strconv"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type BuildpackModelPtr struct {
	HclType       string
	HclObjectName string
	Name          *string
	Path          *string
	State         *string
	Id            *string
	Stack         *string
	Filename      *string
	Position      *int
	Enabled       *bool
	Locked        *bool
	Labels        *string
	Annotations   *string
	CreatedAt     *string
	UpdatedAt     *string
}

func hclBuildpack(bmp *BuildpackModelPtr) string {
	if bmp != nil {
		s := `
		{{.HclType}} "cloudfoundry_buildpack" {{.HclObjectName}} {
			{{- if .Name}}
				name = "{{.Name}}"
			{{- end -}}
			{{if .Id}}
				id = "{{.Id}}"
			{{- end -}}
			{{if .Path}}
				path = "{{.Path}}"
			{{- end -}}
			{{if .State}}
				state = "{{.State}}"
			{{- end -}}
			{{if .Stack}}
				stack = "{{.Stack}}"
			{{- end -}}
			{{if .Filename}}
				filename = {{.Filename}}
			{{- end -}}
			{{if .Position}}
				position = "{{.Position}}"
			{{- end -}}
			{{if .Enabled}}
				enabled = {{.Enabled}}
			{{- end -}}
			{{if .Locked}}
				locked = {{.Locked}}
			{{- end -}}
			{{if .Labels}}
				labels = {{.Labels}}
			{{- end -}}
			{{if .Annotations}}
				annotations = {{.Annotations}}
			{{- end -}}
			{{if .CreatedAt}}
				created_at = "{{.CreatedAt}}"
			{{- end -}}
			{{if .UpdatedAt}}
				updated_at = "{{.UpdatedAt}}"
			{{- end }}
			}`
		tmpl, err := template.New("resource_buildpack").Parse(s)
		if err != nil {
			panic(err)
		}
		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, bmp)
		if err != nil {
			panic(err)
		}
		return buf.String()
	}
	return bmp.HclType + ` "cloudfoundry_buildpack "` + bmp.HclObjectName + ` {}`
}

func TestBuildpackResource_Configure(t *testing.T) {
	var (
		// in staging
		zipFilePath     = "../../assets/cf-sample-app-nodejs.zip"
		resourceName    = "cloudfoundry_buildpack.rs"
		buildpackName   = "hifi"
		buildpackName2  = "bruh"
		position        = 1
		updatedPosition = 2
		stack           = "cflinuxfs3"
		enabled         = false
		locked          = false
		updatedLocked   = true
	)
	t.Parallel()
	t.Run("happy path - create/import/delete buildpack", func(t *testing.T) {

		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_buildpack")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclBuildpack(&BuildpackModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						Name:          &buildpackName,
						Position:      &position,
						Stack:         &stack,
						Enabled:       &enabled,
						Locked:        &locked,
						Path:          &zipFilePath,
						Labels:        &testCreateLabel,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestCheckResourceAttr(resourceName, "enabled", strconv.FormatBool(enabled)),
						resource.TestCheckResourceAttr(resourceName, "position", strconv.Itoa(position)),
						resource.TestCheckResourceAttr(resourceName, "stack", stack),
						resource.TestCheckResourceAttr(resourceName, "locked", strconv.FormatBool(locked)),
						resource.TestCheckResourceAttr(resourceName, "name", buildpackName),
						resource.TestCheckResourceAttr(resourceName, "path", zipFilePath),
						resource.TestCheckResourceAttr(resourceName, "labels.purpose", "testing"),
					),
				},
				{
					Config: hclProvider(nil) + hclBuildpack(&BuildpackModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						Name:          &buildpackName,
						Position:      &updatedPosition,
						Stack:         &stack,
						Enabled:       &enabled,
						Locked:        &updatedLocked,
						Path:          &zipFilePath,
						Labels:        &testUpdateLabel,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestCheckResourceAttr(resourceName, "enabled", strconv.FormatBool(enabled)),
						resource.TestCheckResourceAttr(resourceName, "position", strconv.Itoa(updatedPosition)),
						resource.TestCheckResourceAttr(resourceName, "stack", stack),
						resource.TestCheckResourceAttr(resourceName, "locked", strconv.FormatBool(updatedLocked)),
						resource.TestCheckResourceAttr(resourceName, "name", buildpackName),
						resource.TestCheckResourceAttr(resourceName, "path", zipFilePath),
						resource.TestCheckResourceAttr(resourceName, "labels.purpose", "production"),
					),
				},
				{
					ResourceName:            resourceName,
					ImportStateIdFunc:       getIdForImport(resourceName),
					ImportStateVerifyIgnore: []string{"path", "source_code_hash"},
					ImportState:             true,
					ImportStateVerify:       true,
				},
			},
		})
	})
	t.Run("error path - create/update buildpacks with existing name/invalid path/invalid zip file", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_buildpack_invalid")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclBuildpack(&BuildpackModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						Name:          strtostrptr("hi"),
						Stack:         &stack,
					}),
					ExpectError: regexp.MustCompile(`API Error Creating Buildpack`),
				},
				{
					Config: hclProvider(nil) + hclBuildpack(&BuildpackModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						Name:          strtostrptr("hibro"),
						Path:          strtostrptr("ok"),
					}),
					ExpectError: regexp.MustCompile(`Invalid file or Path given`),
				},
				{
					Config: hclProvider(nil) + hclBuildpack(&BuildpackModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						Name:          strtostrptr("hi"),
						Path:          strtostrptr("resource_buildpack_test.go"),
					}),
					ExpectError: regexp.MustCompile(`API Error Uploading Buildpack`),
				},
				{
					Config: hclProvider(nil) + hclBuildpack(&BuildpackModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						Name:          &buildpackName2,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestCheckResourceAttr(resourceName, "name", buildpackName2),
					),
				},
				{
					Config: hclProvider(nil) + hclBuildpack(&BuildpackModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						Name:          &buildpackName2,
						Path:          strtostrptr("ok"),
					}),
					ExpectError: regexp.MustCompile(`Invalid file or Path given`),
				},
				{
					Config: hclProvider(nil) + hclBuildpack(&BuildpackModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						Name:          &buildpackName2,
						Path:          strtostrptr("resource_buildpack_test.go"),
					}),
					ExpectError: regexp.MustCompile(`API Error Uploading Buildpack`),
				},
			},
		})
	})
}
