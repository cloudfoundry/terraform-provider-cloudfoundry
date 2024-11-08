# Terraform Provider for Cloud Foundry

![Golang](https://img.shields.io/badge/Go-1.23-informational)

## About This Project

The Terraform provider for [Cloud Foundry](https://www.cloudfoundry.org/) allows the management of resources via [Terraform](https://terraform.io/).

This provider makes use of the [go-cfclient](https://github.com/cloudfoundry/go-cfclient) to interact with the Cloud Foundry Cloud Controller [v3 APIs](https://v3-apidocs.cloudfoundry.org/version/3.159.0/index.html) and take advantages of the same. Additionally, the [v2 APIs are deprecated](https://apidocs.cloudfoundry.org/16.22.0/).

You can find usage examples in the [examples folder](examples/) of this repository.

Check the [Authentication](/Authentication.md) documentation for supported approaches.

## Developing & Contributing to the Provider

The [developer documentation](DEVELOPER.md) file is a basic outline on how to build and develop the provider. 

For more information about how to contribute, the project structure, and additional contribution information, see our [Contribution Guidelines](CONTRIBUTING.md).

## Prerequisites and Usage of the Provider

For the best experience using the Terraform Provider for Cloud Foundry, we recommend applying the common best practices for Terraform adoption as described in the [Hashicorp documentation](https://developer.hashicorp.com/well-architected-framework/operational-excellence/operational-excellence-terraform-maturity). For migrating usage from the existing [cloudfoundry-community](https://github.com/cloudfoundry-community/terraform-provider-cloudfoundry) provider to this provider, refer to our [migration-guide](./migration-guide/Readme.md).

## Support and Feedback

❓ - If you have a *question* you can ask it here in the [GitHub Discussions](https://github.com/cloudfoundry/terraform-provider-cloudfoundry/discussions/).

🐞 - If you find a bug, feel free to create a [bug report](https://github.com/cloudfoundry/terraform-provider-cloudfoundry/issues/new?assignees=&labels=bug%2Cneeds-triage&projects=&template=bug_report.yml&title=%5BBUG%5D).

💡 - If you have an idea for improvement or a feature request, please open a [feature request](https://github.com/cloudfoundry/terraform-provider-cloudfoundry/issues/new?assignees=&labels=enhancement%2Cneeds-triage&projects=&template=feature_request.yml&title=%5BFEATURE%5D).

## Security / Disclosure

If you find any bug that may be a security problem, please follow our [instructions](https://www.cloudfoundry.org/security/) on how to report it. Please do not create GitHub issues for security-related doubts or problems.

## Code of Conduct

Members, contributors, and leaders pledge to make participation in our community a harassment-free experience. By participating in this project, you agree to always abide by its [Code of Conduct](https://www.cloudfoundry.org/wp-content/uploads/2015/09/CFF_Code_of_Conduct.pdf).
