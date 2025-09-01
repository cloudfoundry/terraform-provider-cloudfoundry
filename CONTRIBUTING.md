# Contributing

We thank you for taking an interest in our project and contributing to it. Contributions can be in the form of raising bugs or fixing them as well. PRs and discussions for new features and improvements are welcome as well. Discussions can be done in the discussion section.

## Engaging in Our Project

We use GitHub to manage reviews of pull requests.

* If you are a new contributor, see: [Steps to Contribute](#steps-to-contribute)

* Before implementing your change, create an issue that describes the problem you would like to solve or the code that should be enhanced. Please note that you are willing to work on that issue.

* The team will review the issue and decide whether it should be implemented as a pull request. In that case, they will assign the issue to you. If the team decides against picking up the issue, the team will post a comment with an explanation.

## Steps to Contribute

Should you wish to work on an issue, please claim it first by commenting on the GitHub issue that you want to work on. This is to prevent duplicated efforts from other contributors on the same issue.

If you have questions about one of the issues, please comment on them, and one of the maintainers will clarify.

## Contributing Code or Documentation

You are welcome to contribute code in order to fix a bug or to implement a new feature that is logged as an issue.

The following rule governs code contributions:

All contributors to this project must have a signed Contributor License Agreement (**"CLA"**) on file with us. The CLA grants us the permissions we need to use and redistribute your contributions as part of the project; you or your employer retain the copyright to your contribution. Before a PR can pass all required checks, our CLA action will prompt you to accept the agreement.

Head over to the [Linux Foundation EasyCLA](https://api.easycla.lfx.linuxfoundation.org/v2/repository-provider/github/sign/1797134/394751388/618/#/?version=2).


Note: A signed CLA is required even for minor updates. If you see something trivial that needs to be fixed, but are unable or unwilling to sign a CLA, the maintainers will be happy to make the change on your behalf. If you can
describe the change in a bug report, it would be greatly appreciated.

## Code Contribution Workflow 

When contributing code, please follow these steps to ensure consistency and smooth reviews:

1. **Set up your fork and branch**  
   - Fork this repo and clone your fork locally.  
   - Create a new branch for your changes:
     ```bash
     git checkout -b feat/my-feature
     ```

2. **Run formatting and lint checks**  
   ```bash
   make fmt
   make lint
   ```
       
3. **Generate documentation**
   - Ensure resource and data source documentation is up-to-date:
     ```bash
     go generate ./...
     ```
4. **Write and update tests**
   - Unit tests for all new logic.
   - Acceptance tests for new resources/data sources.
   - Ensure tests are written so that **sensitive information (username/password) is never recorded** in fixtures.  
   - Use environment variables for credentials, and provide **redacted fallbacks** when environment variables are not set.  
   - The recommended way to authenticate is via the **Cloud Foundry CLI (`cf login`)**. The provider can reuse your existing session instead of requiring explicit username and password.
   - Example Test Snippet when using explicit username and password

```go
package provider

import (
   "os"
   "testing"
   "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDatasourceServicePlan(t *testing.T) {
   datasourceName := "data.cloudfoundry_service_plan.test"
   
   endpoint := strtostrptr(os.Getenv("TEST_CF_API_URL"))
   user := strtostrptr(os.Getenv("TEST_CF_USER"))
   password := strtostrptr(os.Getenv("TEST_CF_PASSWORD"))
   origin := strtostrptr(os.Getenv("TEST_CF_ORIGIN"))
   
   // Redact credentials if not provided
   if *endpoint == "" || *user == "" || *password == "" || *origin == "" {
      t.Logf("\nATTENTION: Using redacted user credentials since credentials not set as env.\nMake sure you are not triggering a recording else test will fail.")
      endpoint = redactedTestUser.Endpoint
      user = redactedTestUser.User
      password = redactedTestUser.Password
      origin = redactedTestUser.Origin
   }
   
   cfg := CloudFoundryProviderConfigPtr{
   Endpoint: endpoint,
   User:     user,
   Password: password,
   Origin:   origin,
   }

   t.Parallel()
   t.Run("error path - get unavailable service plan", func(t *testing.T) {
   rec := cfg.SetupVCR(t, "fixtures/datasource_service_plan_invalid")
   defer stopQuietly(rec)
   
   // test steps here
   })
}
```

5. **Record VCR fixtures**
   - We use [go-vcr](https://pkg.go.dev/github.com/dnaeon/go-vcr) to record and replay Cloud Foundry API interactions.
   - First run hits a live CF environment and records responses into fixture files.
   - Future runs replay fixtures for deterministic testing.
   - Re-record fixtures when APIs, schemas, or resources change

---
**Deliverables for New Entities**
If you are adding a new Cloud Foundry entity:
- Resources
    - Full CRUD support (Create, Read, Update, Delete).
    - Import functionality (if possible).
    - Unit + acceptance tests.
    - VCR fixtures recorded.
    - Documentation generated (go generate).

- Data Sources
    - Data source for the new entity.
    - Unit + acceptance tests.
    - VCR fixtures recorded.
    - Documentation generated. 
---
**Quick Checklist**
   - make fmt and make lint pass.
   - ```go generate ./... ```
   - Tests and fixtures updated.
   - Commit message follow the semantic commit principle (feat|fix|chore)
   - CLA signed.
   - Docs for resources/data sources included.
## Running Tests

The Cloud Foundry Terraform provider uses both **unit tests** and **acceptance tests** to validate functionality.  
Before opening a pull request, please ensure that your changes are covered by tests and that all existing tests pass.

1. **Running all the tests**  
   ```bash
   make test
   ```

2. **Running specific test for a package**
   - You can target a specific package (e.g., the provider):
      ```bash
      go test ./cloudfoundry/provider
      ```
3. **Running a Single Test**
   ```bash
   go test -timeout 30s -run ^TestServiceCredentialBindingDataSource$ github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider
   ```
   - ```-timeout 30s``` → sets a max runtime per test.
   - ```-run ^TestName$``` → matches only the specified test function.
        
4. **Running Tests Matching a Pattern**
   - To run all tests that match a substring or prefix:
      ```bash
      go test -run ServiceCredentialBinding github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider
      ```
  
## Issues and Planning

* We use GitHub issues to track bugs and enhancement requests.

* Please provide as much context as possible when you open an issue. The information you provide must be comprehensive enough to reproduce that issue for the assignee.
