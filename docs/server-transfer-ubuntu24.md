# Server Transfer Runbook (Ubuntu 24.04)

Цель: быстро перенести `mudro` на новый VPS и подключить Vercel к внешней БД.

## 1) Базовая подготовка VPS

Под root (первый вход по паролю):

```bash
apt-get update
apt-get install -y ca-certificates curl git make ufw fail2ban docker.io docker-compose-plugin
systemctl enable --now docker
```

Создать рабочего пользователя:

```bash
adduser admin
usermod -aG sudo,docker admin
```

Добавить SSH-ключ пользователю `admin`:

```bash
mkdir -p /home/admin/.ssh
chmod 700 /home/admin/.ssh
cat > /home/admin/.ssh/authorized_keys <<'EOF'
ssh-ed25519 <PUBLIC_KEY> <comment>
EOF
chmod 600 /home/admin/.ssh/authorized_keys
chown -R admin:admin /home/admin/.ssh
```

## 2) Минимальный SSH hardening

```bash
sed -i 's/^#\?PasswordAuthentication.*/PasswordAuthentication no/' /etc/ssh/sshd_config
sed -i 's/^#\?PermitRootLogin.*/PermitRootLogin no/' /etc/ssh/sshd_config
systemctl restart ssh
```

Проверить вход новым пользователем:

```bash
ssh admin@<VPS_IP>
```

## 3) Firewall (черновой, но безопасный минимум)

```bash
ufw allow OpenSSH
ufw allow 80/tcp
ufw allow 443/tcp
ufw allow 5433/tcp
ufw --force enable
ufw status
```

Если БД не нужна снаружи напрямую, лучше НЕ открывать `5433/tcp`.

## 4) Деплой проекта на VPS

Под `admin`:

```bash
mkdir -p ~/projects
cd ~/projects
git clone git@github.com:ilariongoritski/mudro.git
cd mudro
```

Заполнить переменные окружения:
- `.env` (быстрый путь), либо
- `env/common.env`, `env/api.env`, `env/agent.env`, `env/reporter.env`, `env/bot.env`, `env/db.env`.

Поднять локальный контур:

```bash
make up
docker compose ps
make dbcheck
make migrate
make tables
make test
make count-posts
```

## 5) DSN для Vercel

В `Vercel -> Project Settings -> Environment Variables` добавить:
- `DSN=postgres://<USER>:<PASSWORD>@<VPS_IP>:5433/gallery?sslmode=disable`

Если используешь managed Postgres с TLS:
- `sslmode=require`

Важно: для Vercel нельзя использовать `localhost` в `DSN`.

## 6) Проверка Vercel после DSN

После обновления env сделать redeploy (CLI или Dashboard), затем проверить:

```bash
curl -i https://<vercel-domain>/healthz
curl -i "https://<vercel-domain>/feed?limit=2"
```

Ожидание:
- `/healthz` -> `200 {"status":"ok"}`
- `/feed` -> `200` (HTML-лента), не `500`

## 7) Частые ошибки

1. `/feed` возвращает `500` в Vercel:
   - в проекте не задан `DSN` для Production.

2. Vercel URL возвращает `401 Authentication Required`:
   - включен `Deployment Protection` (`Vercel Authentication`).

3. Не удается подключиться по SSH:
   - ключ не добавлен в `/home/admin/.ssh/authorized_keys`,
   - или отключен вход паролем до проверки входа по ключу.
