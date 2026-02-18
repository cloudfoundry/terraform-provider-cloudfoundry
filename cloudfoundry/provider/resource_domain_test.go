package provider

import (
	"bytes"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
)

type DomainModelPtr struct {
	HclType            string
	HclObjectName      string
	Name               *string
	Internal           *bool
	RouterGroup        *string
	Id                 *string
	Org                *string
	SharedOrgs         *string
	SupportedProtocols *string
	Labels             *string
	Annotations        *string
	CreatedAt          *string
	UpdatedAt          *string
}

func hclDomain(dmp *DomainModelPtr) string {
	if dmp != nil {
		s := `
		{{.HclType}} "cloudfoundry_domain" {{.HclObjectName}} {
			{{- if .Name}}
				name = "{{.Name}}"
			{{- end -}}
			{{if .Id}}
				id = "{{.Id}}"
			{{- end -}}
			{{if .Internal}}
				internal = {{.Internal}}
			{{- end -}}
			{{if .RouterGroup}}
				router_group = "{{.RouterGroup}}"
			{{- end -}}
			{{if .Org}}
				org = "{{.Org}}"
			{{- end -}}
			{{if .SharedOrgs}}
				shared_orgs = {{.SharedOrgs}}
			{{- end -}}
			{{if .SupportedProtocols}}
				supported_protocols = "{{.SupportedProtocols}}"
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
		tmpl, err := template.New("resource_domain").Parse(s)
		if err != nil {
			panic(err)
		}
		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, dmp)
		if err != nil {
			panic(err)
		}
		return buf.String()
	}
	return dmp.HclType + ` "cloudfoundry_domain "` + dmp.HclObjectName + ` {}`
}

func TestDomainResource_Configure(t *testing.T) {
	var (
		// in staging
		existingDomainName     = "cert.cfapps.stagingazure.hanavlab.ondemand.com"
		testSharedDomainName   = "test.internal"
		testPrivateDomainName  = "test.cfapps.stagingazure.hanavlab.ondemand.com"
		domainOwnerOrg         = "23919ba5-f9b6-4128-a1fb-69890818d25c"
		domainSharedOrgs       = `["537e7b58-b3e0-4464-9cad-2deae6120a80","30edf44a-2d4c-432c-9680-9a61123edcf1"]`
		domainUpdateSharedOrgs = `["ca721b24-e24d-4171-83e1-1ef6bd836b38","30edf44a-2d4c-432c-9680-9a61123edcf1"]`
		resourceName           = "cloudfoundry_domain.rs"
		testOrgGUID            = "b4da43cd-2055-4d4d-ae6e-4066ce53a8b9"
	)
	t.Parallel()
	t.Run("happy path - create/import/delete shared domain", func(t *testing.T) {

		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_domain_shared_crud")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclDomain(&DomainModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						Name:          &testSharedDomainName,
						Internal:      new(true),
						Labels:        &testCreateLabel,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestCheckResourceAttr(resourceName, "internal", "true"),
						resource.TestCheckResourceAttr(resourceName, "name", testSharedDomainName),
						resource.TestCheckResourceAttr(resourceName, "labels.purpose", "testing"),
						resource.TestCheckResourceAttr(resourceName, "supported_protocols.0", "http"),
					),
				},
				{
					Config: hclProvider(nil) + hclDomain(&DomainModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						Name:          &testSharedDomainName,
						Internal:      new(true),
						Labels:        &testUpdateLabel,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestCheckResourceAttr(resourceName, "internal", "true"),
						resource.TestCheckResourceAttr(resourceName, "name", testSharedDomainName),
						resource.TestCheckResourceAttr(resourceName, "labels.purpose", "production"),
						resource.TestCheckResourceAttr(resourceName, "supported_protocols.0", "http"),
					),
				},
				{
					ResourceName:      resourceName,
					ImportStateIdFunc: getIdForImport(resourceName),
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})

	t.Run("happy path - import with identity ", func(t *testing.T) {
		resourceName = "cloudfoundry_domain.rs"
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_domain_import_with_identity")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclDomain(&DomainModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						Name:          &testPrivateDomainName,
						Org:           &testOrgGUID,
						Labels:        &testCreateLabel,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestCheckResourceAttr(resourceName, "name", testPrivateDomainName),
						resource.TestCheckResourceAttr(resourceName, "labels.purpose", "testing"),
					),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectIdentity("cloudfoundry_domain.rs", map[string]knownvalue.Check{
							"domain_guid": knownvalue.NotNull(),
						}),
					},
				},
				{
					ResourceName:    resourceName,
					ImportState:     true,
					ImportStateKind: resource.ImportBlockWithResourceIdentity,
				},
			},
		})
	})
	t.Run("happy path - create/import/delete private domain", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_domain_private_crud")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclDomain(&DomainModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						Name:          &testPrivateDomainName,
						Org:           &domainOwnerOrg,
						SharedOrgs:    &domainSharedOrgs,
						Labels:        &testCreateLabel,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestCheckResourceAttr(resourceName, "internal", "false"),
						resource.TestCheckResourceAttr(resourceName, "name", testPrivateDomainName),
						resource.TestCheckResourceAttr(resourceName, "labels.purpose", "testing"),
						resource.TestCheckResourceAttr(resourceName, "supported_protocols.0", "http"),
						resource.TestCheckResourceAttr(resourceName, "org", domainOwnerOrg),
						resource.TestCheckResourceAttr(resourceName, "shared_orgs.#", "2"),
					),
				},
				{
					Config: hclProvider(nil) + hclDomain(&DomainModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						Name:          &testPrivateDomainName,
						Org:           &domainOwnerOrg,
						SharedOrgs:    &domainUpdateSharedOrgs,
						Labels:        &testUpdateLabel,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestCheckResourceAttr(resourceName, "name", testPrivateDomainName),
						resource.TestCheckResourceAttr(resourceName, "labels.purpose", "production"),
						resource.TestCheckResourceAttr(resourceName, "supported_protocols.0", "http"),
						resource.TestCheckResourceAttr(resourceName, "org", domainOwnerOrg),
						resource.TestCheckResourceAttr(resourceName, "shared_orgs.#", "2"),
					),
				},
				{
					Config: hclProvider(nil) + hclDomain(&DomainModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						Name:          &testPrivateDomainName,
						Org:           &domainOwnerOrg,
						Labels:        &testUpdateLabel,
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
						resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
						resource.TestCheckResourceAttr(resourceName, "name", testPrivateDomainName),
						resource.TestCheckResourceAttr(resourceName, "labels.purpose", "production"),
						resource.TestCheckResourceAttr(resourceName, "supported_protocols.0", "http"),
						resource.TestCheckResourceAttr(resourceName, "org", domainOwnerOrg),
						resource.TestCheckNoResourceAttr(resourceName, "shared_orgs"),
					),
				},
				{
					ResourceName:      resourceName,
					ImportStateIdFunc: getIdForImport(resourceName),
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})
	t.Run("error path - create domain with existing name", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_domain_invalid")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclDomain(&DomainModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						Name:          &existingDomainName,
					}),
					ExpectError: regexp.MustCompile(`API Error Creating Domain`),
				},
			},
		})
	})
}
