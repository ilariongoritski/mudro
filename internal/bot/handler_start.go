package bot

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/goritskimihail/mudro/internal/config"
)

// handleStart — enhanced with Telegram auth + profile link
func handleStart(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	from := update.Message.From

	// Build TelegramUser from message
	tgUser := TelegramUser{
		ID:           from.ID,
		FirstName:    from.FirstName,
		LastName:     from.LastName,
		Username:     from.UserName,
		LanguageCode: from.LanguageCode,
	}

	// Get runner instance (assume global or inject)
	// For simplicity we call the auth function directly
	userID, err := AuthOrLinkTelegramUser(context.Background(), tgUser)
	if err != nil {
		log.Printf("telegram auth failed: %v", err)
		msg := tgbotapi.NewMessage(chatID, "Ошибка привязки аккаунта. Попробуйте позже.")
		bot.Send(msg)
		return
	}

	welcome := fmt.Sprintf(
		"✅ Аккаунт привязан!\n\n"+
			"Telegram: @%s (ID: %d)\n"+
			"Ваш Mudro ID: %d\n\n"+
			"Открыть профиль: /profile\n"+
			"Или используйте Mini App для заполнения имени, возраста, bio и аватара.",
		tgUser.Username, tgUser.ID, userID,
	)

	msg := tgbotapi.NewMessage(chatID, welcome)

	// Add inline button to open profile or Mini App
	btn := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("Открыть профиль", "https://mudro.local/profile/"+strconv.FormatInt(userID, 10)),
			tgbotapi.NewInlineKeyboardButtonURL("Mini App", "https://t.me/your_bot?startapp=profile"),
		),
	)
	msg.ReplyMarkup = btn

	if _, err := bot.Send(msg); err != nil {
		log.Printf("send /start: %v", err)
	}
}
