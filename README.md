# keyword-market-mcp

A Go-based **Model Context Protocol (MCP) server** for validating product ideas
using Google Ads Keyword Planner data — keyword demand, competition, commercial
intent, and market-opportunity signals.

It is **local-first** and **API/data-driven** (no LLM inside the server). Connect
it to an AI assistant (Claude, ChatGPT, Cursor, Codex, …) and ask things like:

> "Validate this product idea: a portable espresso maker for campers in Turkey."

The assistant can then call tools that answer:

- Is there search demand?
- Which related keywords are people searching, and at what volumes?
- Is competition low, medium, or high?
- Are there commercial-intent and long-tail keywords?
- Which product angles look promising, and which risks should be considered?
- What should be researched next?

> Keyword volume is a **directional signal**, not proof of sales or
> product-market fit. The server's wording is deliberately careful about this.

---

## 1. What this project does

`keyword-market-mcp` turns a free-text product idea into structured market
signals:

1. Derives seed keywords from the idea (deterministic heuristics).
2. Calls Google Ads `GenerateKeywordIdeas` (and `GenerateKeywordHistoricalMetrics`).
3. Normalizes, deduplicates, scores, classifies intent, and clusters keywords.
4. Produces a verdict (`strong` / `promising` / `weak` / `inconclusive`) with
   demand, competition, commercial-intent, and opportunity scores.
5. Persists every run in SQLite so it can be re-read, compared, and exported.

A **mock adapter** lets the whole server run with **no credentials**, which is
ideal for local development, demos, and tests.

## 2. Why MCP is used

MCP is the open protocol that lets AI assistants call external **tools**, read
**resources**, and use **prompts**. By exposing this functionality as an MCP
server, any MCP-compatible assistant can validate product ideas conversationally,
without bespoke integrations. The server is the product; the HTTP API is only for
health checks and local debugging.

## 3. Tech Stack

* Go
* Fiber
* GORM
* SQLite
* Model Context Protocol
* Google Ads API
* slog or zerolog
* testify

(This implementation uses Go's standard-library `slog` for logging and a pure-Go
SQLite driver, so no CGO toolchain is required.)

## 4. Requirements

- Go 1.24+
- For live data: a Google Ads API developer token and OAuth credentials.
- For mock mode: nothing beyond Go.

## 5. Google Ads API credential setup

To use real data you need a Google Ads API setup:

1. A Google Ads **Developer Token** (from a Google Ads Manager account →
   API Center).
2. An OAuth 2.0 **Client ID** and **Client Secret** (Google Cloud Console).
3. A **Refresh Token** for an account authorized to use the Keyword Planner.
4. The **Customer ID** to query, and (if using a manager account) the
   **Login Customer ID**.

References:
- Google Ads API: https://developers.google.com/google-ads/api/docs/start
- OAuth setup: https://developers.google.com/google-ads/api/docs/oauth/overview

Put these in `.env` (see below). Credentials stay server-side and are **never**
returned through MCP responses, resources, logs, or debug endpoints.

## 6. Environment variables

Copy `.env.example` to `.env` and fill in values:

| Variable | Default | Description |
| --- | --- | --- |
| `GOOGLE_ADS_DEVELOPER_TOKEN` | — | Google Ads developer token |
| `GOOGLE_ADS_CLIENT_ID` | — | OAuth client ID |
| `GOOGLE_ADS_CLIENT_SECRET` | — | OAuth client secret |
| `GOOGLE_ADS_REFRESH_TOKEN` | — | OAuth refresh token |
| `GOOGLE_ADS_LOGIN_CUSTOMER_ID` | — | Manager account ID (optional) |
| `GOOGLE_ADS_CUSTOMER_ID` | — | Target customer ID |
| `GOOGLE_ADS_MOCK` | `false` | `true` to use deterministic mock data |
| `GOOGLE_ADS_API_VERSION` | `v18` | Google Ads REST API version |
| `DATABASE_DSN` | `keyword_market.db` | SQLite path |
| `LOG_LEVEL` | `info` | `debug`/`info`/`warn`/`error` |
| `APP_ENV` | `development` | `development`/`test`/`production` |
| `KEYWORD_PLANNING_RATE_LIMIT_RPS` | `1` | Keyword-planning calls per second |
| `CACHE_KEYWORD_IDEAS_DAYS` | `7` | TTL for keyword ideas |
| `CACHE_HISTORICAL_METRICS_DAYS` | `30` | TTL for historical metrics |
| `MCP_SERVER_NAME` | `keyword-market` | MCP server name |
| `MCP_SERVER_VERSION` | `0.1.0` | MCP server version |
| `HTTP_PORT` | `8080` | Debug HTTP server port |
| `DEBUG_API_KEY` | — | Protects `/debug/*` outside development |

Secrets are never printed or logged.

## 7. Install instructions

```bash
git clone https://github.com/alparslanahmed/keyword-market-mcp
cd keyword-market-mcp
cp .env.example .env       # then edit, or set GOOGLE_ADS_MOCK=true
go mod download
make build                 # produces ./bin/keyword-market-mcp and ./bin/research
```

## 8. SQLite database setup

The database is created automatically on first run via GORM `AutoMigrate`. To run
migrations explicitly:

```bash
make migrate
```

The DB path is controlled by `DATABASE_DSN` (default `keyword_market.db`).

## 9. Running in stdio MCP mode

stdio is the primary transport for local MCP clients. JSON-RPC messages are read
from stdin and responses written to stdout; **all logs go to stderr** so the
JSON-RPC stream is never corrupted.

```bash
# Mock mode (no credentials):
GOOGLE_ADS_MOCK=true ./bin/keyword-market-mcp --stdio

# Or via make:
make run
```

## 10. Running the debug HTTP server (Fiber)

```bash
./bin/keyword-market-mcp --http        # or: make run-http
```

Endpoints:

| Method | Path | Description |
| --- | --- | --- |
| GET | `/health` | Liveness check |
| GET | `/version` | Server name/version |
| GET | `/debug/recent-runs` | Recent research runs |
| GET | `/debug/cache-stats` | Cache statistics |
| POST | `/debug/validate-product-idea` | Same input as the MCP tool |

Outside `APP_ENV=development`, `/debug/*` requires an `X-API-Key` header matching
`DEBUG_API_KEY` (and is disabled if no key is set).

```bash
curl localhost:8080/health
curl -X POST localhost:8080/debug/validate-product-idea \
  -H 'Content-Type: application/json' \
  -d '{"idea":"portable espresso maker for campers","country":"Turkey","language":"Turkish"}'
```

## 11. Connecting to MCP clients

Build the binary (`make build`) and point your MCP client at its absolute path
with the `--stdio` argument. Provide credentials via the `env` block (or set
`GOOGLE_ADS_MOCK` to `true`).

## 12. Example MCP client config

```json
{
  "mcpServers": {
    "keyword-market": {
      "command": "/absolute/path/to/keyword-market-mcp",
      "args": ["--stdio"],
      "env": {
        "GOOGLE_ADS_DEVELOPER_TOKEN": "...",
        "GOOGLE_ADS_CLIENT_ID": "...",
        "GOOGLE_ADS_CLIENT_SECRET": "...",
        "GOOGLE_ADS_REFRESH_TOKEN": "...",
        "GOOGLE_ADS_LOGIN_CUSTOMER_ID": "...",
        "GOOGLE_ADS_CUSTOMER_ID": "...",
        "DATABASE_DSN": "/absolute/path/to/keyword_market.db"
      }
    }
  }
}
```

## 13. Example tool calls

**Tools:** `validate_product_idea`, `discover_keywords`, `get_keyword_metrics`,
`score_keyword_opportunities`, `cluster_keywords`, `compare_product_ideas`,
`export_research_run`.

**Resources:** `research-runs://recent`, `research-run://{runId}`,
`keyword-cache://stats`, `config://targeting`.

**Prompts:** `product_validation_brief`, `product_idea_comparison`,
`keyword_opportunity_analysis`.

All tool responses use a standard envelope:

```json
{ "ok": true, "data": { } }
```
```json
{ "ok": false, "error": { "code": "string", "message": "string", "retryable": true } }
```

CLI (validates via the same service the MCP tool uses):

```bash
go run ./cmd/research --idea "portable espresso maker for campers" --country Turkey --language Turkish
GOOGLE_ADS_MOCK=true go run ./cmd/research --idea "portable espresso maker for campers"
```

`tools/call` example (`validate_product_idea`):

```json
{
  "jsonrpc": "2.0", "id": 1, "method": "tools/call",
  "params": {
    "name": "validate_product_idea",
    "arguments": { "idea": "portable espresso maker for campers", "country": "Turkey", "language": "Turkish" }
  }
}
```

## 14. Cache behavior

- Keyword ideas are cached for `CACHE_KEYWORD_IDEAS_DAYS` (default 7).
- Historical metrics are cached for `CACHE_HISTORICAL_METRICS_DAYS` (default 30).
- Cache keys incorporate method, normalized+sorted keywords/seeds, URL, country,
  language, and network.
- Every tool that hits the API accepts `forceRefresh: true` to bypass the cache.
- If a live call fails but a (possibly stale) cached entry exists, the server can
  fall back to the stale entry with a warning rather than failing the request.

## 15. Rate limit behavior

- Keyword-planning calls are serialized to `KEYWORD_PLANNING_RATE_LIMIT_RPS`
  (default 1 request/second).
- Retries use exponential backoff (up to 3 retries) for `RESOURCE_EXHAUSTED` and
  transient network/server errors.
- Invalid-credential, invalid-customer-ID, and validation errors are **not**
  retried.
- Every Google Ads call is recorded in the `ApiCallLog` table (no secrets).

## 16. Known limitations

- Seed generation and clustering are deterministic heuristics (token-based, no
  embeddings). They are isolated behind clean interfaces so embedding-based
  clustering can be added later.
- For validation, keyword-idea metrics (which already include monthly volumes)
  are used directly; a separate historical-metrics call is available via
  `get_keyword_metrics`.
- The real Google Ads adapter uses the REST API; the configured API version may
  need updating over time.
- Geo/language constants in `config/targeting.yaml` should be verified against the
  current Google Ads reference and extended as needed.

## 17. Troubleshooting

- **`MISSING_CREDENTIALS`** — set the Google Ads env vars, or set
  `GOOGLE_ADS_MOCK=true` to use mock data.
- **MCP client sees corrupted output** — ensure nothing writes to stdout; logs go
  to stderr by design. Use the provided binary and `--stdio`.
- **`INVALID_TARGETING`** — the country/language isn't in `config/targeting.yaml`.
- **Empty results** — try `forceRefresh: true`, broaden the idea, or add
  `seedKeywords`.
- **Rate limiting** — lower request volume or adjust
  `KEYWORD_PLANNING_RATE_LIMIT_RPS`.

## 18. Future roadmap

- `GenerateKeywordForecastMetrics`
- Google Trends integration
- Amazon / Etsy / Trendyol marketplace signals
- SERP / SEO-difficulty signals (where legally and technically appropriate)
- Embedding-based keyword clustering
- Streamable HTTP MCP transport (the Fiber server is already in place for this)

---

## Development

```bash
make test     # run tests (no credentials required)
make lint     # go vet + gofmt check
make build    # build binaries into ./bin
make research IDEA="reusable coffee filter"
```

## Project layout

```
cmd/keyword-market-mcp   MCP server + HTTP debug server entrypoint
cmd/research             CLI for local validation
internal/app             config, logger, wiring
internal/mcp             MCP JSON-RPC server, stdio transport, tools/resources/prompts
internal/http            Fiber debug server
internal/googleads       adapter interface, mock + real (REST) adapters, targeting
internal/validation      validation flow, verdict, insights, scoring/clustering glue
internal/scoring         opportunity & trend scoring
internal/intent          search-intent classification
internal/clustering      keyword clustering
internal/cache           SQLite-backed TTL cache
internal/ratelimit       rate limiter + retry/backoff
internal/db              GORM models, migrations, repositories
internal/export          run export (JSON/CSV)
internal/utils           normalization, money, hashing, URL, typed errors
config/targeting.yaml    country/language → Google Ads constants
testdata                 sample API payloads
tests                    cross-package tests
```
