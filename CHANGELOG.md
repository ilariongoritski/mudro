# CHANGELOG

## [Unreleased]

### Added

- Release/showcase path documented around `make health`, `make demo-up`, and `make demo-check`.
- `scripts/release-demo.sh` added as a thin helper for the showcase flow.
- Casino smoke and wallet-sync checks are now called out explicitly in the release docs.

### Changed

- README now presents a single canonical release/showcase entry path instead of mixed canonical/transitioning wording.
- Casino roulette payout/normalization updates are aligned with the current dozen bets and green payout rules.
- Local Windows build artifacts are no longer left in the working tree.

### Fixed

- Removed stale release/showcase ambiguity from the top-level documentation.
- Clarified that casino integration smoke should run against an isolated migrated casino database.

## [0.1.0-mvp] - 2026-04-04

### Added

- **Admin Panel**: Base implementation for user management at `/admin`.
- **Infrastructure Scripts**: `scripts/vps/rotate-secrets.sh` and `scripts/vps/setup-minio.sh`.
- **CI/CD**: GitHub Actions workflow for backend and frontend.
- **DevEx**: `golangci-lint` and `air` configurations.

### Changed
- **Premium UI**: Complete overhaul of Login and Register pages with dark glassmorphism.
- **Mobile Responsiveness**: Critical polish for Feed and Casino pages on mobile devices.
- **Nginx Config**: Prepared `mudro.conf` with HTTPS redirection and SSL placeholders.

### Fixed

- Unused imports and types in `feed-api` and `casino` services.
- Invalid linter configurations in `.golangci.yml`.
- Component naming consistency in `AdminPage`.
