# Server Transfer Runbook (Ubuntu 24.04, VPS-First)

Цель: безопасно подготовить новый VPS под основной runtime `MUDRO` без хранения секретов в репозитории.

## 0) Сначала ротация секретов
- старый пароль сервера и API key панели считаются скомпрометированными
- перевыпустить их вне репозитория
- не сохранять новые значения в tracked-файлах

## 1) Базовая подготовка VPS
Первый вход допустим под `root`, но только как bootstrap-шаг.

```bash
apt-get update
apt-get install -y ca-certificates curl git make rsync ufw fail2ban docker.io docker-compose-plugin nginx certbot python3-certbot-nginx
systemctl enable --now docker
```

## 2) Создать `admin` и подготовить общий путь проекта

```bash
adduser admin
usermod -aG sudo,docker admin
mkdir -p /srv/mudro
chown admin:admin /srv/mudro
```

## 3) Добавить SSH-ключ пользователю `admin`

```bash
mkdir -p /home/admin/.ssh
chmod 700 /home/admin/.ssh
cat > /home/admin/.ssh/authorized_keys <<'EOF'
ssh-ed25519 <PUBLIC_KEY> <comment>
EOF
chmod 600 /home/admin/.ssh/authorized_keys
chown -R admin:admin /home/admin/.ssh
```

Проверить:

```bash
ssh admin@<VPS_IP>
```

## 4) Только после проверки ключа отключить root/password SSH

```bash
sed -i 's/^#\?PasswordAuthentication.*/PasswordAuthentication no/' /etc/ssh/sshd_config
sed -i 's/^#\?PermitRootLogin.*/PermitRootLogin no/' /etc/ssh/sshd_config
systemctl restart ssh
```

## 5) Firewall: только `22/80/443`

```bash
ufw allow OpenSSH
ufw allow 80/tcp
ufw allow 443/tcp
ufw --force enable
ufw status verbose
```

`5433/tcp` наружу не открывать.

## 6) Деплой проекта в `/srv/mudro`

Под `admin`:

```bash
git clone git@github.com:goritskimihail/mudro.git /srv/mudro
cd /srv/mudro
```

## 7) Подготовить runtime env-файлы
Создать локальные серверные файлы из шаблонов:
- `env/common.env`
- `env/api.env`
- `env/agent.env`
- `env/reporter.env`
- `env/db.env`
- `env/casino.env`
- `env/storage.env`

Минимальный способ:

```bash
cp env/common.env.example env/common.env
cp env/api.env.example env/api.env
cp env/agent.env.example env/agent.env
cp env/reporter.env.example env/reporter.env
cp env/db.env.example env/db.env
cp env/casino.env.example env/casino.env
cp env/storage.env.example env/storage.env
```

После этого вручную заполнить реальные секреты и DSN внутри `env/*.env`.

## 8) Поднять основной runtime

```bash
cd /srv/mudro
docker compose -f docker-compose.prod.yml up -d
docker compose -f docker-compose.prod.yml ps
curl -fsS http://127.0.0.1:8080/healthz
```

Ожидание:
- `db`, `api`, `agent`, `redis`, `kafka`, `minio`, `casino-*` подняты
- `healthz` отвечает `{"status":"ok"}`

## 9) Развернуть frontend через nginx

Если frontend собирается прямо на VPS, сначала нужен установленный Node.js.

```bash
cd /srv/mudro/frontend
npm run build
cd /srv/mudro
sudo bash scripts/ops/deploy_vps_frontend.sh
```

Проверка:

```bash
curl -fsS http://127.0.0.1/healthz
curl -I http://127.0.0.1/
```

## 10) Закрыть БД в loopback-only и выделить app-role

```bash
export MUDRO_DB_APP_PASSWORD='<strong password>'
export MUDRO_DB_SUPERUSER_PASSWORD='<strong password>'
sudo -E bash scripts/ops/harden_vps_db_auth.sh
```

Проверка:

```bash
docker compose -f docker-compose.prod.yml ps db api agent
curl -fsS http://127.0.0.1:8080/healthz
sudo ss -lntp | grep 5433
```

## 11) HTTPS
Когда домен уже указывает на VPS:

```bash
sudo certbot --nginx -d <domain> --non-interactive --agree-tos -m <email>
```

## 12) Частые ошибки
1. `docker compose -f docker-compose.prod.yml up -d` падает:
   - отсутствует один из `env/*.env`
2. `curl http://127.0.0.1:8080/healthz` не отвечает:
   - проверить `docker compose -f docker-compose.prod.yml logs api --tail=100`
3. `curl http://127.0.0.1/healthz` не отвечает:
   - проверить `sudo systemctl status nginx --no-pager`
4. `ssh admin@<VPS_IP>` не работает:
   - не отключать парольный вход до успешной проверки ключа
