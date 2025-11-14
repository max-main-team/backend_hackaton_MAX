package bot

import (
	"context"
	"strings"

	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

func (b *Bot) handleUpdate(ctx context.Context, update schemes.UpdateInterface) {
	updateType := update.GetUpdateType()

	b.logger.Print(ctx, "Received update",
		"type", updateType,
		"user_id", update.GetUserID(),
		"chat_id", update.GetChatID(),
	)

	// Обрабатываем событие первого запуска бота
	if botStartedUpdate, ok := update.(*schemes.BotStartedUpdate); ok {
		b.handleBotStarted(ctx, botStartedUpdate)
		return
	}

	// Обрабатываем обычные сообщения
	if messageUpdate, ok := update.(*schemes.MessageCreatedUpdate); ok {
		b.handleMessage(ctx, messageUpdate)
		return
	}
}

func (b *Bot) handleBotStarted(ctx context.Context, botStartedUpdate *schemes.BotStartedUpdate) {
	chatID := botStartedUpdate.GetChatID()
	userID := botStartedUpdate.GetUserID()
	userName := botStartedUpdate.User.Name

	b.logger.Print(ctx, "Bot started by user",
		"user_id", userID,
		"chat_id", chatID,
		"user_name", userName,
	)

	// Отправляем приветственное сообщение при первом запуске
	b.sendWelcomeMessage(ctx, chatID, userName)
}

func (b *Bot) handleMessage(ctx context.Context, messageUpdate *schemes.MessageCreatedUpdate) {
	message := messageUpdate.Message

	if message.Body.Text == "" {
		return
	}

	b.logger.Print(ctx, "Received message",
		"from", message.Sender.Name,
		"text", message.Body.Text,
	)

	if b.isCommand(message.Body.Text) {
		b.handleCommand(ctx, messageUpdate)
	}
}

func (b *Bot) isCommand(text string) bool {
	return strings.HasPrefix(text, "/")
}

func (b *Bot) handleCommand(ctx context.Context, messageUpdate *schemes.MessageCreatedUpdate) {
	text := messageUpdate.Message.Body.Text
	command := strings.TrimPrefix(text, "/")
	command = strings.Split(command, " ")[0]
	command = strings.ToLower(command)

	switch command {
	case "start":
		b.handleStartCommand(ctx, messageUpdate)
	default:
		b.logger.Print(ctx, "Unknown command", "command", command)
	}
}
