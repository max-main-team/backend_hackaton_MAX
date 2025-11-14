package bot

import (
	"context"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

func (b *Bot) handleStartCommand(ctx context.Context, messageUpdate *schemes.MessageCreatedUpdate) {
	chatID := messageUpdate.Message.Recipient.ChatId
	userID := messageUpdate.Message.Sender.UserId
	userName := messageUpdate.Message.Sender.Name

	b.logger.Print(ctx, "Preparing to send message",
		"chat_id", chatID,
		"user_id", userID,
		"to", userName,
	)

	// Отправляем сообщение в чат (как в примере)
	msg := maxbot.NewMessage().
		SetChat(chatID).
		SetText("Привет")

	resp, err := b.api.Messages.Send(ctx, msg)
	if err != nil {
		b.logger.Errorf("Failed to send message: %v (chat_id=%d, user_id=%d)", err, chatID, userID)
		return
	}

	b.logger.Print(ctx, "Sent greeting message successfully",
		"to", userName,
		"response", resp,
	)
}
