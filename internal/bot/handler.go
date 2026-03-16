package bot

import (
	"fmt"
	"log"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/goritskimihail/mudro/internal/config"
	"github.com/goritskimihail/mudro/internal/reporter"
)

// RegisterBotCommands registers bot commands visible in the Telegram UI.
func RegisterBotCommands(botAPI *tgbotapi.BotAPI) error {
	commands := []tgbotapi.BotCommand{
		{Command: "start", Description: "Запуск бота"},
		{Command: "help", Description: "Справка по командам"},
		{Command: "mudro", Description: "Вопрос к LLM по проекту"},
		{Command: "now", Description: "Краткий итог изменений с запуска"},
		{Command: "todo", Description: "Показать цели и улучшения"},
		{Command: "todoadd", Description: "Добавить цель в TODO"},
		{Command: "top10", Description: "Топ-10 ключевых изменений"},
		{Command: "repo", Description: "Структура репозитория"},
		{Command: "find", Description: "Найти улучшения по репозиторию"},
		{Command: "time", Description: "Суммарное время работы"},
		{Command: "rab", Description: "Авто-работяга по TODO"},
		{Command: "memento", Description: "Полная синхронизация памяти"},
		{Command: "tglog", Description: "История управления из Telegram"},
		{Command: "chat", Description: "Режим обычного чата on/off/status"},
		{Command: "reportnow", Description: "Отправить репорт прямо сейчас"},
		{Command: "feed5", Description: "Лента из API: 5 постов"},
		{Command: "health", Description: "Здоровье и итоги за день"},
		{Command: "logs", Description: "Последние логи БД"},
		{Command: "actions10", Description: "Задача/проблемы/выполнения/доработки"},
		{Command: "actions1h", Description: "Путь за час: проблемы и решения"},
		{Command: "commits3", Description: "3 коммита: суть изменений на русском"},
	}

	config := tgbotapi.NewSetMyCommands(commands...)
	if _, err := botAPI.Request(config); err != nil {
		return fmt.Errorf("ошибка при регистрации команд: %v", err)
	}
	return nil
}

// HandleCommands routes incoming bot commands.
func HandleCommands(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	if update.Message == nil {
		return
	}
	key := updateKey(update.Message.Chat.ID, update.Message.MessageID)
	markCommandStart(key)

	if !isAllowedUser(update) {
		denyUnauthorized(bot, update)
		return
	}
	if !update.Message.IsCommand() {
		handleTextFollowup(bot, update)
		return
	}

	switch update.Message.Command() {
	case "start":
		handleStart(bot, update)
	case "help":
		handleHelp(bot, update)
	case "mudro":
		handleMudro(bot, update)
	case "now":
		handleNow(bot, update)
	case "todo":
		handleTodo(bot, update)
	case "todoadd":
		handleTodoAdd(bot, update)
	case "top10":
		handleTop10(bot, update)
	case "repo":
		handleRepo(bot, update)
	case "find":
		handleFind(bot, update)
	case "time":
		handleTime(bot, update)
	case "rab":
		handleRab(bot, update)
	case "memento":
		handleMemento(bot, update)
	case "tglog":
		handleTGLog(bot, update)
	case "chat":
		handleChatMode(bot, update)
	case "reportnow":
		handleReportNow(bot, update)
	case "feed5":
		handleFeed5(bot, update)
	case "health":
		handleHealth(bot, update)
	case "logs":
		handleLogs(bot, update)
	case "actions10":
		handleActions10(bot, update)
	case "actions1h":
		handleActions1H(bot, update)
	case "commits3":
		handleCommits3(bot, update)
	default:
		handleUnknown(bot, update)
	}
}

func isAllowedUser(update tgbotapi.Update) bool {
	if update.Message == nil || update.Message.From == nil {
		return false
	}
	allowed := config.TelegramAllowedUsername()
	if allowed == "" {
		return false
	}
	return strings.EqualFold(update.Message.From.UserName, allowed)
}

func denyUnauthorized(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "unauthorized")
	if _, err := bot.Send(msg); err != nil {
		log.Printf("send /unauthorized: %v", err)
	}
}

func handleStart(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Добро пожаловать в Мудрот Бот!")
	if _, err := bot.Send(msg); err != nil {
		log.Printf("send /start: %v", err)
	}
}

func handleHelp(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	helpText := strings.Join([]string{
		"Доступные команды:",
		"/start - Запуск бота",
		"/help - Справка по командам",
		"/mudro <вопрос> - Вопрос к LLM по проекту (поддержка выбора 1/2)",
		"/now - Краткий итог изменений с запуска",
		"/todo - Показать будущие цели и улучшения",
		"/todoadd <текст> - Добавить пункт в TODO",
		"/top10 - Топ-10 значимых изменений проекта",
		"/repo - Показать структуру репозитория",
		"/find - Найти потенциальные улучшения по репозиторию",
		"/time - Суммарное время работы за день и всего",
		"/rab - Выполнить простые задачи из TODO и обновить память",
		"/memento - Полная синхронизация памяти проекта",
		"/tglog - Показать историю управления из Telegram",
		"/chat on|off|status - Переключить режим обычного чата",
		"/reportnow - Отправить мгновенный отчет reporter-ботом",
		"/feed5 - Показать 5 постов из API",
		"/health - Здоровье и итоги за день",
		"/logs - Последние логи БД",
		"/actions10 - Задача, проблемы, выполнения, доработки (10 мин)",
		"/actions1h - Путь за час с проблемами и решениями",
		"/commits3 - Краткая суть 3 последних коммитов на русском",
	}, "\n")
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, helpText)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("send /help: %v", err)
	}
}

func handleLogs(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	r := NewRunner()
	out, err := r.Logs()
	sendReply(bot, update, "/logs", out, err)
}

func handleHealth(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	r := NewRunner()
	out, err := r.HealthDaily()
	sendReply(bot, update, "/health", out, err)
}

func handleMudro(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	r := NewRunner()
	query := update.Message.CommandArguments()
	out, err := r.AskMudro(query)
	if err == nil {
		answer := string(out)
		if detectChoicePrompt(answer) {
			savePendingChoice(update.Message.Chat.ID, query, answer)
			answer += "\n\nВыбери вариант: отправь отдельным сообщением `1` или `2`."
			out = []byte(answer)
		} else {
			clearPendingChoice(update.Message.Chat.ID)
		}
	}
	sendReply(bot, update, "/mudro", out, err)
}

func handleTextFollowup(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	choice, ok := isChoiceReply(update.Message.Text)
	if !ok {
		r := NewRunner()
		if err := r.loadChatModes(); err == nil && r.isChatModeEnabled(update.Message.Chat.ID) {
			if query, ok := r.handleChatText(update.Message.Text); ok {
				out, err := r.AskMudro(query)
				if err != nil {
					if strings.Contains(err.Error(), "OPENAI_API_KEY") {
						sendPlainWithCommand(bot, update, "/mudro_chat", "Режим чата включен, но OPENAI_API_KEY не задан.\nДобавь ключ в .env или выключи режим: /chat off")
						return
					}
					sendPlainWithCommand(bot, update, "/mudro_chat", "Ошибка chat-mode: "+err.Error())
					return
				}
				sendPlainWithCommand(bot, update, "/mudro_chat", string(out))
			}
		}
		return
	}
	pending, ok := loadPendingChoice(update.Message.Chat.ID)
	if !ok {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Нет активного выбора. Сначала вызови /mudro с вопросом.")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("send /followup-miss: %v", err)
		}
		return
	}

	r := NewRunner()
	followup := strings.Join([]string{
		"Продолжи по выбранному варианту.",
		"Первичный запрос: " + pending.question,
		"Предыдущий ответ:\n" + pending.answer,
		"Выбор пользователя: " + choice,
		"Дай конкретные шаги запуска/открытия, кратко.",
	}, "\n\n")

	out, err := r.AskMudro(followup)
	clearPendingChoice(update.Message.Chat.ID)
	sendReply(bot, update, "/mudro", out, err)
}

func handleNow(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	r := NewRunner()
	out, err := r.NowSummary()
	sendReply(bot, update, "/now", out, err)
}

func handleTodo(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	r := NewRunner()
	out, err := r.TodoList()
	sendReply(bot, update, "/todo", out, err)
}

func handleTodoAdd(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	r := NewRunner()
	out, err := r.TodoAdd(update.Message.CommandArguments())
	sendReply(bot, update, "/todoadd", out, err)
}

func handleTop10(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	r := NewRunner()
	out, err := r.Top10List()
	sendReply(bot, update, "/top10", out, err)
}

func handleRepo(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	r := NewRunner()
	out, err := r.RepoStructure()
	sendReply(bot, update, "/repo", out, err)
}

func handleFind(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	r := NewRunner()
	out, err := r.FindImprovements()
	sendReply(bot, update, "/find", out, err)
}

func handleTime(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	r := NewRunner()
	out, err := r.TimeSummary()
	sendReply(bot, update, "/time", out, err)
}

func handleRab(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	r := NewRunner()
	out, err := r.RunAutoBacklog()
	sendReply(bot, update, "/rab", out, err)
}

func handleMemento(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	r := NewRunner()
	out, err := r.Memento()
	sendReply(bot, update, "/memento", out, err)
}

func handleTGLog(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	r := NewRunner()
	out, err := r.TelegramControlLog()
	sendReply(bot, update, "/tglog", out, err)
}

func handleChatMode(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	r := NewRunner()
	arg := strings.ToLower(strings.TrimSpace(update.Message.CommandArguments()))
	switch arg {
	case "on", "enable", "1":
		out, err := r.SetChatMode(update.Message.Chat.ID, true)
		if err == nil && config.OpenAIAPIKey() == "" {
			out = append(out, []byte("\nВнимание: OPENAI_API_KEY не задан, ответы LLM недоступны.")...)
		}
		sendReply(bot, update, "/chat", out, err)
	case "off", "disable", "0":
		out, err := r.SetChatMode(update.Message.Chat.ID, false)
		sendReply(bot, update, "/chat", out, err)
	case "status", "":
		out, err := r.ChatModeStatus(update.Message.Chat.ID)
		sendReply(bot, update, "/chat", out, err)
	default:
		sendReply(bot, update, "/chat", []byte("Использование: /chat on | /chat off | /chat status"), nil)
	}
}

func handleReportNow(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	token := config.ReportBotToken()
	if token == "" {
		sendReply(bot, update, "/reportnow", nil, fmt.Errorf("REPORT_BOT_TOKEN is not set"))
		return
	}
	r := NewRunner()
	chatID := reporter.ResolveChatID(r.RepoRoot, config.ReportChatID())
	if chatID <= 0 {
		sendReply(bot, update, "/reportnow", nil, fmt.Errorf("REPORT_CHAT_ID not found"))
		return
	}
	repBot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		sendReply(bot, update, "/reportnow", nil, err)
		return
	}
	s, err := reporter.BuildSummary(r.RepoRoot)
	if err != nil {
		sendReply(bot, update, "/reportnow", nil, err)
		return
	}
	msg := tgbotapi.NewMessage(chatID, s.Text)
	if _, err := repBot.Send(msg); err != nil {
		sendReply(bot, update, "/reportnow", nil, err)
		return
	}
	sendReply(bot, update, "/reportnow", []byte("Отчет отправлен reporter-ботом."), nil)
}

func handleFeed5(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	r := NewRunner()
	out, err := r.Feed5()
	sendReply(bot, update, "/feed5", out, err)
}

func handleActions10(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	r := NewRunner()
	out, err := r.ActionsLast10Min()
	sendReply(bot, update, "/actions10", out, err)
}

func handleActions1H(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	r := NewRunner()
	out, err := r.ActionsLast1H()
	sendReply(bot, update, "/actions1h", out, err)
}

func handleCommits3(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	r := NewRunner()
	out, err := r.Last3Commits()
	sendReply(bot, update, "/commits3", out, err)
}

func handleUnknown(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда. Используйте /help для получения списка доступных команд.")
	if _, err := bot.Send(msg); err != nil {
		log.Printf("send /unknown: %v", err)
	}
}

func sendReply(bot *tgbotapi.BotAPI, update tgbotapi.Update, prefix string, out []byte, err error) {
	text := formatReply(prefix, out, err)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	sendStartedAt := time.Now()
	if _, sendErr := bot.Send(msg); sendErr != nil {
		log.Printf("send %s: %v", prefix, sendErr)
	}
	sendElapsed := time.Since(sendStartedAt)
	finalizeResponseMetrics(update, prefix, out, err, sendStartedAt, sendElapsed)
}

func finalizeResponseMetrics(update tgbotapi.Update, prefix string, out []byte, err error, sendStartedAt time.Time, sendElapsed time.Duration) {
	key := updateKey(update.Message.Chat.ID, update.Message.MessageID)
	if startedAt, ok := popCommandStart(key); ok {
		totalElapsed := time.Since(startedAt)
		processElapsed := sendStartedAt.Sub(startedAt)
		if processElapsed < 0 {
			processElapsed = 0
		}
		r := NewRunner()
		if recErr := recordRuntimeTime(r.RepoRoot, prefix, processElapsed, sendElapsed, totalElapsed); recErr != nil {
			log.Printf("record runtime time %s: %v", prefix, recErr)
		}
	}

	r := NewRunner()
	status := "ok"
	errText := ""
	if err != nil {
		status = "error"
		errText = err.Error()
	}
	username := ""
	if update.Message.From != nil {
		username = update.Message.From.UserName
	}
	if logErr := appendTGControlEvent(r.RepoRoot, tgControlEvent{
		Username:  username,
		ChatID:    update.Message.Chat.ID,
		Command:   prefix,
		Args:      trimForLog(update.Message.CommandArguments(), 120),
		Status:    status,
		Error:     errText,
		ReplyHint: trimForLog(string(out), 120),
	}); logErr != nil {
		log.Printf("tg control log %s: %v", prefix, logErr)
	}
}

func sendPlain(bot *tgbotapi.BotAPI, update tgbotapi.Update, text string) {
	sendPlainWithCommand(bot, update, "/chat", text)
}

func sendPlainWithCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update, command string, text string) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, trimMessage(text, NewRunner().Limit))
	sendStartedAt := time.Now()
	if _, sendErr := bot.Send(msg); sendErr != nil {
		log.Printf("send plain: %v", sendErr)
	}
	sendElapsed := time.Since(sendStartedAt)
	finalizeResponseMetrics(update, command, []byte(text), nil, sendStartedAt, sendElapsed)
}
