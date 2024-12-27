package bot

import (
	"context"
	"edu-portal/app"
	"log"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Users interface {
	CreateUser(ctx context.Context, user *app.User) error
	GetUserByTgId(ctx context.Context, id int64) (*app.User, error)
	ResolveTgToken(ctx context.Context, user_id int, token string) error
}

type Handler struct {
	users Users
}

func New(users Users) *Handler {
	return &Handler{
		users: users,
	}
}

func (h Handler) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	after, found := strings.CutPrefix(update.Message.Text, "/start")
	if !found {
		return
	}
	after = strings.TrimSpace(after)

	user, err := h.users.GetUserByTgId(ctx, update.Message.From.ID)
	if err != nil {
		log.Printf("[ERROR] Get user by tg id: %v", err)
		return
	}
	if user == nil {
		user = &app.User{
			TgId:       &update.Message.From.ID,
			TgUsername: &update.Message.From.Username,
		}
	}

	if err := h.users.CreateUser(ctx, user); err != nil {
		log.Printf("[ERROR] Create new user: %v", err)
		return
	}

	if err := h.users.ResolveTgToken(ctx, user.Id, after); err != nil {
		log.Printf("[ERROR] Resolve tg token: %v", err)
		return
	}

	log.Printf("[INFO] Tg token has been resolved for user id %d", user.Id)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.From.ID,
		Text:   "Авторизация прошла успешно, вернитесь на портал.",
	})
}
