package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type UserGroupsResourceModelPtr struct {
	HclType       string
	HclObjectName string
	User          *string
	Origin        *string
	Groups        *string
}

func hclResourceUserGroups(urmp *UserGroupsResourceModelPtr) string {
	if urmp != nil {
		s := `
		{{.HclType}} "cloudfoundry_user_groups" {{.HclObjectName}} {
			{{- if .User}}
				user = "{{.User}}"
			{{- end -}}
			{{if .Origin}}
				origin = "{{.Origin}}"
			{{- end -}}
			{{if .Groups}}
				groups = {{.Groups}}
			{{- end }}
			}`
		tmpl, err := template.New("resource_user_groups").Parse(s)
		if err != nil {
			panic(err)
		}
		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, urmp)
		if err != nil {
			panic(err)
		}
		return buf.String()
	}
	return urmp.HclType + ` "cloudfoundry_user_groups" ` + urmp.HclObjectName + ` {}`
}

func TestUserResourceGroups_Configure(t *testing.T) {
	t.Parallel()
	var (
		userGUID       = "ad6bb1e0-05f6-4440-9485-6fe20b38c500"
		user2GUID      = "f417c4fb-d555-4c36-86fb-20d24b5da496"
		origin         = "uaa"
		groups         = `["cloud_controller.admin", "scim.read", "scim.write"]`
		groupsUpdated1 = `["cloud_controller.admin", "scim.read"]`
		groupsUpdated2 = `["cloud_controller.admin", "scim.write"]`
		invalidGroups  = `["some_random.group"]`
		groups2        = `["cloud_controller.admin"]`
		groupsUpdated3 = `["cloud_controller.admin", "some_random.group"]`
		groupsUpdated4 = `["cloud_controller.admin", "scim.write"]`
		resourceName   = "cloudfoundry_user_groups.us"
	)
	t.Run("happy path - create/update/delete user group bindings", func(t *testing.T) {

		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_user_groups_crud")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclResourceUserGroups(&UserGroupsResourceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "us",
						User:          &userGUID,
						Origin:        &origin,
						Groups:        &groups,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "user", userGUID),
						resource.TestCheckResourceAttr(resourceName, "origin", origin),
						resource.TestCheckResourceAttr(resourceName, "groups.#", "3"),
					),
				},
				{
					Config: hclProvider(nil) + hclResourceUserGroups(&UserGroupsResourceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "us",
						User:          &userGUID,
						Origin:        &origin,
						Groups:        &groupsUpdated1,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "user", userGUID),
						resource.TestCheckResourceAttr(resourceName, "origin", origin),
						resource.TestCheckResourceAttr(resourceName, "groups.#", "2"),
					),
				},
				{
					Config: hclProvider(nil) + hclResourceUserGroups(&UserGroupsResourceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "us",
						User:          &userGUID,
						Origin:        &origin,
						Groups:        &groupsUpdated2,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "user", userGUID),
						resource.TestCheckResourceAttr(resourceName, "origin", origin),
						resource.TestCheckResourceAttr(resourceName, "groups.#", "2"),
					),
				},
			},
		})
	})
	t.Run("error path - invalid create/update scenarios", func(t *testing.T) {

		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_user_groups_invalid")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclResourceUserGroups(&UserGroupsResourceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "us",
						User:          &user2GUID,
						Origin:        &origin,
						Groups:        &invalidGroups,
					}),
					ExpectError: regexp.MustCompile(`API Error Fetching UAA Group`),
				},
				{
					Config: hclProvider(nil) + hclResourceUserGroups(&UserGroupsResourceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "us",
						User:          &user2GUID,
						Origin:        &origin,
						Groups:        &groups2,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "user", user2GUID),
						resource.TestCheckResourceAttr(resourceName, "origin", origin),
						resource.TestCheckResourceAttr(resourceName, "groups.#", "1"),
					),
				},
				{
					Config: hclProvider(nil) + hclResourceUserGroups(&UserGroupsResourceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "us",
						User:          &user2GUID,
						Origin:        &origin,
						Groups:        &groupsUpdated3,
					}),
					ExpectError: regexp.MustCompile(`API Error Fetching UAA Group`),
				},
				{
					Config: hclProvider(nil) + hclResourceUserGroups(&UserGroupsResourceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "us",
						User:          &user2GUID,
						Origin:        &origin,
						Groups:        &groupsUpdated4,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "user", user2GUID),
						resource.TestCheckResourceAttr(resourceName, "origin", origin),
						resource.TestCheckResourceAttr(resourceName, "groups.#", "2"),
					),
				},
			},
		})
	})
}
