# Contributing

We thank you for taking an interest in our project and contributing to it. Contributions can be in the form of

- raising issues for bugs or feature requests or starting discussions (see [Contributions via issues or discussions](#contributions-via-issues-or-discussions))
- code contributions via PRs to fix bugs, implement new features or improving the documentation (see [Code Contributions](#code-contributions))

In the following sections you will find guidelines on how to contribute effectively to our project focusing on the areas mentioned above.

## Contributions via Issues or Discussions

- We use GitHub issues to track [bugs](https://github.com/cloudfoundry/terraform-provider-cloudfoundry/issues/new?assignees=&labels=bug%2Cneeds-triage&projects=&template=bug_report.yml&title=%5BBUG%5D) and [feature requests](https://github.com/cloudfoundry/terraform-provider-cloudfoundry/issues/new?assignees=&labels=enhancement%2Cneeds-triage&projects=&template=feature_request.yml&title=%5BFEATURE%5D).
- Please provide as much context as possible when you open an issue to help us understand and reproduce the issue.
- You can also start discussions in the [discussions section](https://github.com/cloudfoundry/terraform-provider-cloudfoundry/discussions/) to discuss ideas or ask questions. Sharing is caring!

## Code Contributions

### Prerequisites

Before starting:
- Go installed (see go.mod for minimum version).
- Terraform (versions covered by test matrix; latest stable recommended).
- Cloud Foundry CLI (`cf`) authenticated to an environment you are allowed to use for recordings.
- Make (to run helper targets).

All contributors must have a signed CLA. The EasyCLA bot will guide you in your first PR.

### Propose or Claim Work

1. To work on an existing issue, comment “I will take this” (or similar).
1. To propose new work, open an issue describing problem + proposed approach.
1. Maintainers confirm scope. If declined, rationale will be added in the issue.

> [!NOTE]
> You can of course also implement your change first and open a PR. However, in that case, the maintainers might decide not to merge your PR if they disagree with the proposed change. To avoid duplicated efforts, we recommend following the process above.

### Contributing Code

You are welcome to contribute code to fix a bug or to implement a new feature that is logged as an issue.

Please be aware that all contributors to this project must have a signed Contributor License Agreement (**"CLA"**) on file with us. The CLA grants us the permissions we need to use and redistribute your contributions as part of the project. Before a PR can pass all required checks, our CLA action will prompt you to accept the agreement.

You can sign the CLA by heading over to the [Linux Foundation EasyCLA](https://api.easycla.lfx.linuxfoundation.org/v2/repository-provider/github/sign/1797134/394751388/618/#/?version=2).

#### Code Contribution Workflow

In the following sections we describe the workflow to contribute code to this project. Please follow these steps to ensure consistency and smooth reviews.

##### Fork the repository

1. First you must fork the provider repository into your organization or personal account Set up your fork and branch
1. Do not make changes directly to the `main` branch. Instead, create a new branch for your changes:

   ```bash
   git checkout -b my-branch-for-the-issue
   ```

> [!NOTE]
> Make sure to regularly sync your fork with main repository and rebase your branch. This will help to avoid merge conflicts when you open a pull request.

##### Deliverables per type of change

Depending on the type of change you are making, you might need to update different parts of the codebase (e.g., resources, data sources, tests, documentation). The following table summarizes the deliverables for different types of changes and links them to the relevant sections in this document.

| Type of Change                | Deliverables                                                                                     | Relevant Sections                      |
|-------------------------------|--------------------------------------------------------------------------------------------------|---------------------------------------|
| Bug Fix                       | Code changes, update/enhancement of unit test and recording of VCR test fixtures, generation of documentation  | [Implement your changes](#implement-your-changes), [Add/Update tests](#addupdate-tests), [Add/update documentation](#addupdate-documentation) |
| New Feature                   | Code changes, update/enhancement of unit test and recording of VCR fixtures, generation of documentation  | [Implement your changes](#implement-your-changes) [Implement your changes](#implement-your-changes), [Add/Update tests](#addupdate-tests), [Add/update documentation](#addupdate-documentation) |
| Documentation Improvement     | Update of documentation, generation of documentation                                                   | [Add/update documentation](#addupdate-documentation) |


If your contribution involves a new Cloud Foundry entity, please ensure you deliver the following components:

- **Resources**
    - Full CRUD support namely `create`, `read`, `update` (if possible) and `delete`
    - Import functionality (if possible)
    - Unit tests
    - Test fixtures via VCR
    - Documentation

- **Data Sources**
    - Data sources for the new entity (get and list).
    - Unit tests
    - Test fixtures via VCR
    - Documentation

> [!NOTE]
> We follow the best practices for writing Terraform providers as outlined in the [Terraform Plugin Development documentation](https://developer.hashicorp.com/terraform/plugin/framework). Please refer to this documentation for guidance on generic information on implementing resources and data sources.

##### Implement your changes

> [!NOTE]
> All commit messages of your changes should follow the [conventional commit](https://www.conventionalcommits.org/en/v1.0.0/) format.

Depending on the type of change implement the necessary code changes in your branch.

Once you’re done make sure that the code is properly formatted and linted. We provide a make file for that. Run the following commands to format and lint your code:

```bash
make fmt
make lint
```

> [!IMPORTANT]
> If you are heavily using tools like GitHub Copilot to generate code, please make sure to review the generated code carefully. While these tools can be very helpful, they might also introduce subtle bugs or security issues if not properly vetted.

##### Add/Update tests

Depending on your type of change, you might need to add or update tests to cover your changes. In general tests are required for all new logic.

- Unit tests are written using the standard Go testing framework.
- Acceptance tests are written using the Terraform Plugin Testing framework as well as [go-vcr](https://pkg.go.dev/github.com/dnaeon/go-vcr) to record and replay Cloud Foundry API interactions.

> [!TIP]
> The recording of the CF API interactions is done via VCR fixtures. This allows us to run acceptance tests without requiring a live Cloud Foundry environment for every test run. You do not need to know the details of VCR to contribute code.

###### Writing Tests using VCR

Adding or enhancing tests for resources or data sources involves the test *per se* using the Terraform test framework as well as recording the CF API interactions using VCR.

The easiest way to get started is to look at existing tests in the `cloudfoundry/provider` package. Let us look at a simplified test for the resource [`cloudfoundry_org`](cloudfoundry/provider/resource_org_test.go).

```go
package provider

import (
   "testing"

   "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestResourceOrg(t *testing.T) {
   t.Parallel()
   t.Run("happy path - create/update/delete/import org", func(t *testing.T) {
      cfg := getCFHomeConf()
      resourceName := "cloudfoundry_org.crud_org"
      rec := cfg.SetupVCR(t, "fixtures/resource_org")
      defer stopQuietly(rec)
      resource.Test(t, resource.TestCase{
         IsUnitTest:               true,
         ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
         Steps: []resource.TestStep{
            {
               Config: hclProvider(nil) + hclOrg(&OrgModelPtr{
                  HclType:       hclObjectResource,
                  HclObjectName: "crud_org",
                  Name:          strtostrptr("tf-unit-test"),
                  Labels:        strtostrptr(testCreateLabel),
               }),
               Check: resource.ComposeAggregateTestCheckFunc(
                  resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
                  resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
                  resource.TestMatchResourceAttr(resourceName, "updated_at", regexpValidRFC3999Format),
                  resource.TestMatchResourceAttr(resourceName, "quota", regexpValidUUID),
                  resource.TestCheckResourceAttr(resourceName, "labels.purpose", "testing"),
               ),
            },
            {
               Config: hclProvider(nil) + hclOrg(&OrgModelPtr{
                  HclType:       hclObjectResource,
                  HclObjectName: "crud_org",
                  Name:          strtostrptr("tf-org-test"),
                  Labels:        strtostrptr(testUpdateLabel),
               }),
               Check: resource.ComposeAggregateTestCheckFunc(
                  resource.TestMatchResourceAttr(resourceName, "id", regexpValidUUID),
                  resource.TestMatchResourceAttr(resourceName, "created_at", regexpValidRFC3999Format),
                  resource.TestMatchResourceAttr(resourceName, "updated_at", regexpValidRFC3999Format),
                  resource.TestMatchResourceAttr(resourceName, "quota", regexpValidUUID),
                  resource.TestCheckResourceAttr(resourceName, "labels.purpose", "production"),
                  resource.TestCheckResourceAttr(resourceName, "labels.%", "2"),
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
}
```

The overall structure of the test follows the standard pattern for Terraform provider tests. The key part for recording CF API interactions is the setup of the VCR recorder given by the following lines:

```go
cfg := getCFHomeConf()
resourceName := "cloudfoundry_org.crud_org"
rec := cfg.SetupVCR(t, "fixtures/resource_org")
defer stopQuietly(rec)
```
Here we create a VCR recorder that records all CF API interactions into the specified file (`fixtures/resource_org.yaml` in this case). The recorder is then passed to the provider factory to ensure that all API calls made by the provider during the test are recorded. This is all you need to do to enable VCR recording for your tests.

> [!IMPORTANT]
> Make sure that the path to the fixture directory is unique per test. This ensures that the recorded interactions do not interfere with other tests.

###### Recording Tests using VCR

After you have written or updated tests, you need to record the CF API interactions using VCR. To ensure that the tests are properly recorded set the environment variable `TEST_FORCE_REC` to `true`. Alternatively, you can delete existing fixtures to force re-recording.

As the first recording of the tests will hit a live Cloud Foundry environment, log in to your Cloud Foundry instance using the `cf` CLI and ensure that you have the necessary permissions to create/update/delete the resources involved in your tests.

Trigger the recording by execute the tests you want to record.

As a result, the VCR test fixtures will be created/updated in fixture directory `cloudfoundry/provider/fixtures`. Do not edit the generated fixtures manually.

> [!IMPORTANT]
>Before pushing your changes to your fork review the generated test fixtures carefully to ensure that **sensitive information** (like usernames/password) is redacted in the fixtures. If you find any sensitive information, please open a [bug](https://github.com/cloudfoundry/terraform-provider-cloudfoundry/issues/new?assignees=&labels=bug%2Cneeds-triage&projects=&template=bug_report.yml&title=%5BBUG%5D) as this is an error in our test recording setup.

###### Running Tests Locally

Once you recorded your tests or to validate that your change did not break existing tests, run the tests locally.

> [!IMPORTANT] Make sure that the environment variable `TEST_FORCE_REC` is **not** set to `true` when you run the tests locally to avoid re-recording of the fixtures.

You have several options to run the tests:

- **Running all the tests**

   ```bash
   make test
   ```

- **Running specific test for a package**

      ```bash
      go test ./cloudfoundry/provider
      ```
- **Running a Single Test**

   ```bash
   go test -run ^TestServiceCredentialBindingDataSource$ github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider
   ```

- **Running Tests Matching a Pattern**

   ```bash
   go test -run ServiceCredentialBinding github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider
   ```

##### Add/Update documentation

Depending on your type of change, you might need to add or update the documentation to reflect your changes. The only place to change the documentation is the schema definition of resources and data sources in the provider code.
Once you made the changes to the schema, you need to regenerate the documentation files that will appear in the Terraform Registry. To do so we offer a make target that you must execute:

```bash
make generate
```

After this command has run successfully, the documentation files will be updated in the `docs` directory. Please review the generated documentation files to ensure that everything is correct. If you find any issues, please fix them in the schema definition and regenerate the documentation again. Do not edit the generated documentation files manually.


#### Before submitting the PR

Before submitting your pull request, please ensure the following

- Code is properly formatted and linted (`make fmt`, `make lint`)
- All new and changed tests pass locally
- Your code builds successfully (`make build`)
- Documentation is generated (`make generate`).
- Commit messages follow the [conventional commit](https://www.conventionalcommits.org/en/v1.0.0/) format.
- You have signed the Contributor License Agreement (CLA).

#### Submitting the PR

1. Push your branch to your forked repository.
1. Open a pull request against the `main` branch of the main repository.
1. We provide a PR template. Please fill it out completely to help us understand your changes and ease the review process.

Once the PR is opened several automated checks will run comprising:

- A test if the CLA has been signed.
- Checks that the project builds successfully.
- Execution of all tests via a test matrix against the currently supported Terraform versions.
- Code checks via CodeQL to identify potential security issues.
- Code linting and formatting checks.
- Validation that the documentation was properly generated.

Please address any issues that arise during these checks. If you have questions or need help, feel free to ask in the PR comments.

#### Merging the PR

Once your PR has been reviewed and approved by the project maintainers, and all checks have passed, it will be merged into the `main` branch. Thank you for your contribution!
