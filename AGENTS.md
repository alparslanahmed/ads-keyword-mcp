# AGENTS.md

## Project purpose

This repository provides a local Go MCP server named `keyword-market`. It uses
Google Ads Keyword Planner data to research product ideas, discover keywords,
compare ideas, cluster keywords, and score opportunities.

The default market for this workspace is Turkey / Turkish unless the user asks
for another supported country or language.

## Current local status (2026-07-11)

- The MCP server builds and all Go tests pass on Windows.
- Codex CLI was updated to `0.144.1` during setup.
- Google OAuth desktop credentials and a refresh token are configured locally.
- A Google Ads test manager account is configured as the target account.
- Live `GenerateKeywordIdeas` calls succeeded and returned real Keyword Planner
  search-volume, competition, and bid data.
- A real Codex MCP invocation succeeded with run ID
  `163091ac-a0f7-4059-b3b8-adbc83838d02`.
- The global Codex MCP entry `keyword-market` is installed in live mode through
  `scripts/start-mcp.ps1 -Live`.
- `GOOGLE_ADS_LOGIN_CUSTOMER_ID` is intentionally empty because the authorized
  Google user has direct access to the configured test manager account. Do not
  add a login customer ID unless calls are made through a different manager.
- The Google Ads REST API default is `v24`.

Do not assume the current Codex/ChatGPT Desktop process has reloaded MCP config.
After installation or config changes, restart the client and start a new chat.

## Secrets and credential safety

- `.env` contains live credentials and is ignored by Git.
- OAuth client JSON files matching
  `client_secret_*.apps.googleusercontent.com.json` are ignored by Git.
- `credentials.json` is also ignored by Git.
- Never print, paste, log, commit, or include the following values in responses:
  developer token, OAuth client secret, refresh token, or access token.
- Never include full Google Ads customer IDs in documentation or chat output.
  Mask them if diagnostic output is necessary.
- The developer token was previously visible in a screenshot. Treat only the
  reset/replacement value currently stored in `.env` as valid.
- Google Cloud application-default credentials are stored outside the repo at
  `%APPDATA%\gcloud\application_default_credentials.json`.
- Google Cloud CLI is installed under
  `%LOCALAPPDATA%\Google\Cloud SDK\google-cloud-sdk\bin\gcloud.cmd`.
- `scripts/import-google-oauth.ps1` securely copies OAuth client credentials and
  the refresh token into `.env` without printing their values.

When checking `.env`, report only whether each field is `SET` or `MISSING`.

## Required local configuration

The ignored `.env` must contain these fields:

```dotenv
GOOGLE_ADS_DEVELOPER_TOKEN=<secret>
GOOGLE_ADS_CLIENT_ID=<secret>
GOOGLE_ADS_CLIENT_SECRET=<secret>
GOOGLE_ADS_REFRESH_TOKEN=<secret>
GOOGLE_ADS_LOGIN_CUSTOMER_ID=
GOOGLE_ADS_CUSTOMER_ID=<test or production customer ID>
GOOGLE_ADS_MOCK=false
GOOGLE_ADS_API_VERSION=v24
```

Do not replace `.env` with `.env.example`; that would erase the working local
credentials.

## Build and test

Run from the repository root in PowerShell:

```powershell
go test ./...
go build -o bin\keyword-market-mcp.exe .\cmd\keyword-market-mcp
go build -o bin\research.exe .\cmd\research
```

Direct live smoke test, bypassing all cached results:

```powershell
$env:GOOGLE_ADS_MOCK='false'
.\bin\research.exe `
  --idea "katlanabilir kamp masasi" `
  --country Turkey `
  --language Turkish `
  --max 20 `
  --force-refresh
```

A live test is successful only when the log says `using real Google Ads adapter`
and the command exits with code 0. Do not accept a cached result as proof of a
working Google Ads connection.

## Codex / ChatGPT Desktop MCP setup

Install or refresh the global MCP entry in live mode:

```powershell
.\scripts\install-codex.ps1 -Live
codex mcp get keyword-market
```

Expected MCP properties:

- transport: `stdio`
- launcher: `scripts/start-mcp.ps1 -Live`
- `default_tools_approval_mode: approve`

The ChatGPT desktop app, Codex app, Codex CLI, and IDE extension on the same host
share Codex MCP configuration. Restart the client after reinstalling the MCP.
ChatGPT web cannot directly start this local stdio process; web use would require
a separate HTTPS Streamable HTTP MCP endpoint.

Example prompt for a new chat:

> Use the keyword-market MCP to validate this product idea in Turkey and Turkish,
> force-refresh the data, then summarize demand, competition, commercial intent,
> top clusters, risks, and next experiments: portable solar panel for campers.

Useful MCP tools:

- `validate_product_idea`
- `discover_keywords`
- `get_keyword_metrics`
- `score_keyword_opportunities`
- `cluster_keywords`
- `compare_product_ideas`
- `export_research_run`

## Mock mode

Credential-free mock mode can be installed with:

```powershell
.\scripts\install-codex.ps1
```

Mock mode is for protocol and UI testing only. Never describe mock output as
live Google Ads research. Return to live mode afterward with:

```powershell
.\scripts\install-codex.ps1 -Live
```

## Important cache warning

Mock and live modes currently use the same SQLite database and cache-key scheme.
A previously cached mock response can therefore be returned after switching to
live mode. During connection tests and important research:

- use `forceRefresh: true` in MCP tool arguments, or
- use `--force-refresh` with the research CLI.

The correct future code fix is to partition cache keys by adapter/data source
(`mock` versus `google_ads`). Until that is implemented and tested, always force
a refresh when verifying live data.

## Google Ads account troubleshooting

- `DEVELOPER_TOKEN_NOT_APPROVED` means a test-access developer token was used
  against a production account. Use a genuine test account or obtain Basic /
  Standard API access.
- `CUSTOMER_NOT_FOUND` usually means the configured customer ID is wrong or an
  inaccessible `login-customer-id` header was supplied.
- `ListAccessibleCustomers` only lists accounts directly accessible by the OAuth
  user; it does not automatically list every child under a manager.
- Actual test accounts have `customer.test_account = true`. The phrase “Test
  Account Access” in API Center describes the developer token access level, not
  whether an arbitrary customer account is a test account.
- When changing accounts, first test with a process-level override and
  `--force-refresh`; update `.env` only after the API call succeeds.

## Working conventions for future agents

- Read this file and `README.md` before changing setup or credentials.
- Preserve the user's existing `.env`, database, and unrelated worktree changes.
- Use `apply_patch` for repository file edits.
- Keep logs on stderr; stdout is reserved for MCP JSON-RPC responses.
- After code changes, run `gofmt` on modified Go files and `go test ./...`.
- Do not commit, stage, reset, delete databases, or rotate credentials unless the
  user explicitly requests it.
- Keyword volume is directional evidence, not proof of sales or product-market
  fit. State this limitation in research conclusions.
