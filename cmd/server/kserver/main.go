// Package main содержит улучшенный Telegram бот для интеграции с работягой mudro.
// Улучшения: env для токена, лучшие ошибки, логирование, concurrency-safe, сигналы shutdown.

package main

import (
	"context"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN not set")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = false // Отключить debug для продакшена
	log.Printf("Authorized: %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	updates := bot.GetUpdatesChan(u)

	go handleUpdates(ctx, bot, updates)

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Shutting down")
}

func handleUpdates(ctx context.Context, bot *tgbotapi.BotAPI, updates tgbotapi.UpdatesChannel) {
	for {
		select {
		case <-ctx.Done():
			return
		case update := <-updates:
			if update.Message == nil {
				continue
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
			if update.Message.Text == "/health" {
				ctxCmd, cancelCmd := context.WithTimeout(ctx, 30*time.Second)
				defer cancelCmd()

				cmd := exec.CommandContext(ctxCmd, "bash", "-c", "cd ~/projects/mudro && make up && docker compose ps && make dbcheck && make migrate && make tables && make test && psql \"$(DSN)\" -X -c \"select count(*) from posts;\"")
				out, err := cmd.CombinedOutput()
				if err != nil {
					msg.Text = "Error: " + err.Error() + "\nOutput: " + string(out)
				} else {
					msg.Text = "Health loop: " + string(out)
				}
			} else {
				msg.Text = "Use /health"
			}

			if _, err := bot.Send(msg); err != nil {
				log.Printf("Send error: %v", err)
			}
		}
	}
}
