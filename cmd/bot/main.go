package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/goritskimihail/mudro/internal/bot"
)

var botAPI *tgbotapi.BotAPI

// Инициализация бота и регистрация команд
func init() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is not set")
	}

	var err error
	botAPI, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}

	// Регистрируем команды
	err = bot.RegisterBotCommands(botAPI)
	if err != nil {
		log.Fatalf("Ошибка при регистрации команд: %v", err)
	}

	log.Printf("Авторизован как %s", botAPI.Self.UserName)
}

// Основная функция для запуска
func main() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := botAPI.GetUpdatesChan(u)

	// Обрабатываем обновления
	for update := range updates {
		// Передаем обновления в обработчик команд
		bot.HandleCommands(botAPI, update)
	}
}
