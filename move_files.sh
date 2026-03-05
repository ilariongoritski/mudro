#!/bin/bash

# Убедитесь, что скрипт запускается из корня проекта
if [ ! -f go.mod ]; then
  echo "Вы должны запустить этот скрипт из корня проекта!"
  exit 1
fi

echo "Перемещаем файлы в нужные директории..."

# Функция для перемещения файла, если он существует и не совпадает с целевым
move_file() {
  if [ -f "$1" ]; then
    if [ "$1" != "$2" ]; then  # Проверка на совпадение исходного и целевого пути
      mv "$1" "$2"
      echo "Файл $1 перемещен в $2"
    else
      echo "Файл $1 уже находится в целевой директории"
    fi
  else
    echo "Файл $1 не найден!"
  fi
}

# Перемещение файлов бота
mkdir -p cmd/bot
move_file cmd/main.go cmd/bot/main.go
move_file cmd/main_test.go cmd/bot/main_test.go
move_file cmd/bot/handler.go internal/bot/handler.go
move_file cmd/bot/server.go internal/bot/server.go

# Перемещение серверного файла
mkdir -p cmd/kserver
move_file cmd/kserver/main.go cmd/kserver/main.go

# Перемещение миграций
mkdir -p migrations
move_file migrations/001_init.sql migrations/001_init.sql

# Перемещение скриптов
mkdir -p scripts
move_file scripts/import.sh scripts/import.sh
move_file scripts/migrate.sh scripts/migrate.sh
move_file scripts/stats.sh scripts/stats.sh

# Перемещение логики репозитория (если есть)
mkdir -p internal/repo
move_file internal/repo/repo.go internal/repo/repo.go

# Печать результата
echo "Перемещение файлов завершено."

# Убедитесь, что пути импорта в Go файлах обновлены.
echo "Пожалуйста, обновите пути импорта в Go файлах вручную!"

# Опционально: если нужно, можно сразу обновить зависимости
go mod tidy

echo "Зависимости обновлены с помощью go mod tidy."