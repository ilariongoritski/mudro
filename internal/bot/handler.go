package bot

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Функция для отправки сообщений
func sendMessage(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("Error sending message to chat %d: %v", chatID, err)
	}
}

// Функция обработки команд
func HandleCommands(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	if update.Message != nil && update.Message.IsCommand() {
		switch update.Message.Command() {
		case "start":
			handleStart(bot, update)
		case "help":
			handleHelp(bot, update)
		default:
			handleUnknown(bot, update)
		}
	}
}

// Обработчик команды /start
func handleStart(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	welcomeMessage := "Welcome to the Mudro bot! Use /help for assistance."
	sendMessage(bot, update.Message.Chat.ID, welcomeMessage)
}

// Обработчик команды /help
func handleHelp(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	helpMessage := "This bot can help you with the following commands:\n/start - Start the bot\n/help - Show this help message."
	sendMessage(bot, update.Message.Chat.ID, helpMessage)
}

// Обработчик неизвестных команд
func handleUnknown(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	unknownCommandMessage := "Unknown command. Use /help to get a list of commands."
	sendMessage(bot, update.Message.Chat.ID, unknownCommandMessage)
}