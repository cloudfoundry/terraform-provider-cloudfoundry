package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestMtaResource_Configure(t *testing.T) {
	var (
		resourceName               = "cloudfoundry_mta.rs"
		spaceGuid                  = "02c0cc92-6ecc-44b1-b7b2-096ca19ee143"
		namespace                  = "test"
		mtarPath                   = "../../assets/a.cf.app.mtar"
		mtarPath2                  = "../../assets/my-mta_1.0.0.mtar"
		mtarUrl                    = "https://github.com/Dray56/mtar-archive/releases/download/v1.0.0/a.cf.app.mtar"
		extensionDescriptors       = `["../../assets/prod.mtaext","../../assets/prod-scale-vertically.mtaext"]`
		sourceCodeHash             = "fca8f8d1c499a1d0561c274ab974faf09355d513bb36475fe67577d850562801"
		bgDeploy                   = "blue-green"
		versionRuleAll             = "ALL"
		modules                    = `["my-app"]`
		extensionDescriptorsString = `[
    <<EOT
_schema-version: 3.3.0
ID: my-mta-prod
extends: my-mta
version: 1.0.0

modules:
- name: my-app
  parameters:
    instances: 2

resources:
 - name: my-service
   parameters:
     service-plan: "lite"
EOT
    ,
    <<EOT
_schema-version: 3.3.0
ID: my-mta-prod-scale-vertically
extends: my-mta-prod
version: 1.0.0

modules:
- name: my-app
  parameters:
    memory: 2G
EOT  
  ]`
	)
	t.Parallel()
	t.Run("happy path - create/update/delete mta from path", func(t *testing.T) {

		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_mta")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclResourceMta(&MtaResourceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						MtarPath:      strtostrptr(mtarPath),
						Space:         strtostrptr(spaceGuid),
						Namespace:     strtostrptr(namespace),
						VersionRule:   strtostrptr(versionRuleAll),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "mtar_path", mtarPath),
						resource.TestCheckResourceAttr(resourceName, "space", spaceGuid),
						resource.TestCheckResourceAttr(resourceName, "mta.metadata.namespace", namespace),
					),
				},
				{
					Config: hclProvider(nil) + hclResourceMta(&MtaResourceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						MtarUrl:       strtostrptr(mtarUrl),
						Space:         strtostrptr(spaceGuid),
						Namespace:     strtostrptr(namespace),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "mtar_url", mtarUrl),
						resource.TestCheckNoResourceAttr(resourceName, "mtar_path"),
						resource.TestCheckResourceAttr(resourceName, "space", spaceGuid),
						resource.TestCheckResourceAttr(resourceName, "mta.metadata.namespace", namespace),
					),
				},
				{
					Config: hclProvider(nil) + hclResourceMta(&MtaResourceModelPtr{
						HclType:              hclObjectResource,
						HclObjectName:        "rs",
						MtarPath:             strtostrptr(mtarPath2),
						Space:                strtostrptr(spaceGuid),
						Namespace:            strtostrptr(namespace),
						ExtensionDescriptors: strtostrptr(extensionDescriptors),
					}),
					ExpectError: regexp.MustCompile(`Error: New MTA ID`),
				},
			},
		})
	})

	t.Run("happy path - create/update/delete mta from url", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_mta_url")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclResourceMta(&MtaResourceModelPtr{
						HclType:              hclObjectResource,
						HclObjectName:        "rs",
						MtarPath:             strtostrptr(mtarPath2),
						Space:                strtostrptr(spaceGuid),
						Namespace:            strtostrptr(namespace),
						ExtensionDescriptors: strtostrptr(extensionDescriptors),
						VersionRule:          strtostrptr(versionRuleAll),
						Modules:              strtostrptr(modules),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "mtar_path", mtarPath2),
						resource.TestCheckResourceAttr(resourceName, "space", spaceGuid),
						resource.TestCheckResourceAttr(resourceName, "mta.metadata.namespace", namespace),
					),
				},
				{
					Config: hclProvider(nil) + hclResourceMta(&MtaResourceModelPtr{
						HclType:              hclObjectResource,
						HclObjectName:        "rs",
						MtarPath:             strtostrptr(mtarPath2),
						Space:                strtostrptr(spaceGuid),
						Namespace:            strtostrptr(namespace),
						ExtensionDescriptors: strtostrptr(extensionDescriptors),
						SourceCodeHash:       strtostrptr(sourceCodeHash),
						DeployStrategy:       strtostrptr(bgDeploy),
						VersionRule:          strtostrptr(versionRuleAll),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "mtar_path", mtarPath2),
						resource.TestCheckResourceAttr(resourceName, "space", spaceGuid),
						resource.TestCheckResourceAttr(resourceName, "mta.metadata.namespace", namespace),
						resource.TestCheckResourceAttr(resourceName, "source_code_hash", sourceCodeHash),
					),
				},
				{
					Config: hclProvider(nil) + hclResourceMta(&MtaResourceModelPtr{
						HclType:                    hclObjectResource,
						HclObjectName:              "rs",
						MtarPath:                   strtostrptr(mtarPath2),
						Space:                      strtostrptr(spaceGuid),
						Namespace:                  strtostrptr(namespace),
						ExtensionDescriptorsString: strtostrptr(extensionDescriptorsString),
						SourceCodeHash:             strtostrptr(sourceCodeHash),
						DeployStrategy:             strtostrptr(bgDeploy),
						VersionRule:                strtostrptr(versionRuleAll),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "mtar_path", mtarPath2),
						resource.TestCheckResourceAttr(resourceName, "space", spaceGuid),
						resource.TestCheckResourceAttr(resourceName, "mta.metadata.namespace", namespace),
						resource.TestCheckResourceAttr(resourceName, "source_code_hash", sourceCodeHash),
					),
				},
			},
		})
	})

	t.Run("error path - create mtar from invalid path/file", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_mta_invalid_mta_path")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclResourceMta(&MtaResourceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						Space:         strtostrptr(spaceGuid),
						MtarPath:      strtostrptr(invalidOrgGUID),
					}),
					ExpectError: regexp.MustCompile(`Unable to upload mtar file`),
				},
				{
					Config: hclProvider(nil) + hclResourceMta(&MtaResourceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						Space:         strtostrptr(spaceGuid),
						MtarPath:      strtostrptr(""),
					}),
					ExpectError: regexp.MustCompile(`Unable to upload mtar file`),
				},
				{
					Config: hclProvider(nil) + hclResourceMta(&MtaResourceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						Space:         strtostrptr(spaceGuid),
						MtarPath:      strtostrptr("../../assets/provider-config-local.txt"),
					}),
					ExpectError: regexp.MustCompile(`MTA ID missing`),
				},
				{
					Config: hclProvider(nil) + hclResourceMta(&MtaResourceModelPtr{
						HclType:              hclObjectResource,
						HclObjectName:        "rs",
						Space:                strtostrptr(spaceGuid),
						MtarPath:             strtostrptr(mtarPath),
						ExtensionDescriptors: strtostrptr(`["../../assets/pr"]`),
					}),
					ExpectError: regexp.MustCompile(`Unable to upload mta extension descriptor`),
				},
				{
					Config: hclProvider(nil) + hclResourceMta(&MtaResourceModelPtr{
						HclType:              hclObjectResource,
						HclObjectName:        "rs",
						Space:                strtostrptr(spaceGuid),
						MtarPath:             strtostrptr(mtarPath),
						ExtensionDescriptors: strtostrptr(`[""]`),
					}),
					ExpectError: regexp.MustCompile(`Unable to upload mta extension descriptor`),
				},
				{
					Config: hclProvider(nil) + hclResourceMta(&MtaResourceModelPtr{
						HclType:              hclObjectResource,
						HclObjectName:        "rs",
						Space:                strtostrptr(spaceGuid),
						MtarPath:             strtostrptr(mtarPath),
						ExtensionDescriptors: strtostrptr(`["../../assets/provider-config-local.txt"]`),
					}),
					ExpectError: regexp.MustCompile(`Failure in polling MTA operation`),
				},
			},
		})
	})
	t.Run("error path - create mtar for invalid namespace", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_mta_invalid_namespace")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclResourceMta(&MtaResourceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						MtarPath:      strtostrptr(mtarPath),
						Space:         strtostrptr(spaceGuid),
						Namespace:     strtostrptr("Hello"),
					}),
					ExpectError: regexp.MustCompile(`Failure in polling MTA operation`),
				},
			},
		})
	})
	t.Run("error path - create mtar from empty URL", func(t *testing.T) {
		cfg := getCFHomeConf()
		rec := cfg.SetupVCR(t, "fixtures/resource_mta_invalid_empty_url")
		defer stopQuietly(rec)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
			Steps: []resource.TestStep{
				{
					Config: hclProvider(nil) + hclResourceMta(&MtaResourceModelPtr{
						HclType:       hclObjectResource,
						HclObjectName: "rs",
						Space:         strtostrptr(spaceGuid),
						MtarUrl:       strtostrptr(""),
					}),
					ExpectError: regexp.MustCompile(`Unable to upload remote mtar file`),
				},
			},
		})
	})
}
