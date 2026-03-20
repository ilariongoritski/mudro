# OpenClaw Profile

- Использовать только контуры, отраженные в `services-map.yaml`.
- Основной runtime: `ops/compose/docker-compose.core.yml`.
- Legacy контуры запускать только явной командой через `ops/compose/docker-compose.legacy.yml`.
- Любые risky операции по БД требуют явного подтверждения оператора.
