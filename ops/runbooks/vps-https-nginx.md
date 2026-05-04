# VPS HTTPS / Nginx Runbook

## Scope

Production entrypoint:
- `https://<domain>/` serves `frontend/dist` from `/var/www/mudro/frontend`.
- `https://<domain>/healthz`, `/api/*`, `/media/*` proxy to `127.0.0.1:8080`.
- HTTP remains open only for ACME challenge and redirect to HTTPS after cert issuance.

Do not store domain credentials, registrar credentials, API tokens, or private keys in the repository.

## Prerequisites

- DNS `A` record for `<domain>` points to the VPS public IP.
- Ports `80/tcp` and `443/tcp` are reachable from the internet.
- `mudro-api` is healthy on loopback:
  - `curl -fsS http://127.0.0.1:8080/healthz`
- Frontend is built and present at `/var/www/mudro/frontend`.

## Rollout

From the VPS:

```bash
cd /root/projects/mudro
export MUDRO_DOMAIN='<domain>'
export MUDRO_LETSENCRYPT_EMAIL='<ops email>'
bash ops/scripts/install_vps_https.sh
```

Firewall changes are gated. After explicit approval to change UFW rules, run:

```bash
export MUDRO_CONFIRM_FIREWALL=1
bash ops/scripts/install_vps_https.sh
```

The script:
- installs `nginx` and `certbot`;
- installs the HTTP bootstrap config from `ops/nginx/mudro.conf`;
- requests the Let's Encrypt certificate with webroot challenge;
- renders `ops/nginx/mudro.https.conf.template`;
- reloads nginx;
- opens `80/tcp` + `443/tcp` in UFW only when `MUDRO_CONFIRM_FIREWALL=1`.

## Smoke Checklist

Run from any external machine after DNS propagation:

```bash
curl -I https://<domain>/healthz
curl -I "https://<domain>/api/front?limit=1"
curl -fsS "https://<domain>/api/front?limit=1"
curl -I https://<domain>/
curl -I https://<domain>/feed
```

Expected:
- `/healthz` returns `200`.
- `/api/front?limit=1` returns `200` for headers check and JSON with at least one feed item in a populated runtime.
- `/` returns `200` and includes security headers:
  - `Strict-Transport-Security`
  - `X-Content-Type-Options`
  - `X-Frame-Options`
  - `Referrer-Policy`
  - `Content-Security-Policy`
- Frontend route refresh works:
  - open `https://<domain>/feed`;
  - refresh the browser page;
  - expected result is the React page, not nginx `404`.
- HTTP redirects:
  - `curl -I http://<domain>/` returns `301` to `https://<domain>/`.

## Renewal Check

```bash
systemctl list-timers --all | grep -E 'certbot|snap.certbot'
certbot renew --dry-run
nginx -t
```

If renewal fails, keep the existing certificate in place, inspect `/var/log/letsencrypt/letsencrypt.log`, and verify that `/.well-known/acme-challenge/` is not blocked by firewall or DNS drift.
