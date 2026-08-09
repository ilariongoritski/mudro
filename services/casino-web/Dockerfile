# services/casino-web/Dockerfile
FROM oven/bun:1.2-alpine AS base

WORKDIR /app

# Install dependencies
COPY package.json bun.lock ./
RUN bun install --frozen-lockfile

# Copy source
COPY . .

# Build
RUN bun run build

# Production image
FROM oven/bun:1.2-alpine AS runner
WORKDIR /app

ENV NODE_ENV=production
ENV PORT=3000

COPY --from=base --chown=bun:bun /app/.next/standalone ./
COPY --from=base --chown=bun:bun /app/.next/static ./.next/static
COPY --from=base --chown=bun:bun /app/public ./public

EXPOSE 3000

USER bun
HEALTHCHECK --interval=30s --timeout=5s --start-period=30s --retries=3 CMD bun -e "fetch('http://127.0.0.1:3000').then(r => process.exit(r.ok ? 0 : 1)).catch(() => process.exit(1))"
CMD ["bun", "server.js"]
