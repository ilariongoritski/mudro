package app

import (
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/goritskimihail/mudro/internal/config"
	"github.com/goritskimihail/mudro/internal/reporter"
)

func Run() {
	if err := config.ValidateRequiredEnv("REPORT_BOT_TOKEN", "REPORT_CHAT_ID"); err != nil {
		log.Fatal(err)
	}
	token := config.ReportBotToken()
	repoRoot := config.RepoRoot()
	chatID := reporter.ResolveChatID(repoRoot, config.ReportChatID())

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("new reporter bot: %v", err)
	}
	log.Printf("reporter authorized as %s, chat_id=%d", bot.Self.UserName, chatID)

	send := func() {
		s, err := reporter.BuildSummary(repoRoot)
		if err != nil {
			log.Printf("build summary: %v", err)
			return
		}
		msg := tgbotapi.NewMessage(chatID, s.Text)
		if _, err := bot.Send(msg); err != nil {
			log.Printf("send report: %v", err)
		}
	}

	send()
	interval := time.Duration(config.ReportIntervalMinutes()) * time.Minute
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		send()
	}
}
