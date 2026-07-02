
import paramiko, sys, time

def run(ssh, cmd, timeout=120):
    print(f'>>> {cmd[:80]}')
    stdin, stdout, stderr = ssh.exec_command(cmd, timeout=timeout)
    out = stdout.read().decode().strip()
    err = stderr.read().decode().strip()
    if out: print(out[:500])
    if err: print('ERR:', err[:300])
    return out

ssh = paramiko.SSHClient()
ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
ssh.connect('222.167.208.10', username='root', password='ugVY5%glo0A', timeout=15)
print('CONNECTED\n')

# 1. Install deps
run(ssh, 'apt update -qq && apt install -y -qq nginx ufw git curl > /dev/null 2>&1 && echo DEPS_OK')

# 2. Clone repo
run(ssh, 'if [ ! -d /opt/mudro ]; then git clone https://github.com/goritskimihail/mudro /opt/mudro 2>&1 | tail -2; else echo REPO_EXISTS; fi')

# 3. Check repo
run(ssh, 'ls /opt/mudro/Makefile /opt/mudro/docker-compose.prod.yml 2>/dev/null && echo REPO_OK || echo REPO_MISSING')

# 4. Create env dir
run(ssh, 'mkdir -p /etc/mudro')

# 5. Docker compose up (DBs first)
run(ssh, 'cd /opt/mudro && POSTGRES_PASSWORD=mudropass2026 CASINO_POSTGRES_PASSWORD=casinopass2026 JWT_SECRET=dev-jwt-secret-32chars-long-1234567890 CASINO_INTERNAL_SECRET=vps-casino-secret-2026 MUDRO_APP_DSN=postgres://postgres:mudropass2026@db:5432/gallery?sslmode=disable CASINO_APP_DSN=postgres://postgres:casinopass2026@casino-db:5432/mudro_casino?sslmode=disable MINIO_ROOT_USER=admin MINIO_ROOT_PASSWORD=minioadmin2026 docker compose -f docker-compose.prod.yml up -d 2>&1 | tail -10', timeout=180)

print('\nDONE')
ssh.close()
