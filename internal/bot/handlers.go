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

	if messageUpdate, ok := update.(*schemes.MessageCreatedUpdate); ok {
		b.handleMessage(ctx, messageUpdate)
	}
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
		b.handleCommand(ctx, message)
	}
}

func (b *Bot) isCommand(text string) bool {
	return strings.HasPrefix(text, "/")
}

func (b *Bot) handleCommand(ctx context.Context, message schemes.Message) {
	text := message.Body.Text
	command := strings.TrimPrefix(text, "/")
	command = strings.Split(command, " ")[0]
	command = strings.ToLower(command)

	switch command {
	case "start":
		b.handleStartCommand(ctx, message)
	default:
		b.logger.Print(ctx, "Unknown command", "command", command)
	}
}
