# Documentation RAG MVP

`rag-api` is an internal-only service for answers grounded in MUDRO technical documentation. It binds only to `127.0.0.1`; nginx and the frontend must not expose it.

## Allowed sources

The indexer reads only `README.md`, `docs/`, `ops/runbooks/`, and `contracts/`, accepting `.md`, `.yaml`, and `.yml` files. It does not read `.env`, `env/`, `.codex/`, `data/`, logs, dumps, Telegram content, user posts, or databases.

## Run

1. Set `RAG_LLM_API_KEY` outside Git. The selected OpenAI-compatible provider must support both chat completions and embeddings.
2. Start the normal stack with `docker compose -f docker-compose.prod.yml up -d qdrant rag-api`.
3. Refresh the documentation index after approved documentation changes:

```bash
docker compose -f docker-compose.prod.yml --profile rag-tools run --rm rag-indexer
```

4. Call only locally:

```bash
curl --request POST http://127.0.0.1:8092/internal/rag/ask \
  --header 'Content-Type: application/json' \
  --data '{"question":"How do I run runtime health checks?"}'
```

Every successful response includes source paths. A missing relevant source returns HTTP 422 rather than an invented answer.
