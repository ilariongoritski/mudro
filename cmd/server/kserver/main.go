// Package main содержит улучшенный Telegram бот для интеграции с работягой mudro.
// Улучшения: env для токена, лучшие ошибки, логирование, concurrency-safe, сигналы shutdown.

package main

import (
	"context"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/goritskimihail/mudro/internal/config"
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

				repoRoot := config.RepoRoot()
				dsn := config.DSN()

				out, err := runHealthLoop(ctxCmd, repoRoot, dsn)
				if err != nil {
					msg.Text = formatReply("Error: "+err.Error()+"\nOutput: ", out, config.TelegramMessageLimit())
				} else {
					msg.Text = formatReply("Health loop: ", out, config.TelegramMessageLimit())
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

func runHealthLoop(ctx context.Context, repoRoot, dsn string) ([]byte, error) {
	var out strings.Builder

	steps := [][]string{
		{"make", "up"},
		{"docker", "compose", "ps"},
		{"make", "dbcheck"},
		{"make", "migrate"},
		{"make", "tables"},
		{"make", "test"},
		{"psql", dsn, "-X", "-c", "select count(*) from posts;"},
	}

	for _, step := range steps {
		out.WriteString("$ " + strings.Join(step, " ") + "\n")
		cmd := exec.CommandContext(ctx, step[0], step[1:]...)
		cmd.Dir = repoRoot
		cmd.Env = os.Environ()
		if step[0] == "psql" {
			cmd.Env = append(cmd.Env, "PGCONNECT_TIMEOUT=3")
		}
		b, err := cmd.CombinedOutput()
		out.Write(b)
		if err != nil {
			return []byte(out.String()), err
		}
	}

	return []byte(out.String()), nil
}

func formatReply(prefix string, body []byte, limit int) string {
	if limit <= 0 {
		return prefix
	}
	full := prefix + string(body)
	if len([]rune(full)) <= limit {
		return full
	}

	// Leave space for suffix.
	suffix := "\n...(truncated)"
	avail := limit - len([]rune(suffix))
	if avail <= 0 {
		return prefix + suffix
	}

	r := []rune(full)
	if avail > len(r) {
		avail = len(r)
	}
	return string(r[:avail]) + suffix
}
