package bot

import (
	"fmt"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func handleMovies(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	webAppURL := getMoviesWebAppURL()

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("🎬 Открыть каталог фильмов", webAppURL),
		),
	)

	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		"🎬 *Каталог фильмов MUDROTOP*\n\n"+
			"Откройте каталог для просмотра фильмов с фильтрами:\n"+
			"• По году выпуска\n"+
			"• По длительности\n"+
			"• По жанрам\n"+
			"• С рейтингом и описанием",
	)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	if _, err := bot.Send(msg); err != nil {
		log.Printf("handleMovies send error: %v", err)
	}
}

func getMoviesWebAppURL() string {
	baseURL := os.Getenv("MUDRO_WEB_URL")
	if baseURL == "" {
		baseURL = "https://mudro.vercel.app"
	}
	return fmt.Sprintf("%s/movies", baseURL)
}
