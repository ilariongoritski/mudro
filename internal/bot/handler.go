package bot

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// RegisterBotCommands registers bot commands visible in the Telegram UI.
func RegisterBotCommands(botAPI *tgbotapi.BotAPI) error {
	commands := []tgbotapi.BotCommand{
		{Command: "start", Description: "Запуск бота"},
		{Command: "help", Description: "Справка по командам"},
		{Command: "status", Description: "Текущий статус работы бота"},
		{Command: "health", Description: "Проверка состояния системы"},
		{Command: "migrate", Description: "Запуск миграции базы данных"},
		{Command: "test", Description: "Запуск тестов"},
	}

	config := tgbotapi.NewSetMyCommands(commands...)
	if _, err := botAPI.Request(config); err != nil {
		return fmt.Errorf("ошибка при регистрации команд: %v", err)
	}
	return nil
}

// HandleCommands routes incoming bot commands.
func HandleCommands(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	if update.Message == nil || !update.Message.IsCommand() {
		return
	}

	switch update.Message.Command() {
	case "start":
		handleStart(bot, update)
	case "help":
		handleHelp(bot, update)
	case "status":
		handleStatus(bot, update)
	case "health":
		handleHealth(bot, update)
	case "migrate":
		handleMigrate(bot, update)
	case "test":
		handleTest(bot, update)
	default:
		handleUnknown(bot, update)
	}
}

func handleStart(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Добро пожаловать в Мудрот Бот!")
	if _, err := bot.Send(msg); err != nil {
		log.Printf("send /start: %v", err)
	}
}

func handleHelp(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	helpText := "Доступные команды:\n/start - Запуск бота\n/help - Справка по командам\n/status - Статус работы бота\n/health - Проверка состояния системы\n/migrate - Запуск миграции\n/test - Запуск тестов"
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, helpText)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("send /help: %v", err)
	}
}

func handleStatus(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	status := "Бот работает."
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, status)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("send /status: %v", err)
	}
}

func handleHealth(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Работяга жив и готов к работе!")
	if _, err := bot.Send(msg); err != nil {
		log.Printf("send /health: %v", err)
	}
}

func handleMigrate(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Запуск миграции базы данных...")
	if _, err := bot.Send(msg); err != nil {
		log.Printf("send /migrate: %v", err)
	}
}

func handleTest(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Запуск тестов...")
	if _, err := bot.Send(msg); err != nil {
		log.Printf("send /test: %v", err)
	}
}

func handleUnknown(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда. Используйте /help для получения списка доступных команд.")
	if _, err := bot.Send(msg); err != nil {
		log.Printf("send /unknown: %v", err)
	}
}
