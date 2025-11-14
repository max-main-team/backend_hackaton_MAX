package bot

import (
	"context"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

func (b *Bot) handleStartCommand(ctx context.Context, message schemes.Message) {
	chatID := message.Recipient.ChatId

	msg := maxbot.NewMessage().
		SetBot(b.token).
		SetChat(chatID).
		SetText("Привет")

	if _, err := b.api.Messages.Send(ctx, msg); err != nil {
		b.logger.Errorf("Failed to send message: %v", err)
		return
	}

	b.logger.Print(ctx, "Sent greeting message", "to", message.Sender.Name)
}
