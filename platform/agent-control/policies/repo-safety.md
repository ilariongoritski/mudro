# Repo Safety Policy

- Не коммитить секреты, `.env`, локальные дампы и generated artifacts.
- Не делать destructive git-операции без явного подтверждения.
- Любой неиспользуемый/устаревший код переносить в `legacy/old/*` с записью в `legacy/old/manifest.yaml`.
- Каноничная активная структура: `services/*`, `tools/*`, `ops/*`, `platform/agent-control/*`.
