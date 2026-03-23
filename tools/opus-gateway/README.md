# Opus Gateway

Local HTTP sidecar that runs `Claude Code` SDK against this repository using your local `ANTHROPIC_API_KEY`.

## What It Does

- listens on `127.0.0.1:8788` by default
- runs `Opus` through the official `@anthropic-ai/claude-code` SDK
- defaults to `claude-opus-4-1-20250805` unless overridden by `OPUS_GATEWAY_MODEL` or `ANTHROPIC_MODEL`
- keeps the agent inside the repo root
- supports `read-only` and `edit` modes
- optionally enables a tightly allowlisted `Bash`

## Setup

```bash
cd tools/opus-gateway
npm install
```

Set environment variables:

```bash
ANTHROPIC_API_KEY=...
OPUS_GATEWAY_PORT=8788
OPUS_GATEWAY_MODEL=claude-opus-4-1-20250805
```

Run:

```bash
npm run start
```

## API

Health:

```bash
curl http://127.0.0.1:8788/healthz
```

Run a read-only task:

```bash
curl -X POST http://127.0.0.1:8788/v1/run ^
  -H "Content-Type: application/json" ^
  -d "{\"prompt\":\"Объясни структуру services/feed-api\",\"mode\":\"read-only\"}"
```

Run a task with edits:

```bash
curl -X POST http://127.0.0.1:8788/v1/run ^
  -H "Content-Type: application/json" ^
  -d "{\"prompt\":\"Обнови комментарии в tools/opus-gateway/src/server.ts\",\"mode\":\"edit\"}"
```

Enable allowlisted shell commands:

```bash
curl -X POST http://127.0.0.1:8788/v1/run ^
  -H "Content-Type: application/json" ^
  -d "{\"prompt\":\"Запусти git diff и кратко опиши изменения\",\"mode\":\"read-only\",\"allowBash\":true}"
```

## Notes

- This gateway is not a native extension point for the built-in subagents of this Codex chat.
- It is a separate local HTTP service that you run next to the repo.
- `read-only` requests run in Claude Code `plan` permission mode; `edit` requests use `acceptEdits`.
- Logs are written under `var/log/opus-gateway/`.
