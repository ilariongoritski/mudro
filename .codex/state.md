- Дата/время: 2026-02-26 23:48
- Что запускал: попытка установки gh и dnsutils (apt/sudo), проверка DNS/утилит
- Что прошло: определил окружение и причины блокировки
- Что упало (ошибка 5–15 строк):
  E: Could not open lock file /var/lib/apt/lists/lock - open (13: Permission denied)
  sudo: a terminal is required to read the password
  sudo: a password is required
- Что починил (если было): нет (нужны права пользователя)
- Следующий шаг: пользователь выполняет sudo-установку; потом повторно проверю gh и CI
