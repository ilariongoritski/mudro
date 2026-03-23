# Legacy MudroTop CRA

Старый standalone-контур `MudroTop` больше не считается активной архитектурой.

Активный staging-контур теперь строится как `mini-mudro`:

- `frontend/`
- `services/`
- `contracts/`
- `migrations/`
- `scripts/`
- `docs/`
- `ops/`

Если потребуется переносить старые файлы `src/`, `public/`, корневой `package.json` и старые asset-пути, делать это только как controlled migration, а не как активную базу для разработки.
