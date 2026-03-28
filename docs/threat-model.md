# Overview
MUDRO is a microservices-first monorepo with Go services (`services/*`), a React+Vite frontend (`frontend/`), CLI tooling (`tools/*`), and ops/deploy assets (`ops/*`). The runtime is a mix of new services (feed-api, auth-api, api-gateway, bff-web, casino, agent, bot, orchestration-api, movie-catalog) and legacy/transition code. Postgres is the primary datastore; Redis is used for API rate limiting; Kafka is optional for agent task events. Media files are served directly from a filesystem directory (`/media`) or via Nginx alias/MinIO. A serverless/Vercel adapter (`api/index.go`, `pkg/vercelapi`) exposes the feed API without auth.

Security-relevant functionality includes JWT-based authentication and admin roles, Telegram WebApp authentication for the casino, a Telegram bot that can trigger local operational commands, a provably fair casino ledger, and a local Opus gateway that can execute Claude Code against the repo. The architecture is in flux, so validating which services are exposed in production is part of the security posture.

# Threat model, Trust boundaries and assumptions
## Trust boundaries
- **Internet clients → HTTP APIs**: feed-api (`internal/api/server.go`), auth-api (`services/auth-api`), casino (`internal/casino`), movie-catalog (`services/movie-catalog`), API gateway/BFF (`services/api-gateway`, `services/bff-web`), orchestration proxy (`services/orchestration-api`), serverless handler (`api/index.go`).
- **Browser → API**: browser-based clients store JWTs in localStorage and send `Authorization: Bearer` headers; CORS is permissive in some services.
- **Telegram → Bot/Miniapp**: Telegram updates are untrusted; WebApp `initData` is validated via HMAC with the bot token.
- **Service-to-service**: gateways and BFF proxy to upstream services and forward auth headers.
- **Filesystem/DB boundary**: static media, import/export files, `.codex` logs, and DB data are trusted by services but may include untrusted content.
- **Third-party APIs**: OpenAI and Anthropic/Claude receive prompts and repo metadata from the bot and Opus gateway.

## Attacker-controlled inputs
- HTTP method/path/query/body/headers (including `X-Forwarded-For`, `Origin`, `Authorization`).
- WebSocket query parameters (`userId`).
- Telegram message text and WebApp `initData`.
- Imported files (CSV/JSON/HTML/media) used by CLI tools.
- Potentially malicious content stored in posts/comments/media that is later rendered.

## Operator-controlled inputs
- Environment variables: `JWT_SECRET`, bot tokens, `CASINO_ADMIN_KEY`, `CASINO_ALLOWED_ORIGINS`, DSNs, upstream URLs, Redis/Kafka configs.
- Nginx/compose/systemd configuration (`ops/`).
- Media root path and filesystem permissions.

## Developer-controlled inputs
- Local scripts, tests, TODO files driving agent tasks, CLI tooling.
- Build/CI settings and legacy code paths.

## Assumptions
- Production deployments set strong secrets and do not rely on dev defaults.
- Internal services are not exposed directly unless intended (e.g., casino service behind gateway).
- Opus gateway is bound to `127.0.0.1` and not publicly reachable.
- DB/Redis/Kafka are network-restricted; attackers do not have direct DB access.

# Attack surface, mitigations and attacker stories
## Major attack surfaces
1. **Public HTTP APIs**  
   - Feed API exposes `/api/posts`, `/api/front`, `/feed` HTML, and `/media` static files.  
   - Auth API provides registration/login/JWT issuance; admin endpoints require role checks.  
   - API gateway, BFF, and orchestration proxies forward auth headers to upstreams.  
   - Movie catalog API is public and unauthed.
2. **Authentication & session management**  
   - JWTs signed with HS256; user passwords hashed with bcrypt.  
   - Tokens are stored in localStorage; logout is client-side only.  
   - Telegram auth flow validates initData; auth API enforces 24h expiry.
3. **Casino runtime**  
   - Telegram WebApp initData is the auth mechanism; demo mode bypass exists for local use.  
   - Admin endpoints are guarded by a static admin key.  
   - WebSockets use `userId` query param and origin checks, but no per-user auth.  
   - Ledger uses DB transactions and idempotency keys.
4. **Telegram bot & agent worker**  
   - Bot accepts commands only from `TELEGRAM_ALLOWED_USERNAME`.  
   - Bot/agent can execute local commands (`docker compose logs`, `make test`), access repo data, and call OpenAI.  
   - Agent tasks are queued from `.codex/todo.md`; risky tasks require manual approval.
5. **Importers & tooling**  
   - CLI tools ingest Telegram/VK exports, CSVs, and media directories.  
   - Opus gateway exposes a local HTTP interface to run Claude Code; enforces path and tool allowlists.
6. **Deployment & reverse proxy**  
   - Nginx aliasing `/media` and `X-Forwarded-For` handling affect rate limiting and IP trust.

## Mitigations / robustness observed
- Parameterized SQL throughout repositories; input validation for pagination/filters.
- `config.ValidateRuntime` and required env checks prevent superuser DSNs and missing secrets.
- bcrypt password hashing and JWT expiry.
- Rate limiting in feed-api (Redis or in-memory).
- HTML is rendered with `html/template`; React auto-escapes text to reduce XSS.
- Media URLs normalized to avoid unsafe schemes.
- Casino paytable validation and provably fair RNG (`internal/casino/provablyfair.go`).
- Opus gateway enforces repo-root path checks and allowlisted Bash commands.
- Risky agent TODO items require approval before execution.

## Attacker stories (examples)
1. **JWT secret compromise or default secret** → Forge tokens to access `/api/admin/*` or `/api/orchestration/status`, potentially leaking user lists and internal operational data.  
2. **Token theft via XSS or leaked localStorage** → Attacker uses bearer token to access user data, casino balances, or admin-only endpoints (if role elevation exists).  
3. **CASINO_ADMIN_KEY leaked** → Attacker manipulates RTP profiles, reads casino stats, or alters operational state.  
4. **WebSocket userId spoofing** → Attacker subscribes to another user’s casino events if they can guess IDs; balance and game outcomes leak.  
5. **Opus gateway exposed on a network** → Remote user can read repo files or run allowlisted shell commands, potentially exfiltrating secrets.  
6. **Rate limit bypass** via spoofed `X-Forwarded-For` when not behind a trusted proxy; can amplify DoS or brute force attempts.  
7. **/media or Nginx alias misconfiguration** → Unintended file exposure if MEDIA_ROOT or alias points to sensitive directories.  
8. **Bot username misconfiguration** → Unauthorized Telegram user triggers operational commands, logs, or LLM calls (cost/DoS risk).

### Criticality in context
- CSRF is less critical because auth uses bearer tokens in headers and CORS does not allow credentials by default; it becomes higher risk only if cookies are introduced.  
- SSRF risk is limited because upstream URLs are operator-configured, not user-controlled.  
- Importer vulnerabilities mainly affect operator-run workflows; they are lower risk unless untrusted data is imported into public-facing feeds.

# Criticality calibration (critical, high, medium, low)
**Critical**
- Authentication bypass (forged JWTs, leaked JWT secret).  
- Remote code execution or arbitrary command execution via bot/agent or Opus gateway exposure.  
- Unauthorized casino ledger manipulation or RTP profile changes via leaked admin key.  
- Direct access to DB/secret files through misconfigured media roots or serverless handlers.

**High**
- Privilege escalation to admin role or access to `/api/admin/*`.  
- Cross-user data access (e.g., WebSocket userId spoofing).  
- Persistent XSS leading to token theft in the browser.  
- Exposure of Telegram bot token or initData validation bypass.

**Medium**
- Brute-force login due to missing auth rate limits.  
- Replay of Telegram initData if expiry is not enforced in casino service.  
- Denial-of-service via large requests or heavy queries.  
- Excessive data leakage in orchestration status responses.

**Low**
- Misconfigured CORS or missing security headers without direct auth bypass.  
- Minor information leaks in logs or error responses.  
- Low-impact logic bugs in pagination, sorting, or rate-limiting.
