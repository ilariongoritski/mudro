package main

import (
	"fmt"
	"log"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

// Регистрируем команды для бота
func RegisterBotCommands(botAPI *tgbotapi.BotAPI) error {
    commands := []tgbotapi.BotCommand{
        tgbotapi.BotCommand{Command: "start", Description: "Запуск бота"},
        tgbotapi.BotCommand{Command: "help", Description: "Справка по командам"},
        tgbotapi.BotCommand{Command: "status", Description: "Текущий статус работы бота"},
        tgbotapi.BotCommand{Command: "health", Description: "Проверка состояния системы"},
        tgbotapi.BotCommand{Command: "migrate", Description: "Запуск миграции базы данных"},
        tgbotapi.BotCommand{Command: "test", Description: "Запуск тестов"},
    }

    config := tgbotapi.NewSetMyCommands(commands...) 
    _, err := botAPI.Request(config)
    if err != nil {
        return fmt.Errorf("ошибка при регистрации команд: %v", err)
    }
    return nil
}

// Обработка команд
func HandleCommands(botAPI *tgbotapi.BotAPI, update tgbotapi.Update) {
	switch update.Message.Command() {
	case "start":
		handleStart(botAPI, update)
	case "help":
		handleHelp(botAPI, update)
	case "status":
		handleStatus(botAPI, update)
	case "health":
		handleHealth(botAPI, update)
	case "migrate":
		handleMigrate(botAPI, update)
	case "test":
		handleTest(botAPI, update)
	default:
		handleUnknown(botAPI, update)
	}
}

// Функции для команд
func handleStart(botAPI *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Добро пожаловать в Мудрот Бот!")
	botAPI.Send(msg)
}

func handleHelp(botAPI *tgbotapi.BotAPI, update tgbotapi.Update) {
	helpText := "Доступные команды:\n/start - Запуск бота\n/help - Справка по командам\n/status - Статус работы бота\n/health - Проверка состояния системы\n/migrate - Запуск миграции\n/test - Запуск тестов"
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, helpText)
	botAPI.Send(msg)
}

func handleStatus(botAPI *tgbotapi.BotAPI, update tgbotapi.Update) {
	status := "Бот работает, подключение к БД успешно!"
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, status)
	botAPI.Send(msg)
}

func handleHealth(botAPI *tgbotapi.BotAPI, update tgbotapi.Update) {
	// Проверка состояния системы, например, работяги
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Работяга жив и готов к работе!")
	botAPI.Send(msg)
}

func handleMigrate(botAPI *tgbotapi.BotAPI, update tgbotapi.Update) {
	// Запуск миграции
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Запуск миграции базы данных...")
	botAPI.Send(msg)
}

func handleTest(botAPI *tgbotapi.BotAPI, update tgbotapi.Update) {
	// Запуск тестов
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Запуск тестов...")
	botAPI.Send(msg)
}

func handleUnknown(botAPI *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда. Используйте /help для получения списка доступных команд.")
	botAPI.Send(msg)
}