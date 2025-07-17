# Authentication Mechanisms for Cloudfoundry Terraform Provider

The cloudfoundry terraform provider supports any of the following authentication mechanism currently:

## USERNAME-PASSWORD

Use the env variables `CF_API_URL`, `CF_USER` and `CF_PASSWORD`.

Alternatively,

```hcl
provider "cloudfoundry" {
    api_url = "<CF-API-URL>"
    user = "<USER-ID>"
    password = "<PASSWORD>"
}
```

## CLIENT ID-CLIENT SECRET

Use the env variables `CF_API_URL`, `CF_CF_CLIENT_ID` and `CF_CF_CLIENT_SECRET`.

Alternatively, 

```hcl
provider "cloudfoundry" {
    api_url = "<CF-API-URL>"
    cf_client_id = "<CF-CLIENT-ID>"
    cf_client_secret = "<CF-CLIENT-SECRET>"
}
```

## OAuth JWT Assertion Bearer Flow

Use the env variable `CF_ASSERTION_TOKEN`. Typically for cloudfoundry, this login also would use a custom `origin` (unless the default `origin` is configured to support this method)

Refer [this document](https://docs.secureauth.com/ciam/en/using-jwt-profile-for-oauth-2-0-authorization-flows.html) to understand the JWT Assertion Bearer Flow.

This flow can be used in automated scenarios where the OIDC provider that is trusted by UAA has a secure means of providing assertion tokens. These tokens are short lived. 

A typical example would be using the the Open-ID Connect feature of [github](https://docs.github.com/en/actions/concepts/security/openid-connect) 
In this scenario, an `origin` in UAA would be configured to use Github OIDC as an [identity provider](https://docs.cloudfoundry.org/uaa/identity-providers.html#oidc). Refer this [blog](https://community.sap.com/t5/technology-blog-posts-by-sap/authenticating-github-actions-workflows-deploying-to-the-sap-btp-cloud/ba-p/14075047) where a similar setup is done with the cf cli. **Similarly**, the terraform provider can then use assertion tokens provided by github in a github action to login to Cloud Foundry with that specific `origin`

```hcl
provider "cloudfoundry" {
    api_url = "<CF-API-URL>"
    cf_client_id = "<CF-ASSERTION-TOKEN>"
    origin = "<CF-ORIGIN>"
}
```

## Using cf-cli configuration.

If you have installed the [cf-cli](https://docs.cloudfoundry.org/cf-cli/) and have [logged into the environment](https://docs.cloudfoundry.org/cf-cli/getting-started.html#login), the the cloudfoundry terraform provider can use the default configuration of the cf-cli (present in `~/.cf` folder) to connect to the environment.

```hcl
provider cloudfoundry {}
```

If the provider is initialized without any parameters and no environment variables are set, then the provider will try to connect this way.


