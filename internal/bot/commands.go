package bot

import (
	"context"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

func (b *Bot) handleStartCommand(ctx context.Context, messageUpdate *schemes.MessageCreatedUpdate) {
	chatID := messageUpdate.GetChatID()
	userName := messageUpdate.Message.Sender.Name

	msg := maxbot.NewMessage().
		SetBot(b.token).
		SetChat(chatID).
		SetText("Привет")

	if _, err := b.api.Messages.Send(ctx, msg); err != nil {
		b.logger.Errorf("Failed to send message: %v", err)
		return
	}

	b.logger.Print(ctx, "Sent greeting message", "to", userName)
}
