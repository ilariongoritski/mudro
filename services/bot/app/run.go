package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/goritskimihail/mudro/internal/bot"
	"github.com/goritskimihail/mudro/internal/config"
)

// Run starts the Telegram bot with graceful shutdown support.
func Run() {
	if err := config.ValidateRuntime("bot", "TELEGRAM_BOT_TOKEN", "TELEGRAM_ALLOWED_USERNAME"); err != nil {
		log.Fatal(err)
	}
	token := os.Getenv("TELEGRAM_BOT_TOKEN")

	botAPI, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}

	if err := bot.RegisterBotCommands(botAPI); err != nil {
		log.Fatalf("register commands: %v", err)
	}
	log.Printf("bot authorized as %s", botAPI.Self.UserName)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := botAPI.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			log.Printf("bot shutting down...")
			botAPI.StopReceivingUpdates()
			return
		case update := <-updates:
			bot.HandleCommands(botAPI, update)
		}
	}
}

