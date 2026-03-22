# MCP-базис для mudro

Этот набор фиксирует минимальный MCP-контур для двух сред:

1. локальный `Codex` на ПК;
2. `OpenClaw` на VPS.

`OpenClaw` по текущему плану живет только на VPS. Локальная машина остается основным местом для `Codex` и локальных MCP.

## Что входит в baseline

- `filesystem`
- `git`
- `postgres`
- `github` (готовится через PAT, по умолчанию не включается автоматически)

`Playwright` остается отдельным полезным сервером для UI, но не входит в этот baseline.

Дополнительно можно подключать `Magic` от `21st.dev`, но только локально и только как вспомогательный UI-generator. В серверный OpenClaw профиль его включать не нужно.

## Локальный Codex (Windows)

Wrapper-скрипты лежат в:

- `scripts/mcp/local/mudro-mcp-filesystem.ps1`
- `scripts/mcp/local/mudro-mcp-git.ps1`
- `scripts/mcp/local/mudro-mcp-postgres.ps1`
- `scripts/mcp/local/mudro-mcp-github.ps1`
- `scripts/mcp/local/mudro-mcp-magic.ps1`
- `scripts/mcp/local/mudro-mcp-claude-repo.ps1`

Все секреты и внешние переменные живут вне репозитория, в:

- `%USERPROFILE%\.codex\secrets\mudro-postgres-mcp.local.env`
- `%USERPROFILE%\.codex\secrets\mudro-github-mcp.local.env`
- `%USERPROFILE%\.codex\secrets\magic-21st.local.env`
- `%USERPROFILE%\.codex\secrets\mudro-claude-mcp.local.env`

### Границы доступа

- `filesystem` ограничен только корнем локального репозитория `mudro`.
- `git` ограничен тем же корнем и запускается через отдельный Docker image `mudro-mcp-git:2026.1.14`.
- `postgres` работает только через read-only MCP server.
- `github` по умолчанию должен использовать fine-grained PAT только на нужные репозитории.
- `mudro_claude_repo` ограничен корнем локального репозитория `mudro` и отправляет во внешний Claude-провайдер только явно приложенные файлы или текущий git diff.

### GitHub PAT

Рекомендуемая модель:

1. создать fine-grained PAT;
2. выбрать только нужные репозитории;
3. дать права:
   - Contents: Read-only
   - Issues: Read-only
   - Pull requests: Read-only
   - Actions: Read-only
4. записать токен в `%USERPROFILE%\.codex\secrets\mudro-github-mcp.local.env`

После этого можно включить `mudro_github` в `C:\Users\gorit\.codex\config.toml`.

### Magic (опционально)

`Magic` полезен только для локальной генерации/рефайна React UI. Для его работы нужен ключ `TWENTY_FIRST_API_KEY`, который хранится в `%USERPROFILE%\.codex\secrets\magic-21st.local.env`.

Без ключа сервер стартует, но UI generation endpoints (`fetch-ui`, `create-ui`, `refine-ui`) будут отвечать `401 Unauthorized`.

### Claude Custom Key (опционально)

`mudro_claude_repo` нужен для случаев, когда нужно спросить Claude отдельно от основного Codex-контура, но в привязке к локальному репозиторию.

Сервер дает два инструмента:

- `claude_repo_ask` — отправляет произвольный prompt и выбранные файлы из репозитория;
- `claude_repo_review_worktree` — отправляет текущие `git status` и `git diff`.

Оба инструмента поддерживают `thinking_budget_tokens` для Claude extended thinking на Anthropic-compatible endpoint.

Секреты и настройки лежат в `%USERPROFILE%\.codex\secrets\mudro-claude-mcp.local.env`:

- `MUDRO_CLAUDE_API_KEY`
- `MUDRO_CLAUDE_BASE_URL`
- `MUDRO_CLAUDE_MODEL`
- `MUDRO_CLAUDE_ANTHROPIC_VERSION`

Сервер рассчитан на Anthropic-compatible endpoint и запускается через `scripts/mcp/local/mudro-mcp-claude-repo.ps1`.

## VPS / OpenClaw

Wrapper-скрипты для VPS лежат в:

- `scripts/mcp/vps/mudro-mcp-filesystem.sh`
- `scripts/mcp/vps/mudro-mcp-git.sh`
- `scripts/mcp/vps/mudro-mcp-postgres.sh`
- `scripts/mcp/vps/mudro-mcp-github.sh`

Особенности:

- `filesystem` работает только внутри `/home/node/work/mudro11`;
- `git` использует host Docker и читает только `/root/projects/mudro`;
- `postgres` читает DSN из `/home/node/.openclaw/mcp-secrets/postgres-readonly.env`;
- `github` читает PAT из `/home/node/.openclaw/mcp-secrets/github-readonly.env`.

### Подготовка VPS bundle

На сервере bundle готовится так:

```bash
cd /root/projects/mudro
bash scripts/mcp/vps/install_openclaw_mcp_bundle.sh
```

Что делает install script:

- создает `/root/.openclaw/mcp-secrets`;
- при отсутствии создает `postgres-readonly.env` с текущим серверным DSN;
- создает шаблон `github-readonly.env.example`;
- собирает Docker image `mudro-mcp-git:2026.1.14`.

### Чего bundle не делает автоматически

- не включает GitHub MCP без PAT;
- не расширяет OpenClaw доступ за пределы `mudro`;
- не дает MCP-доступ к `/root/.ssh`, Docker в целом или ко всей файловой системе.

## Проверка baseline

### Локально

- `filesystem` видит `README`, `.codex/*`, исходники;
- `git` видит только состояние репозитория;
- `postgres` проходит `select 1` и `select count(*) from posts`;
- `github` стартует только после заполнения PAT.

### На VPS

- `filesystem` должен читать только `/root/projects/mudro` через контейнерный путь `/home/node/work/mudro11`;
- `git` должен читать текущую ветку/историю `mudro`;
- `postgres` должен выполнять только read-only запросы;
- `github` в v1 должен быть read-only через fine-grained PAT.
