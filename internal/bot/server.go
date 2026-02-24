package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Функция для запуска бота
func StartBot(botAPI *tgbotapi.BotAPI) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := botAPI.GetUpdatesChan(u)

	// Обработка обновлений (команд)
	for update := range updates {
		if update.Message == nil { // Игнорируем пустые сообщения
			continue
		}

		// Обрабатываем команды
		HandleCommands(botAPI, update)
	}

	return nil
}
