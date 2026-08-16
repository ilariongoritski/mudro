# Internal Documentation RAG

`rag-api` is internal-only: it binds to `127.0.0.1`; nginx and all frontend services must not expose it. It answers only from approved MUDRO technical documentation.

## Allowed sources

The indexer reads only `README.md`, `docs/`, `ops/runbooks/`, and `contracts/`, accepting `.md`, `.yaml`, and `.yml` files. It never reads `.env`, `env/`, `.codex/`, `data/`, logs, dumps, Telegram content, user posts, or databases.

## Retrieval policy

- The active collection is the `RAG_QDRANT_COLLECTION` alias (default `mudro_docs_current`).
- The indexer builds a timestamped physical collection and switches the alias atomically only after all approved chunks are uploaded.
- Previous physical collections are retained for rollback; do not delete them until the replacement is verified.
- Results below `RAG_MIN_SCORE` (default `0.65`) are rejected with HTTP 422 instead of generating an unsupported answer.
- Each chunk is tagged `corpus=internal_docs`, `visibility=internal`, `content_hash`, and `indexed_at`.

## Run

1. Set `RAG_LLM_API_KEY` outside Git. The selected OpenAI-compatible provider must support both chat completions and embeddings.
2. Start the internal services:

```bash
docker compose -f docker-compose.prod.yml up -d qdrant rag-api
```

3. Build and activate a new documentation index:

```bash
docker compose -f docker-compose.prod.yml --profile rag-tools run --rm rag-indexer
```

4. Verify process liveness and retrieval readiness locally:

```bash
curl -fsS http://127.0.0.1:8092/healthz
curl -fsS http://127.0.0.1:8092/readyz
```

5. Ask only locally:

```bash
curl --request POST http://127.0.0.1:8092/internal/rag/ask \
  --header 'Content-Type: application/json' \
  --data '{"question":"How do I run runtime health checks?"}'
```

A successful response contains source paths. Missing or weakly relevant documentation returns HTTP 422.

## Rollback

Record the previous physical collection before reindexing. To roll back, atomically repoint `mudro_docs_current` to that previous collection with Qdrant's `POST /collections/aliases` batch API, then confirm `GET /readyz` returns `200`.
