package bot

import (
	"context"
	"fmt"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/vmkteam/embedlog"
)

type Bot struct {
	api    *maxbot.Api
	logger embedlog.Logger
	token  string
}

func New(token string, logger embedlog.Logger) (*Bot, error) {
	api, err := maxbot.New(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	logger.Print(context.Background(), "Bot API initialized successfully")

	return &Bot{
		api:    api,
		logger: logger,
		token:  token,
	}, nil
}

func (b *Bot) Start(ctx context.Context) error {
	b.logger.Print(ctx, "Starting bot...")

	updates := b.api.GetUpdates(ctx)

	for {
		select {
		case <-ctx.Done():
			b.logger.Print(ctx, "Bot stopped")
			return nil
		case update := <-updates:
			if update == nil {
				continue
			}

			b.handleUpdate(ctx, update)
		}
	}
}
