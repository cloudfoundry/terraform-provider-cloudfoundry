# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Terraform provider for Cloud Foundry (CF) using the Terraform Plugin Framework v1 and the CF v3 API via `go-cfclient/v3`. Registry address: `registry.terraform.io/cloudfoundry/cloudfoundry`.

## Build & Development Commands

```bash
make build          # Compile
make test           # Unit tests (10m timeout, uses VCR fixtures — no live CF needed)
make testacc        # Acceptance tests (120m timeout, TF_ACC=1)
make lint           # golangci-lint
make fmt            # gofmt
make generate       # Regenerate docs (tfplugindocs) + format examples
make fix            # go fix
make lefthook       # Install pre-commit hooks (gofmt, golangci-lint, terraform fmt)
```

**Run a single test:**
```bash
go test -run ^TestResourceOrg$ github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider
```

**Run tests matching a pattern:**
```bash
go test -run ServiceCredentialBinding github.com/cloudfoundry/terraform-provider-cloudfoundry/cloudfoundry/provider
```

**Force VCR re-recording** (requires CF CLI login): set `TEST_FORCE_REC=true` before running tests. Unset it for normal test runs.

## Architecture

### Package Layout

```
main.go                              # Entry point, serves provider via RPC
cloudfoundry/provider/
├── provider.go                      # Provider definition, auth, resource/datasource registration
├── resource_*.go                    # ~24 resources (CRUD + import)
├── datasource_*.go                  # ~40 data sources (read-only)
├── types_*.go                       # Type structs + bidirectional mapper functions
├── utils.go                         # Shared schema helpers, metadata, error handling, polling
├── *_test.go                        # Tests (VCR-based)
├── managers/session.go              # CF client session creation
└── fixtures/                        # VCR cassettes (YAML, ~181 files)
internal/
├── mta/                             # Multi-Target Application client (separate HTTP client with CSRF)
├── validation/uuidvalidator.go      # Custom UUID validator
└── version/version.go               # Provider version (injected via ldflags)
```

### Resource Implementation Pattern

Every resource follows this structure:

1. **Struct** implementing `resource.Resource`, `ResourceWithConfigure`, `ResourceWithImportState`, optionally `ResourceWithIdentity`
2. **`Configure()`** receives `*Session` from provider, stores `cfClient`
3. **`Schema()`** defines attributes using shared helpers from `utils.go` (`guidSchema()`, `resourceLabelsSchema()`, `createdAtSchema()`, etc.)
4. **`Create/Read/Update/Delete`** use mapper functions from `types_*.go` to convert between Terraform types and CF API structs
5. **`ImportState()`** via `ImportStatePassthroughID` or identity-based import

Key conventions:
- Async deletions use `pollJob()` (20-min timeout, 2-sec interval)
- `handleReadErrors()` removes resources from state on 404
- Labels/annotations use `setClientMetadataForUpdate()` which removes old then adds new
- `lo.Find`, `lo.Difference`, `lo.Intersect` from samber/lo for collection operations

### Data Source Pattern

Implements `datasource.DataSource` + `DataSourceWithConfigure`. Read-only: builds list options with filters, calls CF API, uses `lo.Find()` to match, maps result to Terraform type.

### Type Mapping Convention

Each entity has mapper functions in `types_*.go`:
- `map<Entity>ValuesToType()` — CF API response → Terraform type struct
- `mapCreate<Entity>TypeToValues()` — Terraform plan → CF Create request
- `mapUpdate<Entity>TypeToValues()` — Terraform plan/state → CF Update request

### Test Pattern (VCR-based)

```go
func TestResourceOrg(t *testing.T) {
    t.Parallel()
    t.Run("happy path", func(t *testing.T) {
        cfg := getCFHomeConf()
        rec := cfg.SetupVCR(t, "fixtures/resource_org")  // unique path per test
        defer stopQuietly(rec)
        resource.Test(t, resource.TestCase{
            IsUnitTest:               true,
            ProtoV6ProviderFactories: getProviders(rec.GetDefaultClient()),
            Steps: []resource.TestStep{...},
        })
    })
}
```

- Fixtures in `cloudfoundry/provider/fixtures/` — one YAML per test, path must be unique
- Tests use `IsUnitTest: true` with VCR (no live CF environment)
- HCL generators: `hclProvider()`, `hclOrg()`, `hclSpace()`, etc. with `*ModelPtr` structs
- Regex validators: `regexpValidUUID`, `regexpValidRFC3999Format`

### Provider Authentication

Supports (in priority order): username/password, client credentials, access/refresh tokens, JWT assertion tokens, CF CLI config fallback (`~/.cf/config.json`). Environment variables: `CF_API_URL`, `CF_USER`, `CF_PASSWORD`, `CF_CLIENT_ID`, `CF_CLIENT_SECRET`.

## Conventions

- **Commits**: Conventional commits required (`fix:`, `feat:`, `refactor!:`)
- **Docs**: Never edit `docs/` manually — generated from schema via `make generate`
- **VCR fixtures**: Never edit manually. Review for leaked credentials before committing.
- **Plan modifiers**: `UseStateForUnknown()` for computed fields, `RequiresReplace()` for immutable fields
- **CI matrix**: Tests run against Terraform 1.12, 1.13, 1.14
