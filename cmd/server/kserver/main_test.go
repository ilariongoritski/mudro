package main

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq" // Драйвер PostgreSQL
	"github.com/stretchr/testify/assert"
)

// Тест на проверку наличия токена Telegram бота
func TestBotToken(t *testing.T) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		t.Skip("TELEGRAM_BOT_TOKEN is not set")
	}
	assert.NotEmpty(t, token, "TELEGRAM_BOT_TOKEN is not set")
}

// Тест на подключение к базе данных
func TestDatabaseConnection(t *testing.T) {
	dsn := os.Getenv("DSN")
	if dsn == "" {
		t.Skip("DSN is not set")
	}

	db, err := sql.Open("postgres", dsn)
	assert.Nil(t, err, "Failed to open database connection")
	defer db.Close()

	err = db.Ping()
	assert.Nil(t, err, "Failed to ping database")
}

// Тест на команды бота (можно использовать Telegram Bot API для тестов)
func TestBotCommands(t *testing.T) {
	botCommands := os.Getenv("TELEGRAM_BOT_COMMANDS")
	if botCommands == "" {
		t.Skip("TELEGRAM_BOT_COMMANDS are not set")
	}
	assert.NotEmpty(t, botCommands, "TELEGRAM_BOT_COMMANDS are not set")

	// Проверяем, что команды бота содержат необходимые команды
	expectedCommands := []string{"/start", "/help", "/setcommands"}
	for _, command := range expectedCommands {
		assert.Contains(t, botCommands, command, "Expected command not found: %s", command)
	}

	// Логирование успешной проверки
	log.Printf("Bot commands are correctly set")
}
