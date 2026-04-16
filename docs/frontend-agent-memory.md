# Frontend Agent Memory

Updated: 2026-04-17

## Design System
- **Aesthetic**: Premium Dark Glassmorphism.
- **Tech**: React + TS + Vite + Tailwind (managed via `@tailwindcss/vite`).
- **State**: RTK Query for all API interactions.

## Key Changes (April 2026)
- **Plinko**: Added wallet synchronization logic to prevent balance drift during drops.
- **Orchestration**: Created `OrchestrationPage` as a control plane for Skaro/Opus/Codex. Dashboard URL is now configurable via `VITE_SKARO_URL`.
- **Proxy**: Vite configuration updated with `changeOrigin: true` for reliable local development connectivity.

## Pending UI Work
1. **Blackjack Integration**: Finalize the mobile-responsive UI for Blackjack TMA.
2. **Feed Socials**: Add threaded comments and reactions UI to the main feed.
3. **Error Handling**: Standardize RTK Query error boundaries across the app.
