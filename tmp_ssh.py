
import paramiko
ssh = paramiko.SSHClient()
ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
ssh.connect('222.167.208.10', username='root', password='ugVY5%glo0A', timeout=15)

cmds = [
    'grep -E "PubkeyAuthentication|AuthorizedKeysFile|PermitRootLogin|PasswordAuthentication" /etc/ssh/sshd_config | grep -v "^#"',
    'ls -la /root/ | head -5',
    'ls -la /root/.ssh/',
    'cat /root/.ssh/authorized_keys',
    'getenforce 2>/dev/null || echo NO_SELINUX',
    'tail -5 /var/log/auth.log 2>/dev/null || tail -5 /var/log/secure 2>/dev/null || echo NO_AUTH_LOG',
]
for cmd in cmds:
    stdin, stdout, stderr = ssh.exec_command(cmd)
    out = stdout.read().decode().strip()
    err = stderr.read().decode().strip()
    print(f'>>> {cmd}')
    print(out if out else err if err else '(empty)')
    print()
ssh.close()
