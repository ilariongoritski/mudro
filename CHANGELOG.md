# CHANGELOG

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
